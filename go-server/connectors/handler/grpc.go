package handler

import (
	"context"
	"errors"
	"fmt"

	"connector-recruitment/go-server/connectors/errs"
	pb "connector-recruitment/go-server/connectors/genproto"
	"connector-recruitment/go-server/connectors/logger"
	"connector-recruitment/go-server/connectors/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errdetails "google.golang.org/genproto/googleapis/rpc/errdetails"
)

// ConnectorsGrpcHandler implements pb.ConnectorServiceServer.
type ConnectorsGrpcHandler struct {
	logger           logger.Logger
	connectorService types.ConnectorService
	pb.UnimplementedConnectorServiceServer
}

// NewGrpcConnectorsService registers a new gRPC handler with the provided server.
func NewGrpcConnectorsService(grpcServer *grpc.Server, connectorService types.ConnectorService, logger logger.Logger) {
	handler := &ConnectorsGrpcHandler{
		logger:           logger,
		connectorService: connectorService,
	}
	pb.RegisterConnectorServiceServer(grpcServer, handler)
}

func (h *ConnectorsGrpcHandler) GetConnector(ctx context.Context, req *pb.GetConnectorRequest) (*pb.GetConnectorResponse, error) {
	// Validate required field.
	if req.ConnectorId == "" {
		br := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "connectorId",
					Description: "missing required field: connectorId",
				},
			},
		}
		st := status.New(codes.InvalidArgument, "missing required field: connectorId")
		stWithDetails, err := st.WithDetails(br)
		if err != nil {
			h.logger.Error("GetConnector: failed to attach error details", "error", err)
			return nil, st.Err()
		}
		return nil, stWithDetails.Err()
	}

	conn, err := h.connectorService.GetConnector(ctx, req.ConnectorId)
	if err != nil {
		if errors.Is(err, errs.ErrConnectorNotFound) {
			h.logger.Warn("GetConnector not found", "id", req.ConnectorId)
			info := &errdetails.ErrorInfo{
				Reason:   "ConnectorNotFound",
				Domain:   "connectors.service",
				Metadata: map[string]string{"connectorId": req.ConnectorId},
			}
			st := status.New(codes.NotFound, fmt.Sprintf("connector with id %s not found", req.ConnectorId))
			stWithDetails, err := st.WithDetails(info)
			if err != nil {
				h.logger.Error("GetConnector: failed to attach error details", "error", err)
				return nil, st.Err()
			}
			return nil, stWithDetails.Err()
		}

		h.logger.Error("GetConnector internal error", "id", req.ConnectorId, "msg", "failed to fetch connector data", "err", err)
		info := &errdetails.ErrorInfo{
			Reason:   "InternalError",
			Domain:   "connectors.service",
			Metadata: map[string]string{"connectorId": req.ConnectorId},
		}
		st := status.New(codes.Internal, "internal server error: failed to fetch connector data")
		stWithDetails, detailsErr := st.WithDetails(info)
		if detailsErr != nil {
			h.logger.Error("GetConnector: failed to attach internal error details", "error", detailsErr)
			return nil, st.Err()
		}
		return nil, stWithDetails.Err()
	}

	res := &pb.GetConnectorResponse{
		Connector: conn,
	}
	return res, nil
}

func (h *ConnectorsGrpcHandler) CreateConnector(ctx context.Context, req *pb.CreateConnectorRequest) (*pb.CreateConnectorResponse, error) {
	// Validate inputs and build up a list of field violations.
	var violations []*errdetails.BadRequest_FieldViolation
	if len(req.SlackToken) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "slackToken",
			Description: "slack token is required",
		})
	}
	if len(req.TenantId) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "tenantId",
			Description: "tenant id is required",
		})
	}
	if len(req.DefaultChannelId) == 0 {
		violations = append(violations, &errdetails.BadRequest_FieldViolation{
			Field:       "defaultChannelId",
			Description: "default channel id is required",
		})
	}
	if len(violations) > 0 {
		br := &errdetails.BadRequest{FieldViolations: violations}
		st := status.New(codes.InvalidArgument, "invalid input parameters")
		stWithDetails, err := st.WithDetails(br)
		if err != nil {
			h.logger.Error("CreateConnector: failed to attach bad request details", "error", err)
			return nil, st.Err()
		}
		return nil, stWithDetails.Err()
	}

	err := h.connectorService.CreateConnector(ctx, req.SlackToken, &pb.Connector{
		TenantId:         req.TenantId,
		DefaultChannelId: req.DefaultChannelId,
	})
	if err != nil {
		h.logger.Error("CreateConnector internal error", "err", err.Error())
		info := &errdetails.ErrorInfo{
			Reason:   "InternalError",
			Domain:   "connectors.service",
			Metadata: map[string]string{},
		}
		st := status.New(codes.Internal, "internal server error: failed to create connector")
		stWithDetails, detailsErr := st.WithDetails(info)
		if detailsErr != nil {
			h.logger.Error("CreateConnector: failed to attach internal error details", "error", detailsErr)
			return nil, st.Err()
		}
		return nil, stWithDetails.Err()
	}

	return &pb.CreateConnectorResponse{}, nil
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
			h.logger.Warn("DeleteConnector not found", "id", req.ConnectorId)
			info := &errdetails.ErrorInfo{
				Reason:   "ConnectorNotFound",
				Domain:   "connectors.service",
				Metadata: map[string]string{"connectorId": req.ConnectorId},
			}
			st := status.New(codes.NotFound, "connector not found")
			stWithDetails, detailsErr := st.WithDetails(info)
			if detailsErr != nil {
				h.logger.Error("DeleteConnector: failed to attach error details", "error", detailsErr)
				return nil, st.Err()
			}
			return nil, stWithDetails.Err()
		}

		h.logger.Error("failed to delete connector", "id", req.ConnectorId, "err", err)
		info := &errdetails.ErrorInfo{
			Reason:   "InternalError",
			Domain:   "connectors.service",
			Metadata: map[string]string{"connectorId": req.ConnectorId},
		}
		st := status.New(codes.Internal, "internal server error: failed to delete connector")
		stWithDetails, detailsErr := st.WithDetails(info)
		if detailsErr != nil {
			h.logger.Error("DeleteConnector: failed to attach internal error details", "error", detailsErr)
			return nil, st.Err()
		}
		return nil, stWithDetails.Err()
	}

	return &pb.DeleteConnectorResponse{}, nil
}
