package errors

import "errors"

// Common errors.
//
// These sentinels are wrapped with %w by the parsers so callers can classify a
// decode failure with errors.Is (e.g. errors.Is(err, ErrInvalidHeader)) without
// matching on error strings. Each layer wraps its own sentinel: the message
// parser wraps the section-level sentinels (ErrInvalidAnswer/Authority/Additional)
// around the ErrInvalidResourceRecord returned by the record parser, so a section
// failure matches both its section sentinel and ErrInvalidResourceRecord.
var (
	ErrInvalidDomainName     = errors.New("invalid domain name")
	ErrInvalidMessage        = errors.New("invalid message format")
	ErrInvalidHeader         = errors.New("invalid header format")
	ErrInvalidQuestion       = errors.New("invalid question format")
	ErrInvalidAnswer         = errors.New("invalid answer format")
	ErrInvalidAuthority      = errors.New("invalid authority format")
	ErrInvalidAdditional     = errors.New("invalid additional format")
	ErrInvalidResourceRecord = errors.New("invalid resource record format")
	ErrNameTooLong           = errors.New("domain name too long")
	ErrLabelTooLong          = errors.New("label too long")
)
