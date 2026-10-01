package trackdrive

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	authHeader string
	httpClient *http.Client
}

func New(authHeader string, httpClient *http.Client) *Client {
	return &Client{
		authHeader: authHeader,
		httpClient: httpClient,
	}
}

func (client *Client) Download(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building download request: %w", err)
	}

	req.Header.Set("Authorization", client.authHeader)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading audio: %w", err)
	}

	return resp.Body, nil
}
