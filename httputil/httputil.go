package httputil

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
)

type RequestOptions struct {
	Method  string
	URL     string
	Body    []byte
	Timeout time.Duration
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
		return nil, errors.Wrap(err, "failed to create request: %w")
	}

	for k, v := range opt.Header {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: opt.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request: %w")
	}

	return resp, nil
}
