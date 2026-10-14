package rpc

import "context"

// Caller runs one program call. Implementations validate the call, never
// retry or replay it on their own, and return a Result aligned with its
// parameters. They may return a non-nil Result together with an error when
// the program ran but reported failure.
type Caller interface {
	Call(context.Context, Call) (*Result, error)
}

// ExclusiveCaller runs a sequence of calls that must not interleave with
// other work on the same session, for example opening, paging and closing a
// list handle that belongs to one server job.
//
// The Caller passed to fn is valid only until fn returns. It must not be
// retained or used from another goroutine. fn must not call methods of the
// ExclusiveCaller itself, which would deadlock.
type ExclusiveCaller interface {
	ExecuteExclusive(context.Context, func(Caller) error) error
}
