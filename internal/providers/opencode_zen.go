package providers

import (
	"context"
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
	var r opencodeZenResponse
	if err := doJSONRequest(ctx, o.endpoint, opencodeZenRequest{
		Model:     o.model,
		MaxTokens: 4096,
		System:    system,
		Messages:  []Message{{Role: "user", Content: userMsg}},
	}, &r, map[string]string{
		"x-api-key":         o.apiKey,
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
