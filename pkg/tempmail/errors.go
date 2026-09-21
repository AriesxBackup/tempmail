package tempmail

import (
	"errors"
	"fmt"
)

var (
	ErrAPIKeyRequired  = errors.New("api key required")
	ErrNoActiveAccount = errors.New("no active account")
	ErrNoDomains       = errors.New("no active domains available")
	ErrNotFound        = errors.New("not found")
	ErrTimeout         = errors.New("timed out waiting for message")
	ErrMailboxNotFound = errors.New("mailbox not found")
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("api error: HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("api error: HTTP %d - %s", e.StatusCode, e.Message)
}

func (e *APIError) Is(target error) bool {
	return target == ErrNotFound && e.StatusCode == 404
}
