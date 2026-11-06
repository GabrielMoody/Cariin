package reranker

import "fmt"

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf(
		"reranker API error (status %d): %s",
		e.StatusCode,
		e.Message,
	)
}
