package main

import (
	"errors"
	"os"
)

type config struct {
	dsn             string
	tdAuthHeader    string
	openaiApiKey    string
	anthropicApiKey string
}

func loadConfig() (config, error) {
	cfg := config{
		dsn:             os.Getenv("DATABASE_URL"),
		tdAuthHeader:    os.Getenv("TD_AUTH_HEADER"),
		openaiApiKey:    os.Getenv("OPENAI_API_KEY"),
		anthropicApiKey: os.Getenv("ANTHROPIC_API_KEY"),
	}

	if cfg.dsn == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.tdAuthHeader == "" {
		return config{}, errors.New("TD_AUTH_HEADER is required")
	}

	if cfg.openaiApiKey == "" {
		return config{}, errors.New("OPENAI_API_KEY is required")
	}

	if cfg.anthropicApiKey == "" {
		return config{}, errors.New("ANTHROPIC_API_KEY is required")
	}

	return cfg, nil
}
