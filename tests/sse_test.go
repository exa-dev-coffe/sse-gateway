package tests

import (
	"fmt"
	"io"
	"testing"
)

func TestSSESuite(t *testing.T) {
	_, teardown := SetupTestRabbitMQ(t)
	defer teardown()

	app := SetupTestApp()
	validToken := GenerateTestJWT(1, "ADMIN")

	t.Run("GET /health - Health Check Endpoint 200", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/health", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 200 {
			t.Fatalf("Expected HTTP 200 OK for /health, got %v", resp.StatusCode)
		}
	})

	t.Run("GET /api/1.0/events - Real RabbitMQ SSE Stream 'order' 200 OK", func(t *testing.T) {
		url := fmt.Sprintf("/api/1.0/events?token=%s&type=order", validToken)
		resp, _ := ExecuteTestRequest(app, "GET", url, nil, "", 200)
		if resp != nil && resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK text/event-stream, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /api/1.0/events - Real RabbitMQ SSE Stream 'update_history_balance' 200 OK", func(t *testing.T) {
		url := fmt.Sprintf("/api/1.0/events?token=%s&type=update_history_balance", validToken)
		resp, _ := ExecuteTestRequest(app, "GET", url, nil, "", 200)
		if resp != nil && resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK text/event-stream, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /api/1.0/events - Missing Token 401", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/events?type=order_status", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 401 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 401 Unauthorized, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /api/1.0/events - Invalid Event Type 400", func(t *testing.T) {
		url := fmt.Sprintf("/api/1.0/events?token=%s&type=INVALID", validToken)
		resp, err := ExecuteTestRequest(app, "GET", url, nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 400 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 400 Bad Request, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /api/1.0/events - Invalid JWT Token 401", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/api/1.0/events?token=invalid_token_str&type=order_status", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 401 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 401 Unauthorized, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /invalid-route - Invalid Endpoint 404", func(t *testing.T) {
		resp, err := ExecuteTestRequest(app, "GET", "/invalid-route", nil, "")
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		if resp.StatusCode != 404 {
			t.Fatalf("Expected HTTP 404 Not Found, got %v", resp.StatusCode)
		}
	})
}
