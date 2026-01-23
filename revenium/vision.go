// Package revenium provides vision content detection for Google GenAI messages
// Detects image content in GenerateContent payloads for metering

package revenium

import (
	"strings"

	"google.golang.org/genai"
)

// VisionDetectionResult contains vision detection statistics
type VisionDetectionResult struct {
	// HasVisionContent indicates whether any vision/image content was found
	HasVisionContent bool
	// ImageCount is the number of images found in content
	ImageCount int
	// TotalImageSizeBytes is the estimated size of image data in bytes
	TotalImageSizeBytes int
	// MediaTypes contains the image media types found
	MediaTypes []string
}

// DetectVisionContent scans Google GenAI content for image content
func DetectVisionContent(contents []*genai.Content) VisionDetectionResult {
	result := VisionDetectionResult{
		HasVisionContent:    false,
		ImageCount:          0,
		TotalImageSizeBytes: 0,
		MediaTypes:          []string{},
	}

	if contents == nil {
		return result
	}

	// Iterate through all content objects
	for _, content := range contents {
		if content == nil || content.Parts == nil {
			continue
		}

		// Check each part for image content
		for _, part := range content.Parts {
			if part == nil {
				continue
			}

			// Check for inline image data (Blob)
			if part.InlineData != nil {
				processInlineData(part.InlineData, &result)
			}

			// Check for file data (could be images)
			if part.FileData != nil {
				processFileData(part.FileData, &result)
			}
		}
	}

	return result
}

// processInlineData extracts information from inline blob data
func processInlineData(blob *genai.Blob, result *VisionDetectionResult) {
	if blob == nil {
		return
	}

	// Check if this is image content based on MIME type
	if !isImageMimeType(blob.MIMEType) {
		return
	}

	result.HasVisionContent = true
	result.ImageCount++

	// Track media type
	if blob.MIMEType != "" && !containsString(result.MediaTypes, blob.MIMEType) {
		result.MediaTypes = append(result.MediaTypes, blob.MIMEType)
	}

	// Calculate size from raw data
	if blob.Data != nil {
		result.TotalImageSizeBytes += len(blob.Data)
	}
}

// processFileData extracts information from file data references
func processFileData(fileData *genai.FileData, result *VisionDetectionResult) {
	if fileData == nil {
		return
	}

	// Check if this is image content based on MIME type
	if !isImageMimeType(fileData.MIMEType) {
		return
	}

	result.HasVisionContent = true
	result.ImageCount++

	// Track media type
	if fileData.MIMEType != "" && !containsString(result.MediaTypes, fileData.MIMEType) {
		result.MediaTypes = append(result.MediaTypes, fileData.MIMEType)
	}

	// Note: We can't calculate size for file references (URL-based)
	// The size will be 0 for these cases
}

// isImageMimeType checks if a MIME type represents an image
func isImageMimeType(mimeType string) bool {
	if mimeType == "" {
		return false
	}

	// Check for image/* MIME types
	mimeTypeLower := strings.ToLower(mimeType)
	return strings.HasPrefix(mimeTypeLower, "image/")
}

// containsString checks if a string slice contains a value
func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// BuildVisionAttributes creates attributes map for metering payload
func BuildVisionAttributes(result VisionDetectionResult) map[string]interface{} {
	if !result.HasVisionContent {
		return nil
	}

	return map[string]interface{}{
		"vision_image_count":      result.ImageCount,
		"vision_total_size_bytes": result.TotalImageSizeBytes,
		"vision_media_types":      result.MediaTypes,
	}
}
