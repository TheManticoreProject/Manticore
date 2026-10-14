package hostserver

import (
	"errors"
	"math"
	"sync"
)

// ErrCorrelationExhausted is returned once every nonzero 32-bit correlation
// ID has been issued. IDs never wrap through zero.
var ErrCorrelationExhausted = errors.New("hostserver: correlation IDs exhausted")

// Correlator issues monotonically increasing, nonzero correlation IDs,
// starting at 1. Zero is reserved for startup and end exchanges. The zero
// value is ready to use and safe for concurrent use.
type Correlator struct {
	mu   sync.Mutex
	last uint32
}

// Next returns the next correlation ID.
func (c *Correlator) Next() (uint32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.last == math.MaxUint32 {
		return 0, ErrCorrelationExhausted
	}
	c.last++
	return c.last, nil
}
