package revenium

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/genai"
)

const (
	videoMeteringEndpoint = "/meter/v2/ai/video"
)

// Videos returns the videos interface for generating videos with metering
func (r *ReveniumGoogle) Videos() *VideosInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &VideosInterface{
		client:   r.client,
		config:   r.config,
		provider: r.provider,
		parent:   r,
	}
}

// VideosInterface provides methods for video generation with metering (Veo)
type VideosInterface struct {
	client   *genai.Client
	config   *Config
	provider Provider
	parent   *ReveniumGoogle
}

// GenerateVideos starts video generation using Google Veo with automatic metering
// Video generation is asynchronous - returns an operation that can be polled
// Use WaitForVideoGeneration to wait for completion with metering
func (v *VideosInterface) GenerateVideos(ctx context.Context, model string, prompt string, image *genai.Image, config *genai.GenerateVideosConfig) (*genai.GenerateVideosOperation, error) {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	// Record start time
	requestTime := time.Now()

	Debug("GenerateVideos called with model: %s, prompt length: %d", model, len(prompt))

	// Get requested video count from config (default is 1)
	requestedCount := 1
	if config != nil && config.NumberOfVideos > 0 {
		requestedCount = int(config.NumberOfVideos)
	}

	// Call Google Veo API
	operation, err := v.client.Models.GenerateVideos(ctx, model, prompt, image, config)

	if err != nil {
		duration := time.Since(requestTime)
		Debug("GenerateVideos error: %v", err)
		v.parent.wg.Add(1)
		go func() {
			defer v.parent.wg.Done()
			v.sendVideoMeteringForError(ctx, model, metadata, duration, requestTime, err.Error(), requestedCount)
		}()
		return nil, err
	}

	// Calculate duration (time to initiate the operation)
	duration := time.Since(requestTime)

	Debug("GenerateVideos operation started in %v, operation name: %s", duration, operation.Name)

	// Send metering data for operation start asynchronously
	// Note: This meters the operation initiation. Use WaitForVideoGeneration for final metering
	v.parent.wg.Add(1)
	go func() {
		defer v.parent.wg.Done()
		v.sendVideoOperationStartMetering(ctx, operation, model, metadata, duration, requestTime, requestedCount, config)
	}()

	return operation, nil
}

// WaitForVideoGeneration polls an operation until complete and meters the final result
func (v *VideosInterface) WaitForVideoGeneration(ctx context.Context, operation *genai.GenerateVideosOperation, model string, pollInterval time.Duration, timeout time.Duration) (*genai.GenerateVideosResponse, error) {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	// Record start time for total wait duration
	waitStartTime := time.Now()

	// Default poll interval
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}

	// Default timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	Debug("WaitForVideoGeneration started, polling every %v with timeout %v", pollInterval, timeout)

	// Create a ticker for polling
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled or timed out
			duration := time.Since(waitStartTime)
			err := ctx.Err()
			Debug("WaitForVideoGeneration timeout/cancelled: %v", err)
			v.parent.wg.Add(1)
			go func() {
				defer v.parent.wg.Done()
				v.sendVideoMeteringForError(ctx, model, metadata, duration, waitStartTime, fmt.Sprintf("operation timeout: %v", err), 0)
			}()
			return nil, err

		case <-ticker.C:
			// Poll operation status
			updatedOp, err := v.client.Operations.GetVideosOperation(ctx, operation, nil)
			if err != nil {
				Error("Failed to get operation status: %v", err)
				continue
			}

			if updatedOp.Done {
				duration := time.Since(waitStartTime)

				// Check for error
				if updatedOp.Error != nil && len(updatedOp.Error) > 0 {
					errStr := fmt.Sprintf("%v", updatedOp.Error)
					Debug("Video generation failed: %s", errStr)
					v.parent.wg.Add(1)
					go func() {
						defer v.parent.wg.Done()
						v.sendVideoMeteringForError(ctx, model, metadata, duration, waitStartTime, errStr, 0)
					}()
					return nil, fmt.Errorf("video generation failed: %v", updatedOp.Error)
				}

				// Success
				if updatedOp.Response != nil {
					actualCount := len(updatedOp.Response.GeneratedVideos)
					Debug("Video generation completed in %v, videos generated: %d", duration, actualCount)
					v.parent.wg.Add(1)
					go func() {
						defer v.parent.wg.Done()
						v.sendVideoCompletionMetering(ctx, updatedOp.Response, model, metadata, duration, waitStartTime)
					}()
					return updatedOp.Response, nil
				}

				// Operation done but no response
				Debug("Video generation completed but no response")
				return nil, fmt.Errorf("video generation completed but no response")
			}

			Debug("Video generation still in progress...")
		}
	}
}

