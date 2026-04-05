# LLM Integration Guide

This project integrates with multiple LLM providers via the `go-llm` library for enhanced rule generation and test case creation.

## Supported Providers

- **OpenAI** (Default) - GPT-3.5-turbo, GPT-4
- **Ollama** - Local LLM inference
- **Anthropic** - Claude models

## Configuration

Set environment variables to configure the LLM provider:

```bash
# Enable LLM features
export LLM_ENABLED=true

# Choose provider: openai, ollama, or anthropic
export LLM_PROVIDER=openai

# Provider-specific settings
export LLM_API_KEY=your-api-key          # For OpenAI/Anthropic
export LLM_BASE_URL=http://localhost:11434  # For Ollama
export LLM_MODEL=gpt-3.5-turbo           # Model to use
```

## Usage Examples

### With Ollama (Local)

```bash
# Start Ollama
ollama serve

# In another terminal, pull a model
ollama pull llama2

# Enable LLM integration
export LLM_ENABLED=true
export LLM_PROVIDER=ollama
export LLM_BASE_URL=http://localhost:11434
export LLM_MODEL=llama2

# Run generate tool (will enhance rules with LLM)
go run ./internal/generate
```

### With OpenAI

```bash
export LLM_ENABLED=true
export LLM_PROVIDER=openai
export LLM_API_KEY=sk-...
export LLM_MODEL=gpt-3.5-turbo

go run ./internal/generate
```

### With Anthropic

```bash
export LLM_ENABLED=true
export LLM_PROVIDER=anthropic
export LLM_API_KEY=sk-ant-...
export LLM_MODEL=claude-3-sonnet-20240229

go run ./internal/generate
```

## Features

### Rule Code Generation
When enabled, the generator will use the LLM to enhance rule implementations with actual working code patterns instead of just templates.

```
Created interface-pointer (Enhanced with LLM)
Created mutex-zero-value (Enhanced with LLM)
```

### Test Case Generation
The `GenerateTests` method generates comprehensive test cases for rules:

```go
llmProvider.GenerateTests(ctx, "interface-pointer", "Discourage pointer receivers for interface definitions")
```

## Implementation Details

- `internal/llm/provider.go` - Core provider interface and config
- `internal/llm/openai.go` - OpenAI implementation
- `internal/llm/ollama.go` - Ollama implementation
- `internal/llm/anthropic.go` - Anthropic implementation

All providers implement the same interface, making it easy to swap between them.

## Architecture

```go
type Provider interface {
    CallChat(ctx context.Context, messages []Message) (string, error)
    GenerateRuleCode(ctx context.Context, ruleName, styleGuideExcerpt string) (string, error)
    GenerateTests(ctx context.Context, ruleName, ruleDescription string) (string, error)
}
```

This abstraction allows different tools to use different providers without coupling to specific implementations.
