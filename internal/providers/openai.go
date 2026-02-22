package providers

import (
	"context"
	"encoding/json"
	"fmt"
)

type openaiRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type OpenAIProvider struct {
	endpoint string
	model    string
	apiKey   string
}

func NewOpenAIProvider(endpoint, model, apiKey string) Provider {
	return &OpenAIProvider{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
	}
}

func (o *OpenAIProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	jsonData, err := json.Marshal(openaiRequest{
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
