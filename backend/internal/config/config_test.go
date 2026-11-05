package config_test

import (
	"os"
	"testing"

	"github.com/FahadAkash/factorypulse/backend/internal/config"
)

func TestConfig_LoadDefaults(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error loading default config: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.WorkerPoolCap != 16 {
		t.Errorf("expected 16 workers, got %d", cfg.WorkerPoolCap)
	}
	if cfg.MQTTPort != 1883 {
		t.Errorf("expected MQTT port 1883, got %d", cfg.MQTTPort)
	}
}

func TestConfig_EnvOverrides(t *testing.T) {
	_ = os.Setenv("FP_PORT", "9090")
	_ = os.Setenv("FP_WORKERS", "32")
	defer func() {
		_ = os.Unsetenv("FP_PORT")
		_ = os.Unsetenv("FP_WORKERS")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error loading overridden config: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.WorkerPoolCap != 32 {
		t.Errorf("expected 32 workers, got %d", cfg.WorkerPoolCap)
	}
}

