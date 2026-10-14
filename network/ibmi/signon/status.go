package signon

import (
	"errors"
	"fmt"
)

// ReturnCode is a Sign-on server return code. The high-order halfword
// selects the family.
type ReturnCode uint32

// Reason classifies a nonzero ReturnCode.
type Reason uint8

const (
	ReasonUnknown Reason = iota
	// ReasonRequestRejected (0x0001xxxx): the server rejected the request
	// itself, such as a bad seed or length. This indicates a client or
	// protocol fault, not a wrong password.
	ReasonRequestRejected
	ReasonProfileUnknown                // 0x00020001
	ReasonProfileDisabled               // 0x00020002
	ReasonProfileMismatch               // 0x00020003
	ReasonProfileRejected               // other 0x0002xxxx
	ReasonPasswordIncorrect             // 0x0003000B
	ReasonPasswordIncorrectNextDisables // 0x0003000C
	ReasonPasswordExpired               // 0x0003000D: correct but expired
	ReasonPasswordNone                  // 0x00030010: password is *NONE
	ReasonPasswordRejected              // other 0x0003xxxx
	ReasonSecurityFailure               // 0x0004xxxx
	ReasonExitProgram                   // 0x0005xxxx
	ReasonTokenFailure                  // 0x0006xxxx
)

var reasonNames = map[Reason]string{
	ReasonUnknown:                       "unknown",
	ReasonRequestRejected:               "request rejected",
	ReasonProfileUnknown:                "user profile unknown",
	ReasonProfileDisabled:               "user profile disabled",
	ReasonProfileMismatch:               "user profile mismatch",
	ReasonProfileRejected:               "user profile rejected",
	ReasonPasswordIncorrect:             "password incorrect",
	ReasonPasswordIncorrectNextDisables: "password incorrect; the next failure disables the profile",
	ReasonPasswordExpired:               "password expired",
	ReasonPasswordNone:                  "password is *NONE",
	ReasonPasswordRejected:              "password rejected",
	ReasonSecurityFailure:               "security failure",
	ReasonExitProgram:                   "exit program denied the request",
	ReasonTokenFailure:                  "token failure",
}

func (r Reason) String() string {
	if n, ok := reasonNames[r]; ok {
		return n
	}
	return fmt.Sprintf("Reason(%d)", uint8(r))
}

// Reason classifies the code.
func (c ReturnCode) Reason() Reason {
	switch c {
	case 0x00020001:
		return ReasonProfileUnknown
	case 0x00020002:
		return ReasonProfileDisabled
	case 0x00020003:
		return ReasonProfileMismatch
	case 0x0003000B:
		return ReasonPasswordIncorrect
	case 0x0003000C:
		return ReasonPasswordIncorrectNextDisables
	case 0x0003000D:
		return ReasonPasswordExpired
	case 0x00030010:
		return ReasonPasswordNone
	}
	switch c >> 16 {
	case 0x0001:
		return ReasonRequestRejected
	case 0x0002:
		return ReasonProfileRejected
	case 0x0003:
		return ReasonPasswordRejected
	case 0x0004:
		return ReasonSecurityFailure
	case 0x0005:
		return ReasonExitProgram
	case 0x0006:
		return ReasonTokenFailure
	}
	return ReasonUnknown
}

var (
	// ErrAuthenticationRejected is matched by every server rejection of the
	// credential (profile, password, security, exit-program and token
	// families). It is never retried.
	ErrAuthenticationRejected = errors.New("signon: authentication rejected")
	// ErrRequestRejected is matched when the server rejected the request
	// format (family 0x0001) or returned an unclassified code.
	ErrRequestRejected = errors.New("signon: request rejected by server")
)

// Error is a nonzero Sign-on return code. Its text contains the stage,
// code and reason, never the profile or any credential material.
type Error struct {
	// Stage is "exchange attributes" (0xF003) or "retrieve information"
	// (0xF004).
	Stage string
	Code  ReturnCode
}

func (e *Error) Error() string {
	return fmt.Sprintf("signon: %s failed with return code 0x%08X (%s)", e.Stage, uint32(e.Code), e.Code.Reason())
}

// Reason classifies the return code.
func (e *Error) Reason() Reason { return e.Code.Reason() }

// Is matches ErrAuthenticationRejected or ErrRequestRejected by family.
func (e *Error) Is(target error) bool {
	switch e.Code.Reason() {
	case ReasonRequestRejected, ReasonUnknown:
		return target == ErrRequestRejected
	}
	return target == ErrAuthenticationRejected
}
