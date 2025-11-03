package main

import (
	"log"

	"eka-dev.cloud/sse-gateway/config"
	"eka-dev.cloud/sse-gateway/lib"
	_ "eka-dev.cloud/sse-gateway/lib"
	"eka-dev.cloud/sse-gateway/middleware"
	"eka-dev.cloud/sse-gateway/modules/sse"
	"eka-dev.cloud/sse-gateway/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load env
	initiator()

}

func initiator() {
	// Initialize the fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173," + config.Config.AllowedOrigin,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	fiberApp.Use(logger.New(logger.Config{
		Format:     "[${time}] ${ip} ${method} ${path} - ${status} (${latency})\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Jakarta",
	}))

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		_, err := lib.GetChannel()
		if err != nil {
			log.Println("RabbitMQ connection failed:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(response.InternalServerError("RabbitMQ connection error", nil))
		}
		return c.Status(fiber.StatusOK).JSON(response.Success("OK", nil))
	})

	// Initialize routes
	// SSE
	sse.NewHandler(fiberApp)

	fiberApp.All("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(response.NotFound("Route not found", nil))
	})

	err := fiberApp.Listen(config.Config.Port)
	if err != nil {
		log.Fatalln("Failed to start server:", err)
		return
	}
}
