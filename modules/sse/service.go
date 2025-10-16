package sse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"

	"eka-dev.cloud/sse-gateway/lib"
	"eka-dev.cloud/sse-gateway/middleware"
	"eka-dev.cloud/sse-gateway/utils/response"
	"github.com/gofiber/fiber/v2"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Service interface {
	Events(c *fiber.Ctx, claims *middleware.Claims, eventType EventType) error
	sendEventUpdateHistoryBalance(w *bufio.Writer, msg amqp.Delivery) error
	heartbeat(w *bufio.Writer) error
	handleEventUpdateHistoryBalance(c *fiber.Ctx, userId int64, ch *amqp.Channel) error
}

type eventService struct {
}

func NewService() Service {
	return &eventService{}
}

func (s *eventService) sendEventUpdateHistoryBalance(w *bufio.Writer, msg amqp.Delivery) error {
	_, err := fmt.Fprintf(w, "data: %s\n\n", msg.Body)
	if err != nil {
		log.Error("Failed to write message:", err)
		return err
	}
	if err := w.Flush(); err != nil {
		log.Error("Failed to flush data:", err)
		return err
	}
	return nil
}

func (s *eventService) heartbeat(w *bufio.Writer) error {
	_, err := fmt.Fprintf(w, ": heartbeat\n\n")
	if err != nil {
		log.Error("Failed to write heartbeat:", err)
		return err
	}
	if err := w.Flush(); err != nil {
		log.Error("Failed to flush heartbeat:", err)
		return err
	}
	return nil
}

func (s *eventService) handleEventUpdateHistoryBalance(c *fiber.Ctx, userId int64, ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		"balance.history.updated", // name
		"fanout",                  // type
		false,                     // durable
		true,                      // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)

	if err != nil {
		log.Error("Failed to declare exchange:", err)
		return response.InternalServerError("Failed to declare exchange", nil)
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Error("Failed to declare queue:", err)
		return response.InternalServerError("Failed to declare queue", nil)
	}

	err = ch.QueueBind(q.Name, "", "balance.history.updated", false, nil)
	if err != nil {
		log.Error("Failed to bind queue:", err)
		return response.InternalServerError("Failed to bind queue", nil)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Error("Failed to register consumer:", err)
		return response.InternalServerError("Failed to consume message", nil)
	}

	ctx := c.Context()

	done := make(chan struct{}) // pakai struct{} lebih ringan

	go func() {
		<-ctx.Done()
		log.Info("Received done signal, shutting down")
		close(done) // aman, bisa diterima di select
	}()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		defer func(ch *amqp.Channel) {
			log.Info("Closing SSE connection for client")

			err := ch.Close()
			if err != nil {
				log.Error("Failed to close channel:", err)
			}
			// di sini bisa close queue atau channel jika perlu
		}(ch)

		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					log.Error("Message channel closed")
					return
				}
				baseMessage, err := extractBaseBodyMessage(msg)
				if err != nil {
					log.Error("Failed to extract message:", err)
					continue
				}
				if baseMessage.UserId == userId {
					err = s.sendEventUpdateHistoryBalance(w, msg)
					if err != nil {
						log.Error("Failed to send event:", err)
						return
					}
				}
			case <-ticker.C:
				err := s.heartbeat(w)
				if err != nil {
					log.Error("Failed to send heartbeat:", err)
					return
				}
			case <-done:
				log.Info("Client disconnected")
				return
			}
		}
	})
	return nil
}

func (s *eventService) Events(c *fiber.Ctx, claims *middleware.Claims, eventType EventType) error {
	// Set headers for SSE
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	ch, err := lib.GetChannel()
	if err != nil {
		log.Error("Failed to open channel:", err)
		return response.InternalServerError("Failed to open channel", nil)
	}

	switch eventType {
	case EventUpdateHistoryBalance:
		return s.handleEventUpdateHistoryBalance(c, claims.UserId, ch)
	default:
		err := ch.Close()
		if err != nil {
			log.Error("Failed to close channel:", err)
			return err
		}
		return response.BadRequest("Invalid event type", nil)
	}
}

func extractBaseBodyMessage(msg amqp.Delivery) (*BaseBodyMessage, error) {
	var message BaseBodyMessage
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		log.Error("Failed to unmarshal message:", err)
		return nil, response.InternalServerError("Internal Server Error", nil)
	}
	return &message, nil
}
