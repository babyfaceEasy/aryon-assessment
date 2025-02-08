package main

import (
	"log"
	"log/slog"
	"net"

	"connector-recruitment/go-server/connectors/db"
	"connector-recruitment/go-server/connectors/handler"
	"connector-recruitment/go-server/connectors/service"
	"connector-recruitment/go-server/connectors/storage"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jackc/pgx/v5"
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
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// get db
	db, err := db.NewPGXStorage(pgx.ConnConfig{}) // TODO: fill these from env variables
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	storage := storage.NewSqlStorage(db)

	// register our grpc services
	connectorService := service.NewConnectorService(storage, s.secretManager)
	handler.NewGrpcConnectorsService(grpcServer, connectorService)

	slog.Info("Starting Slack Connector gRPC server on ", "info", s.addr)

	return grpcServer.Serve(lis)
}
