package revenium

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"sync"
	"time"

	"google.golang.org/genai"
)

const (
	defaultReveniumBaseURL = "https://api.revenium.ai"
	meteringEndpoint       = "/meter/v2/ai/completions"
	defaultCostType        = "AI"
	defaultOperationType   = "CHAT"
)

// ReveniumGoogle is the main middleware client that wraps the Google Genai SDK
// and adds metering capabilities
type ReveniumGoogle struct {
	client   *genai.Client
	config   *Config
	provider Provider
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

var (
	globalClient *ReveniumGoogle
	globalMu     sync.RWMutex
	initialized  bool
)

// createGenaiClient creates a Google Genai client based on the provider configuration
func createGenaiClient(ctx context.Context, cfg *Config, provider Provider) (*genai.Client, error) {
	if provider.IsVertexAI() {
		if cfg.ProjectID == "" {
			return nil, NewConfigError("GOOGLE_CLOUD_PROJECT is required for Vertex AI", nil)
		}
		if cfg.Location == "" {
			return nil, NewConfigError("GOOGLE_CLOUD_LOCATION is required for Vertex AI", nil)
		}
		return genai.NewClient(ctx, &genai.ClientConfig{
			Project:  cfg.ProjectID,
			Location: cfg.Location,
			Backend:  genai.BackendVertexAI,
		})
	}

	if cfg.GoogleAPIKey == "" {
		return nil, NewConfigError("GOOGLE_API_KEY is required for Google AI", nil)
	}
	return genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GoogleAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
}

// Initialize sets up the global Revenium middleware with configuration
func Initialize(opts ...Option) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	if initialized {
		return nil
	}

	InitializeLogger()
	Info("Initializing Revenium middleware...")

	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.loadFromEnv(); err != nil {
		Warn("Failed to load configuration from environment: %v", err)
	}

	SetGlobalDebug(cfg.Debug)

	if cfg.ReveniumAPIKey == "" {
		return NewConfigError("REVENIUM_METERING_API_KEY is required", nil)
	}

	provider := DetectProvider(cfg)

	ctx := context.Background()
	genaiClient, err := createGenaiClient(ctx, cfg, provider)
	if err != nil {
		return NewProviderError("failed to create Google Genai client", err)
	}

	globalClient = &ReveniumGoogle{
		client:   genaiClient,
		config:   cfg,
		provider: provider,
	}

	initialized = true
	Info("Revenium middleware initialized successfully with provider: %s", provider.String())
	return nil
}

// IsInitialized checks if the middleware is properly initialized
func IsInitialized() bool {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return initialized
}

// GetClient returns the global Revenium client
func GetClient() (*ReveniumGoogle, error) {
	globalMu.RLock()
	defer globalMu.RUnlock()

	if !initialized {
		return nil, NewConfigError("middleware not initialized, call Initialize() first", nil)
	}

	return globalClient, nil
}

// NewReveniumGoogle creates a new Revenium client with explicit configuration
func NewReveniumGoogle(cfg *Config) (*ReveniumGoogle, error) {
	if cfg == nil {
		return nil, NewConfigError("config cannot be nil", nil)
	}

	// Validate required fields
	if cfg.ReveniumAPIKey == "" {
		return nil, NewConfigError("REVENIUM_METERING_API_KEY is required", nil)
	}

	provider := DetectProvider(cfg)

	ctx := context.Background()
	genaiClient, err := createGenaiClient(ctx, cfg, provider)
	if err != nil {
		return nil, NewProviderError("failed to create Google Genai client", err)
	}

	return &ReveniumGoogle{
		client:   genaiClient,
		config:   cfg,
		provider: provider,
	}, nil
}

// GetConfig returns the configuration
func (r *ReveniumGoogle) GetConfig() *Config {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// GetProvider returns the detected provider
func (r *ReveniumGoogle) GetProvider() Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.provider
}

// GetGenaiClient returns the underlying Google Genai client
func (r *ReveniumGoogle) GetGenaiClient() *genai.Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.client
}

// Models returns the models interface for generating content
func (r *ReveniumGoogle) Models() *ModelsInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &ModelsInterface{
		client:   r.client,
		config:   r.config,
		provider: r.provider,
		parent:   r,
	}
}

// Flush waits for all pending metering requests to complete
// This should be called before the application exits to ensure all metering data is sent
func (r *ReveniumGoogle) Flush() {
	Debug("Flushing pending metering requests...")
	r.wg.Wait()
	Debug("All metering requests completed")
}

// Close closes the client and cleans up resources
// It waits for all pending metering requests to complete before returning
func (r *ReveniumGoogle) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Wait for all pending metering requests
	r.Flush()

	// Google Genai client doesn't have a Close method
	return nil
}

