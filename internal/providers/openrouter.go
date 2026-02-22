package providers

import (
	"context"
	"encoding/json"
	"fmt"
)

type openrouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenRouterProvider struct {
	endpoint string
	model    string
	apiKey   string
}

func NewOpenRouterProvider(endpoint, model, apiKey string) Provider {
	return &OpenRouterProvider{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
	}
}

func (o *OpenRouterProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	jsonData, err := json.Marshal(openrouterRequest{
		Model:    o.model,
		Messages: buildMessages(system, userMsg),
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + o.apiKey,
	}

	body, err := doRequest(ctx, o.endpoint, jsonData, headers)
	if err != nil {
		return "", err
	}

	var r openaiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if r.Error != nil {
		return "", fmt.Errorf("%s", r.Error.Message)
	}
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content, nil
	}
	return "", nil
}
