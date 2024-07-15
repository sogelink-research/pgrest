package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"bufio"

	"github.com/andybalholm/brotli"
	"github.com/jackc/pgx/v5"
	"github.com/sogelink-research/pgrest/errors"
	"github.com/sogelink-research/pgrest/models"
	"github.com/sogelink-research/pgrest/service"
	"github.com/sogelink-research/pgrest/settings"
	"github.com/sogelink-research/pgrest/utils"
)

// QueryHandler handles the HTTP request for executing a database query.
// It takes in the HTTP response writer, the HTTP request and the database connection configuration.
// It returns an error if there was an issue connecting to the database or executing the query.
func QueryHandler(config settings.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connectionName, err := utils.GetConnectionNameFromRequest(r)
		if err != nil {
			HandleError(w, err)
			return
		}

		// Find the requested database connection
		connection, err := config.GetConnectionConfig(connectionName)
		if err != nil {
			apiError := errors.NewAPIError(http.StatusBadRequest, fmt.Sprintf("Requested connection '%s' not found", connectionName), nil)
			HandleError(w, apiError)
			return
		}

		body, err := getBodyData(r)
		if err != nil {
			HandleError(w, err)
			return
		}

		rows, columns, err := service.QueryPostgres(body.Query, connection)
		if err != nil {
			HandleError(w, err)
			return
		}

		defer rows.Close()

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
		case models.DefaultFormat:
			handleFormatDefault(rows, columns, writer, encoder)
		case models.DataArrayFormat:
			handleFormatDataArray(rows, columns, writer, encoder)
		default:
			handleFormatDefault(rows, columns, writer, encoder)
		}
	}
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

// getBodyData reads the request body from the provided http.Request and returns
// the decoded JSON payload as a QueryRequestBody struct.
// If there is an error reading the request body or decoding the JSON payload, an APIError is returned.
// The request body is reset with the original data before returning.
func getBodyData(r *http.Request) (*models.QueryRequestBody, error) {
	var requestBody models.QueryRequestBody

	// Decode the incoming JSON payload
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		details := err.Error()
		return nil, errors.NewAPIError(http.StatusBadRequest, "Invalid request body", &details)
	}

	return &requestBody, nil
}
