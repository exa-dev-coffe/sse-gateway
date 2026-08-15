# SSE Gateway (sse-gateway)

SSE Gateway is a real-time notification microservice built with **Go** and **Fiber**. It maintains persistent Server-Sent Events (SSE) connections with the frontend to push real-time updates.

## 🚀 Technologies

*   **Language**: Go 1.25
*   **Framework**: Fiber v2 (with SSE support)
*   **Message Broker**: RabbitMQ (`amqp091-go`)
*   **Observability**: OpenTelemetry
*   **Logging**: `log/slog`

## 📦 Features

*   **Real-time Notifications**: Pushes events to web clients efficiently using SSE.
*   **Event Subscription**: Consumes specific events from RabbitMQ and broadcasts them to connected users.
*   **Lightweight**: Designed specifically for high-concurrency, long-lived connections.

## ⚙️ Environment Variables

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

## 🚀 How to Run

1.  **Download Dependencies:**
    ```bash
    go mod download
    ```

2.  **Run Locally:**
    ```bash
    go run main.go
    ```
