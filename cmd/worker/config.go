package main

import (
	"errors"
	"os"
)

type config struct {
	dsn          string
	tdAuthHeader string
}

func loadConfig() (config, error) {
	cfg := config{
		dsn:          os.Getenv("DATABASE_URL"),
		tdAuthHeader: os.Getenv("TD_AUTH_HEADER"),
	}

	if cfg.dsn == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.tdAuthHeader == "" {
		return config{}, errors.New("TD_AUTH_HEADER is required")
	}

	return cfg, nil
}
