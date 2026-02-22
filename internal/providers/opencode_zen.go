package providers

import (
	"context"
	"encoding/json"
	"fmt"
)

type opencodeZenRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
}

type opencodeZenResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type OpencodeZenProvider struct {
	endpoint string
	model    string
	apiKey   string
}

func NewOpencodeZenProvider(endpoint, model, apiKey string) Provider {
	return &OpencodeZenProvider{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
	}
}

func (o *OpencodeZenProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	jsonData, err := json.Marshal(opencodeZenRequest{
		Model:     o.model,
		MaxTokens: 4096,
		System:    system,
		Messages:  []Message{{Role: "user", Content: userMsg}},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := map[string]string{
		"x-api-key":         o.apiKey,
		"anthropic-version": "2023-06-01",
	}

	body, err := doRequest(ctx, o.endpoint, jsonData, headers)
	if err != nil {
		return "", err
	}

	var r opencodeZenResponse
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
