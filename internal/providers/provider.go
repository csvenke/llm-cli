package providers

import "context"

// Provider defines the strategy interface for LLM chat completions.
// Each provider implementation encapsulates its own configuration
// (endpoint, model, API key) and handles the complete request lifecycle.
type Provider interface {
	Complete(ctx context.Context, system, userMsg string) (string, error)
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
