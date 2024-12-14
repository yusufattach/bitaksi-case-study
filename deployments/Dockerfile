# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and build dependencies
RUN apk add --no-cache git tzdata ca-certificates

# Copy only go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy only necessary source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY deployments/.env ./deployments/.env

# Build the application with security flags
ARG SERVICE_NAME
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/service ./cmd/${SERVICE_NAME}/main.go

# Final stage
FROM alpine:3.18

# Add non-root user
RUN adduser -D -g '' appuser && \
    apk add --no-cache tzdata ca-certificates

# Set timezone
ENV TZ=UTC

# Set working directory
WORKDIR /app

# Copy binary and config from builder
COPY --from=builder /app/bin/service .
COPY --from=builder /app/deployments/.env ./deployments/.env

# Set ownership to non-root user
RUN chown -R appuser:appuser /app

# Use non-root user
USER appuser

# Expose port (will be overridden by docker-compose)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Run the service
CMD ["./service"] 