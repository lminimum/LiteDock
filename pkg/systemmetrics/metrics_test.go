package systemmetrics_test

import (
	"testing"
	"time"

	"github.com/lminimum/LiteDock/pkg/systemmetrics"
)

func TestGetSystemMetrics_Success(t *testing.T) {
	t.Parallel()

	sm, err := systemmetrics.GetSystemMetrics()
	if err != nil {
		t.Fatalf("GetSystemMetrics() unexpected error: %v", err)
	}

	if sm.CPU < 0 {
		t.Errorf("CPU = %f, want >= 0", sm.CPU)
	}
	if sm.Memory < 0 || sm.Memory > 100 {
		t.Errorf("Memory = %f, want [0, 100]", sm.Memory)
	}
	if sm.Disk < 0 || sm.Disk > 100 {
		t.Errorf("Disk = %f, want [0, 100]", sm.Disk)
	}

	if sm.At.IsZero() {
		t.Error("At should not be zero")
	}
	if time.Since(sm.At) > 5*time.Second {
		t.Errorf("At is too far in the past: %v ago", time.Since(sm.At))
	}
}
