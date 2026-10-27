package analyzer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/jtzuccarelli/archie/internal/retry"
)

type Analysis struct {
	Flags []string `json:"flags"`
}

var analysisSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"flags": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
	},
	"required":             []string{"flags"},
	"additionalProperties": false,
}

type Client struct {
	client anthropic.Client
}

func New(apiKey string, httpClient *http.Client) *Client {
	return &Client{
		client: anthropic.NewClient(
			option.WithAPIKey(apiKey),
			option.WithHTTPClient(httpClient),
			option.WithMaxRetries(0),
		),
	}
}

func (c *Client) Analyze(ctx context.Context, transcript string) (Analysis, error) {
	message, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet5_5,
		MaxTokens: 4096,
		System:    []anthropic.TextBlockParam{{Text: instructions}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(transcript)),
		},
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{Schema: analysisSchema},
		},
	})
	if err != nil {
		var apiErr *anthropic.Error
		if errors.As(err, &apiErr) {
			if apiErr.StatusCode == http.StatusRequestTimeout || apiErr.StatusCode == http.StatusConflict || apiErr.StatusCode == http.StatusTooManyRequests || apiErr.StatusCode >= 500 {
				return Analysis{}, retry.Retryable(fmt.Errorf("anthropic analysis: %w", err))
			}
			return Analysis{}, retry.Permanent(fmt.Errorf("anthropic analysis: %w", err))
		}
		return Analysis{}, retry.Retryable(fmt.Errorf("sending anthropic analysis request: %w", err))
	}

	if message.StopReason != anthropic.StopReasonEndTurn {
		return Analysis{}, retry.Permanent(fmt.Errorf("anthropic analysis stopped: %s", message.StopReason))
	}

	for _, block := range message.Content {
		if block.Type != "text" {
			continue
		}
		var analysis Analysis
		if err := json.Unmarshal([]byte(block.Text), &analysis); err != nil {
			return Analysis{}, retry.Permanent(fmt.Errorf("decoding anthropic analysis output: %w", err))
		}
		return analysis, nil
	}

	return Analysis{}, retry.Permanent(errors.New("anthropic analysis response has no text"))
}
