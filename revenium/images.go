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
	imageMeteringEndpoint = "/meter/v2/ai/images"
)

// Images returns the images interface for generating images with metering
func (r *ReveniumGoogle) Images() *ImagesInterface {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &ImagesInterface{
		client:   r.client,
		config:   r.config,
		provider: r.provider,
		parent:   r,
	}
}

// ImagesInterface provides methods for image generation with metering (Imagen)
type ImagesInterface struct {
	client   *genai.Client
	config   *Config
	provider Provider
	parent   *ReveniumGoogle
}

// GenerateImages generates images using Google Imagen with automatic metering
func (i *ImagesInterface) GenerateImages(ctx context.Context, model string, prompt string, config *genai.GenerateImagesConfig) (*genai.GenerateImagesResponse, error) {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	// Record start time
	requestTime := time.Now()

	// Get requested image count from config (default is 1)
	requestedCount := 1
	if config != nil && config.NumberOfImages > 0 {
		requestedCount = int(config.NumberOfImages)
	}

	Debug("GenerateImages called with model: %s, prompt length: %d", model, len(prompt))

	// Call Google Imagen API
	resp, err := i.client.Models.GenerateImages(ctx, model, prompt, config)

	if err != nil {
		duration := time.Since(requestTime)
		Debug("GenerateImages error: %v", err)
		i.parent.wg.Add(1)
		go func() {
			defer i.parent.wg.Done()
			i.sendImageMeteringForError(ctx, model, metadata, duration, requestTime, err.Error(), requestedCount)
		}()
		return nil, err
	}

	// Calculate duration
	duration := time.Since(requestTime)

	// Count actual images returned
	actualCount := 0
	if resp != nil && resp.GeneratedImages != nil {
		actualCount = len(resp.GeneratedImages)
	}

	Debug("GenerateImages completed in %v, images generated: %d", duration, actualCount)

	// Send metering data asynchronously
	i.parent.wg.Add(1)
	go func() {
		defer i.parent.wg.Done()
		i.sendImageMeteringData(ctx, resp, model, metadata, duration, requestTime, requestedCount, config)
	}()

	return resp, nil
}

// EditImage edits images using Google Imagen with automatic metering
func (i *ImagesInterface) EditImage(ctx context.Context, model, prompt string, referenceImages []genai.ReferenceImage, config *genai.EditImageConfig) (*genai.EditImageResponse, error) {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	// Record start time
	requestTime := time.Now()

	// Get requested image count from config (default is 1)
	requestedCount := 1
	if config != nil && config.NumberOfImages > 0 {
		requestedCount = int(config.NumberOfImages)
	}

	Debug("EditImage called with model: %s, prompt length: %d", model, len(prompt))

	// Call Google Imagen Edit API
	resp, err := i.client.Models.EditImage(ctx, model, prompt, referenceImages, config)

	if err != nil {
		duration := time.Since(requestTime)
		Debug("EditImage error: %v", err)
		i.parent.wg.Add(1)
		go func() {
			defer i.parent.wg.Done()
			i.sendImageMeteringForError(ctx, model, metadata, duration, requestTime, err.Error(), requestedCount)
		}()
		return nil, err
	}

	// Calculate duration
	duration := time.Since(requestTime)

	// Count actual images returned
	actualCount := 0
	if resp != nil && resp.GeneratedImages != nil {
		actualCount = len(resp.GeneratedImages)
	}

	Debug("EditImage completed in %v, images generated: %d", duration, actualCount)

	// Send metering data asynchronously
	i.parent.wg.Add(1)
	go func() {
		defer i.parent.wg.Done()
		i.sendEditImageMeteringData(ctx, resp, model, metadata, duration, requestTime, requestedCount, config)
	}()

	return resp, nil
}

