package pgrest

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"bufio"

	"github.com/andybalholm/brotli"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// QueryHandler handles the HTTP request for executing a database query.
// It takes in the HTTP response writer, the HTTP request, the database connection configuration,
// and the request body containing the query to be executed.
// It returns an error if there was an issue connecting to the database or executing the query.
func QueryHandler(w http.ResponseWriter, r *http.Request, connection *ConnectionConfig, body *QueryRequestBody) error {
	pool, err := GetDBPool(connection.Name, connection.ConnectionString)
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, fmt.Sprintf("Error connecting to database: %v", connection.Name), nil)
	}

	rows, err := pool.Query(context.Background(), body.Query)
	if err != nil {
		errorMessage := fmt.Sprintf("%v", err.Error())
		return NewAPIError(http.StatusBadRequest, "Error executing query", &errorMessage)
	}
	defer rows.Close()

	columns := getColumnNames(rows.FieldDescriptions())
	w.Header().Set("Content-Type", "application/json")

	var encoder *json.Encoder
	var writer io.Writer

	// Use buffered writer for more efficient I/O
	const bufferSize = 64 * 1024 // 64 KB
	bw := bufio.NewWriterSize(w, bufferSize)
	defer bw.Flush()

	acceptEncoding := r.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEncoding, "br") {
		w.Header().Set("Content-Encoding", "br")
		brWriter := brotli.NewWriterLevel(bw, brotli.DefaultCompression)
		defer brWriter.Close()
		writer = brWriter
		encoder = json.NewEncoder(brWriter)
	} else if strings.Contains(acceptEncoding, "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		var gz *gzip.Writer
		gz, _ = gzip.NewWriterLevel(bw, gzip.DefaultCompression)
		defer gz.Close()
		writer = gz
		encoder = json.NewEncoder(gz)
	} else {
		writer = bw
		encoder = json.NewEncoder(bw)
	}

	switch body.Format {
	case DefaultFormat:
		return handleFormatDefault(rows, columns, writer, encoder)
	case DataArrayFormat:
		return handleFormatDataArray(rows, columns, writer, encoder)
	}

	return nil
}

// getColumnNames returns a slice of column names extracted from the given pgconn.FieldDescription slice.
func getColumnNames(columns []pgconn.FieldDescription) []string {
	names := make([]string, len(columns))
	for i, col := range columns {
		names[i] = string(col.Name)
	}
	return names
}

// handleFormatDefault writes the query result in the default JSON format to the provided writer.
// It takes the rows returned by the query, the column names, the writer to write the JSON output,
// and the encoder to encode the JSON data.
// The function iterates over the rows, converts them into a map with column names as keys,
// and writes the JSON-encoded rows to the writer.
// The resulting JSON is wrapped in a data array.
func handleFormatDefault(rows pgx.Rows, columns []string, writer io.Writer, encoder *json.Encoder) error {
	jsonStart := []byte(`{"data":[`)
	writer.Write(jsonStart)

	first := true
	for rows.Next() {
		values, _ := rows.Values()

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}

		if !first {
			writer.Write([]byte(`,`))
		}
		first = false

		encoder.Encode(row)
	}

	writer.Write([]byte(`]}`))
	return nil
}

// handleFormatDataArray formats the data from the given rows and writes it to the writer in JSON format.
// It includes the fields and rows information in the JSON output.
func handleFormatDataArray(rows pgx.Rows, columns []string, writer io.Writer, encoder *json.Encoder) error {
	jsonStart := []byte(`{"data": {`)
	writer.Write(jsonStart)

	jsonFields := []byte(fmt.Sprintf(`"fields":["%s"],`, strings.Join(columns, `","`)))
	writer.Write(jsonFields)

	jsonRows := []byte(`"rows":[`)
	writer.Write(jsonRows)

	first := true
	for rows.Next() {
		values, _ := rows.Values()

		if !first {
			writer.Write([]byte(`,`))
		}
		first = false

		encoder.Encode(values)
	}

	writer.Write([]byte(`]}}`))
	return nil
}
