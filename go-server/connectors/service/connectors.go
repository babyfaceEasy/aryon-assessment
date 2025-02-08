package service

import (
	"context"
	"errors"

	pb "connector-recruitment/go-server/connectors/genproto"
	"connector-recruitment/go-server/connectors/storage"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ConnectorService struct {
	storage  storage.Storage
	smClient *secretsmanager.SecretsManager
}

func NewConnectorService(storage storage.Storage, smClient *secretsmanager.SecretsManager) *ConnectorService {
	return &ConnectorService{
		storage:  storage,
		smClient: smClient,
	}
}

func (s *ConnectorService) GetConnector(ctx context.Context, ID string) (*pb.Connector, error) {
	connectorRow, err := s.storage.GetConnectorByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	return &pb.Connector{
		Id:               connectorRow.ID,
		TenantId:         connectorRow.WorkspaceID,
		DefaultChannelId: connectorRow.DefaultChannelID,
		CreatedAt:        timestamppb.New(connectorRow.CreatedAt),
		UpdatedAt:        timestamppb.New(connectorRow.UpdatedAt),
	}, nil
}

func (s *ConnectorService) CreateConnector(ctx context.Context, slackToken string, connector *pb.Connector) error {
	_, err := s.storage.SaveConnector(ctx, &storage.Connector{
		WorkspaceID:      connector.TenantId,
		DefaultChannelID: connector.DefaultChannelId,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *ConnectorService) GetConnectors(ctx context.Context) []*pb.Connector {
	conns, err := s.storage.GetAllConnectors(ctx)
	if err != nil {
		return nil
	}

	var result []*pb.Connector

	for _, conn := range conns {
		pbConnector := &pb.Connector{
			Id:               conn.ID,
			TenantId:         conn.WorkspaceID,
			DefaultChannelId: conn.DefaultChannelID,
			CreatedAt:        timestamppb.New(conn.CreatedAt),
			UpdatedAt:        timestamppb.New(conn.UpdatedAt),
		}

		result = append(result, pbConnector)
	}

	return result
}

func (s *ConnectorService) DeleteConnector(ctx context.Context, ID string) error {
	err := s.storage.DeleteConnector(ctx, ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ConnectorService) SendMessage(ctx context.Context, connectorID string, message string) error {
	/*
		Get the default channel from the repository and get the slack token from secret manager
		call the slack client to push the message to the default channel
	*/

	return errors.New("not implemented yet")
}
