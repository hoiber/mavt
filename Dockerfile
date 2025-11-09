# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o mavt ./cmd/mavt

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/mavt .

# Create data directory
RUN mkdir -p /app/data

# Use non-root user
RUN addgroup -g 1000 mavt && \
    adduser -D -u 1000 -G mavt mavt && \
    chown -R mavt:mavt /app

USER mavt

# Set default environment variables
ENV MAVT_DATA_DIR=/app/data
ENV MAVT_CHECK_INTERVAL=1h
ENV MAVT_LOG_LEVEL=info

# Volume for persistent data
VOLUME ["/app/data"]

# Default command: run in daemon mode
ENTRYPOINT ["./mavt"]
CMD ["-daemon"]