// sendVideoOperationStartMetering sends metering data for video generation operation start
func (v *VideosInterface) sendVideoOperationStartMetering(ctx context.Context, operation *genai.GenerateVideosOperation, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, requestedCount int, config *genai.GenerateVideosConfig) {
	defer func() {
		if r := recover(); r != nil {
			Error("Video metering goroutine panic: %v", r)
		}
	}()

	// Build payload
	payload := v.buildVideoOperationStartPayload(operation, model, metadata, duration, requestTime, requestedCount, config)

	Debug("[METERING] Sending video operation start metering data...")
	if err := v.sendVideoMeteringRequest(payload); err != nil {
		Error("Failed to send video operation start metering data: %v", err)
	} else {
		Debug("[METERING] Video operation start metering data sent successfully")
	}
}

// sendVideoCompletionMetering sends metering data for completed video generation
func (v *VideosInterface) sendVideoCompletionMetering(ctx context.Context, resp *genai.GenerateVideosResponse, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time) {
	defer func() {
		if r := recover(); r != nil {
			Error("Video metering goroutine panic: %v", r)
		}
	}()

	// Build payload
	payload := v.buildVideoCompletionPayload(resp, model, metadata, duration, requestTime)

	Debug("[METERING] Sending video completion metering data...")
	if err := v.sendVideoMeteringRequest(payload); err != nil {
		Error("Failed to send video completion metering data: %v", err)
	} else {
		Debug("[METERING] Video completion metering data sent successfully")
	}
}

// sendVideoMeteringForError sends metering data for failed video generation
func (v *VideosInterface) sendVideoMeteringForError(ctx context.Context, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, errorReason string, requestedCount int) {
	defer func() {
		if r := recover(); r != nil {
			Error("Video error metering goroutine panic: %v", r)
		}
	}()

	payload := v.buildVideoErrorMeteringPayload(model, metadata, duration, requestTime, errorReason, requestedCount)

	Debug("[METERING] Sending video error metering data...")
	if err := v.sendVideoMeteringRequest(payload); err != nil {
		Error("Failed to send video error metering data: %v", err)
	} else {
		Debug("[METERING] Video error metering data sent successfully")
	}
}

// buildVideoOperationStartPayload builds the metering payload for video operation start
func (v *VideosInterface) buildVideoOperationStartPayload(operation *genai.GenerateVideosOperation, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, requestedCount int, config *genai.GenerateVideosConfig) map[string]interface{} {
	responseTime := time.Now().UTC()
	responseTimeISO := responseTime.Format(time.RFC3339)
	requestTimeISO := requestTime.UTC().Format(time.RFC3339)

	// Build attributes
	attributes := make(map[string]interface{})
	attributes["operationPhase"] = "start"
	if operation != nil && operation.Name != "" {
		attributes["operationName"] = operation.Name
	}
	if config != nil {
		if config.AspectRatio != "" {
			attributes["aspectRatio"] = config.AspectRatio
		}
	}

	payload := map[string]interface{}{
		"stopReason":          "PENDING", // Operation started but not complete
		"costType":            defaultCostType,
		"operationType":       "VIDEO",
		"model":               model,
		"provider":            v.provider.String(),
		"transactionId":       generateRequestID(),
		"requestTime":         requestTimeISO,
		"responseTime":        responseTimeISO,
		"requestDuration":     duration.Milliseconds(),
		"middlewareSource":    GetMiddlewareSource(),
		// Video-specific billing fields
		"actualVideoCount":    0, // Not complete yet
		"requestedVideoCount": requestedCount,
		"attributes":          attributes,
	}

	// Add metadata fields
	addGoogleMetadataToPayload(payload, metadata)

	return payload
}

