package llm

import (
	"context"
	"fmt"

	"github.com/tmc/go-llm/llms"
	"github.com/tmc/go-llm/llms/anthropic"
)

// AnthropicProvider wraps the go-llm Anthropic implementation
type AnthropicProvider struct {
	llm llms.LanguageModel
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(cfg *Config) (*AnthropicProvider, error) {
	opts := []anthropic.Option{}

	if cfg.APIKey != "" {
		opts = append(opts, anthropic.WithAPIKey(cfg.APIKey))
	}

	if cfg.Model != "" {
		opts = append(opts, anthropic.WithModel(cfg.Model))
	}

	llm, err := anthropic.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic client: %w", err)
	}

	return &AnthropicProvider{llm: llm}, nil
}

// CallChat sends a chat message to Anthropic
func (p *AnthropicProvider) CallChat(ctx context.Context, messages []Message) (string, error) {
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
		return "", fmt.Errorf("no choices returned from Anthropic")
	}

	return response.Choices[0].Content, nil
}

// GenerateRuleCode generates Go code for a linting rule
func (p *AnthropicProvider) GenerateRuleCode(ctx context.Context, ruleName string, styleGuideExcerpt string) (string, error) {
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
func (p *AnthropicProvider) GenerateTests(ctx context.Context, ruleName string, ruleDescription string) (string, error) {
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
