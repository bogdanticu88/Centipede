package parsers

import (
	"testing"
)

func TestGenericJSONParser(t *testing.T) {
	data := []byte(`[
  {
    "timestamp": "2026-03-28T10:00:00Z",
    "tenant_id": "tenant1",
    "endpoint": "/api/test",
    "method": "GET",
    "status_code": 200,
    "payload_size": 1024,
    "response_ms": 50,
    "user_agent": "test",
    "source_ip": "127.0.0.1"
  }
]`)

	parser := &GenericJSONParser{}
	calls, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(calls))
	}

	if calls[0].TenantID != "tenant1" {
		t.Errorf("expected tenant1, got %s", calls[0].TenantID)
	}

	if calls[0].StatusCode != 200 {
		t.Errorf("expected status 200, got %d", calls[0].StatusCode)
	}
}

func TestGenericJSONParserMultipleCalls(t *testing.T) {
	data := []byte(`[
  {
    "timestamp": "2026-03-28T10:00:00Z",
    "tenant_id": "tenant1",
    "endpoint": "/api/test",
    "method": "GET",
    "status_code": 200,
    "payload_size": 1024,
    "response_ms": 50,
    "user_agent": "test",
    "source_ip": "127.0.0.1"
  },
  {
    "timestamp": "2026-03-28T10:00:01Z",
    "tenant_id": "tenant2",
    "endpoint": "/api/other",
    "method": "POST",
    "status_code": 201,
    "payload_size": 2048,
    "response_ms": 100,
    "user_agent": "test",
    "source_ip": "127.0.0.2"
  }
]`)

	parser := &GenericJSONParser{}
	calls, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 2 {
		t.Errorf("expected 2 calls, got %d", len(calls))
	}
}

func TestGenericJSONParserInvalidJSON(t *testing.T) {
	data := []byte(`invalid json`)

	parser := &GenericJSONParser{}
	_, err := parser.Parse(data)

	if err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}
