package servicemapper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
)

// DefaultTimeout bounds one lookup, connect included, when Client.Timeout is
// 0.
const DefaultTimeout = 10 * time.Second

// Client performs mapper lookups. The zero value is ready to use.
type Client struct {
	// Dial opens the mapper connection; nil uses net.Dialer.DialContext.
	Dial ibminet.DialFunc
	// Timeout bounds each lookup; 0 means DefaultTimeout.
	Timeout time.Duration
}

// Lookup asks the mapper at host:port (0 means DefaultPort) for the port of
// one service name, over its own connection.
func (c *Client) Lookup(ctx context.Context, host string, port int, name string) (int, error) {
	req, err := EncodeRequest(name)
	if err != nil {
		return 0, err
	}
	if host == "" {
		return 0, errors.New("servicemapper: empty host")
	}
	if port == 0 {
		port = DefaultPort
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("servicemapper: mapper port %d outside 1-65535", port)
	}
	if c.Timeout < 0 {
		return 0, fmt.Errorf("servicemapper: negative timeout %s", c.Timeout)
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dial := c.Dial
	if dial == nil {
		dial = (&net.Dialer{}).DialContext
	}
	conn, err := dial(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return 0, fmt.Errorf("servicemapper: dial mapper for %s: %w", name, err)
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Unix(1, 0)) })
	defer stop()
	if d, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(d)
	}

	if _, err := conn.Write(req); err != nil {
		return 0, fmt.Errorf("servicemapper: send %s: %w", name, ctxErr(ctx, err))
	}
	reply := make([]byte, ReplyLength)
	if _, err := io.ReadFull(conn, reply[:1]); err != nil {
		return 0, fmt.Errorf("%w: no reply for %s: %w", ErrProtocol, name, ctxErr(ctx, err))
	}
	if reply[0] != replySuccess {
		return DecodeReply(name, reply[:1])
	}
	if n, err := io.ReadFull(conn, reply[1:]); err != nil {
		return 0, fmt.Errorf("%w: %d-byte reply for %s: %w", ErrProtocol, 1+n, name, ctxErr(ctx, err))
	}
	return DecodeReply(name, reply)
}

func ctxErr(ctx context.Context, err error) error {
	if e := ctx.Err(); e != nil {
		return fmt.Errorf("%w (%w)", e, err)
	}
	return err
}

// DiscoverOptions selects what Discover resolves.
type DiscoverOptions struct {
	Host string
	// MapperPort is the mapper port; 0 means DefaultPort.
	MapperPort int
	// Policy must be TLSRequired or CleartextExplicit.
	Policy ibminet.TransportPolicy
	// SignonPort and RemoteCommandPort, when nonzero, are used as-is and
	// skip only that service's lookup. They do not change Policy.
	SignonPort        int
	RemoteCommandPort int
}

// Discover resolves the Sign-on and Remote Command endpoints under
// opts.Policy, one mapper connection per lookup. It may return a partial
// catalog together with a joined error naming each failed service; a
// failed TLS lookup is never replaced by a cleartext service.
func (c *Client) Discover(ctx context.Context, opts DiscoverOptions) (ibminet.ServiceCatalog, error) {
	if err := opts.Policy.Validate(); err != nil {
		return ibminet.ServiceCatalog{}, err
	}
	if opts.Host == "" {
		return ibminet.ServiceCatalog{}, errors.New("servicemapper: empty host")
	}
	useTLS := opts.Policy.TLS()
	var catalog ibminet.ServiceCatalog
	var errs []error
	for _, svc := range ibminet.Services() {
		explicit := opts.SignonPort
		if svc == ibminet.ServiceRemoteCommand {
			explicit = opts.RemoteCommandPort
		}
		ep := ibminet.Endpoint{Service: svc, Name: svc.MapperName(useTLS), Host: opts.Host, TLS: useTLS}
		if explicit != 0 {
			ep.Port, ep.Source = explicit, ibminet.EndpointExplicit
		} else {
			port, err := c.Lookup(ctx, opts.Host, opts.MapperPort, ep.Name)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", svc, err))
				catalog.Missing = append(catalog.Missing, svc)
				continue
			}
			ep.Port, ep.Source = port, ibminet.EndpointServiceMapper
		}
		if err := ep.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", svc, err))
			catalog.Missing = append(catalog.Missing, svc)
			continue
		}
		if svc == ibminet.ServiceSignon {
			catalog.Signon = ep
		} else {
			catalog.RemoteCommand = ep
		}
	}
	return catalog, errors.Join(errs...)
}
