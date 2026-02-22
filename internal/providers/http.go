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

func doJSONRequest(ctx context.Context, endpoint string, req, resp any, headers map[string]string) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	body, err := doRequest(ctx, endpoint, jsonData, headers)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

func buildMessages(system, user string) []Message {
	msgs := make([]Message, 0, 2)
	if system != "" {
		msgs = append(msgs, Message{Role: "system", Content: system})
	}
	msgs = append(msgs, Message{Role: "user", Content: user})
	return msgs
}
