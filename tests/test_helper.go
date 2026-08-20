package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eka-dev.cloud/sse-gateway/config"
	"eka-dev.cloud/sse-gateway/middleware"
	"eka-dev.cloud/sse-gateway/modules/sse"
	"eka-dev.cloud/sse-gateway/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestRabbitMQ(t *testing.T) (string, func()) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3-alpine",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForLog("Server startup complete").WithStartupTimeout(45 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("Skipping RabbitMQ container integration test: Docker/Testcontainers unavailable: %v", err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5672")
	amqpURL := fmt.Sprintf("amqp://guest:guest@%s:%s/", host, port.Port())

	config.Config.RabbitmqUrl = amqpURL

	teardown := func() {
		_ = container.Terminate(ctx)
	}

	return amqpURL, teardown
}

func SetupTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(response.Success("OK", nil))
	})

	// Real SSE Module Handler (No Mocks!)
	sse.NewHandler(app)

	app.All("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(response.NotFound("Route not found", nil))
	})

	return app
}

func GenerateTestJWT(userId int64, role string) string {
	secret := config.Config.SecretJwt
	if secret == "" {
		secret = "super-secret-jwt-key"
		config.Config.SecretJwt = secret
	}
	claims := middleware.Claims{
		UserId: userId,
		Type:   "ACCESS",
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func ExecuteTestRequest(app *fiber.App, method, url string, body []byte, token string, msTimeout ...int) (*http.Response, error) {
	var req *http.Request
	if len(body) > 0 {
		req = httptest.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	timeout := 1000
	if len(msTimeout) > 0 {
		timeout = msTimeout[0]
	}

	return app.Test(req, timeout)
}
