# ---- Build Stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o detector-app .

# ---- Runtime Stage ----
FROM alpine:3.19

WORKDIR /app

# Install iptables (required for blocking IPs)
RUN apk add --no-cache iptables

# Copy binary from builder
COPY --from=builder /app/detector-app .

# Copy config and static assets
COPY config.yaml .
COPY dashboard/static ./dashboard/static

# Expose dashboard port
EXPOSE 8081

ENTRYPOINT ["./detector-app"]
