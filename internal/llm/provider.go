package llm

import (
	"context"
	"fmt"
	"os"
)

// Message represents a chat message
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// Provider is the interface for LLM backends
type Provider interface {
	CallChat(ctx context.Context, messages []Message) (string, error)
	GenerateRuleCode(ctx context.Context, ruleName string, styleGuideExcerpt string) (string, error)
	GenerateTests(ctx context.Context, ruleName string, ruleDescription string) (string, error)
}

// Config holds LLM provider configuration
type Config struct {
	Provider string // "openai", "ollama", "anthropic", "cohere", "huggingface"
	APIKey   string
	BaseURL  string // For Ollama: http://localhost:11434
	Model    string
	Enabled  bool
}

// NewConfig loads configuration from environment
func NewConfig() *Config {
	return &Config{
		Provider: getEnv("LLM_PROVIDER", "openai"),
		APIKey:   getEnv("LLM_API_KEY", ""),
		BaseURL:  getEnv("LLM_BASE_URL", ""),
		Model:    getEnv("LLM_MODEL", "gpt-3.5-turbo"),
		Enabled:  getEnv("LLM_ENABLED", "false") == "true",
	}
}

// NewProvider creates a provider based on configuration
func NewProvider(cfg *Config) (Provider, error) {
	if !cfg.Enabled {
		return &NoOpProvider{}, nil
	}

	switch cfg.Provider {
	case "openai":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("OpenAI provider requires LLM_API_KEY environment variable")
		}
		return NewOpenAIProvider(cfg)
	case "ollama":
		if cfg.BaseURL == "" {
			return nil, fmt.Errorf("Ollama provider requires LLM_BASE_URL environment variable (e.g., http://localhost:11434)")
		}
		return NewOllamaProvider(cfg)
	case "anthropic":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("Anthropic provider requires LLM_API_KEY environment variable")
		}
		return NewAnthropicProvider(cfg)
	default:
		return &NoOpProvider{}, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

// NoOpProvider is a no-op implementation for when LLM is disabled
type NoOpProvider struct{}

func (n *NoOpProvider) CallChat(ctx context.Context, messages []Message) (string, error) {
	return "", nil
}

func (n *NoOpProvider) GenerateRuleCode(ctx context.Context, ruleName string, styleGuideExcerpt string) (string, error) {
	return "", nil
}

func (n *NoOpProvider) GenerateTests(ctx context.Context, ruleName string, ruleDescription string) (string, error) {
	return "", nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
