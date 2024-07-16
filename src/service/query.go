package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sogelink-research/pgrest/database"
	"github.com/sogelink-research/pgrest/errors"
	"github.com/sogelink-research/pgrest/settings"
)

// QueryPostgres executes a query on a PostgreSQL database using the provided connection configuration.
// It returns the result rows, column names, and an API error if any.
func QueryPostgres(query string, connection *settings.ConnectionConfig) (pgx.Rows, []string, error) {
	pool, err := database.GetDBPool(connection.Name, connection.ConnectionString)
	if err != nil {
		return nil, nil, errors.NewAPIError(http.StatusInternalServerError, fmt.Sprintf("Error connecting to database: %v", connection.Name), nil)
	}

	rows, err := pool.Query(context.Background(), query)
	if err != nil {
		errorMessage := fmt.Sprintf("%v", err.Error())
		return nil, nil, errors.NewAPIError(http.StatusBadRequest, "Error executing query", &errorMessage)
	}

	columns := getColumnNames(rows.FieldDescriptions())

	return rows, columns, nil
}

// getColumnNames returns a slice of column names extracted from the given pgconn.FieldDescription slice.
func getColumnNames(columns []pgconn.FieldDescription) []string {
	names := make([]string, len(columns))
	for i, col := range columns {
		names[i] = string(col.Name)
	}
	return names
}
