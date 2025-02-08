package storage

import "database/sql"

type SqlStorage struct {
	db *sql.DB
}

func NewSqlStorage(db *sql.DB) *SqlStorage {
	return &SqlStorage{db: db}
}

func (s *SqlStorage) SaveConnector(connector *Connector) error {
	_, err := s.db.Exec("INSERT INTO connectors (workspace_id, default_channel_id, created_at, updated_at) VALUES ($1, $2, $3, $4)", connector.WorkspaceID, connect.DefaultChannelID, connector.CreatedAt, connector.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqlStorage) GetConnectorByID(connectorID string) (*Connector, error) {
	rows, err := s.db.Query("SELECT * FROM connectors WHERE id = $1", connectorID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	c := new(Connector)
	for rows.Next() {
		c, err = scanRowsIntoConnector(rows)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func (s *SqlStorage) GetAllConnectors() ([]*Connector, error) {
	return nil, nil
}

func (s *SqlStorage) DeleteConnector(ID string) error {
	return nil
}

func scanRowsIntoConnector(rows *sql.Rows) (*Connector, error) {
	connector := new(Connector)

	err := rows.Scan(
		&connector.ID,
		&connector.WorkspaceID,
		&connector.DefaultChannelID,
		&connector.CreatedAt,
		&connector.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return connector, nil
}