// UpscaleImage upscales images using Google Imagen with automatic metering
func (i *ImagesInterface) UpscaleImage(ctx context.Context, model string, image *genai.Image, upscaleFactor string, config *genai.UpscaleImageConfig) (*genai.UpscaleImageResponse, error) {
	// Extract metadata from context
	metadata := GetUsageMetadata(ctx)

	// Record start time
	requestTime := time.Now()

	Debug("UpscaleImage called with model: %s, upscaleFactor: %s", model, upscaleFactor)

	// Call Google Imagen Upscale API
	resp, err := i.client.Models.UpscaleImage(ctx, model, image, upscaleFactor, config)

	if err != nil {
		duration := time.Since(requestTime)
		Debug("UpscaleImage error: %v", err)
		i.parent.wg.Add(1)
		go func() {
			defer i.parent.wg.Done()
			i.sendImageMeteringForError(ctx, model, metadata, duration, requestTime, err.Error(), 1)
		}()
		return nil, err
	}

	// Calculate duration
	duration := time.Since(requestTime)

	Debug("UpscaleImage completed in %v", duration)

	// Send metering data asynchronously
	i.parent.wg.Add(1)
	go func() {
		defer i.parent.wg.Done()
		i.sendUpscaleMeteringData(ctx, resp, model, metadata, duration, requestTime, upscaleFactor)
	}()

	return resp, nil
}

// sendImageMeteringData sends metering data for image generation
func (i *ImagesInterface) sendImageMeteringData(ctx context.Context, resp *genai.GenerateImagesResponse, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, requestedCount int, config *genai.GenerateImagesConfig) {
	defer func() {
		if r := recover(); r != nil {
			Error("Image metering goroutine panic: %v", r)
		}
	}()

	// Build payload
	payload := i.buildImageMeteringPayload(resp, model, metadata, duration, requestTime, requestedCount, config)

	Debug("[METERING] Sending image metering data...")
	if err := i.sendImageMeteringRequest(payload); err != nil {
		Error("Failed to send image metering data: %v", err)
	} else {
		Debug("[METERING] Image metering data sent successfully")
	}
}

// sendEditImageMeteringData sends metering data for image editing
func (i *ImagesInterface) sendEditImageMeteringData(ctx context.Context, resp *genai.EditImageResponse, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, requestedCount int, config *genai.EditImageConfig) {
	defer func() {
		if r := recover(); r != nil {
			Error("Image metering goroutine panic: %v", r)
		}
	}()

	// Build payload
	payload := i.buildEditImageMeteringPayload(resp, model, metadata, duration, requestTime, requestedCount, config)

	Debug("[METERING] Sending edit image metering data...")
	if err := i.sendImageMeteringRequest(payload); err != nil {
		Error("Failed to send edit image metering data: %v", err)
	} else {
		Debug("[METERING] Edit image metering data sent successfully")
	}
}

// sendUpscaleMeteringData sends metering data for image upscaling
func (i *ImagesInterface) sendUpscaleMeteringData(ctx context.Context, resp *genai.UpscaleImageResponse, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, upscaleFactor string) {
	defer func() {
		if r := recover(); r != nil {
			Error("Image metering goroutine panic: %v", r)
		}
	}()

	// Build payload
	payload := i.buildUpscaleMeteringPayload(resp, model, metadata, duration, requestTime, upscaleFactor)

	Debug("[METERING] Sending upscale metering data...")
	if err := i.sendImageMeteringRequest(payload); err != nil {
		Error("Failed to send upscale metering data: %v", err)
	} else {
		Debug("[METERING] Upscale metering data sent successfully")
	}
}

// sendImageMeteringForError sends metering data for failed image generation
func (i *ImagesInterface) sendImageMeteringForError(ctx context.Context, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, errorReason string, requestedCount int) {
	defer func() {
		if r := recover(); r != nil {
			Error("Image error metering goroutine panic: %v", r)
		}
	}()

	payload := i.buildImageErrorMeteringPayload(model, metadata, duration, requestTime, errorReason, requestedCount)

	Debug("[METERING] Sending image error metering data...")
	if err := i.sendImageMeteringRequest(payload); err != nil {
		Error("Failed to send image error metering data: %v", err)
	} else {
		Debug("[METERING] Image error metering data sent successfully")
	}
}

