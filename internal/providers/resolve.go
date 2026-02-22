package providers

import (
	"fmt"
	"os"
)

// ResolveByAPIKey checks environment variables in order of precedence and returns
// a fully configured provider ready to make API calls.
// Priority: OPENROUTER_API_KEY > OPENCODE_ZEN_API_KEY > ANTHROPIC_API_KEY > OPENAI_API_KEY
func ResolveByAPIKey() (Provider, error) {
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		return NewOpenRouterProvider(
			"https://openrouter.ai/api/v1/chat/completions",
			"anthropic/claude-3.5-haiku",
			apiKey,
		), nil
	}

	if apiKey := os.Getenv("OPENCODE_ZEN_API_KEY"); apiKey != "" {
		return NewOpencodeZenProvider(
			"https://opencode.ai/zen/v1/messages",
			"claude-3-5-haiku",
			apiKey,
		), nil
	}

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		return NewAnthropicProvider(
			"https://api.anthropic.com/v1/messages",
			"claude-haiku-4-5",
			apiKey,
		), nil
	}

	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		return NewOpenAIProvider(
			"https://api.openai.com/v1/chat/completions",
			"gpt-4o-mini",
			apiKey,
		), nil
	}

	return nil, fmt.Errorf("no API key found. Set OPENROUTER_API_KEY, OPENCODE_ZEN_API_KEY, ANTHROPIC_API_KEY, or OPENAI_API_KEY")
}
