package llm

import (
	"context"
	"fmt"

	"github.com/tmc/go-llm/llms"
	"github.com/tmc/go-llm/llms/ollama"
)

// OllamaProvider wraps the go-llm Ollama implementation
type OllamaProvider struct {
	llm llms.LanguageModel
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(cfg *Config) (*OllamaProvider, error) {
	opts := []ollama.Option{}

	if cfg.BaseURL != "" {
		opts = append(opts, ollama.WithServerURL(cfg.BaseURL))
	}

	if cfg.Model != "" {
		opts = append(opts, ollama.WithModel(cfg.Model))
	}

	llm, err := ollama.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	return &OllamaProvider{llm: llm}, nil
}

// CallChat sends a chat message to Ollama
func (p *OllamaProvider) CallChat(ctx context.Context, messages []Message) (string, error) {
	// Convert our Message type to go-llm format
	llmMessages := make([]llms.MessageContent, len(messages))
	for i, msg := range messages {
		llmMessages[i] = llms.MessageContent{
			Role: msg.Role,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: msg.Content},
			},
		}
	}

	response, err := p.llm.GenerateContent(ctx, llmMessages)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from Ollama")
	}

	return response.Choices[0].Content, nil
}

// GenerateRuleCode generates Go code for a linting rule
func (p *OllamaProvider) GenerateRuleCode(ctx context.Context, ruleName string, styleGuideExcerpt string) (string, error) {
	prompt := fmt.Sprintf(`Generate a Go revive linting rule for the Uber Go Style Guide.

Rule Name: %s

Style Guide Excerpt:
%s

Requirements:
1. Implement the revive Rule interface
2. Include proper error messages
3. Handle nil safely
4. Follow the pattern in existing rules
5. Return lint.Failure for each violation

Generate only the Go code, no markdown.`, ruleName, styleGuideExcerpt)

	messages := []Message{
		{Role: "system", Content: "You are an expert Go developer specializing in code linting and style guides."},
		{Role: "user", Content: prompt},
	}

	return p.CallChat(ctx, messages)
}

// GenerateTests generates test cases for a rule
func (p *OllamaProvider) GenerateTests(ctx context.Context, ruleName string, ruleDescription string) (string, error) {
	prompt := fmt.Sprintf(`Generate comprehensive Go test cases for a revive linting rule.

Rule: %s
Description: %s

Generate:
1. Bad code examples (should trigger the rule)
2. Good code examples (should not trigger the rule)
3. Edge cases

Format as Go test table entries with expected failure counts.
Generate only the Go code, no markdown.`, ruleName, ruleDescription)

	messages := []Message{
		{Role: "system", Content: "You are an expert Go developer specializing in code linting."},
		{Role: "user", Content: prompt},
	}

	return p.CallChat(ctx, messages)
}
