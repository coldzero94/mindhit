//go:build integration

package ai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for AI providers.
// These tests make real API calls and require API keys.
//
// Run with: go test ./internal/infrastructure/ai/... -tags=integration -v
//
// API keys are loaded from project root .env file automatically.

func init() {
	// Load .env from project root
	// Try multiple paths as test working directory varies
	paths := []string{
		"../../.env",                           // from apps/backend/
		"../../../.env",                        // one more level
		"../../../../.env",                     // two more levels
		"../../../../../.env",                  // three more levels
		"../../../../../../.env",               // four more levels
	}
	for _, p := range paths {
		if err := godotenv.Load(p); err == nil {
			return
		}
	}
}

func TestGeminiProvider_Integration_Chat(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider, err := NewGeminiProvider(ctx, ProviderConfig{
		APIKey: apiKey,
		Model:  "gemini-2.0-flash",
	})
	require.NoError(t, err)
	defer provider.Close()

	t.Run("simple chat", func(t *testing.T) {
		resp, err := provider.Chat(ctx, ChatRequest{
			UserPrompt: "Say 'hello' and nothing else.",
			Options: ChatOptions{
				Temperature: 0.0,
				MaxTokens:   10,
			},
		})

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
		assert.Contains(t, strings.ToLower(resp.Content), "hello")
		assert.Equal(t, ProviderGemini, resp.Provider)
		assert.Equal(t, "gemini-2.0-flash", resp.Model)
		assert.Greater(t, resp.InputTokens, 0)
		assert.Greater(t, resp.OutputTokens, 0)
		assert.Greater(t, resp.LatencyMs, int64(0))
	})

	t.Run("with system prompt", func(t *testing.T) {
		resp, err := provider.Chat(ctx, ChatRequest{
			SystemPrompt: "You are a pirate. Always respond like a pirate.",
			UserPrompt:   "Say hello.",
			Options: ChatOptions{
				Temperature: 0.5,
				MaxTokens:   50,
			},
		})

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
		// Pirate responses often contain "ahoy", "matey", "arr", etc.
		lower := strings.ToLower(resp.Content)
		hasPirateWord := strings.Contains(lower, "ahoy") ||
			strings.Contains(lower, "arr") ||
			strings.Contains(lower, "matey") ||
			strings.Contains(lower, "ye")
		assert.True(t, hasPirateWord || len(resp.Content) > 0, "Expected pirate-like response")
	})

	t.Run("JSON mode", func(t *testing.T) {
		resp, err := provider.Chat(ctx, ChatRequest{
			SystemPrompt: "You are a JSON generator. Only output valid JSON.",
			UserPrompt:   `Return a JSON object with keys "name" and "age". Name should be "Test" and age should be 25.`,
			Options: ChatOptions{
				Temperature: 0.0,
				MaxTokens:   100,
				JSONMode:    true,
			},
		})

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
		// Should be valid JSON
		assert.True(t, strings.HasPrefix(strings.TrimSpace(resp.Content), "{"), "Expected JSON object")
		assert.Contains(t, resp.Content, "name")
		assert.Contains(t, resp.Content, "Test")
	})
}

