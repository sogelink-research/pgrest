package models

import (
	"encoding/json"
	"fmt"
)

// QueryRequestBody represents the structure of the incoming JSON payload
type QueryRequestBody struct {
	Connection string     `json:"connection"`
	Query      string     `json:"query"`
	Format     FormatType `json:"format,omitempty"`
}

type FormatType string

const (
	DefaultFormat   FormatType = "default"
	DataArrayFormat FormatType = "dataArray"
)

// UnmarshalJSON unmarshals the JSON data into the QueryRequestBody struct.
// It sets default values for Connections and Format fields if they are empty.
// It also validates the Format field and returns an error if it is not a supported format.
func (rb *QueryRequestBody) UnmarshalJSON(data []byte) error {
	// Create a secondary type to avoid recursion
	type Alias QueryRequestBody
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(rb),
	}

	// Unmarshal the data into the auxiliary struct
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Set the default value if Connection is empty
	if rb.Connection == "" {
		rb.Connection = "default"
	}

	// Set the default value if Format is empty
	if rb.Format == "" {
		rb.Format = DefaultFormat
	} else if rb.Format != DefaultFormat && rb.Format != DataArrayFormat {
		return fmt.Errorf(fmt.Sprintf("invalid format type '%s', supported formats: 'default', 'dataArray'", rb.Format))
	}

	return nil
}
