package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type SqlStorage struct {
	db *sql.DB
}

func NewSqlStorage(db *sql.DB) *SqlStorage {
	return &SqlStorage{db: db}
}

func (s *SqlStorage) SaveConnector(ctx context.Context, connector *Connector) (string, error) {
	var connectorID string
	query := `
		INSERT INTO connectors (workspace_id, default_channel_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`

	err := s.db.QueryRowContext(ctx, query,
		connector.WorkspaceID,
		connector.DefaultChannelID,
		connector.CreatedAt,
		connector.UpdatedAt,
	).Scan(&connectorID)
	if err != nil {
		return "", fmt.Errorf("failed to save connector: %w", err)
	}

	return connectorID, nil
}

func (s *SqlStorage) GetConnectorByID(ctx context.Context, connectorID string) (*Connector, error) {
	c := &Connector{}

	query := "SELECT id, workspace_id, default_channel_id, created_at, updated_at FROM connectors WHERE id = $1"
	err := s.db.QueryRowContext(ctx, query, connectorID).Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.DefaultChannelID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("connector with ID %s not found: %w", connectorID, err)
		}
		return nil, fmt.Errorf("failed to get connector by ID: %w", err)
	}

	return c, nil
}

func (s *SqlStorage) GetAllConnectors(ctx context.Context) ([]*Connector, error) {
	query := "SELECT * FROM connectors"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query connectors: %w", err)
	}
	defer rows.Close()

	var connectors []*Connector
	for rows.Next() {
		c, err := scanRowsIntoConnector(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		connectors = append(connectors, c)
	}

	// Check for errors after iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return connectors, nil
}

func (s *SqlStorage) DeleteConnector(ctx context.Context, ID string) error {
	query := "DELETE FROM connectors WHERE id = $1"

	result, err := s.db.ExecContext(ctx, query, ID)
	if err != nil {
		return fmt.Errorf("failed to delete connector with ID %s: %w", ID, err)
	}

	// Check if a row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no connector found with ID %s", ID)
	}

	return nil
}

func scanRowsIntoConnector(rows *sql.Rows) (*Connector, error) {
	connector := &Connector{}

	err := rows.Scan(
		&connector.ID,
		&connector.WorkspaceID,
		&connector.DefaultChannelID,
		&connector.CreatedAt,
		&connector.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan connector row: %w", err)
	}

	return connector, nil
}
