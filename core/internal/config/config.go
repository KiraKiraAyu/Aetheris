package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr           string
	DatabaseURL        string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	QueueType          string
	APIKeys            map[string]string
	CORSAllowedOrigins []string
	RequestMaxBytes    int64
	RateLimitEnabled   bool
	RateLimitPerMinute int
	QueueName          string
	UniqueTTL          time.Duration
	WorkerConcurrency  int
}

func Load() Config {
	loadRootDotenv()

	return Config{
		HTTPAddr:           getenv("HTTP_ADDR", ":8080"),
		DatabaseURL:        getenv("DATABASE_URL", "aetheris.db"),
		RedisAddr:          getenv("REDIS_ADDR", ""),
		RedisPassword:      getenv("REDIS_PASSWORD", ""),
		RedisDB:            getenvInt("REDIS_DB", 0),
		QueueType:          getenv("QUEUE_TYPE", "db"),
		APIKeys:            getenvAPIKeys("API_KEYS"),
		CORSAllowedOrigins: getenvCSV("CORS_ALLOWED_ORIGINS"),
		RequestMaxBytes:    int64(getenvInt("REQUEST_MAX_BYTES", 1<<20)),
		RateLimitEnabled:   getenvBool("RATE_LIMIT_ENABLED", false),
		RateLimitPerMinute: getenvInt("RATE_LIMIT_PER_MINUTE", 600),
		QueueName:          getenv("QUEUE_NAME", "notifications"),
		UniqueTTL:          getenvDuration("ASYNQ_UNIQUE_TTL", 5*time.Minute),
		WorkerConcurrency:  getenvInt("WORKER_CONCURRENCY", 10),
	}
}

func loadRootDotenv() {
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")
}

func getenvAPIKeys(key string) map[string]string {
	value := os.Getenv(key)
	if value == "" {
		return map[string]string{}
	}
	result := map[string]string{}
	for _, pair := range strings.Split(value, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		apiKey := strings.TrimSpace(parts[0])
		tenantID := strings.TrimSpace(parts[1])
		if apiKey != "" && tenantID != "" {
			result[apiKey] = tenantID
		}
	}
	return result
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvJSONMap(key string) map[string]string {
	value := os.Getenv(key)
	if value == "" {
		return map[string]string{}
	}
	var parsed map[string]string
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return map[string]string{}
	}
	if parsed == nil {
		return map[string]string{}
	}
	return parsed
}

func getenvCSV(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}


