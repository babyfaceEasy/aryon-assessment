# Dockerfile.prod

# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# package is in go-server/cmd/server/.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o grpcserver ./go-server/cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/grpcserver .

EXPOSE 50051

# Start the gRPC server.
CMD ["./grpcserver"]
