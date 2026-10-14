package client

import (
	"context"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
)

var (
	_ rpc.Caller          = (*Client)(nil)
	_ rpc.ExclusiveCaller = (*Client)(nil)
)

// Call runs one IBM i program call on the session. It implements
// rpc.Caller; see rmtcmd.Session.Call for the result and error contract.
// A failure that leaves the stream unusable moves the client to poisoned,
// after which every operation fails with ErrNotReady; Close still releases
// the socket.
func (c *Client) Call(ctx context.Context, call rpc.Call) (*rpc.Result, error) {
	var res *rpc.Result
	err := c.exclusive(ctx, func(r *reservation) error {
		s, err := r.Session()
		if err != nil {
			return err
		}
		res, err = s.Call(ctx, call)
		return err
	})
	return res, err
}

// RunCommand executes raw CL source on the session.
//
// This is a privileged operation: command reaches IBM i verbatim, so
// caller-controlled text inside it is CL injection. Prefer a typed API built
// on Call. A nonzero return code returns the result together with an
// *rmtcmd.NativeError; see rmtcmd.Session.RunCommand.
func (c *Client) RunCommand(ctx context.Context, command string, opts rmtcmd.RunCommandOptions) (*rmtcmd.CommandResult, error) {
	var res *rmtcmd.CommandResult
	err := c.exclusive(ctx, func(r *reservation) error {
		s, err := r.Session()
		if err != nil {
			return err
		}
		res, err = s.RunCommand(ctx, command, opts)
		return err
	})
	return res, err
}

// scopedCaller is the rpc.Caller handed to ExecuteExclusive callbacks.
type scopedCaller struct{ r *reservation }

func (s scopedCaller) Call(ctx context.Context, call rpc.Call) (*rpc.Result, error) {
	session, err := s.r.Session()
	if err != nil {
		return nil, err
	}
	return session.Call(ctx, call)
}

// ExecuteExclusive runs fn with a Caller bound to this session while holding
// the lifecycle lock, so the calls fn makes cannot be interleaved with any
// other operation or with Close. It implements rpc.ExclusiveCaller.
//
// The Caller is valid only until fn returns; afterwards it fails with
// ErrReservationReleased. It must not be used from goroutines that outlive
// fn. fn must not call methods of this Client, which would deadlock. If fn
// panics the lock is released and the panic propagates unchanged.
func (c *Client) ExecuteExclusive(ctx context.Context, fn func(rpc.Caller) error) error {
	return c.exclusive(ctx, func(r *reservation) error { return fn(scopedCaller{r: r}) })
}
