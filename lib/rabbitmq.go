package lib

import (
	"log"
	"sync"
	"time"

	"eka-dev.cloud/sse-gateway/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn     *amqp.Connection
	connOnce sync.Once
	mu       sync.Mutex
)

// GetConnection -> dapet koneksi dengan auto-retry
func GetConnection() *amqp.Connection {
	mu.Lock()
	defer mu.Unlock()

	if conn != nil && !conn.IsClosed() {
		return conn
	}

	// retry loop kalau gagal
	for {
		c, err := amqp.Dial(config.Config.RabbitmqUrl)
		if err != nil {
			log.Println("❌ Failed to connect to RabbitMQ, retrying in 5s:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		conn = c
		log.Println("✅ Connected to RabbitMQ")
		break
	}

	return conn
}

// GetChannel -> bikin channel baru (safe untuk goroutine)
func GetChannel() (*amqp.Channel, error) {
	c := GetConnection()
	return c.Channel()
}

func HealthCheck() error {
	c := GetConnection()
	if c.IsClosed() {
		return amqp.ErrClosed
	}
	return nil
}
