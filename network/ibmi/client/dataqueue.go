package client

import (
	"context"
	"errors"
	"fmt"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/dataqueue"
	"github.com/TheManticoreProject/Manticore/network/ibmi/servicemapper"
)

// OpenDataQueue starts a Data Queue server job for cred, which must be the
// credential Connect validated (a password or a multiple-use profile
// token). It needs a ready Client for the password family and system CCSID
// learned from Sign-on, resolves the Data Queue endpoint under the
// configured transport policy, and returns an independent Session: closing
// the Client does not close it. cred is not retained.
func (c *Client) OpenDataQueue(ctx context.Context, cred ibminet.Credential) (*dataqueue.Session, error) {
	if cred == nil {
		return nil, errors.New("client: nil credential")
	}
	c.mu.Lock()
	ready, info := c.state == stateReady, c.info
	c.mu.Unlock()
	if !ready {
		return nil, ErrNotReady
	}
	catalog, err := (&servicemapper.Client{Dial: c.cfg.Dial, Timeout: c.cfg.DialTimeout}).Discover(ctx, servicemapper.DiscoverOptions{
		Host:          c.cfg.Host,
		MapperPort:    c.cfg.MapperPort,
		Policy:        c.cfg.Transport,
		Services:      []ibminet.Service{ibminet.ServiceDataQueue},
		DataQueuePort: c.cfg.DataQueuePort,
	})
	if err != nil {
		return nil, fmt.Errorf("client: discover data queue: %w", err)
	}
	conn, err := c.cfg.dialer().DialEndpoint(ctx, catalog.DataQueue)
	if err != nil {
		return nil, err
	}
	return dataqueue.Start(ctx, conn, cred, info.PasswordFamily, dataqueue.StartOptions{SystemCCSID: info.SystemCCSID})
}
