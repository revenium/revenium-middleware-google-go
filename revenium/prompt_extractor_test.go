package revenium

import (
	"strings"
	"testing"

	"google.golang.org/genai"
)

func TestExtractPromptsFromRequest(t *testing.T) {
	tests := []struct {
		name          string
		contents      []*genai.Content
		config        *genai.GenerateContentConfig
		wantSystem    bool
		wantInput     bool
		wantTruncated bool
	}{
		{
			name:          "empty contents",
			contents:      []*genai.Content{},
			config:        nil,
			wantSystem:    false,
			wantInput:     false,
			wantTruncated: false,
		},
		{
			name:          "nil contents",
			contents:      nil,
			config:        nil,
			wantSystem:    false,
			wantInput:     false,
			wantTruncated: false,
		},
		{
			name: "with system instruction and user content",
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: "Hello, how are you?"},
					},
				},
			},
			config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: "You are a helpful assistant."},
					},
				},
			},
			wantSystem:    true,
			wantInput:     true,
			wantTruncated: false,
		},
		{
			name: "user only content",
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: "What is the weather?"},
					},
				},
			},
			config:        nil,
			wantSystem:    false,
			wantInput:     true,
			wantTruncated: false,
		},
		{
			name: "multiple system instruction parts",
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: "Test"},
					},
				},
			},
			config: &genai.GenerateContentConfig{
				SystemInstruction: &genai.Content{
					Parts: []*genai.Part{
						{Text: "System block 1."},
						{Text: "System block 2."},
					},
				},
			},
			wantSystem:    true,
			wantInput:     true,
			wantTruncated: false,
		},
		{
			name: "content with no role defaults to user",
			contents: []*genai.Content{
				{
					Parts: []*genai.Part{
						{Text: "No role specified"},
					},
				},
			},
			config:        nil,
			wantSystem:    false,
			wantInput:     true,
			wantTruncated: false,
		},
		{
			name: "multi-turn conversation",
			contents: []*genai.Content{
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: "Hello"},
					},
				},
				{
					Role: "model",
					Parts: []*genai.Part{
						{Text: "Hi there! How can I help you?"},
					},
				},
				{
					Role: "user",
					Parts: []*genai.Part{
						{Text: "What time is it?"},
					},
				},
			},
			config:        nil,
			wantSystem:    false,
			wantInput:     true,
			wantTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPromptsFromRequest(tt.contents, tt.config)

			hasSystem := result.SystemPrompt != ""
			hasInput := result.InputMessages != ""

			if hasSystem != tt.wantSystem {
				t.Errorf("SystemPrompt: got %v, want %v (value: %q)", hasSystem, tt.wantSystem, result.SystemPrompt)
			}
			if hasInput != tt.wantInput {
				t.Errorf("InputMessages: got %v, want %v (value: %q)", hasInput, tt.wantInput, result.InputMessages)
			}
			if result.PromptsTruncated != tt.wantTruncated {
				t.Errorf("PromptsTruncated: got %v, want %v", result.PromptsTruncated, tt.wantTruncated)
			}
		})
	}
}

func TestExtractResponseContent(t *testing.T) {
	tests := []struct {
		name          string
		resp          *genai.GenerateContentResponse
		truncated     bool
		wantOutput    bool
		wantTruncated bool
	}{
		{
			name:          "nil response",
			resp:          nil,
			truncated:     false,
			wantOutput:    false,
			wantTruncated: false,
		},
		{
			name: "empty candidates",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{},
			},
			truncated:     false,
			wantOutput:    false,
			wantTruncated: false,
		},
		{
			name: "with text content",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: "Hello! I'm doing great."},
							},
						},
					},
				},
			},
			truncated:     false,
			wantOutput:    true,
			wantTruncated: false,
		},
		{
			name: "preserves input truncation flag",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: "Short response"},
							},
						},
					},
				},
			},
			truncated:     true,
			wantOutput:    true,
			wantTruncated: true,
		},
		{
			name: "multiple text parts",
			resp: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: "First part."},
								{Text: "Second part."},
							},
						},
					},
				},
			},
			truncated:     false,
			wantOutput:    true,
			wantTruncated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractResponseContent(tt.resp, tt.truncated)

			hasOutput := result.OutputResponse != ""
			if hasOutput != tt.wantOutput {
				t.Errorf("OutputResponse: got %v, want %v (value: %q)", hasOutput, tt.wantOutput, result.OutputResponse)
			}
			if result.PromptsTruncated != tt.wantTruncated {
				t.Errorf("PromptsTruncated: got %v, want %v", result.PromptsTruncated, tt.wantTruncated)
			}
		})
	}
}

