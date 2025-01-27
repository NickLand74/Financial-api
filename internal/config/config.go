package config

import (
	"context"
	"os"
)

type Config struct {
	DatabaseURL string
	Context     context.Context
}

func LoadConfig() (*Config, error) {
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Context:     context.Background(),
	}, nil
}
