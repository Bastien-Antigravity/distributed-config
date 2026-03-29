# === BUILD STAGE ===
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev ca-certificates tzdata

WORKDIR /distributed-config

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /distributed-config/distributed-config ./cmd/main

# === RUNTIME STAGE ===
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /distributed-config

# Copy the binary from the build stage
COPY --from=builder /distributed-config/distributed-config /distributed-config/distributed-config

# Set the entrypoint
ENTRYPOINT ["/distributed-config/distributed-config"]
