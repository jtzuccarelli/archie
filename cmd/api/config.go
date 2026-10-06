package main

import (
	"errors"
	"os"
)

type config struct {
	port     string
	dsn      string
	qvSecret string
}

func loadConfig() (config, error) {
	cfg := config{
		port:     os.Getenv("PORT"),
		dsn:      os.Getenv("DATABASE_URL"),
		qvSecret: os.Getenv("QV_WEBHOOK_SECRET"),
	}

	if cfg.port == "" {
		cfg.port = "4000"
	}

	if cfg.dsn == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.qvSecret == "" {
		return config{}, errors.New("QV_WEBHOOK_SECRET is required")
	}

	return cfg, nil
}
