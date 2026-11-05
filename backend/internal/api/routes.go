package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/incident"
	"github.com/FahadAkash/factorypulse/backend/internal/shift"
	"github.com/FahadAkash/factorypulse/backend/internal/ws"
)

type APIHandler struct {
	hub       *ws.Hub
	incidents *incident.Service
}

func NewAPIHandler(hub *ws.Hub, incidents *incident.Service) *APIHandler {
	return &APIHandler{
		hub:       hub,
		incidents: incidents,
	}
}

func (h *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ws", h.handleWebSocket)
	mux.HandleFunc("/api/v1/factory", h.handleFactoryOverview)
	mux.HandleFunc("/api/v1/incidents", h.handleIncidents)
	mux.HandleFunc("/api/v1/incidents/", h.handleIncidentActions)
	mux.HandleFunc("/api/v1/shift/forecast", h.handleShiftForecast)
}

func (h *APIHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if h.hub != nil {
		h.hub.ServeWS(w, r)
	} else {
		http.Error(w, "WebSocket hub unavailable", http.StatusServiceUnavailable)
	}
}

func (h *APIHandler) handleFactoryOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"factory_id": "factory-01",
		"name":       "Nova Central Plant",
		"lines_count": 3,
		"machines_count": 12,
		"active_shift": "Shift 1 (06:00 - 14:00)",
		"shift_target": 500,
		"server_time": time.Now().UTC(),
	})
}

func (h *APIHandler) handleIncidents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if h.incidents != nil {
		list, err := h.incidents.ListActiveIncidents(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(list)
	} else {
		_ = json.NewEncoder(w).Encode([]any{})
	}
}

func (h *APIHandler) handleIncidentActions(w http.ResponseWriter, r *http.Request) {
	// Path format: /api/v1/incidents/{id}/acknowledge OR /api/v1/incidents/{id}/resolve
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	incidentID := pathParts[2]
	action := pathParts[3]

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.incidents == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch action {
	case "acknowledge":
		var req struct {
			Assignee string `json:"assignee"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		assignee := req.Assignee
		if assignee == "" {
			assignee = "Operator/Tech"
		}

		err := h.incidents.AcknowledgeIncident(r.Context(), incidentID, assignee, time.Now().UTC())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "acknowledged"})

	case "resolve":
		var req struct {
			RootCause string `json:"root_cause"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		rootCause := req.RootCause
		if rootCause == "" {
			rootCause = "RESOLVED_MANUAL"
		}

		err := h.incidents.ResolveIncident(r.Context(), incidentID, rootCause, time.Now().UTC())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})

	default:
		http.Error(w, "unknown incident action", http.StatusBadRequest)
	}
}

func (h *APIHandler) handleShiftForecast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 8 hour shift, example mid-shift forecast calculation
	forecast := shift.ComputeShiftForecast(500, 260, 8*time.Hour, 4*time.Hour, 65)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(forecast)
}