// buildVideoCompletionPayload builds the metering payload for completed video generation
func (v *VideosInterface) buildVideoCompletionPayload(resp *genai.GenerateVideosResponse, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time) map[string]interface{} {
	responseTime := time.Now().UTC()
	responseTimeISO := responseTime.Format(time.RFC3339)
	requestTimeISO := requestTime.UTC().Format(time.RFC3339)

	// Count actual videos returned
	actualCount := 0
	requestedCount := 0
	if resp != nil && resp.GeneratedVideos != nil {
		actualCount = len(resp.GeneratedVideos)
		requestedCount = actualCount // Best estimate
	}

	// Build attributes
	attributes := make(map[string]interface{})
	attributes["operationPhase"] = "complete"

	// Add RAI filtering info if present
	if resp != nil && resp.RAIMediaFilteredCount > 0 {
		attributes["raiFilteredCount"] = resp.RAIMediaFilteredCount
		attributes["raiFilteredReasons"] = resp.RAIMediaFilteredReasons
	}

	payload := map[string]interface{}{
		"stopReason":          "END",
		"costType":            defaultCostType,
		"operationType":       "VIDEO",
		"model":               model,
		"provider":            v.provider.String(),
		"transactionId":       generateRequestID(),
		"requestTime":         requestTimeISO,
		"responseTime":        responseTimeISO,
		"requestDuration":     duration.Milliseconds(),
		"middlewareSource":    GetMiddlewareSource(),
		// Video-specific billing fields
		"actualVideoCount":    actualCount,
		"requestedVideoCount": requestedCount,
		"attributes":          attributes,
	}

	// Add metadata fields
	addGoogleMetadataToPayload(payload, metadata)

	return payload
}

// buildVideoErrorMeteringPayload builds the metering payload for failed video generation
func (v *VideosInterface) buildVideoErrorMeteringPayload(model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, errorReason string, requestedCount int) map[string]interface{} {
	responseTime := time.Now().UTC()
	responseTimeISO := responseTime.Format(time.RFC3339)
	requestTimeISO := requestTime.UTC().Format(time.RFC3339)

	payload := map[string]interface{}{
		"stopReason":          "ERROR",
		"costType":            defaultCostType,
		"operationType":       "VIDEO",
		"model":               model,
		"provider":            v.provider.String(),
		"transactionId":       generateRequestID(),
		"requestTime":         requestTimeISO,
		"responseTime":        responseTimeISO,
		"requestDuration":     duration.Milliseconds(),
		"middlewareSource":    GetMiddlewareSource(),
		"errorReason":         errorReason,
		// Video-specific billing fields
		"actualVideoCount":    0,
		"requestedVideoCount": requestedCount,
	}

	// Add metadata fields
	addGoogleMetadataToPayload(payload, metadata)

	return payload
}

// sendVideoMeteringRequest sends the metering request to the video endpoint
func (v *VideosInterface) sendVideoMeteringRequest(payload map[string]interface{}) error {
	const maxRetries = 3
	const initialBackoff = 100 * time.Millisecond

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2
		}

		err := v.doVideoMeteringRequest(payload)
		if err == nil {
			return nil
		}

		lastErr = err

		if IsValidationError(err) {
			return err
		}
	}

	return NewMeteringError(fmt.Sprintf("video metering failed after %d retries", maxRetries), lastErr)
}

// doVideoMeteringRequest sends a single metering request
func (v *VideosInterface) doVideoMeteringRequest(payload map[string]interface{}) error {
	baseURL := v.config.ReveniumBaseURL
	if baseURL == "" {
		baseURL = defaultReveniumBaseURL
	}
	url := baseURL + videoMeteringEndpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return NewMeteringError("failed to marshal video metering payload", err)
	}

	Debug("Sending video metering request to %s", url)
	Debug("Payload: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return NewMeteringError("failed to create video metering request", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("x-api-key", v.config.ReveniumAPIKey)
	req.Header.Set("User-Agent", GetUserAgent())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return NewNetworkError("video metering request failed", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return NewValidationError(
				fmt.Sprintf("video metering API returned %d: %s", resp.StatusCode, string(body)),
				nil,
			)
		}
		return NewMeteringError("video metering API error", fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
	}

	Debug("Video metering request successful")
	return nil
}
