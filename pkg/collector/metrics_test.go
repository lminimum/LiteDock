package collector_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lminimum/LiteDock/internal/entity"
	"github.com/lminimum/LiteDock/pkg/collector"
	"github.com/lminimum/LiteDock/pkg/systemmetrics"
)

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

// mockMetricsRepo implements repo.SystemMetricsRepo for testing.
type mockMetricsRepo struct {
	mu           sync.Mutex
	createCalled bool
	lastMetric   *entity.SystemMetric
	deleteCalled bool
}

func (m *mockMetricsRepo) Create(_ context.Context, metric *entity.SystemMetric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalled = true
	m.lastMetric = metric
	return nil
}

func (m *mockMetricsRepo) GetHistory(_ context.Context, _ time.Time) ([]entity.SystemMetric, error) {
	return nil, nil
}

func (m *mockMetricsRepo) DeleteOlderThan(_ context.Context, _ time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalled = true
	return nil
}

func (m *mockMetricsRepo) wasCreateCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.createCalled
}

func (m *mockMetricsRepo) getLastMetric() *entity.SystemMetric {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastMetric
}

func (m *mockMetricsRepo) wasDeleteCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.deleteCalled
}

// mockLogger implements logger.Interface for testing.
type mockLogger struct{}

func (m *mockLogger) Debug(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Info(_ string, _ ...interface{})       {}
func (m *mockLogger) Warn(_ string, _ ...interface{})       {}
func (m *mockLogger) Error(_ interface{}, _ ...interface{}) {}
func (m *mockLogger) Fatal(_ interface{}, _ ...interface{}) {}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// startAndStop runs mc.Start() in a background goroutine, calls mc.Stop()
// after the given delay, and waits for Start() to return (with timeout).
func startAndStop(t *testing.T, mc *collector.MetricsCollector, stopAfter time.Duration) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		mc.Start()
		close(done)
	}()

	time.Sleep(stopAfter)
	mc.Stop()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Start() did not return after Stop()")
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestMetricsCollector_StartStop(t *testing.T) {
	t.Parallel()

	repo := &mockMetricsRepo{}
	log := &mockLogger{}
	mc := collector.NewMetricsCollector(repo, log, time.Hour)

	startAndStop(t, mc, 50*time.Millisecond)
}

func TestMetricsCollector_Collect_SavesMetric(t *testing.T) {
	// Skip if system metrics collection is unavailable (e.g. restricted container).
	if _, err := systemmetrics.GetSystemMetrics(); err != nil {
		t.Skipf("system metrics not available, skipping collect test: %v", err)
	}

	repo := &mockMetricsRepo{}
	log := &mockLogger{}
	// Use a long interval so the collection ticker never fires during the test.
	mc := collector.NewMetricsCollector(repo, log, time.Hour)

	done := make(chan struct{})
	go func() {
		mc.Start()
		close(done)
	}()

	// Wait for the initial collect() call to complete.
	// collect() uses cpu.Percent(200ms), so 500ms gives a safe margin.
	time.Sleep(500 * time.Millisecond)

	mc.Stop()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Start() did not return after Stop()")
	}

	if !repo.wasCreateCalled() {
		t.Fatal("expected repo.Create to be called by collect()")
	}

	metric := repo.getLastMetric()
	if metric == nil {
		t.Fatal("expected metric to be non-nil")
	}
	if metric.RecordedAt.IsZero() {
		t.Error("expected RecordedAt to be set")
	}
}

func TestMetricsCollector_CleanupOld(t *testing.T) {
	// Verify that the cleanup goroutine:
	//   1. starts inside Start() (indirectly: Start/Stop lifecycle does not hang)
	//   2. exits cleanly when Stop() is called
	//   3. does NOT call DeleteOlderThan before the 6-hour ticker fires

	repo := &mockMetricsRepo{}
	log := &mockLogger{}
	mc := collector.NewMetricsCollector(repo, log, time.Hour)

	startAndStop(t, mc, 100*time.Millisecond)

	// The 6-hour cleanup ticker hasn't fired, so DeleteOlderThan should never be called.
	if repo.wasDeleteCalled() {
		t.Error("expected DeleteOlderThan NOT to be called (6h cleanup ticker hasn't fired yet)")
	}
}