// buildImageMeteringPayload builds the metering payload for image generation
func (i *ImagesInterface) buildImageMeteringPayload(resp *genai.GenerateImagesResponse, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, requestedCount int, config *genai.GenerateImagesConfig) map[string]interface{} {
	responseTime := time.Now().UTC()
	responseTimeISO := responseTime.Format(time.RFC3339)
	requestTimeISO := requestTime.UTC().Format(time.RFC3339)

	// Count actual images returned
	actualCount := 0
	if resp != nil && resp.GeneratedImages != nil {
		actualCount = len(resp.GeneratedImages)
	}

	// Build attributes for additional info
	attributes := make(map[string]interface{})
	if config != nil {
		if config.AspectRatio != "" {
			attributes["aspectRatio"] = config.AspectRatio
		}
		if config.OutputMIMEType != "" {
			attributes["outputMimeType"] = config.OutputMIMEType
		}
		if config.PersonGeneration != "" {
			attributes["personGeneration"] = string(config.PersonGeneration)
		}
	}

	payload := map[string]interface{}{
		"stopReason":          "END",
		"costType":            defaultCostType,
		"operationType":       "IMAGE",
		"model":               model,
		"provider":            i.provider.String(),
		"transactionId":       generateRequestID(),
		"requestTime":         requestTimeISO,
		"responseTime":        responseTimeISO,
		"requestDuration":     duration.Milliseconds(),
		"middlewareSource":    GetMiddlewareSource(),
		// Image-specific billing fields (TOP LEVEL per API contract)
		"actualImageCount":    actualCount,
		"requestedImageCount": requestedCount,
	}

	// Add attributes if any
	if len(attributes) > 0 {
		payload["attributes"] = attributes
	}

	// Add metadata fields
	addGoogleMetadataToPayload(payload, metadata)

	return payload
}

// buildEditImageMeteringPayload builds the metering payload for image editing
func (i *ImagesInterface) buildEditImageMeteringPayload(resp *genai.EditImageResponse, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, requestedCount int, config *genai.EditImageConfig) map[string]interface{} {
	responseTime := time.Now().UTC()
	responseTimeISO := responseTime.Format(time.RFC3339)
	requestTimeISO := requestTime.UTC().Format(time.RFC3339)

	// Count actual images returned
	actualCount := 0
	if resp != nil && resp.GeneratedImages != nil {
		actualCount = len(resp.GeneratedImages)
	}

	// Build attributes for additional info
	attributes := make(map[string]interface{})
	attributes["operationSubtype"] = "edit"
	if config != nil {
		if config.OutputMIMEType != "" {
			attributes["outputMimeType"] = config.OutputMIMEType
		}
		if config.PersonGeneration != "" {
			attributes["personGeneration"] = string(config.PersonGeneration)
		}
	}

	payload := map[string]interface{}{
		"stopReason":          "END",
		"costType":            defaultCostType,
		"operationType":       "IMAGE",
		"model":               model,
		"provider":            i.provider.String(),
		"transactionId":       generateRequestID(),
		"requestTime":         requestTimeISO,
		"responseTime":        responseTimeISO,
		"requestDuration":     duration.Milliseconds(),
		"middlewareSource":    GetMiddlewareSource(),
		// Image-specific billing fields
		"actualImageCount":    actualCount,
		"requestedImageCount": requestedCount,
	}

	// Add attributes if any
	if len(attributes) > 0 {
		payload["attributes"] = attributes
	}

	// Add metadata fields
	addGoogleMetadataToPayload(payload, metadata)

	return payload
}

// buildUpscaleMeteringPayload builds the metering payload for image upscaling
func (i *ImagesInterface) buildUpscaleMeteringPayload(resp *genai.UpscaleImageResponse, model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, upscaleFactor string) map[string]interface{} {
	responseTime := time.Now().UTC()
	responseTimeISO := responseTime.Format(time.RFC3339)
	requestTimeISO := requestTime.UTC().Format(time.RFC3339)

	// Upscale returns 1 image
	actualCount := 0
	if resp != nil && resp.GeneratedImages != nil {
		actualCount = len(resp.GeneratedImages)
	}

	// Build attributes for additional info
	attributes := make(map[string]interface{})
	attributes["operationSubtype"] = "upscale"
	attributes["upscaleFactor"] = upscaleFactor

	payload := map[string]interface{}{
		"stopReason":          "END",
		"costType":            defaultCostType,
		"operationType":       "IMAGE",
		"model":               model,
		"provider":            i.provider.String(),
		"transactionId":       generateRequestID(),
		"requestTime":         requestTimeISO,
		"responseTime":        responseTimeISO,
		"requestDuration":     duration.Milliseconds(),
		"middlewareSource":    GetMiddlewareSource(),
		// Image-specific billing fields
		"actualImageCount":    actualCount,
		"requestedImageCount": 1,
		"attributes":          attributes,
	}

	// Add metadata fields
	addGoogleMetadataToPayload(payload, metadata)

	return payload
}

