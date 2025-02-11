package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"connector-recruitment/go-server/connectors/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SqlStorage is responsible for database and secrets operations.
type SqlStorage struct {
	logger   logger.Logger
	db       *pgxpool.Pool
	smClient *secretsmanager.SecretsManager
}

// NewSqlStorage creates a new SqlStorage instance.
func NewSqlStorage(db *pgxpool.Pool, smClient *secretsmanager.SecretsManager, logger logger.Logger) *SqlStorage {
	return &SqlStorage{db: db, smClient: smClient, logger: logger}
}

// SaveConnector inserts a new connector into the database and creates its secret in AWS Secrets Manager atomically.
func (s *SqlStorage) SaveConnector(ctx context.Context, connector *Connector) (connectorID string, err error) {
	// Begin a transaction.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	// Ensure rollback if commit is not reached.
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Insert connector record using the transaction.
	query := `
		INSERT INTO connectors (workspace_id, default_channel_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`
	err = tx.QueryRow(ctx, query,
		connector.WorkspaceID,
		connector.DefaultChannelID,
		connector.CreatedAt,
		connector.UpdatedAt,
	).Scan(&connectorID)
	if err != nil {
		return "", fmt.Errorf("failed to save connector: %w", err)
	}

	// Create secret in AWS Secrets Manager.
	_, err = s.smClient.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(connectorID),
		SecretString: aws.String(connector.Token),
	})
	if err != nil {
		return "", fmt.Errorf("failed to save slack token: %w", err)
	}

	// Commit the transaction.
	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Debug("Created secret", "secret-name", connectorID)
	return connectorID, nil
}

// GetConnectorByID retrieves a connector by its ID and also fetches its secret token.
func (s *SqlStorage) GetConnectorByID(ctx context.Context, connectorID string) (*Connector, error) {
	c := &Connector{}
	query := `
		SELECT id, workspace_id, default_channel_id, created_at, updated_at 
		FROM connectors 
		WHERE id = $1`
	err := s.db.QueryRow(ctx, query, connectorID).Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.DefaultChannelID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("connector with ID %s not found: %w", connectorID, err)
		}
		return nil, fmt.Errorf("failed to get connector by ID: %w", err)
	}

	// Fetch secret value from AWS Secrets Manager.
	result, err := s.smClient.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(c.ID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value for connector %s: %w", connectorID, err)
	}
	c.Token = aws.StringValue(result.SecretString)
	return c, nil
}

// scanRowsIntoConnector scans a pgx.Rows result into a Connector struct.
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

// GetAllConnectors retrieves all connectors and then concurrently fetches their secret tokens.
func (s *SqlStorage) GetAllConnectors(ctx context.Context) ([]*Connector, error) {
	query := `
		SELECT id, workspace_id, default_channel_id, created_at, updated_at 
		FROM connectors`
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

	// Fetch secrets concurrently for each connector.
	var wg sync.WaitGroup
	for _, c := range connectors {
		wg.Add(1)
		go func(c *Connector) {
			defer wg.Done()
			res, err := s.smClient.GetSecretValueWithContext(ctx, &secretsmanager.GetSecretValueInput{
				SecretId: aws.String(c.ID),
			})
			if err != nil {
				s.logger.Error("Failed to get secret", "ID", c.ID, "error", err)
				return
			}
			c.Token = aws.StringValue(res.SecretString)
		}(c)
	}
	wg.Wait()
	return connectors, nil
}

// DeleteConnector removes a connector by ID from the database and deletes its secret from AWS Secrets Manager atomically.
func (s *SqlStorage) DeleteConnector(ctx context.Context, ID string) error {
	// Begin a transaction.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete connector from the database.
	query := `DELETE FROM connectors WHERE id = $1`
	result, err := tx.Exec(ctx, query, ID)
	if err != nil {
		return fmt.Errorf("failed to delete connector with ID %s: %w", ID, err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("no connector found with ID %s", ID)
	}

	// Delete the secret from AWS Secrets Manager.
	_, err = s.smClient.DeleteSecretWithContext(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(ID),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to delete secret for ID %s: %w", ID, err)
	}

	// Commit the transaction.
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
