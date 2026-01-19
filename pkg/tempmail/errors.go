package tempmail

import "errors"

var (
	ErrNotAuthenticated     = errors.New("not authenticated")
	ErrNoActiveAccount      = errors.New("no active account")
	ErrNoDomains            = errors.New("no domains available")
	ErrCouldNotGetMessages  = errors.New("could not get messages")
	ErrCouldNotGetAccount   = errors.New("could not get account")
	ErrCouldNotUpdateMessage = errors.New("could not update message")
	ErrAddressRequired      = errors.New("address and password required")
)
