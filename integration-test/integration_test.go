package integration_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	// Base settings
	host     = "app"
	attempts = 20

	// Attempts connection
	httpURL        = "http://" + host + ":8080"
	healthPath     = httpURL + "/healthz"
	requestTimeout = 5 * time.Second
)

var errHealthCheck = fmt.Errorf("url %s is not available", healthPath)

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
	for attempts > 0 {
		statusCode, err := getHealthCheck(healthPath)
		if err != nil {
			return err
		}

		if statusCode == http.StatusOK {
			return nil
		}

		fmt.Printf("Integration tests: url %s is not available, attempts left: %d\n", healthPath, attempts)

		time.Sleep(time.Second)
		attempts--
	}

	return errHealthCheck
}

func TestMain(m *testing.M) {
	err := healthCheck(attempts)
	if err != nil {
		fmt.Printf("Integration tests: httpURL %s is not available: %s\n", httpURL, err)
		os.Exit(1)
	}

	fmt.Printf("Integration tests: httpURL %s is available\n", httpURL)

	code := m.Run()
	os.Exit(code)
}

// HTTP GET: /healthz
func TestHealthCheck(t *testing.T) {
	statusCode, err := getHealthCheck(healthPath)
	if err != nil {
		t.Fatalf("Failed to check health: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, statusCode)
	}
}
