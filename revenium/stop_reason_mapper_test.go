package revenium

import (
	"testing"

	"google.golang.org/genai"
)

func TestMapGoogleFinishReason(t *testing.T) {
	tests := []struct {
		name           string
		finishReason   genai.FinishReason
		defaultReason  ReveniumStopReason
		expectedReason ReveniumStopReason
	}{
		// Natural completion
		{
			name:           "STOP maps to END",
			finishReason:   "STOP",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonEnd,
		},
		// Token limits
		{
			name:           "MAX_TOKENS maps to TOKEN_LIMIT",
			finishReason:   "MAX_TOKENS",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonTokenLimit,
		},
		// Safety and content filtering
		{
			name:           "SAFETY maps to ERROR",
			finishReason:   "SAFETY",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonError,
		},
		{
			name:           "RECITATION maps to ERROR",
			finishReason:   "RECITATION",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonError,
		},
		{
			name:           "BLOCKLIST maps to ERROR",
			finishReason:   "BLOCKLIST",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonError,
		},
		{
			name:           "PROHIBITED_CONTENT maps to ERROR",
			finishReason:   "PROHIBITED_CONTENT",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonError,
		},
		{
			name:           "SPII maps to ERROR",
			finishReason:   "SPII",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonError,
		},
		{
			name:           "MODEL_ARMOR maps to ERROR",
			finishReason:   "MODEL_ARMOR",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonError,
		},
		// Function call errors
		{
			name:           "MALFORMED_FUNCTION_CALL maps to ERROR",
			finishReason:   "MALFORMED_FUNCTION_CALL",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonError,
		},
		{
			name:           "UNEXPECTED_TOOL_CALL maps to ERROR",
			finishReason:   "UNEXPECTED_TOOL_CALL",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonError,
		},
		// Cancellation
		{
			name:           "CANCELLED maps to CANCELLED",
			finishReason:   "CANCELLED",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonCancelled,
		},
		{
			name:           "CANCELED maps to CANCELLED (American spelling)",
			finishReason:   "CANCELED",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonCancelled,
		},
		// Unspecified or other
		{
			name:           "FINISH_REASON_UNSPECIFIED uses default",
			finishReason:   "FINISH_REASON_UNSPECIFIED",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonEnd,
		},
		{
			name:           "OTHER uses default",
			finishReason:   "OTHER",
			defaultReason:  StopReasonTokenLimit,
			expectedReason: StopReasonTokenLimit,
		},
		// Empty finish reason
		{
			name:           "Empty finish reason uses default",
			finishReason:   "",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonEnd,
		},
		// Unknown finish reason
		{
			name:           "Unknown finish reason uses default",
			finishReason:   "UNKNOWN_FUTURE_REASON",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonEnd,
		},
		// Case insensitivity
		{
			name:           "Lowercase stop maps to END",
			finishReason:   "stop",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonEnd,
		},
		{
			name:           "Mixed case max_tokens maps to TOKEN_LIMIT",
			finishReason:   "Max_Tokens",
			defaultReason:  StopReasonEnd,
			expectedReason: StopReasonTokenLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapGoogleFinishReason(tt.finishReason, tt.defaultReason)
			if result != tt.expectedReason {
				t.Errorf("MapGoogleFinishReason(%q, %q) = %q, want %q",
					tt.finishReason, tt.defaultReason, result, tt.expectedReason)
			}
		})
	}
}

func TestExtractFinishReason(t *testing.T) {
	tests := []struct {
		name         string
		response     *genai.GenerateContentResponse
		expectedReason genai.FinishReason
	}{
		{
			name:         "Nil response returns empty",
			response:     nil,
			expectedReason: "",
		},
		{
			name: "Response with candidates returns finish reason",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{FinishReason: "STOP"},
				},
			},
			expectedReason: "STOP",
		},
		{
			name: "Response with empty candidates returns empty",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{},
			},
			expectedReason: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFinishReason(tt.response)
			if result != tt.expectedReason {
				t.Errorf("ExtractFinishReason() = %q, want %q", result, tt.expectedReason)
			}
		})
	}
}

