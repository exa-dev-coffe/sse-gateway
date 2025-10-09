package sse

import (
	"eka-dev.cloud/sse-gateway/middleware"
	"eka-dev.cloud/sse-gateway/utils/enum"
	"eka-dev.cloud/sse-gateway/utils/response"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type Handler interface {
	Events(c *fiber.Ctx) error
}

type handler struct {
	service Service
	db      *sqlx.DB
}

func NewHandler(app *fiber.App) Handler {
	service := NewService()
	handler := &handler{service: service}

	// mapping routes
	routes := app.Group("/api/1.0/events")
	routes.Get("", handler.Events)

	return handler
}

func (h *handler) Events(c *fiber.Ctx) error {
	claims, err := middleware.ValidateTokenQuery(c)
	if err != nil {
		return err
	}

	typeQuery := c.Query("type")
	val, ok := enum.ParseGeneric[EventType](typeQuery, eventTypeMap)
	if !ok {
		return response.BadRequest("Invalid event type", nil)
	}
	return h.service.Events(c, claims, val)
}

//package main
//
//import (
//"fmt"
//"log"
//"time"
//
//"github.com/gofiber/fiber/v2"
//"github.com/streadway/amqp"
//)
//
//func main() {
//	// Connect ke RabbitMQ
//	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
//	if err != nil {
//		log.Fatal("Failed to connect to RabbitMQ:", err)
//	}
//	defer conn.Close()
//
//	ch, err := conn.Channel()
//	if err != nil {
//		log.Fatal("Failed to open channel:", err)
//	}
//	defer ch.Close()
//
//	// Pastikan exchange ada
//	err = ch.ExchangeDeclare(
//		"events",
//		"fanout",
//		true,
//		false,
//		false,
//		false,
//		nil,
//	)
//	if err != nil {
//		log.Fatal("Failed to declare exchange:", err)
//	}
//
//	app := fiber.New()
//
//	// Endpoint SSE
//	app.Get("/sse", func(c *fiber.Ctx) error {
//		// Set SSE headers
//		c.Set("Content-Type", "text/event-stream")
//		c.Set("Cache-Control", "no-cache")
//		c.Set("Connection", "keep-alive")
//
//		// Buat temporary queue per-client
//		q, err := ch.QueueDeclare(
//			"",    // auto-name
//			false, // durable
//			true,  // auto-delete
//			true,  // exclusive
//			false,
//			nil,
//		)
//		if err != nil {
//			return c.Status(500).SendString("Queue declare failed")
//		}
//
//		// Bind ke exchange
//		err = ch.QueueBind(q.Name, "", "events", false, nil)
//		if err != nil {
//			return c.Status(500).SendString("Queue bind failed")
//		}
//
//		// Consume message dari queue
//		msgs, err := ch.Consume(q.Name, "", true, true, false, false, nil)
//		if err != nil {
//			return c.Status(500).SendString("Consume failed")
//		}
//
//		// Streaming ke client
//		return c.Context().SetBodyStreamWriter(func(w *fiber.Writer) {
//			ticker := time.NewTicker(15 * time.Second)
//			defer ticker.Stop()
//
//			for {
//				select {
//				case msg := <-msgs:
//					fmt.Fprintf(w, "data: %s\n\n", msg.Body)
//					w.Flush()
//
//				case <-ticker.C:
//					// Heartbeat (komentar SSE)
//					fmt.Fprintf(w, ": ping\n\n")
//					w.Flush()
//
//				case <-c.Context().Done():
//					log.Println("Client disconnected")
//					return
//				}
//			}
//		})
//	})
//
//	log.Println("SSE server running on :8080")
//	log.Fatal(app.Listen(":8080"))
//}
