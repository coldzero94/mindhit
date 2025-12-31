package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateJSONResponse_ValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		jsonMode bool
		wantErr  bool
	}{
		{
			name:     "valid JSON object",
			content:  `{"key": "value"}`,
			jsonMode: true,
			wantErr:  false,
		},
		{
			name:     "valid JSON array",
			content:  `[1, 2, 3]`,
			jsonMode: true,
			wantErr:  false,
		},
		{
			name:     "valid nested JSON",
			content:  `{"topics": [{"id": "1", "label": "test"}]}`,
			jsonMode: true,
			wantErr:  false,
		},
		{
			name:     "invalid JSON with jsonMode true",
			content:  `not valid json`,
			jsonMode: true,
			wantErr:  true,
		},
		{
			name:     "invalid JSON with jsonMode false - should pass",
			content:  `not valid json`,
			jsonMode: false,
			wantErr:  false,
		},
		{
			name:     "empty string with jsonMode false",
			content:  ``,
			jsonMode: false,
			wantErr:  false,
		},
		{
			name:     "empty string with jsonMode true",
			content:  ``,
			jsonMode: true,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJSONResponse(tt.content, tt.jsonMode)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidJSON)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildMessages(t *testing.T) {
	tests := []struct {
		name     string
		req      ChatRequest
		expected []Message
	}{
		{
			name: "system prompt only",
			req: ChatRequest{
				SystemPrompt: "You are a helpful assistant",
			},
			expected: []Message{
				{Role: RoleSystem, Content: "You are a helpful assistant"},
			},
		},
		{
			name: "user prompt only",
			req: ChatRequest{
				UserPrompt: "Hello",
			},
			expected: []Message{
				{Role: RoleUser, Content: "Hello"},
			},
		},
		{
			name: "both system and user prompt",
			req: ChatRequest{
				SystemPrompt: "You are a helpful assistant",
				UserPrompt:   "Hello",
			},
			expected: []Message{
				{Role: RoleSystem, Content: "You are a helpful assistant"},
				{Role: RoleUser, Content: "Hello"},
			},
		},
		{
			name: "with existing messages",
			req: ChatRequest{
				SystemPrompt: "System",
				Messages: []Message{
					{Role: RoleUser, Content: "First"},
					{Role: RoleAssistant, Content: "Response"},
				},
				UserPrompt: "Second",
			},
			expected: []Message{
				{Role: RoleSystem, Content: "System"},
				{Role: RoleUser, Content: "First"},
				{Role: RoleAssistant, Content: "Response"},
				{Role: RoleUser, Content: "Second"},
			},
		},
		{
			name:     "empty request",
			req:      ChatRequest{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildMessages(tt.req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseProvider_Type(t *testing.T) {
	bp := BaseProvider{providerType: ProviderGemini}
	assert.Equal(t, ProviderGemini, bp.Type())
}

func TestBaseProvider_Model(t *testing.T) {
	bp := BaseProvider{model: "test-model"}
	assert.Equal(t, "test-model", bp.Model())
}

func TestBaseProvider_Close(t *testing.T) {
	bp := BaseProvider{}
	err := bp.Close()
	assert.NoError(t, err)
}

func TestDefaultChatOptions(t *testing.T) {
	opts := DefaultChatOptions()
	assert.Equal(t, 0.7, opts.Temperature)
	assert.Equal(t, 4096, opts.MaxTokens)
	assert.Equal(t, 1.0, opts.TopP)
}

func TestProviderType_Constants(t *testing.T) {
	assert.Equal(t, ProviderType("openai"), ProviderOpenAI)
	assert.Equal(t, ProviderType("gemini"), ProviderGemini)
	assert.Equal(t, ProviderType("claude"), ProviderClaude)
}

func TestDefaultModels(t *testing.T) {
	assert.Equal(t, "gpt-4o", DefaultOpenAIModel)
	assert.Equal(t, "gemini-2.0-flash", DefaultGeminiModel)
	assert.Equal(t, "claude-sonnet-4-20250514", DefaultClaudeModel)
}

func TestTaskType_Constants(t *testing.T) {
	assert.Equal(t, TaskType("tag_extraction"), TaskTagExtraction)
	assert.Equal(t, TaskType("mindmap"), TaskMindmap)
	assert.Equal(t, TaskType("general"), TaskGeneral)
}

func TestRole_Constants(t *testing.T) {
	assert.Equal(t, Role("system"), RoleSystem)
	assert.Equal(t, Role("user"), RoleUser)
	assert.Equal(t, Role("assistant"), RoleAssistant)
}
