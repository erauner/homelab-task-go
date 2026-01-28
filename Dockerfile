# Multi-stage build for minimal production image
# Build stage: compile Go binary
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install git for module downloads
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o taskkit ./cmd/taskkit

# Runtime stage: minimal image
FROM alpine:3.19 AS runtime

WORKDIR /app

# Create non-root user for security
RUN adduser -D -u 1000 taskuser

# Copy binary from builder
COPY --from=builder --chown=taskuser:taskuser /build/taskkit /app/taskkit

# Copy workflow definitions
COPY --chown=taskuser:taskuser workflows/ /app/workflows/

USER taskuser

# Default entrypoint
ENTRYPOINT ["/app/taskkit"]
CMD ["--help"]
