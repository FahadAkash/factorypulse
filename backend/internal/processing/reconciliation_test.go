package processing_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/FahadAkash/factorypulse/backend/internal/domain"
)

// In-Memory Test Store to verify reconciliation and replay without external DB
type memoryStore struct {
	mu           sync.Mutex
	events       map[string]domain.EventEnvelope
	duplicates   int
	machineState map[string]string
	machineSeq   map[string]int64
	incidents    []domain.Incident
	production   map[string]int // machineId -> count
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		events:       make(map[string]domain.EventEnvelope),
		machineState: make(map[string]string),
		machineSeq:   make(map[string]int64),
		production:   make(map[string]int),
	}
}

func (s *memoryStore) Ingest(env domain.EventEnvelope) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Idempotency test: check event_id
	if _, exists := s.events[env.EventID]; exists {
		s.duplicates++
		return false // duplicate ignored
	}
	s.events[env.EventID] = env

	// 2. Out-of-order check: only advance state if seq > lastSeq
	if env.EventType == domain.EventMachineStateChanged {
		var p domain.StatePayload
		_ = json.Unmarshal(env.Payload, &p)

		lastSeq := s.machineSeq[env.MachineID]
		if env.Seq >= lastSeq {
			s.machineState[env.MachineID] = string(p.State)
			s.machineSeq[env.MachineID] = env.Seq
		}

		// Double STOP check: 1 active incident per machine
		if p.State == domain.StateStopped {
			hasActive := false
			for _, inc := range s.incidents {
				if inc.MachineID == env.MachineID && inc.Status != domain.IncidentStatusResolved {
					hasActive = true
					break
				}
			}
			if !hasActive {
				s.incidents = append(s.incidents, domain.Incident{
					ID:        env.EventID,
					MachineID: env.MachineID,
					Status:    domain.IncidentStatusOpen,
					OpenedAt:  env.OccurredAt,
				})
			}
		}
	} else if env.EventType == domain.EventProductionUnit {
		s.production[env.MachineID]++
	}

	return true
}

func (s *memoryStore) Replay() (map[string]int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Recompute aggregates by scanning immutable raw events
	recomputedProd := make(map[string]int)
	for _, env := range s.events {
		if env.EventType == domain.EventProductionUnit {
			recomputedProd[env.MachineID]++
		}
	}
	return recomputedProd, len(s.events)
}

func TestReconciliation_IdempotentDuplicateDelivery(t *testing.T) {
	store := newMemoryStore()

	ev := domain.EventEnvelope{
		SchemaVersion: 1,
		EventID:       "uuid-001",
		EventType:     domain.EventProductionUnit,
		FactoryID:     "factory-01",
		LineID:        "line-refrig-01",
		MachineID:     "M-REF-04",
		RunID:         "run-001",
		Seq:           1,
		OccurredAt:    time.Now().UTC(),
		Payload:       json.RawMessage(`{"outcome":"good","cycle_time_ms":54000}`),
	}

	// Deliver identical event 10 times (simulating MQTT duplicate QoS 1 retry)
	for i := 0; i < 10; i++ {
		store.Ingest(ev)
	}

	if len(store.events) != 1 {
		t.Fatalf("expected exactly 1 stored event, got %d", len(store.events))
	}
	if store.duplicates != 9 {
		t.Errorf("expected 9 duplicate events detected, got %d", store.duplicates)
	}
	if store.production["M-REF-04"] != 1 {
		t.Errorf("expected production count 1, got %d", store.production["M-REF-04"])
	}
}

func TestReconciliation_OutOfOrderStateArrival(t *testing.T) {
	store := newMemoryStore()

	evNewer := domain.EventEnvelope{
		SchemaVersion: 1,
		EventID:       "uuid-seq-2",
		EventType:     domain.EventMachineStateChanged,
		MachineID:     "M-REF-01",
		Seq:           2,
		OccurredAt:    time.Now().UTC(),
		Payload:       json.RawMessage(`{"state":"RUNNING"}`),
	}

	evOlder := domain.EventEnvelope{
		SchemaVersion: 1,
		EventID:       "uuid-seq-1",
		EventType:     domain.EventMachineStateChanged,
		MachineID:     "M-REF-01",
		Seq:           1,
		OccurredAt:    time.Now().UTC().Add(-10 * time.Second),
		Payload:       json.RawMessage(`{"state":"STOPPED","reason_code":"JAM"}`),
	}

	// Seq 2 arrives first, then Seq 1 arrives late
	store.Ingest(evNewer)
	store.Ingest(evOlder)

	// Final state must remain RUNNING (newer seq 2), not regressed by seq 1
	if store.machineState["M-REF-01"] != "RUNNING" {
		t.Errorf("expected machine state to remain RUNNING, got %s", store.machineState["M-REF-01"])
	}
}

func TestReconciliation_SingleActiveIncidentUnderDoubleStop(t *testing.T) {
	store := newMemoryStore()

	evStop1 := domain.EventEnvelope{
		SchemaVersion: 1,
		EventID:       "uuid-stop-1",
		EventType:     domain.EventMachineStateChanged,
		MachineID:     "M-REF-02",
		Seq:           1,
		OccurredAt:    time.Now().UTC(),
		Payload:       json.RawMessage(`{"state":"STOPPED","reason_code":"MOTOR_OVERLOAD"}`),
	}

	evStop2 := domain.EventEnvelope{
		SchemaVersion: 1,
		EventID:       "uuid-stop-2",
		EventType:     domain.EventMachineStateChanged,
		MachineID:     "M-REF-02",
		Seq:           2,
		OccurredAt:    time.Now().UTC().Add(2 * time.Second),
		Payload:       json.RawMessage(`{"state":"STOPPED","reason_code":"MOTOR_OVERLOAD"}`),
	}

	store.Ingest(evStop1)
	store.Ingest(evStop2)

	activeIncidents := 0
	for _, inc := range store.incidents {
		if inc.MachineID == "M-REF-02" && inc.Status != domain.IncidentStatusResolved {
			activeIncidents++
		}
	}

	if activeIncidents != 1 {
		t.Fatalf("expected exactly 1 active incident, got %d", activeIncidents)
	}
}

func TestReconciliation_ReplayFromRawEvents(t *testing.T) {
	store := newMemoryStore()

	for i := 1; i <= 50; i++ {
		ev := domain.EventEnvelope{
			SchemaVersion: 1,
			EventID:       string(rune(i + 1000)),
			EventType:     domain.EventProductionUnit,
			MachineID:     "M-REF-04",
			Seq:           int64(i),
			OccurredAt:    time.Now().UTC(),
			Payload:       json.RawMessage(`{"outcome":"good","cycle_time_ms":54000}`),
		}
		store.Ingest(ev)
	}

	liveProd := store.production["M-REF-04"]
	recomputedProd, totalEvents := store.Replay()

	if totalEvents != 50 {
		t.Errorf("expected 50 raw events replayed, got %d", totalEvents)
	}
	if recomputedProd["M-REF-04"] != liveProd {
		t.Fatalf("replay reconciliation discrepancy! live: %d, replayed: %d", liveProd, recomputedProd["M-REF-04"])
	}
}
