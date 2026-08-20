package main

import (
	"net/http/httptest"
	"testing"

	"eka-dev.cloud/sse-gateway/config"
	"eka-dev.cloud/sse-gateway/middleware"
	"github.com/golang-jwt/jwt/v5"
)

func generateTestJWT() string {
	secret := config.Config.SecretJwt
	if secret == "" {
		secret = "super-secret-jwt-key"
	}
	claims := middleware.Claims{
		UserId: 1,
		Type:   "access",
		Role:   "ADMIN",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestSSEEventsSuite(t *testing.T) {
	app := setupTestApp()
	validToken := generateTestJWT()

	t.Run("Health Check Endpoint 200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 200 {
			t.Fatalf("Expected 200 OK for /health, got %v", resp.StatusCode)
		}
	})

	t.Run("SSE Events - Missing Token 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/1.0/events?type=order_status", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 401 {
			t.Fatalf("Expected 401 Unauthorized, got %v", resp.StatusCode)
		}
	})

	t.Run("SSE Events - Invalid Event Type 400", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/1.0/events?token="+validToken+"&type=INVALID", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 400 {
			t.Fatalf("Expected 400 Bad Request, got %v", resp.StatusCode)
		}
	})

	t.Run("SSE Events - Invalid JWT Token 401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/1.0/events?token=invalid_token_str&type=order_status", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 401 {
			t.Fatalf("Expected 401 Unauthorized, got %v", resp.StatusCode)
		}
	})

	t.Run("Invalid Endpoint 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/invalid-route", nil)
		resp, _ := app.Test(req)
		if resp.StatusCode != 404 {
			t.Fatalf("Expected 404 Not Found, got %v", resp.StatusCode)
		}
	})
}
