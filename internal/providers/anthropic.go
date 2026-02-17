package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Anthropic Messages API types
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// AnthropicProvider implements Provider for Anthropic Messages API.
type AnthropicProvider struct {
	endpoint string
	model    string
	apiKey   string
}

// NewAnthropicProvider creates a new Anthropic provider with the given configuration.
func NewAnthropicProvider(endpoint, model, apiKey string) Provider {
	return &AnthropicProvider{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
	}
}

func (a *AnthropicProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	jsonData, _ := json.Marshal(anthropicRequest{
		Model:     a.model,
		MaxTokens: 4096,
		System:    system,
		Messages:  []Message{{Role: "user", Content: userMsg}},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", a.endpoint, bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var r anthropicResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}
	if r.Error != nil {
		return "", fmt.Errorf("%s", r.Error.Message)
	}

	if len(r.Content) > 0 {
		return r.Content[0].Text, nil
	}
	return "", nil
}
