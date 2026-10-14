package client

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
)

// reservation is the callback-scoped handle passed to exclusive. It is valid
// only until the callback returns; afterwards every use fails with
// ErrReservationReleased, including from goroutines the callback started.
type reservation struct {
	session  *rmtcmd.Session
	released atomic.Bool
}

func (r *reservation) Session() (*rmtcmd.Session, error) {
	if r.released.Load() {
		return nil, ErrReservationReleased
	}
	return r.session, nil
}

// exclusive runs fn while holding the lifecycle lock, so no other operation
// and no Close can interleave with it. When fn returns or panics, the
// reservation is released, the lock is released, and a session whose stream
// became unusable moves the client to poisoned. A panic is then propagated
// unchanged. fn must not call locking Client methods, which would deadlock.
func (c *Client) exclusive(ctx context.Context, fn func(*reservation) error) error {
	if err := c.acquire(ctx); err != nil {
		return err
	}
	defer c.release()
	if st := c.getState(); st != stateReady {
		return fmt.Errorf("%w (state %s)", ErrNotReady, st)
	}
	c.mu.Lock()
	r := &reservation{session: c.session}
	c.mu.Unlock()
	defer func() {
		r.released.Store(true)
		if r.session.Err() != nil {
			c.setState(statePoisoned)
		}
	}()
	return fn(r)
}
