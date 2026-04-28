package entity

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNetworkMarshalJSON(t *testing.T) {
	t.Helper()

	network := Network{
		ID:         "abc123",
		Name:       "test-network",
		Driver:     "bridge",
		Scope:      "local",
		Internal:   false,
		Attachable: false,
		EnableIPv6: true,
		Created:    "2024-01-01T00:00:00Z",
		Labels: map[string]string{
			"env": "test",
		},
		CachedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(network)
	if err != nil {
		t.Fatalf("failed to marshal Network: %v", err)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify field names match Docker API conventions (lowercase_with_underscore)
	expectedFields := []string{
		"id", "name", "driver", "scope", "internal",
		"attachable", "enable_ipv6", "created", "labels", "cached_at",
	}

	for _, field := range expectedFields {
		if _, ok := got[field]; !ok {
			t.Errorf("expected field %q not found in JSON output", field)
		}
	}
}

func TestNetworkUnmarshalJSON(t *testing.T) {
	t.Helper()

	jsonStr := `{
		"id": "abc123",
		"name": "test-network",
		"driver": "bridge",
		"scope": "local",
		"internal": false,
		"attachable": true,
		"enable_ipv6": true,
		"created": "2024-01-01T00:00:00Z",
		"labels": {"env": "test"},
		"cached_at": "2024-01-01T00:00:00Z"
	}`

	var network Network
	if err := json.Unmarshal([]byte(jsonStr), &network); err != nil {
		t.Fatalf("failed to unmarshal JSON to Network: %v", err)
	}

	if network.ID != "abc123" {
		t.Errorf("expected ID 'abc123', got %q", network.ID)
	}
	if network.Name != "test-network" {
		t.Errorf("expected Name 'test-network', got %q", network.Name)
	}
	if network.Driver != "bridge" {
		t.Errorf("expected Driver 'bridge', got %q", network.Driver)
	}
	if network.Scope != "local" {
		t.Errorf("expected Scope 'local', got %q", network.Scope)
	}
	if network.Internal != false {
		t.Errorf("expected Internal false, got %v", network.Internal)
	}
	if network.Attachable != true {
		t.Errorf("expected Attachable true, got %v", network.Attachable)
	}
	if network.EnableIPv6 != true {
		t.Errorf("expected EnableIPv6 true, got %v", network.EnableIPv6)
	}
	if network.Created != "2024-01-01T00:00:00Z" {
		t.Errorf("expected Created '2024-01-01T00:00:00Z', got %q", network.Created)
	}
	if network.Labels["env"] != "test" {
		t.Errorf("expected Labels['env'] 'test', got %q", network.Labels["env"])
	}
}