// buildImageErrorMeteringPayload builds the metering payload for failed image generation
func (i *ImagesInterface) buildImageErrorMeteringPayload(model string, metadata map[string]interface{}, duration time.Duration, requestTime time.Time, errorReason string, requestedCount int) map[string]interface{} {
	responseTime := time.Now().UTC()
	responseTimeISO := responseTime.Format(time.RFC3339)
	requestTimeISO := requestTime.UTC().Format(time.RFC3339)

	payload := map[string]interface{}{
		"stopReason":          "ERROR",
		"costType":            defaultCostType,
		"operationType":       "IMAGE",
		"model":               model,
		"provider":            i.provider.String(),
		"transactionId":       generateRequestID(),
		"requestTime":         requestTimeISO,
		"responseTime":        responseTimeISO,
		"requestDuration":     duration.Milliseconds(),
		"middlewareSource":    GetMiddlewareSource(),
		"errorReason":         errorReason,
		// Image-specific billing fields
		"actualImageCount":    0,
		"requestedImageCount": requestedCount,
	}

	// Add metadata fields
	addGoogleMetadataToPayload(payload, metadata)

	return payload
}

// sendImageMeteringRequest sends the metering request to the images endpoint
func (i *ImagesInterface) sendImageMeteringRequest(payload map[string]interface{}) error {
	const maxRetries = 3
	const initialBackoff = 100 * time.Millisecond

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(backoff)
			backoff *= 2
		}

		err := i.doImageMeteringRequest(payload)
		if err == nil {
			return nil
		}

		lastErr = err

		if IsValidationError(err) {
			return err
		}
	}

	return NewMeteringError(fmt.Sprintf("image metering failed after %d retries", maxRetries), lastErr)
}

// doImageMeteringRequest sends a single metering request
func (i *ImagesInterface) doImageMeteringRequest(payload map[string]interface{}) error {
	baseURL := i.config.ReveniumBaseURL
	if baseURL == "" {
		baseURL = defaultReveniumBaseURL
	}
	url := baseURL + imageMeteringEndpoint

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return NewMeteringError("failed to marshal image metering payload", err)
	}

	Debug("Sending image metering request to %s", url)
	Debug("Payload: %s", string(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return NewMeteringError("failed to create image metering request", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("x-api-key", i.config.ReveniumAPIKey)
	req.Header.Set("User-Agent", GetUserAgent())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return NewNetworkError("image metering request failed", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return NewValidationError(
				fmt.Sprintf("image metering API returned %d: %s", resp.StatusCode, string(body)),
				nil,
			)
		}
		return NewMeteringError("image metering API error", fmt.Errorf("status %d: %s", resp.StatusCode, string(body)))
	}

	Debug("Image metering request successful")
	return nil
}

// addGoogleMetadataToPayload adds metadata fields to the payload
func addGoogleMetadataToPayload(payload map[string]interface{}, metadata map[string]interface{}) {
	if metadata == nil {
		return
	}

	metadataFields := []string{
		// Core tracking fields
		"organizationId", "productId", "taskType", "taskId", "agent", "subscriptionId",
		"traceId", "transactionId", "subscriber", "responseQualityScore",
		"modelSource", "temperature", "mediationLatency",
		// Trace visualization fields
		"traceType", "traceName", "environment", "region",
		"retryNumber", "credentialAlias", "parentTransactionId",
	}

	for _, field := range metadataFields {
		if value, ok := metadata[field]; ok {
			payload[field] = value
		}
	}
}
