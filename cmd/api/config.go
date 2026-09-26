package main

import (
	"errors"
	"os"
)

type config struct {
	port string
	dsn  string
}

func loadConfig() (config, error) {
	cfg := config{
		port: os.Getenv("PORT"),
		dsn:  os.Getenv("DATABASE_URL"),
	}

	if cfg.port == "" {
		cfg.port = "4000"
	}

	if cfg.dsn == "" {
		return config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}
