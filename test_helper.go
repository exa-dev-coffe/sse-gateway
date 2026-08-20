package main

import (
	"context"
	"testing"
	"time"

	"eka-dev.cloud/sse-gateway/middleware"
	"eka-dev.cloud/sse-gateway/modules/sse"
	"eka-dev.cloud/sse-gateway/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestRedis(t *testing.T) (string, func()) {
	ctx := context.Background()
	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Skipf("Skipping integration test: Docker/Testcontainers unavailable: %v", err)
	}

	endpoint, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get redis endpoint: %v", err)
	}

	teardown := func() {
		_ = redisContainer.Terminate(ctx)
	}

	return endpoint, teardown
}

func setupTestApp() *fiber.App {
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
