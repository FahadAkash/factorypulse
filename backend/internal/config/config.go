package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port          string
	PostgresDSN   string
	RedisAddr     string
	MQTTURL       string
	MQTTHost      string
	MQTTPort      int
	WorkerPoolCap int
	LogLevel      string
}

func Load() (*Config, error) {
	port := getEnv("FP_PORT", "8080")
	pgDSN := getEnv("FP_PG_DSN", "postgres://fp:fp_dev_only@localhost:5432/factorypulse?sslmode=disable")
	redisAddr := getEnv("FP_REDIS_ADDR", "localhost:6379")
	mqttURL := getEnv("FP_MQTT_URL", "tcp://localhost:1883")
	mqttHost := getEnv("FP_MQTT_HOST", "localhost")
	mqttPortStr := getEnv("FP_MQTT_PORT", "1883")

	mqttPort, err := strconv.Atoi(mqttPortStr)
	if err != nil {
		return nil, fmt.Errorf("invalid FP_MQTT_PORT %q: %w", mqttPortStr, err)
	}

	workersStr := getEnv("FP_WORKERS", "16")
	workers, err := strconv.Atoi(workersStr)
	if err != nil || workers <= 0 {
		return nil, fmt.Errorf("invalid FP_WORKERS %q: must be positive integer", workersStr)
	}

	cfg := &Config{
		Port:          port,
		PostgresDSN:   pgDSN,
		RedisAddr:     redisAddr,
		MQTTURL:       mqttURL,
		MQTTHost:      mqttHost,
		MQTTPort:      mqttPort,
		WorkerPoolCap: workers,
		LogLevel:      getEnv("FP_LOG_LEVEL", "INFO"),
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
