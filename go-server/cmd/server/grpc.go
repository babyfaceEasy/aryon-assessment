package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/signal"
	"syscall"
	"time"

	"connector-recruitment/go-server/connectors/config"
	"connector-recruitment/go-server/connectors/db"
	"connector-recruitment/go-server/connectors/handler"
	"connector-recruitment/go-server/connectors/interceptors"
	"connector-recruitment/go-server/connectors/logger"
	"connector-recruitment/go-server/connectors/service"
	"connector-recruitment/go-server/connectors/storage"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type gRPCServer struct {
	db            *db.Service
	addr          string
	secretManager *secretsmanager.SecretsManager
	logger        logger.Logger
}

func NewGRPCServer(addr string, secretManager *secretsmanager.SecretsManager, logger logger.Logger, db *db.Service) *gRPCServer {
	return &gRPCServer{addr: ":" + addr, secretManager: secretManager, logger: logger, db: db}
}

func (s *gRPCServer) Run(env config.Env) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create a new gRPC server instance
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.LoggingUnaryInterceptor(s.logger)),
	)

	// health server
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Setup storage and register gRPC services
	storage := storage.NewSqlStorage(s.db.DBPool, s.secretManager, s.logger)
	connectorService := service.NewConnectorService(storage, s.secretManager, s.logger)
	handler.NewGrpcConnectorsService(grpcServer, connectorService, s.logger)

	s.logger.Info("Starting gRPC server", "addr", lis.Addr().String())

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErrCh := make(chan error, 1)
	go func() {
		err := grpcServer.Serve(lis)
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			serverErrCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("Shutdown signal received, gracefully stopping gRPC server")
		timer := time.AfterFunc(time.Duration(env.RPCGracefulShutdownTimeout)*time.Second, func() {
			s.logger.Warn("gRPC server couldn't stop gracefully in time. Doing force stop.")
			grpcServer.Stop()
		})
		defer timer.Stop()

		startTime := time.Now()
		grpcServer.GracefulStop()
		elapsed := time.Since(startTime)

		s.logger.Info("gRPC server gracefully stopped", "elapsed", elapsed)
	case err := <-serverErrCh:
		if err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
	}

	s.logger.Info("gRPC server has been gracefully stopped")
	return nil
}
