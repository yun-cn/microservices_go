# Build
FROM golang:1.25-alpine AS builder

# Install git
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to strip debug info and reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/loan-engine \
    ./cmd/server/main.go

# Build goose for migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.1

# Stage 2: Runtime
FROM alpine:3.21

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/loan-engine /app/loan-engine

# Copy goose for running migrations in k8s
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy database migrations
COPY --from=builder /app/db/migrations /app/db/migrations

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose gRPC port
EXPOSE 50051

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/loan-engine", "--health-check"] || exit 1

# Run the application
CMD ["/app/loan-engine"]
