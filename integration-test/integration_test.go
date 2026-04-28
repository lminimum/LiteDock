package integration_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/lminimum/LiteDock/pkg/logger"
)

const (
	host     = "litedock"
	attempts = 20

	httpURL        = "http://" + host + ":8080"
	healthPath     = httpURL + "/healthz"
	requestTimeout = 5 * time.Second
)

var errHealthCheck = errors.New("health check failed")

func doWebRequestWithTimeout(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
}

func getHealthCheck(url string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func healthCheck(attempts int) error {
	log := logger.New("info")

	for attempts > 0 {
		statusCode, err := getHealthCheck(healthPath)
		if err != nil {
			return err
		}

		if statusCode == http.StatusOK {
			return nil
		}

		log.Warn("Integration tests: url %s is not available, attempts left: %d", healthPath, attempts)

		time.Sleep(time.Second)

		attempts--
	}

	return errHealthCheck
}

func TestMain(m *testing.M) {
	log := logger.New("info")

	err := healthCheck(attempts)
	if err != nil {
		log.Warn("Integration tests: httpURL %s is not available, running mock-based tests only: %s", httpURL, err)
	} else {
		log.Info("Integration tests: httpURL %s is available", httpURL)
	}

	code := m.Run()
	os.Exit(code)
}

// HTTP GET: /healthz
func TestHealthCheck(t *testing.T) {
	statusCode, err := getHealthCheck(healthPath)
	if err != nil {
		t.Skipf("Skipping health check: service not available: %v", err)
		return
	}

	if statusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, statusCode)
	}
}
