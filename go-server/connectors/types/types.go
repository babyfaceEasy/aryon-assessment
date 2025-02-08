package types

import (
	"context"

	pb "connector-recruitment/go-server/connectors/genproto"
)

type ConnectorService interface {
	GetConnector(context.Context, string) (*pb.Connector, error)
	CreateConnector(context.Context, string, *pb.Connector) error
	GetConnectors(context.Context) []*pb.Connector
	DeleteConnector(context.Context, string) error
	SendMessage(context.Context, string, string) error
}