// ModelsInterface provides methods for generating content with metering
type ModelsInterface struct {
	client   *genai.Client
	config   *Config
	provider Provider
	parent   *ReveniumGoogle // Reference to parent for WaitGroup access
}

// GenerateContent generates content with automatic metering
func (m *ModelsInterface) GenerateContent(
	ctx context.Context,
	model string,
	contents []*genai.Content,
	config *genai.GenerateContentConfig,
) (*genai.GenerateContentResponse, error) {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	Debug("GenerateContent called with model: %s", model)

	// Record start time for duration calculation
	requestTime := time.Now()

	// Call Google Genai API
	resp, err := m.client.Models.GenerateContent(ctx, model, contents, config)

	// Record completion time
	completionStartTime := time.Now()
	responseTime := completionStartTime

	if err != nil {
		Debug("GenerateContent error: %v", err)
		// Send metering for failed request
		m.parent.wg.Add(1)
		go func() {
			defer m.parent.wg.Done()
			m.sendMeteringDataWithTiming(ctx, nil, model, metadata, false, requestTime, completionStartTime, responseTime, config, err)
		}()
		return nil, err
	}

	// Calculate duration
	duration := time.Since(requestTime)

	Debug("GenerateContent completed in %v, tokens: %d", duration, resp.UsageMetadata.TotalTokenCount)

	// Send metering data asynchronously (fire-and-forget)
	m.parent.wg.Add(1)
	go func() {
		defer m.parent.wg.Done()
		m.sendMeteringDataWithTiming(ctx, resp, model, metadata, false, requestTime, completionStartTime, responseTime, config, nil)
	}()

	return resp, nil
}

// GenerateContentStream generates streaming content with automatic metering
func (m *ModelsInterface) GenerateContentStream(
	ctx context.Context,
	model string,
	contents []*genai.Content,
	config *genai.GenerateContentConfig,
) iter.Seq2[*genai.GenerateContentResponse, error] {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	Debug("GenerateContentStream called with model: %s", model)

	// Record start time for duration calculation
	requestTime := time.Now()

	// Call Google Genai API
	stream := m.client.Models.GenerateContentStream(ctx, model, contents, config)

	// Wrap the stream to capture usage metadata and send metering
	return func(yield func(*genai.GenerateContentResponse, error) bool) {
		var lastUsage *genai.GenerateContentResponseUsageMetadata
		var completionStartTime time.Time
		var firstTokenReceived bool
		chunkCount := 0

		for resp, err := range stream {
			if err != nil {
				Debug("Stream error after %d chunks: %v", chunkCount, err)
				responseTime := time.Now()
				// Send metering before yielding error
				if lastUsage != nil {
					if !firstTokenReceived {
						completionStartTime = responseTime
					}
					m.parent.wg.Add(1)
					go func() {
						defer m.parent.wg.Done()
						m.sendMeteringDataWithTiming(ctx, &genai.GenerateContentResponse{UsageMetadata: lastUsage}, model, metadata, true, requestTime, completionStartTime, responseTime, config, err)
					}()
				} else {
					// No usage data, but still send error metering
					m.parent.wg.Add(1)
					go func() {
						defer m.parent.wg.Done()
						m.sendMeteringDataWithTiming(ctx, nil, model, metadata, true, requestTime, responseTime, responseTime, config, err)
					}()
				}
				if !yield(nil, err) {
					return
				}
				return
			}

			chunkCount++

			// Record time of first token
			if !firstTokenReceived {
				completionStartTime = time.Now()
				firstTokenReceived = true
			}

			// Capture usage metadata
			if resp.UsageMetadata != nil {
				lastUsage = resp.UsageMetadata
			}

			// Yield the response
			if !yield(resp, nil) {
				// Stream was stopped, send metering
				Debug("Stream stopped by consumer after %d chunks", chunkCount)
				responseTime := time.Now()
				if lastUsage != nil {
					m.parent.wg.Add(1)
					go func() {
						defer m.parent.wg.Done()
						m.sendMeteringDataWithTiming(ctx, &genai.GenerateContentResponse{UsageMetadata: lastUsage}, model, metadata, true, requestTime, completionStartTime, responseTime, config, nil)
					}()
				}
				return
			}
		}

		// Stream completed successfully, send final metering
		responseTime := time.Now()
		if lastUsage != nil {
			duration := time.Since(requestTime)
			Debug("Stream completed: %d chunks, %d total tokens in %v", chunkCount, lastUsage.TotalTokenCount, duration)
			m.parent.wg.Add(1)
			go func() {
				defer m.parent.wg.Done()
				m.sendMeteringDataWithTiming(ctx, &genai.GenerateContentResponse{UsageMetadata: lastUsage}, model, metadata, true, requestTime, completionStartTime, responseTime, config, nil)
			}()
		}
	}
}

