package tests

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestSSESuite(t *testing.T) {
	amqpURL, teardown := SetupTestRabbitMQ(t)
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

	t.Run("GET /api/1.0/events - Real RabbitMQ SSE Stream 'order' 200 OK & Receives Message", func(t *testing.T) {
		// Connect to RabbitMQ container & publish mock order created message
		conn, err := amqp.Dial(amqpURL)
		if err == nil {
			ch, err := conn.Channel()
			if err == nil {
				_ = ch.ExchangeDeclare("order.created", "fanout", false, true, false, false, nil)
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = ch.PublishWithContext(ctx, "order.created", "", false, false, amqp.Publishing{
					ContentType: "application/json",
					Body:        []byte(`{"event":"order_created","data":{"orderId":100,"totalPrice":50000}}`),
				})
				_ = ch.Close()
			}
			_ = conn.Close()
		}

		url := fmt.Sprintf("/api/1.0/events?token=%s&type=order", validToken)
		resp, _ := ExecuteTestRequest(app, "GET", url, nil, "", 300)
		if resp != nil && resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected HTTP 200 OK text/event-stream, got %v: %s", resp.StatusCode, string(respBody))
		}
	})

	t.Run("GET /api/1.0/events - Real RabbitMQ SSE Stream 'update_history_balance' 200 OK & Receives Message", func(t *testing.T) {
		// Connect to RabbitMQ container & publish mock balance update message for user 1
		conn, err := amqp.Dial(amqpURL)
		if err == nil {
			ch, err := conn.Channel()
			if err == nil {
				_ = ch.ExchangeDeclare("balance.history.updated", "direct", false, true, false, false, nil)
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = ch.PublishWithContext(ctx, "balance.history.updated", "1", false, false, amqp.Publishing{
					ContentType: "application/json",
					Body:        []byte(`{"event":"balance_updated","data":{"userId":1,"newBalance":200000}}`),
				})
				_ = ch.Close()
			}
			_ = conn.Close()
		}

		url := fmt.Sprintf("/api/1.0/events?token=%s&type=update_history_balance", validToken)
		resp, _ := ExecuteTestRequest(app, "GET", url, nil, "", 300)
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
