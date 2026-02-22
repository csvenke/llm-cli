package providers

import (
	"context"
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
	var r openaiResponse
	if err := doJSONRequest(ctx, o.endpoint, openrouterRequest{
		Model:    o.model,
		Messages: buildMessages(system, userMsg),
	}, &r, map[string]string{
		"Authorization": "Bearer " + o.apiKey,
	}); err != nil {
		return "", err
	}

	if r.Error != nil {
		return "", fmt.Errorf("%s", r.Error.Message)
	}
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content, nil
	}
	return "", nil
}
