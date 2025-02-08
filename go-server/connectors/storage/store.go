package storage

import "time"

type Connector struct {
	ID               string
	WorkspaceID      string
	DefaultChannelID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Storage interface {
	SaveConnector(tenant *Connector) error
	GetConnectorByID(ID string) (*Connector, error)
	GetAllConnectors() ([]*Connector, error)
	DeleteConnector(ID string) error
}
