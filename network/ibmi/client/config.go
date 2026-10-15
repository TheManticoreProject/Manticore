package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
	"github.com/TheManticoreProject/Manticore/network/ibmi/servicemapper"
)

// Defaults applied to zero Config fields, and the upper bounds enforced.
const (
	DefaultDialTimeout      = 10 * time.Second
	DefaultOperationTimeout = 30 * time.Second
	DefaultCleanupTimeout   = 5 * time.Second

	MaxDialTimeout      = 5 * time.Minute
	MaxOperationTimeout = time.Hour
)

// ErrConfig is matched (via errors.Is) by every Config validation failure.
var ErrConfig = errors.New("client: invalid configuration")

// Config describes how to reach and talk to one IBM i system. Zero
// durations and sizes select the defaults; negative or out-of-range values
// are rejected rather than normalized.
type Config struct {
	// Host is the IBM i host name or address. Required.
	Host string
	// MapperPort is the service-mapper port (default 449).
	MapperPort int
	// SignonPort, RemoteCommandPort and DataQueuePort, when nonzero, skip
	// the mapper lookup for that service only. They do not change the
	// transport policy.
	SignonPort        int
	RemoteCommandPort int
	DataQueuePort     int
	// Transport must be TransportTLSRequired or TransportCleartextExplicit.
	Transport ibminet.TransportPolicy
	// TLS configures verification for TLSRequired.
	TLS hostserver.TLSOptions

	DialTimeout      time.Duration
	OperationTimeout time.Duration
	// CleanupTimeout bounds end-server requests; it may not exceed
	// OperationTimeout.
	CleanupTimeout time.Duration
	// MaxFrame caps every host-server frame (default 16 MiB, 20 B..64 MiB).
	MaxFrame uint32

	// NLV is the four-digit national-language version (default 2924).
	NLV string
	// JobCCSIDOverride replaces the negotiated Remote Command job CCSID; it
	// must be a supported CCSID and always adds a SessionInfo warning.
	JobCCSIDOverride uint32

	// MaxReceiver caps each program-call output (default 16 MiB, at most
	// 64 MiB). MaxMessageText caps each message's text (default 1 MiB).
	// MaxRLEExpansion caps one decompressed output (default 16 MiB, at
	// most 64 MiB). MaxMessages caps the messages in one reply (default
	// 256, at most 65535).
	MaxReceiver     uint32
	MaxMessageText  uint32
	MaxRLEExpansion uint32
	MaxMessages     int

	// Dial replaces net.Dialer.DialContext, for tests and custom routing.
	Dial ibminet.DialFunc
	// Tracer receives connection metadata events (never payloads).
	Tracer hostserver.Tracer
}

func checkPort(name string, port int, optional bool) error {
	if optional && port == 0 {
		return nil
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("%w: %s %d outside 1-65535", ErrConfig, name, port)
	}
	return nil
}

func resolveDuration(name string, d, def, max time.Duration) (time.Duration, error) {
	switch {
	case d == 0:
		return def, nil
	case d < 0 || d > max:
		return 0, fmt.Errorf("%w: %s %s outside (0, %s]", ErrConfig, name, d, max)
	}
	return d, nil
}

// resolve validates c and returns it with defaults applied.
func (c Config) resolve() (Config, error) {
	if c.Host == "" {
		return Config{}, fmt.Errorf("%w: Host is required", ErrConfig)
	}
	if err := c.Transport.Validate(); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrConfig, err)
	}
	if c.MapperPort == 0 {
		c.MapperPort = servicemapper.DefaultPort
	}
	if err := checkPort("MapperPort", c.MapperPort, false); err != nil {
		return Config{}, err
	}
	if err := checkPort("SignonPort", c.SignonPort, true); err != nil {
		return Config{}, err
	}
	if err := checkPort("RemoteCommandPort", c.RemoteCommandPort, true); err != nil {
		return Config{}, err
	}
	if err := checkPort("DataQueuePort", c.DataQueuePort, true); err != nil {
		return Config{}, err
	}
	var err error
	if c.DialTimeout, err = resolveDuration("DialTimeout", c.DialTimeout, DefaultDialTimeout, MaxDialTimeout); err != nil {
		return Config{}, err
	}
	if c.OperationTimeout, err = resolveDuration("OperationTimeout", c.OperationTimeout, DefaultOperationTimeout, MaxOperationTimeout); err != nil {
		return Config{}, err
	}
	if c.CleanupTimeout, err = resolveDuration("CleanupTimeout", c.CleanupTimeout, DefaultCleanupTimeout, c.OperationTimeout); err != nil {
		return Config{}, err
	}
	if c.MaxFrame, err = hostserver.ResolveMaxFrame(c.MaxFrame); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrConfig, err)
	}
	if c.NLV == "" {
		c.NLV = message.DefaultNLV
	}
	if _, err := message.ExchangeAttributesFrame(1, c.NLV); err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrConfig, err)
	}
	limits, err := c.limits().Resolve()
	if err != nil {
		return Config{}, fmt.Errorf("%w: %w", ErrConfig, err)
	}
	if limits.MaxMessageText > c.MaxFrame {
		return Config{}, fmt.Errorf("%w: MaxMessageText %d exceeds MaxFrame %d", ErrConfig, limits.MaxMessageText, c.MaxFrame)
	}
	c.MaxReceiver, c.MaxMessageText, c.MaxRLEExpansion, c.MaxMessages = limits.MaxReceiver, limits.MaxMessageText, limits.MaxRLEExpansion, limits.MaxMessages
	if c.JobCCSIDOverride != 0 {
		if _, err := ccsid.Lookup(ccsid.CCSID(c.JobCCSIDOverride)); err != nil {
			return Config{}, fmt.Errorf("%w: JobCCSIDOverride: %w", ErrConfig, err)
		}
	}
	if c.Transport.TLS() {
		if _, err := c.TLS.Config(c.Host); err != nil {
			return Config{}, fmt.Errorf("%w: %w", ErrConfig, err)
		}
	}
	return c, nil
}

func (c Config) limits() rmtcmd.Limits {
	return rmtcmd.Limits{MaxReceiver: c.MaxReceiver, MaxMessageText: c.MaxMessageText, MaxRLEExpansion: c.MaxRLEExpansion, MaxMessages: c.MaxMessages}
}

func (c Config) dialer() *hostserver.Dialer {
	return &hostserver.Dialer{
		Dial:    c.Dial,
		Timeout: c.DialTimeout,
		TLS:     c.TLS,
		Conn:    hostserver.Options{MaxFrame: c.MaxFrame, OperationTimeout: c.OperationTimeout, Tracer: c.Tracer},
	}
}
