package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
    DBUser string
    DBPass string
    DBName string
    DBHost string
    DBPort string
    JudgeDataDir  string
}

func Load() Config {
    return Config{
        DBUser: mustEnv("DB_USER"),
        DBPass: mustEnv("DB_PASSWORD"),
        DBName: mustEnv("DB_NAME"),
        DBHost: envOrDefault("DB_HOST", "localhost"),
        DBPort: envOrDefault("DB_PORT", "5432"),
        JudgeDataDir: mustEnv("JUDGE_DATA_DIR"),
    }
}

func (c Config) DatabaseUrl() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
        c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}

func mustEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        log.Fatalf("missing required env var: %s", key)
    }
    return v
}

func envOrDefault(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}