package threads

import (
	"errors"
	"fmt"
)

// Error represents an upstream Threads / Graph API error. Meta returns a
// nested "error" object in JSON; this struct flattens it.
type Error struct {
	StatusCode    int    // HTTP status code returned by the API.
	Type          string // Graph "error.type" (e.g. "OAuthException").
	Code          int    // Graph "error.code".
	Subcode       int    // Graph "error.error_subcode".
	Message       string // Human-readable description from the API.
	FBTraceID     string // Meta-side trace identifier for support tickets.
	Body          []byte // Raw response body, retained for diagnostics.
}

func (e *Error) Error() string {
	return fmt.Sprintf("threads: [%d/%s] %s (code=%d subcode=%d trace=%s)", e.StatusCode, e.Type, e.Message, e.Code, e.Subcode, e.FBTraceID)
}

// IsAuthError reports whether the error indicates an authentication failure.
func (e *Error) IsAuthError() bool {
	if e.StatusCode == 401 {
		return true
	}
	switch e.Code {
	case 190, 102:
		return true
	}
	return e.Type == "OAuthException"
}

// IsRateLimited reports whether the error indicates a rate-limit response.
func (e *Error) IsRateLimited() bool {
	return e.StatusCode == 429 || e.Code == 4
}

// IsNotFound reports whether the requested resource was not found.
func (e *Error) IsNotFound() bool { return e.StatusCode == 404 }

// IsServerError reports whether the upstream returned a 5xx-style failure.
func (e *Error) IsServerError() bool { return e.StatusCode >= 500 }

// AsError unwraps an error into a *Error if applicable.
func AsError(err error) (*Error, bool) {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
