package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"bufio"

	"github.com/andybalholm/brotli"
	"github.com/jackc/pgx/v5"
	"github.com/sogelink-research/pgrest/errors"
	"github.com/sogelink-research/pgrest/models"
	"github.com/sogelink-research/pgrest/service"
	"github.com/sogelink-research/pgrest/settings"
	"github.com/sogelink-research/pgrest/utils"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/endian"
	"github.com/apache/arrow/go/v18/arrow/ipc"
	"github.com/apache/arrow/go/v18/arrow/memory"
	"github.com/apache/arrow/go/v18/parquet"
	"github.com/apache/arrow/go/v18/parquet/compress"
	"github.com/apache/arrow/go/v18/parquet/pqarrow"
)

// QueryHandler handles the HTTP request for executing a database query.
// It takes in the HTTP response writer, the HTTP request and the database connection configuration.
// It returns an error if there was an issue connecting to the database or executing the query.
func QueryHandler(config settings.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doneChan := make(chan struct{})
		go func() {
			defer close(doneChan)

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

			var encoder *json.Encoder
			var writer io.Writer

			const bufferSize = 64 * 1024 // 64 KB

			bw := bufio.NewWriterSize(w, bufferSize)
			defer bw.Flush()

			acceptEncoding := r.Header.Get("Accept-Encoding")
			if strings.Contains(acceptEncoding, "br") {
				w.Header().Set("Content-Encoding", "br")
				brWriter := brotli.NewWriterLevel(bw, brotli.DefaultCompression)
				defer brWriter.Close()
				writer = brWriter
			} else if strings.Contains(acceptEncoding, "gzip") {
				w.Header().Set("Content-Encoding", "gzip")
				var gz *gzip.Writer
				gz, _ = gzip.NewWriterLevel(bw, gzip.DefaultCompression)
				defer gz.Close()
				writer = gz
			} else {
				writer = bw
			}

			encoder = json.NewEncoder(writer)

			switch body.Format {
			case models.JSONFormat:
				handleFormatJSON(w, rows, columns, writer, encoder)
			case models.JSONDataArrayFormat:
				handleFormatJSONDataArray(w, rows, columns, writer, encoder)
			case models.ArrowFormat:
				handleFormatArrow(w, rows, writer, 1000)
			case models.CSVFormat:
				handleFormatCSV(w, rows, columns, writer)
			case models.ParquetFormat:
				handleFormatParquet(w, rows, 1000, writer)
			default:
				handleFormatJSON(w, rows, columns, writer, encoder)
			}
		}()

		select {
		case <-r.Context().Done():
			err := r.Context().Err()
			if err == context.Canceled {
				err = errors.NewAPIError(http.StatusRequestTimeout, "Request canceled", nil)
				HandleError(w, err)
				return
			} else if err == context.DeadlineExceeded {
				err = errors.NewAPIError(http.StatusGatewayTimeout, "Processing too slow", nil)
				HandleError(w, err)
				return
			}
			return
		case <-doneChan:
		}
	}
}

