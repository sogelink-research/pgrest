package pgrest

// RequestBody represents the structure of the incoming JSON payload
type RequestBody struct {
	Database string `json:"database"`
	Query    string `json:"query"`
	Format   string `json:"format,omitempty"`
}
