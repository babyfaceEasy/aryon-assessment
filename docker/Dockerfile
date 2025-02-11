ARG GO_VERSION=1.23

# Build stage
FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o grpcserver ./go-server/cmd/server

# Intermediate stage for certificates and grpc_health_probe
FROM alpine:latest AS certs
RUN apk add --no-cache ca-certificates wget && \
    wget -q -O /usr/local/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.11/grpc_health_probe-linux-amd64 && \
    chmod +x /usr/local/bin/grpc_health_probe

# Final stage using scratch
FROM scratch

# Copy CA certificates and the grpc_health_probe binary.
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=certs /usr/local/bin/grpc_health_probe /usr/local/bin/grpc_health_probe

# Copy the static binary.
COPY --from=builder /app/grpcserver /grpcserver
COPY  --from=builder /app/.env .

EXPOSE 50051

# Run binary directly.
CMD ["/grpcserver"]
