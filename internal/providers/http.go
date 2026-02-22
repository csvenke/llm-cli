package providers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// doRequest executes an HTTP POST request with the given endpoint, body, and headers.
// It applies a 30-second timeout and returns the response body or an error.
func doRequest(ctx context.Context, endpoint string, body []byte, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// buildMessages creates a message slice with optional system prompt and user message.
func buildMessages(system, user string) []Message {
	msgs := make([]Message, 0, 2)
	if system != "" {
		msgs = append(msgs, Message{Role: "system", Content: system})
	}
	msgs = append(msgs, Message{Role: "user", Content: user})
	return msgs
}
