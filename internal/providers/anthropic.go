package providers

import (
	"context"
	"encoding/json"
	"fmt"
)

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

type AnthropicProvider struct {
	endpoint string
	model    string
	apiKey   string
}

func NewAnthropicProvider(endpoint, model, apiKey string) Provider {
	return &AnthropicProvider{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
	}
}

func (a *AnthropicProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	jsonData, err := json.Marshal(anthropicRequest{
		Model:     a.model,
		MaxTokens: 4096,
		System:    system,
		Messages:  []Message{{Role: "user", Content: userMsg}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := map[string]string{
		"x-api-key":         a.apiKey,
		"anthropic-version": "2023-06-01",
	}

	body, err := doRequest(ctx, a.endpoint, jsonData, headers)
	if err != nil {
		return "", err
	}

	var r anthropicResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if r.Error != nil {
		return "", fmt.Errorf("%s", r.Error.Message)
	}
	if len(r.Content) > 0 {
		return r.Content[0].Text, nil
	}
	return "", nil
}
