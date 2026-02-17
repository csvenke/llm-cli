package providers

import (
	"fmt"
	"os"
)

// ResolveByAPIKey checks environment variables in order of precedence and returns
// a fully configured provider ready to make API calls.
// Priority: OPENCODE_ZEN_API_KEY > ANTHROPIC_API_KEY > OPENAI_API_KEY
func ResolveByAPIKey() (Provider, error) {
	// Priority 1: OpenCode Zen (uses Anthropic Messages format)
	if apiKey := os.Getenv("OPENCODE_ZEN_API_KEY"); apiKey != "" {
		return NewAnthropicProvider(
			"https://opencode.ai/zen/v1/messages",
			"claude-3-5-haiku",
			apiKey,
		), nil
	}

	// Priority 2: Anthropic Direct
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		return NewAnthropicProvider(
			"https://api.anthropic.com/v1/messages",
			"claude-3-5-haiku",
			apiKey,
		), nil
	}

	// Priority 3: OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		return NewOpenAIProvider(
			"https://api.openai.com/v1/chat/completions",
			"gpt-4o-mini",
			apiKey,
		), nil
	}

	return nil, fmt.Errorf("no API key found. Set OPENCODE_ZEN_API_KEY, ANTHROPIC_API_KEY, or OPENAI_API_KEY")
}
