package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"

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

	grpcServer := grpc.NewServer()

	// Create a connection pool
	pool, err := pgxpool.New(context.Background(), configs.Envs.DBUrl)
	if err != nil {
		return fmt.Errorf("Unable to connect to database: %w", err)
	}
	defer pool.Close()

	// create a new storage
	storage := storage.NewSqlStorage(pool, s.secretManager)

	// register our grpc services
	connectorService := service.NewConnectorService(storage, s.secretManager)
	handler.NewGrpcConnectorsService(grpcServer, connectorService)

	slog.Info("Starting Slack Connector gRPC server on ", "info", s.addr)

	return grpcServer.Serve(lis)
}
