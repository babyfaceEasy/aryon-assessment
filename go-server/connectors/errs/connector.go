package errs

import (
	"errors"
	"fmt"
)

// ErrConnectorNotFound is the base error for not found connectors
var ErrConnectorNotFound = errors.New("connector not found")

type ConnectorNotFoundError struct {
	ID string
}

func (e *ConnectorNotFoundError) Error() string {
	return fmt.Sprintf("connector with key %s not found", e.ID)
}

// NewConnectorNotFoundError creates a new error with the given ID
func NewConnectorNotFoundError(ID string) error {
	return fmt.Errorf("%w: %s", ErrConnectorNotFound, &ConnectorNotFoundError{ID: ID})
}

// ErrConnectorExistAlready is the base error for already exist connectors
var ErrConnectorExistAlready = errors.New("connector already exist")

type ConnectorExistAlreadyError struct {
	ID string
}

func (e *ConnectorExistAlreadyError) Error() string {
	return fmt.Sprintf("connector with ID %s exist already", e.ID)
}

// NewConnectorAlreadyExistError creates a new error with the given ID
func NewConnectorAlreadyExistError(ID string) error {
	return fmt.Errorf("%w: %s", ErrConnectorExistAlready, &ConnectorExistAlreadyError{ID: ID})
}
