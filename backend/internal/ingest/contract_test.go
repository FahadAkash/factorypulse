package ingest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FahadAkash/factorypulse/backend/internal/ingest"
)

func TestContract_GoldenFixtures(t *testing.T) {
	// Verify that contracts/event-schema.json exists and is readable
	schemaPath := filepath.Join("..", "..", "..", "contracts", "event-schema.json")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("unable to read contract schema from %s: %v", schemaPath, err)
	}
	if len(schemaBytes) == 0 {
		t.Fatalf("contracts/event-schema.json is empty")
	}

	fixtures := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "StateChanged_Valid",
			json: `{
				"schema_version": 1,
				"event_id": "c71a3964-b6d3-57c2-9b04-f58c70a8d38a",
				"event_type": "machine.state_changed",
				"factory_id": "factory-01",
				"line_id": "line-refrig-01",
				"machine_id": "M-REF-02",
				"run_id": "run-001",
				"seq": 10,
				"occurred_at": "2026-11-05T12:00:00Z",
				"payload": {
					"state": "STOPPED",
					"reason_code": "MOTOR_OVERLOAD",
					"upstream_cause_machine_id": "M-REF-02"
				}
			}`,
			wantErr: false,
		},
		{
			name: "ProductionUnit_Valid",
			json: `{
				"schema_version": 1,
				"event_id": "d82b4975-c7e4-68d3-ac15-e69d81b9e49b",
				"event_type": "production.unit",
				"factory_id": "factory-01",
				"line_id": "line-refrig-01",
				"machine_id": "M-REF-04",
				"run_id": "run-001",
				"seq": 25,
				"occurred_at": "2026-11-05T12:00:54Z",
				"payload": {
					"outcome": "good",
					"cycle_time_ms": 54000,
					"product_type": "refrigerator-standard"
				}
			}`,
			wantErr: false,
		},
		{
			name: "Heartbeat_Valid",
			json: `{
				"schema_version": 1,
				"event_id": "e93c5086-d8f5-79e4-bd26-f70e92cae5ac",
				"event_type": "machine.heartbeat",
				"factory_id": "factory-01",
				"line_id": "line-refrig-01",
				"machine_id": "M-REF-01",
				"run_id": "run-001",
				"seq": 100,
				"occurred_at": "2026-11-05T12:01:00Z",
				"payload": {}
			}`,
			wantErr: false,
		},
		{
			name: "SensorReading_Valid",
			json: `{
				"schema_version": 1,
				"event_id": "fa4d6197-e906-8af5-ce37-081fa3dbf6bd",
				"event_type": "sensor.reading",
				"factory_id": "factory-01",
				"line_id": "line-refrig-01",
				"machine_id": "M-REF-02",
				"run_id": "run-001",
				"seq": 101,
				"occurred_at": "2026-11-05T12:01:05Z",
				"payload": {
					"sensor": "temperature_celsius",
					"value": 68.4,
					"unit": "C"
				}
			}`,
			wantErr: false,
		},
		{
			name: "InvalidSchemaVersion",
			json: `{
				"schema_version": 2,
				"event_id": "fa4d6197-e906-8af5-ce37-081fa3dbf6bd",
				"event_type": "sensor.reading",
				"factory_id": "factory-01",
				"line_id": "line-refrig-01",
				"machine_id": "M-REF-02",
				"run_id": "run-001",
				"seq": 1,
				"occurred_at": "2026-11-05T12:01:05Z",
				"payload": {}
			}`,
			wantErr: true,
		},
	}

	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ingest.ValidateEvent([]byte(tc.json))
			if (err != nil) != tc.wantErr {
				t.Fatalf("fixture %s: got error %v, wantErr %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
