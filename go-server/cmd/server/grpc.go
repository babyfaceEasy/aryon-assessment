package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	"connector-recruitment/go-server/connectors/configs"
	"connector-recruitment/go-server/connectors/handler"
	"connector-recruitment/go-server/connectors/service"
	"connector-recruitment/go-server/connectors/storage"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type gRPCServer struct {
	addr          string
	secretManager *secretsmanager.SecretsManager
}

func NewGRPCServer(addr string, secretManager *secretsmanager.SecretsManager) *gRPCServer {
	return &gRPCServer{addr: addr, secretManager: secretManager}
}

func (s *gRPCServer) Run() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create a new gRPC server instance
	grpcServer := grpc.NewServer()

	// Create a connection pool
	pool, err := pgxpool.New(context.Background(), configs.Envs.DBUrl)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	defer pool.Close()

	// Setup storage and register gRPC services
	storage := storage.NewSqlStorage(pool, s.secretManager)
	connectorService := service.NewConnectorService(storage, s.secretManager)
	handler.NewGrpcConnectorsService(grpcServer, connectorService)

	slog.Info("Starting Slack Connector gRPC server on ", "info", s.addr)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErrCh := make(chan error, 1)
	go func() {
		// Serve will block until the server is stopped
		err := grpcServer.Serve(lis)
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serverErrCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("Shutdown signal received, gracefully stopping gRPC server")
		timer := time.AfterFunc(time.Duration(configs.Envs.GRPCGracefulShutdownTimeout)*time.Second, func() {
			slog.Warn("gRPC server couldn't stop gracefully in time. Doing force stop.")
			grpcServer.Stop()
		})
		defer timer.Stop()

		startTime := time.Now()
		grpcServer.GracefulStop()
		elapsed := time.Since(startTime)

		slog.Info("gRPC server gracefully stopped", "elapsed", elapsed)
	case err := <-serverErrCh:
		if err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
	}

	slog.Info("gRPC server has been gracefully stopped")
	return nil
}