func TestExtractStreamingResponseContent(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		truncated     bool
		wantOutput    bool
		wantTruncated bool
	}{
		{
			name:          "empty content",
			content:       "",
			truncated:     false,
			wantOutput:    false,
			wantTruncated: false,
		},
		{
			name:          "normal content",
			content:       "This is a streaming response.",
			truncated:     false,
			wantOutput:    true,
			wantTruncated: false,
		},
		{
			name:          "preserves input truncation",
			content:       "Some content",
			truncated:     true,
			wantOutput:    true,
			wantTruncated: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractStreamingResponseContent(tt.content, tt.truncated)

			hasOutput := result.OutputResponse != ""
			if hasOutput != tt.wantOutput {
				t.Errorf("OutputResponse: got %v, want %v", hasOutput, tt.wantOutput)
			}
			if result.PromptsTruncated != tt.wantTruncated {
				t.Errorf("PromptsTruncated: got %v, want %v", result.PromptsTruncated, tt.wantTruncated)
			}
		})
	}
}

func TestTruncation(t *testing.T) {
	// Create content longer than MaxPromptLength
	longContent := strings.Repeat("a", MaxPromptLength+1000)

	result := ExtractStreamingResponseContent(longContent, false)

	if !result.PromptsTruncated {
		t.Error("Expected PromptsTruncated to be true for long content")
	}

	if len(result.OutputResponse) != MaxPromptLength {
		t.Errorf("Expected truncated length %d, got %d", MaxPromptLength, len(result.OutputResponse))
	}

	if !strings.HasSuffix(result.OutputResponse, TruncationMarker) {
		t.Error("Expected truncated content to end with truncation marker")
	}
}

func TestAddPromptDataToPayload(t *testing.T) {
	payload := make(map[string]interface{})

	data := PromptData{
		SystemPrompt:     "You are helpful",
		InputMessages:    `[{"role":"user","content":"Hi"}]`,
		OutputResponse:   "Hello there!",
		PromptsTruncated: true,
	}

	AddPromptDataToPayload(payload, data)

	if payload["systemPrompt"] != "You are helpful" {
		t.Error("systemPrompt not added to payload")
	}
	if payload["inputMessages"] != `[{"role":"user","content":"Hi"}]` {
		t.Error("inputMessages not added to payload")
	}
	if payload["outputResponse"] != "Hello there!" {
		t.Error("outputResponse not added to payload")
	}
	if payload["promptsTruncated"] != true {
		t.Error("promptsTruncated not added to payload")
	}
}

func TestAddPromptDataToPayload_EmptyFields(t *testing.T) {
	payload := make(map[string]interface{})

	data := PromptData{
		SystemPrompt:     "",
		InputMessages:    "",
		OutputResponse:   "",
		PromptsTruncated: false,
	}

	AddPromptDataToPayload(payload, data)

	// Empty fields should not be added
	if _, ok := payload["systemPrompt"]; ok {
		t.Error("empty systemPrompt should not be added to payload")
	}
	if _, ok := payload["inputMessages"]; ok {
		t.Error("empty inputMessages should not be added to payload")
	}
	if _, ok := payload["outputResponse"]; ok {
		t.Error("empty outputResponse should not be added to payload")
	}
	if _, ok := payload["promptsTruncated"]; ok {
		t.Error("false promptsTruncated should not be added to payload")
	}
}

func TestExtractContentText(t *testing.T) {
	tests := []struct {
		name    string
		content *genai.Content
		want    string
	}{
		{
			name:    "nil content",
			content: nil,
			want:    "",
		},
		{
			name: "empty parts",
			content: &genai.Content{
				Parts: []*genai.Part{},
			},
			want: "",
		},
		{
			name: "single text part",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "Hello world"},
				},
			},
			want: "Hello world",
		},
		{
			name: "multiple text parts",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "First part."},
					{Text: "Second part."},
				},
			},
			want: "First part.\nSecond part.",
		},
		{
			name: "nil part in array",
			content: &genai.Content{
				Parts: []*genai.Part{
					{Text: "Before"},
					nil,
					{Text: "After"},
				},
			},
			want: "Before\nAfter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractContentText(tt.content)
			if got != tt.want {
				t.Errorf("extractContentText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSystemPromptTruncation(t *testing.T) {
	// Create a system prompt longer than MaxPromptLength
	longSystemPrompt := strings.Repeat("s", MaxPromptLength+500)

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: longSystemPrompt},
			},
		},
	}

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: "Hi"},
			},
		},
	}

	result := ExtractPromptsFromRequest(contents, config)

	if !result.PromptsTruncated {
		t.Error("Expected PromptsTruncated to be true for long system prompt")
	}

	if len(result.SystemPrompt) != MaxPromptLength {
		t.Errorf("Expected truncated system prompt length %d, got %d", MaxPromptLength, len(result.SystemPrompt))
	}

	if !strings.HasSuffix(result.SystemPrompt, TruncationMarker) {
		t.Error("Expected truncated system prompt to end with truncation marker")
	}
}

func TestInputMessagesTruncation(t *testing.T) {
	// Create a message with content longer than halfLimit
	halfLimit := MaxPromptLength / 2
	longContent := strings.Repeat("m", halfLimit+500)

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: longContent},
			},
		},
	}

	result := ExtractPromptsFromRequest(contents, nil)

	if !result.PromptsTruncated {
		t.Error("Expected PromptsTruncated to be true for long message content")
	}

	// The individual message content should be truncated
	if !strings.Contains(result.InputMessages, TruncationMarker) {
		t.Error("Expected truncated input messages to contain truncation marker")
	}
}
