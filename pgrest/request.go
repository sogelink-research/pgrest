package pgrest

import (
	"encoding/json"
	"fmt"
)

// RequestBody represents the structure of the incoming JSON payload
type RequestBody struct {
	Database string     `json:"database"`
	Query    string     `json:"query"`
	Format   FormatType `json:"format,omitempty"`
}

type FormatType string

const (
	DefaultFormat   FormatType = "default"
	DataArrayFormat FormatType = "dataArray"
)

// UnmarshalJSON unmarshals the JSON data into the RequestBody struct.
// It sets default values for Database and Format fields if they are empty.
// It also validates the Format field and returns an error if it is not a supported format.
func (rb *RequestBody) UnmarshalJSON(data []byte) error {
	// Create a secondary type to avoid recursion
	type Alias RequestBody
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(rb),
	}

	// Unmarshal the data into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Set the default value if Database is empty
	if rb.Database == "" {
		rb.Database = "default"
	}

	// Set the default value if Format is empty
	if rb.Format == "" {
		rb.Format = DefaultFormat
	} else if rb.Format != DefaultFormat && rb.Format != DataArrayFormat {
		return fmt.Errorf(fmt.Sprintf("invalid format type '%s', supported formats: 'default', 'dataArray'", rb.Format))
	}

	return nil
}
