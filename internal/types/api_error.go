package types

// APIError represents an error with an associated HTTP status code and message.
// It implements the error interface.
type APIError struct {
	StatusCode int    `json:"status_code"` // HTTP status code for the error
	Message    string `json:"message"`     // Human-readable error message
}

// Error returns the error message string.
// This method makes APIError implement the error interface.
func (e *APIError) Error() string {
	return e.Message
}
