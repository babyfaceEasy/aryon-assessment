package handler

import (
	"context"
	"errors"
	"log/slog"

	"connector-recruitment/go-server/connectors/errs"
	pb "connector-recruitment/go-server/connectors/genproto"
	"connector-recruitment/go-server/connectors/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConnectorsGrpcHandler struct {
	connectorService types.ConnectorService
	pb.UnimplementedConnectorServiceServer
}

func NewGrpcConnectorsService(grpc *grpc.Server, connectorService types.ConnectorService) {
	gRPCHandler := &ConnectorsGrpcHandler{
		connectorService: connectorService,
	}

	// register the ConnectorServiceServer
	pb.RegisterConnectorServiceServer(grpc, gRPCHandler)
}

func (h *ConnectorsGrpcHandler) GetConnector(ctx context.Context, req *pb.GetConnectorRequest) (*pb.GetConnectorResponse, error) {
	if req.ConnectorId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "missing required field: connectorId")
	}
	conn, err := h.connectorService.GetConnector(ctx, req.ConnectorId)
	if err != nil {

		if errors.Is(err, errs.ErrConnectorNotFound) {
			slog.Warn("GetConnector not found", "id", req.ConnectorId)
			return nil, status.Errorf(codes.NotFound, "connector with id %s not found", req.ConnectorId)
		}

		slog.Error("GetConnector internal error", "id", req.ConnectorId, "error", "failed to fetch connector data")
		return nil, status.Errorf(codes.Internal, "internal server error: failed to fetch connector data")
	}

	res := &pb.GetConnectorResponse{
		Connector: conn,
	}

	return res, nil
}

func (h *ConnectorsGrpcHandler) CreateConnector(ctx context.Context, req *pb.CreateConnectorRequest) (*pb.CreateConnectorResponse, error) {
	// validation
	if len(req.SlackToken) <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "slack token is required")
	}

	if len(req.TenantId) <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "tenant id is required")
	}

	if len(req.DefaultChannelId) <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "default channel id is required")
	}

	err := h.connectorService.CreateConnector(ctx, req.SlackToken, &pb.Connector{
		TenantId:         req.TenantId,
		DefaultChannelId: req.DefaultChannelId,
	})
	if err != nil {
		slog.Error("CreateConnector internal error", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "internal server error: failed to create connector")
	}

	return nil, nil
}

func (h *ConnectorsGrpcHandler) GetConnectors(ctx context.Context, req *pb.GetConnectorsRequest) (*pb.GetConnectorsResponse, error) {
	conns := h.connectorService.GetConnectors(ctx)
	res := &pb.GetConnectorsResponse{
		Connectors: conns,
	}

	return res, nil
}

func (h *ConnectorsGrpcHandler) DeleteConnector(ctx context.Context, req *pb.DeleteConnectorRequest) (*pb.DeleteConnectorResponse, error) {
	err := h.connectorService.DeleteConnector(ctx, req.ConnectorId)
	if err != nil {
		if errors.Is(err, errs.ErrConnectorNotFound) {
			return nil, nil
		}

		slog.Error("failed to delete connector", "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error: failed to delete connector")
	}

	return &pb.DeleteConnectorResponse{}, nil
}
