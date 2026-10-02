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

const maxAudioBytes = 25 << 20

func (client *Client) Download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building download request: %w", err)
	}

	req.Header.Set("Authorization", client.authHeader)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading audio: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("TrackDrive returned HTTP status %d", resp.StatusCode)
	}

	audio, err := io.ReadAll(io.LimitReader(resp.Body, maxAudioBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading audio body: %w", err)
	}

	if len(audio) > maxAudioBytes {
		return nil, fmt.Errorf("audio exceeds %d bytes", maxAudioBytes)
	}

	return audio, nil
}
