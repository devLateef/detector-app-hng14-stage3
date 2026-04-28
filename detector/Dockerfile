# ---- Build Stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o detector-app .

# ---- Runtime Stage ----
FROM alpine:3.19

WORKDIR /app

# iptables is required for blocking IPs
RUN apk add --no-cache iptables

COPY --from=builder /app/detector-app .
COPY config.yaml .
COPY dashboard/static ./dashboard/static

EXPOSE 8081

ENTRYPOINT ["./detector-app"]
