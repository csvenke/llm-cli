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

// OpenAI Chat Completions API types
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

// OpenAIProvider implements Provider for OpenAI Chat Completions API.
type OpenAIProvider struct {
	endpoint string
	model    string
	apiKey   string
}

// NewOpenAIProvider creates a new OpenAI provider with the given configuration.
func NewOpenAIProvider(endpoint, model, apiKey string) Provider {
	return &OpenAIProvider{
		endpoint: endpoint,
		model:    model,
		apiKey:   apiKey,
	}
}

func (o *OpenAIProvider) Complete(ctx context.Context, system, userMsg string) (string, error) {
	msgs := []Message{}
	if system != "" {
		msgs = append(msgs, Message{Role: "system", Content: system})
	}
	msgs = append(msgs, Message{Role: "user", Content: userMsg})

	jsonData, err := json.Marshal(openaiRequest{
		Model:    o.model,
		Messages: msgs,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

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

	var r openaiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%d: %s", resp.StatusCode, string(body))
	}
	if r.Error != nil {
		return "", fmt.Errorf("%s", r.Error.Message)
	}

	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content, nil
	}
	return "", nil
}
