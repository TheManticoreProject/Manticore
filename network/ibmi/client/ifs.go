package client

import (
	"context"
	"errors"
	"fmt"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/ifs"
	"github.com/TheManticoreProject/Manticore/network/ibmi/servicemapper"
)

// OpenIFS starts a File server session (package network/ibmi/ifs) for
// cred, which must be the credential Connect validated (a password or a
// multiple-use profile token). Like OpenDataQueue it needs a ready Client,
// resolves the File endpoint under the configured transport policy and
// returns an independent Session that closing the Client does not close.
// cred is not retained.
func (c *Client) OpenIFS(ctx context.Context, cred ibminet.Credential) (*ifs.Session, error) {
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
		Host:       c.cfg.Host,
		MapperPort: c.cfg.MapperPort,
		Policy:     c.cfg.Transport,
		Services:   []ibminet.Service{ibminet.ServiceFile},
		FilePort:   c.cfg.FilePort,
	})
	if err != nil {
		return nil, fmt.Errorf("client: discover file server: %w", err)
	}
	conn, err := c.cfg.dialer().DialEndpoint(ctx, catalog.File)
	if err != nil {
		return nil, err
	}
	return ifs.Start(ctx, conn, cred, info.PasswordFamily, ifs.StartOptions{SystemCCSID: info.SystemCCSID})
}
