package handyman

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

type RequestOptions struct {
	Method  string
	URL     string
	Body    []byte
	timeOut time.Duration
	Header  map[string]string
}

// SendRequest sends a request to the given URL with the given body and method.
// It returns the response and an error if any.
func SendRequest(ctx context.Context, opt RequestOptions) (*http.Response, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		opt.Method,
		opt.URL,
		bytes.NewBuffer(opt.Body),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range opt.Header {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: opt.timeOut}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}
