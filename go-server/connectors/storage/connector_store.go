package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SqlStorage struct {
	db       *pgxpool.Pool
	smClient *secretsmanager.SecretsManager
}

func NewSqlStorage(db *pgxpool.Pool, smClient *secretsmanager.SecretsManager) *SqlStorage {
	return &SqlStorage{db: db, smClient: smClient}
}

// SaveConnector inserts a new connector and returns its ID
func (s *SqlStorage) SaveConnector(ctx context.Context, connector *Connector) (string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		// Ensure rollback if commit wasn't called
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	var connectorID string
	query := `
		INSERT INTO connectors (workspace_id, default_channel_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`

	err = s.db.QueryRow(ctx, query,
		connector.WorkspaceID,
		connector.DefaultChannelID,
		connector.CreatedAt,
		connector.UpdatedAt,
	).Scan(&connectorID)
	if err != nil {
		return "", fmt.Errorf("failed to save connector: %w", err)
	}

	// save slack token
	result, err := s.smClient.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(connectorID),
		SecretString: aws.String(connector.Token),
	})
	if err != nil {
		return "", fmt.Errorf("failed to save slack token: %w", err)
	}

	// Commit transaction if everything succeeds
	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("secret manager result: %+v", result)

	return connectorID, nil
}

// GetConnectorByID retrieves a connector by its ID
func (s *SqlStorage) GetConnectorByID(ctx context.Context, connectorID string) (*Connector, error) {
	c := &Connector{}

	query := `SELECT id, workspace_id, default_channel_id, created_at, updated_at FROM connectors WHERE id = $1`
	err := s.db.QueryRow(ctx, query, connectorID).Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.DefaultChannelID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("connector with ID %s not found", connectorID)
		}
		return nil, fmt.Errorf("failed to get connector by ID: %w", err)
	}

	result, err := s.smClient.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(c.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}

	c.Token = aws.StringValue(result.SecretString) // Store secret in struct

	// fmt.Printf("secret manager get result: %+v\n", result)

	return c, nil
}

// GetAllConnectors retrieves all connectors
func (s *SqlStorage) GetAllConnectorsOLD(ctx context.Context) ([]*Connector, error) {
	query := `SELECT id, workspace_id, default_channel_id, created_at, updated_at FROM connectors`
	rows, err := s.db.Query(ctx, query)
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

func (s *SqlStorage) GetAllConnectors(ctx context.Context) ([]*Connector, error) {
	query := `SELECT id, workspace_id, default_channel_id, created_at, updated_at FROM connectors`
	rows, err := s.db.Query(ctx, query)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	// Fetch secrets AFTER fetching rows using goroutines (better performance)
	var wg sync.WaitGroup
	for _, c := range connectors {
		wg.Add(1)
		go func(c *Connector) {
			defer wg.Done()
			result, err := s.smClient.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
				SecretId: aws.String(c.ID),
			})
			if err != nil {
				slog.Error("Failed to get secret for", "ID", c.ID, "err", err)
				return
			}
			c.Token = aws.StringValue(result.SecretString)
		}(c)
	}
	wg.Wait()

	return connectors, nil
}

// DeleteConnector removes a connector by ID
func (s *SqlStorage) DeleteConnectorOLD(ctx context.Context, ID string) error {
	query := `DELETE FROM connectors WHERE id = $1`

	result, err := s.db.Exec(ctx, query, ID)
	if err != nil {
		return fmt.Errorf("failed to delete connector with ID %s: %w", ID, err)
	}

	// Check if a row was actually deleted
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no connector found with ID %s", ID)
	}

	return nil
}

// DeleteConnector removes a connector by ID and deletes its secret from AWS Secrets Manager
func (s *SqlStorage) DeleteConnector(ctx context.Context, ID string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Ensure rollback if something fails

	query := `DELETE FROM connectors WHERE id = $1`
	result, err := tx.Exec(ctx, query, ID)
	if err != nil {
		return fmt.Errorf("failed to delete connector with ID %s: %w", ID, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no connector found with ID %s", ID)
	}

	_, err = s.smClient.DeleteSecretWithContext(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(ID),
		ForceDeleteWithoutRecovery: aws.Bool(true), // Immediately deletes the secret
	})
	if err != nil {
		return fmt.Errorf("failed to delete secret for ID %s: %w", ID, err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// scanRowsIntoConnector scans a row into a Connector struct
func scanRowsIntoConnector(rows pgx.Rows) (*Connector, error) {
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
