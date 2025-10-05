# Stage 1: Build Go binary
FROM golang:1.24.5-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o sse-gateway .

# Stage 2: Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/sse-gateway .
COPY db/migrations ./db/migrations

EXPOSE 8003
CMD ["./sse-gateway"]
