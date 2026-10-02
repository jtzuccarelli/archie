package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	endpoint = "https://api.openai.com/v1/responses"
	model    = "gpt-4o-mini"
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

type Analysis struct {
	Flags   []string `json:"flags"`
	IsPushy bool     `json:"is_pushy"`
	Score   int      `json:"score"`
}

type responsesRequest struct {
	Model        string       `json:"model"`
	Instructions string       `json:"instructions"`
	Input        string       `json:"input"`
	Store        bool         `json:"store"`
	Text         textSettings `json:"text"`
}

type textSettings struct {
	Format textFormat `json:"format"`
}

type textFormat struct {
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Strict bool           `json:"strict"`
	Schema map[string]any `json:"schema"`
}

type responsesResponse struct {
	Status string `json:"status"`
	Output []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

var analysisSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"flags":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"is_pushy": map[string]any{"type": "boolean"},
		"score":    map[string]any{"type": "integer"},
	},
	"required":             []string{"flags", "is_pushy", "score"},
	"additionalProperties": false,
}

func (c *Client) Analyze(ctx context.Context, transcript string) (Analysis, error) {
	payload := responsesRequest{
		Model:        model,
		Instructions: instructions,
		Input:        transcript,
		Store:        false,
		Text: textSettings{
			Format: textFormat{
				Type:   "json_schema",
				Name:   "call_analysis",
				Strict: true,
				Schema: analysisSchema,
			},
		},
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return Analysis{}, fmt.Errorf("encoding analysis request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return Analysis{}, fmt.Errorf("building analysis request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Analysis{}, fmt.Errorf("sending analysis request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Analysis{}, fmt.Errorf("openai returned status %d: %s", resp.StatusCode, msg)
	}

	var out responsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Analysis{}, fmt.Errorf("decoding analysis response: %w", err)
	}

	if out.Status != "completed" {
		return Analysis{}, fmt.Errorf("analysis response status %q", out.Status)
	}

	text, err := outputText(out)
	if err != nil {
		return Analysis{}, err
	}

	var analysis Analysis
	if err := json.Unmarshal([]byte(text), &analysis); err != nil {
		return Analysis{}, fmt.Errorf("decoding analysis output: %w", err)
	}

	return analysis, nil
}

func outputText(r responsesResponse) (string, error) {
	for _, item := range r.Output {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			if content.Type == "output_text" {
				return content.Text, nil
			}
		}
	}

	return "", errors.New("analysis response has no output_text")
}