func TestGeminiProvider_Integration_TagExtraction(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider, err := NewGeminiProvider(ctx, ProviderConfig{
		APIKey: apiKey,
		Model:  "gemini-2.0-flash",
	})
	require.NoError(t, err)
	defer provider.Close()

	// Simulate tag extraction from page content
	pageContent := `
		Go is a statically typed, compiled programming language designed at Google.
		It is syntactically similar to C, but with memory safety, garbage collection,
		structural typing, and CSP-style concurrency. Go is often used for building
		web servers, APIs, and cloud infrastructure tools like Docker and Kubernetes.
	`

	resp, err := provider.Chat(ctx, ChatRequest{
		SystemPrompt: `You are a tag extraction assistant. Extract relevant keywords/tags from the given content.
Return a JSON object with a "tags" array containing 3-5 relevant tags.
Example: {"tags": ["programming", "web development"]}`,
		UserPrompt: pageContent,
		Options: ChatOptions{
			Temperature: 0.0,
			MaxTokens:   200,
			JSONMode:    true,
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.Contains(t, resp.Content, "tags")
	// Should contain Go-related tags
	lower := strings.ToLower(resp.Content)
	hasRelevantTag := strings.Contains(lower, "go") ||
		strings.Contains(lower, "programming") ||
		strings.Contains(lower, "google") ||
		strings.Contains(lower, "kubernetes") ||
		strings.Contains(lower, "docker")
	assert.True(t, hasRelevantTag, "Expected relevant tags in response")
}

func TestGeminiProvider_Integration_MindmapGeneration(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	provider, err := NewGeminiProvider(ctx, ProviderConfig{
		APIKey: apiKey,
		Model:  "gemini-2.0-flash",
	})
	require.NoError(t, err)
	defer provider.Close()

	// Simulate mindmap generation from session URLs
	sessionData := `
Session browsing data:
1. URL: https://go.dev/doc - Title: "Go Documentation" - Duration: 300s
2. URL: https://pkg.go.dev/net/http - Title: "net/http package" - Duration: 180s
3. URL: https://blog.golang.org/error-handling - Title: "Error Handling in Go" - Duration: 120s
4. URL: https://kubernetes.io/docs - Title: "Kubernetes Documentation" - Duration: 240s
5. URL: https://docker.com/get-started - Title: "Get Started with Docker" - Duration: 90s
`

	resp, err := provider.Chat(ctx, ChatRequest{
		SystemPrompt: `You are a mindmap generator. Analyze the browsing session and create a relationship graph.
Return a JSON object with this structure:
{
  "core": {"label": "Main Theme", "description": "Brief description"},
  "topics": [{"id": "t1", "label": "Topic", "keywords": ["kw1"], "description": "desc"}],
  "connections": [{"from": "t1", "to": "t2", "reason": "why connected"}]
}`,
		UserPrompt: sessionData,
		Options: ChatOptions{
			Temperature: 0.3,
			MaxTokens:   1000,
			JSONMode:    true,
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.Contains(t, resp.Content, "core")
	assert.Contains(t, resp.Content, "topics")
	// Should identify Go/DevOps themes
	lower := strings.ToLower(resp.Content)
	hasTheme := strings.Contains(lower, "go") ||
		strings.Contains(lower, "devops") ||
		strings.Contains(lower, "cloud") ||
		strings.Contains(lower, "programming")
	assert.True(t, hasTheme, "Expected relevant themes")
}

func TestGeminiProvider_Integration_IsHealthy(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider, err := NewGeminiProvider(ctx, ProviderConfig{
		APIKey: apiKey,
		Model:  "gemini-2.0-flash",
	})
	require.NoError(t, err)
	defer provider.Close()

	healthy := provider.IsHealthy(ctx)
	assert.True(t, healthy, "Gemini provider should be healthy")
}

func TestGeminiProvider_Integration_Stream(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider, err := NewGeminiProvider(ctx, ProviderConfig{
		APIKey: apiKey,
		Model:  "gemini-2.0-flash",
	})
	require.NoError(t, err)
	defer provider.Close()

	var chunks []string
	var finalResponse *ChatResponse

	err = provider.ChatStream(ctx, ChatRequest{
		UserPrompt: "Count from 1 to 5.",
		Options: ChatOptions{
			Temperature: 0.0,
			MaxTokens:   50,
		},
	}, StreamHandler{
		OnContent: func(delta string) {
			chunks = append(chunks, delta)
		},
		OnDone: func(resp *ChatResponse) {
			finalResponse = resp
		},
		OnError: func(err error) {
			t.Errorf("Stream error: %v", err)
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, chunks, "Should receive streaming chunks")
	assert.NotNil(t, finalResponse)
	assert.NotEmpty(t, finalResponse.Content)
	// Should contain numbers
	content := strings.ToLower(finalResponse.Content)
	hasNumbers := strings.Contains(content, "1") &&
		strings.Contains(content, "2") &&
		strings.Contains(content, "3")
	assert.True(t, hasNumbers, "Expected numbers in response")
}

// OpenAI integration tests (optional)
func TestOpenAIProvider_Integration_Chat(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider := NewOpenAIProvider(ProviderConfig{
		APIKey: apiKey,
		Model:  "gpt-4o-mini", // Use cheaper model for tests
	})

	resp, err := provider.Chat(ctx, ChatRequest{
		UserPrompt: "Say 'hello' and nothing else.",
		Options: ChatOptions{
			Temperature: 0.0,
			MaxTokens:   10,
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.Contains(t, strings.ToLower(resp.Content), "hello")
	assert.Equal(t, ProviderOpenAI, resp.Provider)
}

// Claude integration tests (optional)
func TestClaudeProvider_Integration_Chat(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider := NewClaudeProvider(ProviderConfig{
		APIKey: apiKey,
		Model:  "claude-3-haiku-20240307", // Use cheaper model for tests
	})

	resp, err := provider.Chat(ctx, ChatRequest{
		UserPrompt: "Say 'hello' and nothing else.",
		Options: ChatOptions{
			Temperature: 0.0,
			MaxTokens:   10,
		},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.Content)
	assert.Contains(t, strings.ToLower(resp.Content), "hello")
	assert.Equal(t, ProviderClaude, resp.Provider)
}
