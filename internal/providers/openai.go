package providers

import (
	"context"
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
	var r openaiResponse
	if err := doJSONRequest(ctx, o.endpoint, openaiRequest{
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
