package api_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FahadAkash/factorypulse/backend/internal/api"
)

func TestHealthAndReadinessEndpoints(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	srv := api.NewServer("0", nil, nil, logger)

	// Test /healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected /healthz 200 OK, got %d", resp.StatusCode)
	}

	var healthBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&healthBody); err != nil {
		t.Fatalf("failed to decode healthz response: %v", err)
	}
	if healthBody["status"] != "UP" {
		t.Errorf("expected status UP, got %s", healthBody["status"])
	}

	// Test /readyz
	reqReady := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	wReady := httptest.NewRecorder()
	srv.ServeHTTP(wReady, reqReady)

	respReady := wReady.Result()
	if respReady.StatusCode != http.StatusOK {
		t.Fatalf("expected /readyz 200 OK, got %d", respReady.StatusCode)
	}

	var readyBody map[string]string
	if err := json.NewDecoder(respReady.Body).Decode(&readyBody); err != nil {
		t.Fatalf("failed to decode readyz response: %v", err)
	}
	if readyBody["status"] != "READY" {
		t.Errorf("expected status READY, got %s", readyBody["status"])
	}

	// Test /api/v1/factory
	reqFactory := httptest.NewRequest(http.MethodGet, "/api/v1/factory", nil)
	wFactory := httptest.NewRecorder()
	srv.ServeHTTP(wFactory, reqFactory)

	respFactory := wFactory.Result()
	if respFactory.StatusCode != http.StatusOK {
		t.Fatalf("expected /api/v1/factory 200 OK, got %d", respFactory.StatusCode)
	}

	// Test /api/v1/shift/forecast
	reqShift := httptest.NewRequest(http.MethodGet, "/api/v1/shift/forecast", nil)
	wShift := httptest.NewRecorder()
	srv.ServeHTTP(wShift, reqShift)

	respShift := wShift.Result()
	if respShift.StatusCode != http.StatusOK {
		t.Fatalf("expected /api/v1/shift/forecast 200 OK, got %d", respShift.StatusCode)
	}
}

