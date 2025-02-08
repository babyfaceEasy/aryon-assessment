package storage

import (
	"context"
	"time"
)

type Connector struct {
	ID               string
	WorkspaceID      string
	DefaultChannelID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Storage interface {
	SaveConnector(context.Context, *Connector) (string, error)
	GetConnectorByID(context.Context, string) (*Connector, error)
	GetAllConnectors(context.Context) ([]*Connector, error)
	DeleteConnector(context.Context, string) error
}