// handleFormatJSON writes the query result in the default JSON format to the provided writer.
// It takes the rows returned by the query, the column names, the writer to write the JSON output,
// and the encoder to encode the JSON data.
// The function iterates over the rows, converts them into a map with column names as keys,
// and writes the JSON-encoded rows to the writer.
// The resulting JSON is wrapped in a data array.
func handleFormatJSON(w http.ResponseWriter, rows pgx.Rows, columns []string, writer io.Writer, encoder *json.Encoder) error {
	w.Header().Set("Content-Type", "application/json")
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

// handleFormatJSONDataArray formats the data from the given rows and writes it to the writer in JSON format.
// It includes the fields and rows information in the JSON output.
func handleFormatJSONDataArray(w http.ResponseWriter, rows pgx.Rows, columns []string, writer io.Writer, encoder *json.Encoder) {
	w.Header().Set("Content-Type", "application/json")
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
}

// handleFormatCSV writes the given rows and columns to the provided writer in CSV format.
// It sets the appropriate Content-Type header and handles any errors that occur during the process.
func handleFormatCSV(w http.ResponseWriter, rows pgx.Rows, columns []string, writer io.Writer) {
	w.Header().Set("Content-Type", "text/csv")
	csvWriter := csv.NewWriter(writer)

	// Write the CSV header
	if err := csvWriter.Write(columns); err != nil {
		apiError := errors.NewAPIError(http.StatusInternalServerError, "Error writing CSV header", nil)
		HandleError(w, apiError)
		return
	}

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			apiError := errors.NewAPIError(http.StatusInternalServerError, "Error reading row values", nil)
			HandleError(w, apiError)
			return
		}

		strValues := make([]string, len(values))
		for i, val := range values {
			if val != nil {
				strValues[i] = fmt.Sprintf("%v", val)
			} else {
				strValues[i] = ""
			}
		}

		if err := csvWriter.Write(strValues); err != nil {
			apiError := errors.NewAPIError(http.StatusInternalServerError, "Error writing CSV row", nil)
			HandleError(w, apiError)
			return
		}
	}

	// Flush the CSV writer to ensure all data is written
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		apiError := errors.NewAPIError(http.StatusInternalServerError, "Error flushing CSV writer", nil)
		HandleError(w, apiError)
		return
	}
}

// createArrowSchema creates Arrow schema from column descriptions
func createArrowSchema(rows pgx.Rows) (*arrow.Schema, error) {
	columns := rows.FieldDescriptions()
	fields := make([]arrow.Field, len(columns))
	for i, col := range columns {
		arrowType, err := utils.PGTypeToArrowType(col.DataTypeOID)
		if err != nil {
			return nil, fmt.Errorf("unsupported column type %d for column %s: %v", col.DataTypeOID, col.Name, err)
		}
		fields[i] = arrow.Field{Name: col.Name, Type: arrowType, Nullable: true}
	}

	return arrow.NewSchemaWithEndian(fields, nil, endian.NativeEndian), nil
}

// Create RecordBuilder for a given schema
func createRecordBuilder(schema *arrow.Schema) *array.RecordBuilder {
	pool := memory.NewGoAllocator()
	return array.NewRecordBuilder(pool, schema)
}

// Append values to the appropriate RecordBuilder field
func appendArrowValues(recordBuilder *array.RecordBuilder, values []interface{}) error {
	for i, value := range values {
		switch builder := recordBuilder.Field(i).(type) {
		case *array.Int64Builder:
			if value == nil {
				builder.AppendNull()
			} else {
				builder.Append(value.(int64))
			}
		case *array.Int32Builder:
			if value == nil {
				builder.AppendNull()
			} else {
				builder.Append(value.(int32))
			}
		case *array.Int16Builder:
			if value == nil {
				builder.AppendNull()
			} else {
				builder.Append(value.(int16))
			}
		case *array.Float32Builder:
			if value == nil {
				builder.AppendNull()
			} else {
				builder.Append(value.(float32))
			}
		case *array.Float64Builder:
			if value == nil {
				builder.AppendNull()
			} else {
				builder.Append(value.(float64))
			}
		case *array.BooleanBuilder:
			if value == nil {
				builder.AppendNull()
			} else {
				builder.Append(value.(bool))
			}
		case *array.StringBuilder:
			if value == nil {
				builder.AppendNull()
			} else {
				builder.Append(value.(string))
			}
		case *array.TimestampBuilder:
			if value == nil {
				builder.AppendNull()
			} else {
				timestamp := value.(time.Time)
				millis := timestamp.UnixNano() / int64(time.Millisecond)
				builder.Append(arrow.Timestamp(millis))
			}
		case *array.BinaryBuilder:
			if value == nil {
				builder.AppendNull()
			} else {
				builder.Append(value.([]byte))
			}
		default:
			return fmt.Errorf("unsupported builder type %T", builder)
		}
	}
	return nil
}

