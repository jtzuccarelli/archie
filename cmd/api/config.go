package main

import (
	"errors"
	"os"
)

type config struct {
	port   string
	dsn    string
	apiKey string
}

func loadConfig() (config, error) {
	cfg := config{
		port:   os.Getenv("PORT"),
		dsn:    os.Getenv("DATABASE_URL"),
		apiKey: os.Getenv("QV_WEBHOOK_SECRET"),
	}

	if cfg.port == "" {
		cfg.port = "4000"
	}

	if cfg.dsn == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.apiKey == "" {
		return config{}, errors.New("QV_WEBHOOK_SECRET is required")
	}

	return cfg, nil
}
