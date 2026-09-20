package config

import (
	"os"
)

type Config struct {
	DBPath string
}

func Load() *Config {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/finance.db"
	}
	return &Config{DBPath: dbPath}
}