// sendMeteringDataWithTiming sends metering data with precise timing information
func (m *ModelsInterface) sendMeteringDataWithTiming(
	ctx context.Context,
	resp *genai.GenerateContentResponse,
	model string,
	metadata map[string]interface{},
	isStreamed bool,
	requestTime time.Time,
	completionStartTime time.Time,
	responseTime time.Time,
	config *genai.GenerateContentConfig,
	err error,
) {
	defer func() {
		if r := recover(); r != nil {
			Error("Metering goroutine panic: %v", r)
		}
	}()

	// Build metering payload with precise timing
	payload := buildGoogleMeteringPayloadWithTiming(
		resp,
		model,
		metadata,
		isStreamed,
		requestTime,
		completionStartTime,
		responseTime,
		m.provider.String(),
		config,
		err,
	)

	// Send to Revenium API with retry logic
	Debug("[METERING] About to send metering data...")
	if err := sendMeteringWithRetry(m.config, payload); err != nil {
		Error("Failed to send metering data: %v", err)
	} else {
		Debug("[METERING] Metering data sent successfully")
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

// buildGoogleMeteringPayloadWithTiming builds a metering payload with precise timing information
func buildGoogleMeteringPayloadWithTiming(
	resp *genai.GenerateContentResponse,
	model string,
	metadata map[string]interface{},
	isStreamed bool,
	requestTime time.Time,
	completionStartTime time.Time,
	responseTime time.Time,
	provider string,
	config *genai.GenerateContentConfig,
	err error,
) map[string]interface{} {
	// Format timestamps as ISO 8601
	requestTimeISO := requestTime.UTC().Format(time.RFC3339)
	completionStartTimeISO := completionStartTime.UTC().Format(time.RFC3339)
	responseTimeISO := responseTime.UTC().Format(time.RFC3339)

	// Calculate durations
	requestDuration := responseTime.Sub(requestTime).Milliseconds()
	timeToFirstToken := completionStartTime.Sub(requestTime).Milliseconds()

	// Initialize token counts
	var inputTokens, outputTokens, totalTokens, cachedTokens, thinkingTokens int64

	// Extract usage metadata
	var usage *genai.GenerateContentResponseUsageMetadata
	if resp != nil {
		usage = resp.UsageMetadata
	}

	if usage != nil {
		// Calculate total tokens (fields are int32, not pointers)
		// Use TotalTokenCount if available, otherwise calculate
		totalTokens = int64(usage.TotalTokenCount)
		if totalTokens == 0 {
			totalTokens = int64(usage.PromptTokenCount + usage.CandidatesTokenCount)
		}

		// Get individual token counts
		inputTokens = int64(usage.PromptTokenCount)
		outputTokens = int64(usage.CandidatesTokenCount)
		cachedTokens = int64(usage.CachedContentTokenCount)
		thinkingTokens = int64(usage.ThoughtsTokenCount)
	}

	// Extract and map finish reason to stop reason
	finishReason := ExtractFinishReason(resp)
	stopReason := string(MapGoogleFinishReason(finishReason, StopReasonEnd))

	payload := map[string]interface{}{
		"stopReason":              stopReason,
		"costType":                defaultCostType,
		"isStreamed":              isStreamed,
		"operationType":           defaultOperationType,
		"inputTokenCount":         inputTokens,
		"outputTokenCount":        outputTokens,
		"reasoningTokenCount":     thinkingTokens,
		"cacheCreationTokenCount": int64(0),
		"cacheReadTokenCount":     cachedTokens,
		"totalTokenCount":         totalTokens,
		"model":                   model,
		"transactionId":           generateRequestID(),
		"responseTime":            responseTimeISO,
		"requestDuration":         requestDuration,
		"provider":                provider,
		"requestTime":             requestTimeISO,
		"completionStartTime":     completionStartTimeISO,
		"timeToFirstToken":        timeToFirstToken,
		"middlewareSource":        GetMiddlewareSource(),
	}

	// Add error reason if there was an error
	if err != nil {
		payload["errorReason"] = err.Error()
	}

	// Add temperature if available in config
	if config != nil && config.Temperature != nil {
		payload["temperature"] = *config.Temperature
	}

	// Add metadata fields if they exist
	if metadata != nil {
		if organizationId, ok := metadata["organizationId"]; ok {
			payload["organizationId"] = organizationId
		}
		if productId, ok := metadata["productId"]; ok {
			payload["productId"] = productId
		}
		if taskType, ok := metadata["taskType"]; ok {
			payload["taskType"] = taskType
		}
		if agent, ok := metadata["agent"]; ok {
			payload["agent"] = agent
		}
		if subscriptionId, ok := metadata["subscriptionId"]; ok {
			payload["subscriptionId"] = subscriptionId
		}
		if traceId, ok := metadata["traceId"]; ok {
			payload["traceId"] = traceId
		}
		if subscriber, ok := metadata["subscriber"]; ok {
			payload["subscriber"] = subscriber
		}
		if responseQualityScore, ok := metadata["responseQualityScore"]; ok {
			payload["responseQualityScore"] = responseQualityScore
		}
		if taskId, ok := metadata["taskId"]; ok {
			payload["taskId"] = taskId
		}
		if modelSource, ok := metadata["modelSource"]; ok {
			payload["modelSource"] = modelSource
		}
		if mediationLatency, ok := metadata["mediationLatency"]; ok {
			payload["mediationLatency"] = mediationLatency
		}
		// Temperature from metadata (overrides config if set)
		if temperature, ok := metadata["temperature"]; ok {
			payload["temperature"] = temperature
		}

		// Trace visualization fields (10 fields for distributed tracing)
		if transactionId, ok := metadata["transactionId"]; ok {
			payload["transactionId"] = transactionId
		}
		if traceType, ok := metadata["traceType"]; ok {
			payload["traceType"] = traceType
		}
		if traceName, ok := metadata["traceName"]; ok {
			payload["traceName"] = traceName
		}
		if environment, ok := metadata["environment"]; ok {
			payload["environment"] = environment
		}
		if region, ok := metadata["region"]; ok {
			payload["region"] = region
		}
		// NOTE: operationType is fixed to "CHAT" and should NOT be overridden by metadata
		// (API only accepts: CHAT, GENERATE, EMBED, CLASSIFY, SUMMARIZE, TRANSLATE, OTHER)
		// operationSubtype is auto-detected, not user-provided
		if retryNumber, ok := metadata["retryNumber"]; ok {
			payload["retryNumber"] = retryNumber
		}
		if credentialAlias, ok := metadata["credentialAlias"]; ok {
			payload["credentialAlias"] = credentialAlias
		}
		if parentTransactionId, ok := metadata["parentTransactionId"]; ok {
			payload["parentTransactionId"] = parentTransactionId
		}
	}

	return payload
}

// sendMeteringWithRetry sends metering data with exponential backoff retry
func sendMeteringWithRetry(config *Config, payload map[string]interface{}) error {
	const maxRetries = 3
	const initialBackoff = 100 * time.Millisecond

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		err := sendMeteringRequest(config, payload)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on validation errors
		if IsValidationError(err) {
			return err
		}
	}

	return NewMeteringError("metering failed after retries", fmt.Errorf("retries: %d, last error: %w", maxRetries, lastErr))
}

// sendMeteringRequest sends a single metering request to Revenium API
func sendMeteringRequest(config *Config, payload map[string]interface{}) error {
	if config == nil || config.ReveniumAPIKey == "" {
		return NewConfigError("metering not configured", nil)
	}

	url := config.ReveniumBaseURL + meteringEndpoint

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return NewMeteringError("failed to marshal metering payload", err)
	}

	// Log the exact payload being sent
	Debug("[METERING] Sending payload to %s: %s", url, string(jsonData))

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return NewMeteringError("failed to create metering request", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("x-api-key", config.ReveniumAPIKey)
	req.Header.Set("User-Agent", GetUserAgent())

	// Send request with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		Error("[METERING] Network error: %v", err)
		return NewNetworkError("metering request failed", err)
	}
	defer resp.Body.Close()

	// Read response body for error details
	body, _ := io.ReadAll(resp.Body)

	Debug("[METERING] Response status: %d, body: %s", resp.StatusCode, string(body))

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Log response for debugging
		Error("[METERING] API error response (status %d): %s", resp.StatusCode, string(body))
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			// Validation error - don't retry
			return NewValidationError(
				fmt.Sprintf("metering API returned %d: %s", resp.StatusCode, string(body)),
				nil,
			)
		}
		return NewMeteringError("metering API error", fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
	}

	Debug("[METERING] Successfully sent metering data (status %d)", resp.StatusCode)
	return nil
}

// Reset resets the global middleware state for testing
func Reset() {
	globalMu.Lock()
	defer globalMu.Unlock()

	if globalClient != nil {
		globalClient.Close()
		globalClient = nil
	}

	initialized = false
}