// handleFormatParquet writes the given rows to the provided writer in Parquet format.
func handleFormatParquet(w http.ResponseWriter, rows pgx.Rows, batchSize int, writer io.Writer) {
	w.Header().Set("Content-Type", "application/octet-stream")

	schema, err := createArrowSchema(rows)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	writerProps := parquet.NewWriterProperties(parquet.WithBatchSize(int64(batchSize)), parquet.WithCompression(compress.Codecs.Snappy))
	pw, err := pqarrow.NewFileWriter(schema, buf, writerProps, pqarrow.DefaultWriterProps())
	if err != nil {
		details := err.Error()
		err = errors.NewAPIError(http.StatusInternalServerError, "Error creating Parquet writer", &details)
		HandleError(w, err)
		return
	}

	recordBuilder := createRecordBuilder(schema)
	defer recordBuilder.Release()

	// Iterate over rows
	var recordCounter int
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			detail := err.Error()
			err := errors.NewAPIError(http.StatusInternalServerError, "Error reading row values", &detail)
			HandleError(w, err)
			return
		}

		if err := appendArrowValues(recordBuilder, values); err != nil {
			details := err.Error()
			err := errors.NewAPIError(http.StatusInternalServerError, "Error appending arrow value", &details)
			HandleError(w, err)
			return
		}

		recordCounter++

		if recordCounter >= batchSize {
			record := recordBuilder.NewRecord()
			if err := pw.Write(record); err != nil {
				details := err.Error()
				err := errors.NewAPIError(http.StatusInternalServerError, "Error writing Parquet record", &details)
				HandleError(w, err)
				return
			}
			recordCounter = 0
			record.Release()
		}
	}

	if recordCounter > 0 {
		record := recordBuilder.NewRecord()
		if err := pw.Write(record); err != nil {
			details := err.Error()
			err := errors.NewAPIError(http.StatusInternalServerError, "Error writing Parquet record", &details)
			HandleError(w, err)
			return
		}

		record.Release()
	}

	err = pw.Close()
	if err != nil {
		err := errors.NewAPIError(http.StatusInternalServerError, "Error closing Parquet writer", nil)
		HandleError(w, err)
		return
	}

	_, err = writer.Write(buf.Bytes())
	if err != nil {
		err := errors.NewAPIError(http.StatusInternalServerError, "Error writing Parquet data", nil)
		HandleError(w, err)
		return
	}
}

// handleFormatArrow handles the formatting of the query results in Apache Arrow format.
func handleFormatArrow(w http.ResponseWriter, rows pgx.Rows, writer io.Writer, batchSize int) {
	w.Header().Set("Content-Type", "application/vnd.apache.arrow.stream")

	schema, err := createArrowSchema(rows)
	if err != nil {
		details := err.Error()
		apiErr := errors.NewAPIError(http.StatusInternalServerError, "Error creating Arrow schema", &details)
		HandleError(w, apiErr)
		return
	}

	arrWriter := ipc.NewWriter(writer, ipc.WithSchema(schema))
	defer arrWriter.Close()

	recordBuilder := createRecordBuilder(schema)
	defer recordBuilder.Release()

	var recordCounter int
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			details := err.Error()
			apiErr := errors.NewAPIError(http.StatusInternalServerError, "Error reading row values", &details)
			HandleError(w, apiErr)
			return
		}

		if err := appendArrowValues(recordBuilder, values); err != nil {
			details := err.Error()
			err := errors.NewAPIError(http.StatusInternalServerError, "Error appending arrow value", &details)
			HandleError(w, err)
			return
		}

		recordCounter++

		if recordCounter >= batchSize {
			record := recordBuilder.NewRecord()

			if err := arrWriter.Write(record); err != nil {
				details := err.Error()
				apiErr := errors.NewAPIError(http.StatusInternalServerError, "Error writing record batch", &details)
				HandleError(w, apiErr)
				return
			}

			record.Release()
			recordCounter = 0
		}
	}

	if recordCounter > 0 {
		record := recordBuilder.NewRecord()
		defer record.Release()

		if err := arrWriter.Write(record); err != nil {
			details := err.Error()
			apiErr := errors.NewAPIError(http.StatusInternalServerError, "Error writing record batch", &details)
			HandleError(w, apiErr)
		}
	}
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
