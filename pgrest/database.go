package pgrest

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgx/v4/pgxpool"

	log "github.com/sirupsen/logrus"
)

//var db *pgxpool.Pool

func QueryHandler(w http.ResponseWriter, r *http.Request, connection ConnectionConfig, body RequestBody) {
	db, err := pgxpool.Connect(context.Background(), connection.ConnectionString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(context.Background(), body.Query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", err), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	// Retrieve column names
	fieldDescriptions := rows.FieldDescriptions()
	columns := getColumnNames(fieldDescriptions)

	// Prepare the response writer for gzip compression
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "application/json")
	gz := gzip.NewWriter(w)
	defer gz.Close()

	// Stream the JSON response
	encoder := json.NewEncoder(gz)

	// Write column names first
	queryResponse := map[string]interface{}{
		"columns": columns,
		"rows":    [][]interface{}{},
	}

	if err := encoder.Encode(queryResponse); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}

	// Stream rows
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading row: %v", err), http.StatusInternalServerError)
			return
		}

		// Encode each row individually
		row := map[string]interface{}{
			"rows": [][]interface{}{values},
		}
		if err := encoder.Encode(row); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding row: %v", err), http.StatusInternalServerError)
			return
		}
	}

	if rows.Err() != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", rows.Err()), http.StatusInternalServerError)
		return
	}
}

func getColumnNames(columns []pgproto3.FieldDescription) []string {
	names := make([]string, len(columns))
	for i, col := range columns {
		names[i] = string(col.Name)
	}
	return names
}
