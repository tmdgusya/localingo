package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrNotFound indicates the requested record was not found
	ErrNotFound = errors.New("record not found")

	// ErrConflict indicates a conflict with existing data (e.g., unique constraint violation)
	ErrConflict = errors.New("record already exists")

	// ErrInvalidInput indicates invalid input parameters
	ErrInvalidInput = errors.New("invalid input")

	// ErrDatabaseConnection indicates a database connection error
	ErrDatabaseConnection = errors.New("database connection error")
)

// WrapDBError converts database-specific errors to domain errors
func WrapDBError(err error) error {
	if err == nil {
		return nil
	}

	// Handle pgx no rows error
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w", ErrNotFound)
	}

	// Handle PostgreSQL-specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("%w: %s", ErrConflict, pgErr.Detail)
		case "23503": // foreign_key_violation
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Detail)
		case "23502": // not_null_violation
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Detail)
		case "23514": // check_violation
			return fmt.Errorf("%w: %s", ErrInvalidInput, pgErr.Detail)
		}
	}

	// Return the original error wrapped
	return fmt.Errorf("database error: %w", err)
}
