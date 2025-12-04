package config

import (
	"log/slog"

	"github.com/joeshaw/envdecode"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Host string `env:"SERVER_HOST,default=localhost"`
	Port string `env:"SERVER_PORT,default=8080"`
}

type DatabaseConfig struct {
	DBHost     string `env:"PGHOST,default=localhost"`
	DBPort     string `env:"PGPORT,default=5432"`
	DBUser     string `env:"PGUSER,default=testuser"`
	DBPassword string `env:"PGPASSWORD,default=123456"`
	DBName     string `env:"PGDATABASE,default=postgres"`
}

func New() (*AppConfig, error) {
	if err := godotenv.Load(".env"); err != nil {
		slog.Error("dotenv load failed", err)
		return nil, err
	}

	var c AppConfig
	if err := envdecode.StrictDecode(&c); err != nil {
		slog.Error("failed to load config", err)
		return nil, err
	}

	return &c, nil
}