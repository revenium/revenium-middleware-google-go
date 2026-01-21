package revenium

import (
	"encoding/json"
	"unicode/utf8"

	"google.golang.org/genai"
)

// truncateUTF8Safe truncates a string to maxBytes while preserving UTF-8 validity.
// It ensures we don't cut in the middle of a multi-byte character.
func truncateUTF8Safe(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	// Find the last valid UTF-8 character boundary before maxBytes
	for maxBytes > 0 && maxBytes < len(s) {
		// If this byte is a continuation byte (10xxxxxx), back up
		if (s[maxBytes] & 0xC0) == 0x80 {
			maxBytes--
		} else {
			break
		}
	}

	// Validate the result is valid UTF-8
	result := s[:maxBytes]
	if !utf8.ValidString(result) {
		// If still invalid, use rune iteration as fallback
		runes := []rune(s)
		result = ""
		for _, r := range runes {
			if len(result)+utf8.RuneLen(r) > maxBytes {
				break
			}
			result += string(r)
		}
	}

	return result
}

const (
	// MaxPromptLength is the maximum length for captured prompts/responses
	// Fields exceeding this limit will be truncated
	MaxPromptLength = 50000

	// TruncationMarker is appended to truncated content
	TruncationMarker = "...[TRUNCATED]"
)

// PromptData holds extracted prompt information for metering
type PromptData struct {
	// SystemPrompt contains the system instruction content (if any)
	SystemPrompt string
	// InputMessages contains JSON-serialized user/model messages
	InputMessages string
	// OutputResponse contains the model's response content
	OutputResponse string
	// PromptsTruncated indicates if any field was truncated
	PromptsTruncated bool
}

// ExtractPromptsFromRequest extracts system prompt and input messages from Google AI request
func ExtractPromptsFromRequest(contents []*genai.Content, config *genai.GenerateContentConfig) PromptData {
	data := PromptData{}

	// Extract system instruction from config if present
	if config != nil && config.SystemInstruction != nil {
		systemContent := extractContentText(config.SystemInstruction)
		if systemContent != "" {
			// Apply truncation if needed
			if len(systemContent) > MaxPromptLength {
				markerLen := len(TruncationMarker)
				truncateAt := MaxPromptLength - markerLen
				systemContent = truncateUTF8Safe(systemContent, truncateAt) + TruncationMarker
				data.PromptsTruncated = true
				Debug("System prompt truncated to %d characters", MaxPromptLength)
			}
			data.SystemPrompt = systemContent
		}
	}

	// Extract input messages
	if len(contents) > 0 {
		var messages []map[string]interface{}
		halfLimit := MaxPromptLength / 2
		markerLen := len(TruncationMarker)

		for _, content := range contents {
			if content == nil {
				continue
			}

			role := content.Role
			if role == "" {
				role = "user" // Default to user if not specified
			}

			text := extractContentText(content)
			messageMap := map[string]interface{}{
				"role":    role,
				"content": text,
			}

			// Truncate individual message content if too long
			if len(text) > halfLimit {
				truncateAt := halfLimit - markerLen
				messageMap["content"] = truncateUTF8Safe(text, truncateAt) + TruncationMarker
				data.PromptsTruncated = true
			}

			messages = append(messages, messageMap)
		}

		if len(messages) > 0 {
			jsonBytes, err := json.Marshal(messages)
			if err != nil {
				Warn("Failed to serialize input messages to JSON: %v", err)
			} else {
				// Note: Individual messages are already truncated above.
				// We avoid truncating the final JSON to prevent invalid JSON.
				data.InputMessages = string(jsonBytes)
			}
		}
	}

	return data
}

// extractContentText extracts all text from a Content object
func extractContentText(content *genai.Content) string {
	if content == nil || len(content.Parts) == 0 {
		return ""
	}

	var result string
	for _, part := range content.Parts {
		if part == nil {
			continue
		}
		if part.Text != "" {
			if result != "" {
				result += "\n"
			}
			result += part.Text
		}
	}
	return result
}

// ExtractResponseContent extracts output response from Google AI response
func ExtractResponseContent(resp *genai.GenerateContentResponse, promptsTruncated bool) PromptData {
	data := PromptData{
		PromptsTruncated: promptsTruncated,
	}

	if resp == nil {
		return data
	}

	// Use the Text() method to get the response content
	content := resp.Text()

	if content == "" {
		return data
	}

	// Apply truncation if needed
	if len(content) > MaxPromptLength {
		markerLen := len(TruncationMarker)
		truncateAt := MaxPromptLength - markerLen
		content = truncateUTF8Safe(content, truncateAt) + TruncationMarker
		data.PromptsTruncated = true
		Debug("Output response truncated to %d characters", MaxPromptLength)
	}

	data.OutputResponse = content
	return data
}

// ExtractStreamingResponseContent extracts output from accumulated streaming content
func ExtractStreamingResponseContent(accumulatedContent string, promptsTruncated bool) PromptData {
	data := PromptData{
		PromptsTruncated: promptsTruncated,
	}

	if accumulatedContent == "" {
		return data
	}

	content := accumulatedContent

	// Apply truncation if needed
	if len(content) > MaxPromptLength {
		markerLen := len(TruncationMarker)
		truncateAt := MaxPromptLength - markerLen
		content = truncateUTF8Safe(content, truncateAt) + TruncationMarker
		data.PromptsTruncated = true
		Debug("Streaming output response truncated to %d characters", MaxPromptLength)
	}

	data.OutputResponse = content
	return data
}

// AddPromptDataToPayload adds prompt capture fields to a metering payload
func AddPromptDataToPayload(payload map[string]interface{}, data PromptData) {
	if data.SystemPrompt != "" {
		payload["systemPrompt"] = data.SystemPrompt
	}
	if data.InputMessages != "" {
		payload["inputMessages"] = data.InputMessages
	}
	if data.OutputResponse != "" {
		payload["outputResponse"] = data.OutputResponse
	}
	if data.PromptsTruncated {
		payload["promptsTruncated"] = true
	}
}
