package providers

import (
	"context"
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
	var r anthropicResponse
	if err := doJSONRequest(ctx, a.endpoint, anthropicRequest{
		Model:     a.model,
		MaxTokens: 4096,
		System:    system,
		Messages:  []Message{{Role: "user", Content: userMsg}},
	}, &r, map[string]string{
		"x-api-key":         a.apiKey,
		"anthropic-version": "2023-06-01",
	}); err != nil {
		return "", err
	}

	if r.Error != nil {
		return "", fmt.Errorf("%s", r.Error.Message)
	}
	if len(r.Content) > 0 {
		return r.Content[0].Text, nil
	}
	return "", nil
}
