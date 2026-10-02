package transcriber

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	endpoint = "https://api.openai.com/v1/audio/transcriptions"
	model    = "gpt-4o-transcribe"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func New(apiKey string, httpClient *http.Client) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

type transcriptionResponse struct {
	Text string `json:"text"`
}

func (c *Client) Transcribe(ctx context.Context, audio []byte) (string, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	if err := mw.WriteField("model", model); err != nil {
		return "", fmt.Errorf("writing model field: %w", err)
	}

	part, err := mw.CreateFormFile("file", "recording.mp3")
	if err != nil {
		return "", fmt.Errorf("creating file field: %w", err)
	}

	if _, err := part.Write(audio); err != nil {
		return "", fmt.Errorf("writing audio: %w", err)
	}

	if err := mw.Close(); err != nil {
		return "", fmt.Errorf("closing multipart body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return "", fmt.Errorf("building transcription request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending transcription request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, msg)
	}

	var out transcriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decoding transcription response: %w", err)
	}

	return out.Text, nil
}
