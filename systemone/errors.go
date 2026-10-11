package systemone

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the SystemOne API.
type APIError struct {
	StatusCode int           `json:"-"`
	Status     string        `json:"-"`
	Code       string        `json:"code"`
	Message    string        `json:"message"`
	Details    []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail provides additional context for validation errors.
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("systemone: %s (%d)", e.Message, e.StatusCode)
	}
	return fmt.Sprintf("systemone: API error (%d)", e.StatusCode)
}

// IsUnauthorized returns true if the error is a 401 Unauthorized.
func IsUnauthorized(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsUnprocessable returns true if the error is a 422 Unprocessable Entity.
func IsUnprocessable(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusUnprocessableEntity
	}
	return false
}

// IsRateLimited returns true if the error is a 429 Too Many Requests.
func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// IsOverloaded returns true if the error is a 529 Overloaded.
func IsOverloaded(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == 529
	}
	return false
}

// IsRetryable returns true if the request should be retried with backoff.
func IsRetryable(err error) bool {
	return IsRateLimited(err) || IsOverloaded(err)
}

func parseError(resp *http.Response, body []byte) error {
	errResp := &APIError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}

	if len(body) > 0 {
		_ = json.Unmarshal(body, errResp)
	}

	if errResp.Message == "" {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			errResp.Message = "missing or invalid API key"
		case http.StatusUnprocessableEntity:
			errResp.Message = "request validation failed"
		case http.StatusTooManyRequests:
			errResp.Message = "rate limit exceeded"
		case 529:
			errResp.Message = "service temporarily overloaded"
		default:
			errResp.Message = http.StatusText(resp.StatusCode)
		}
	}

	return errResp
}
