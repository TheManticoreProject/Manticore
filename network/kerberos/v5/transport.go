// Package kerberos provides a native Kerberos client implementation for
// Active Directory authentication, without external dependencies.
// It supports RC4-HMAC and AES-CTS-HMAC-SHA1-96 encryption types.
package kerberos

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// defaultKDCPort is the standard Kerberos port.
const defaultKDCPort = 88

// defaultTimeout is the dial and I/O timeout for KDC connections.
const defaultTimeout = 10 * time.Second

// udpMaxSize is the maximum message size for UDP Kerberos. Messages larger
// than this are sent via TCP only (RFC 4120 Section 7.2.1).
const udpMaxSize = 1400

// krbErrorTag is the APPLICATION[30] tag byte that identifies a KRB-ERROR message.
const krbErrorTag = 0x7e

// kdcEndpoint is a single candidate KDC: a host (name or IP literal) and the
// port to reach it on. Discovery (DNS SRV) and explicit configuration both
// produce these; the transport fails over across a list of them in order.
type kdcEndpoint struct {
	host string
	port int
}

// String renders the endpoint as host:port for diagnostics.
func (e kdcEndpoint) String() string {
	return net.JoinHostPort(e.host, strconv.Itoa(e.port))
}

// kdcSendEndpoints sends msg to the first KDC endpoint that answers, failing over
// across the list in order. Failover is triggered only by transport-level errors
// (DNS failure, connection refused, timeout); a KDC that answers with a
// KRB-ERROR has still "answered" and that reply is returned to the caller — the
// next endpoint is not tried, because a protocol error (e.g. PREAUTH_REQUIRED)
// would be identical from every replica.
func kdcSendEndpoints(resolver *net.Resolver, endpoints []kdcEndpoint, msg []byte) ([]byte, error) {
	return kdcSendEndpointsContext(context.Background(), resolver, endpoints, msg, defaultTimeout)
}

func kdcSendEndpointsContext(ctx context.Context, resolver *net.Resolver, endpoints []kdcEndpoint, msg []byte, timeout time.Duration) ([]byte, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("kerberos: no KDC endpoints to contact")
	}
	ctx, cancel := kdcOperationContext(ctx, timeout)
	defer cancel()
	var lastErr error
	for _, ep := range endpoints {
		resp, err := kdcSendContext(ctx, resolver, ep.host, ep.port, msg, 0)
		if err == nil {
			return resp, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		lastErr = fmt.Errorf("kerberos: KDC %s: %w", ep, err)
	}
	return nil, fmt.Errorf("kerberos: all %d KDC endpoint(s) failed, last error: %w", len(endpoints), lastErr)
}

// kdcSend sends a Kerberos message to a single KDC host, resolving it to one or
// more IP addresses (A and AAAA — IPv4 and IPv6) and trying each until one
// answers. For each address it tries UDP first for small messages, then falls
// back to TCP.
//
// TCP is attempted when the UDP datagram fails, comes back empty, or carries
// KRB_ERR_RESPONSE_TOO_BIG, the KDC's explicit signal that its answer did not
// fit a datagram and must be retried over TCP (RFC 4120 Section 7.2.1). Other
// KRB-ERROR replies are complete protocol responses and are returned directly.
// UDP has no length prefix; TCP uses the RFC 4120 4-byte big-endian prefix.
func kdcSend(resolver *net.Resolver, kdc_host string, kdc_port int, msg []byte) ([]byte, error) {
	return kdcSendContext(context.Background(), resolver, kdc_host, kdc_port, msg, defaultTimeout)
}

func kdcSendContext(ctx context.Context, resolver *net.Resolver, kdc_host string, kdc_port int, msg []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := kdcOperationContext(ctx, timeout)
	defer cancel()
	addrs, err := resolveKDCAddrsContext(ctx, resolver, kdc_host)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, addr := range addrs {
		resp, err := kdcSendAddrContext(ctx, addr, kdc_port, msg, 0)
		if err == nil {
			return resp, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		lastErr = err
	}
	return nil, lastErr
}

// kdcSendAddr sends msg to a single already-resolved IP address, applying the
// UDP-first / TCP-fallback policy described on kdcSend.
func kdcSendAddr(ip string, kdc_port int, msg []byte) ([]byte, error) {
	return kdcSendAddrContext(context.Background(), ip, kdc_port, msg, defaultTimeout)
}

func kdcSendAddrContext(ctx context.Context, ip string, kdc_port int, msg []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := kdcOperationContext(ctx, timeout)
	defer cancel()
	if len(msg) <= udpMaxSize {
		resp, err := kdcSendUDPContext(ctx, ip, kdc_port, msg, 0)
		if !shouldRetryOverTCP(resp, err) {
			return resp, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	return kdcSendTCPContext(ctx, ip, kdc_port, msg, 0)
}

// shouldRetryOverTCP decides, from a UDP attempt's result, whether the request
// must be re-sent over TCP. It is true when UDP failed outright, returned an
// empty datagram, or the KDC answered with KRB_ERR_RESPONSE_TOO_BIG (RFC 4120
// Section 7.2.1). It is factored out so the UDP→TCP decision can be unit-tested
// without a live KDC.
func shouldRetryOverTCP(udpResp []byte, udpErr error) bool {
	if udpErr != nil || len(udpResp) == 0 {
		return true
	}
	return responseTooBigOverUDP(udpResp)
}

// responseTooBigOverUDP reports whether a reply is a KRB-ERROR carrying
// KRB_ERR_RESPONSE_TOO_BIG (RFC 4120 Section 7.2.1) — the KDC's explicit signal
// that the datagram answer did not fit and the request must be retried over TCP.
func responseTooBigOverUDP(resp []byte) bool {
	code, ok := krbErrorCode(resp)
	return ok && code == messages.ErrResponseTooBig
}

// krbErrorCode returns the error code of a KRB-ERROR reply, or ok=false when the
// bytes are not a KRB-ERROR (e.g. an AS-REP/TGS-REP).
func krbErrorCode(resp []byte) (int, bool) {
	if len(resp) == 0 || resp[0] != krbErrorTag {
		return 0, false
	}
	var krbErr messages.KRBError
	if _, err := krbErr.Unmarshal(resp); err != nil {
		return 0, false
	}
	return krbErr.ErrorCode, true
}

// resolveKDCAddrs resolves a KDC host to a list of IP-literal dial targets. An
// input that is already an IP literal is returned unchanged. Hostnames are
// resolved through the supplied resolver (nil uses net.DefaultResolver),
// returning both IPv4 (A) and IPv6 (AAAA) addresses so either family can be
// reached.
func resolveKDCAddrs(resolver *net.Resolver, host string) ([]string, error) {
	ctx, cancel := kdcOperationContext(context.Background(), defaultTimeout)
	defer cancel()
	return resolveKDCAddrsContext(ctx, resolver, host)
}

func resolveKDCAddrsContext(ctx context.Context, resolver *net.Resolver, host string) ([]string, error) {
	if ip := net.ParseIP(host); ip != nil {
		return []string{host}, nil
	}
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	ipAddrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("kerberos: resolve KDC host %q: %w", host, err)
	}
	if len(ipAddrs) == 0 {
		return nil, fmt.Errorf("kerberos: KDC host %q resolved to no addresses", host)
	}
	addrs := make([]string, 0, len(ipAddrs))
	for _, ia := range ipAddrs {
		addrs = append(addrs, ia.IP.String())
	}
	return addrs, nil
}

// kdcSendUDP sends msg over UDP and returns the raw response (no length prefix).
func kdcSendUDP(kdc_host string, kdc_port int, msg []byte) ([]byte, error) {
	return kdcSendUDPContext(context.Background(), kdc_host, kdc_port, msg, defaultTimeout)
}

func kdcSendUDPContext(ctx context.Context, kdc_host string, kdc_port int, msg []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := kdcOperationContext(ctx, timeout)
	defer cancel()
	addr := net.JoinHostPort(kdc_host, strconv.Itoa(kdc_port))
	conn, err := (&net.Dialer{}).DialContext(ctx, "udp", addr)
	if err != nil {
		return nil, contextTransportError(ctx, err)
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	setContextDeadline(ctx, conn)

	if _, err := conn.Write(msg); err != nil {
		return nil, fmt.Errorf("kerberos: UDP send: %w", contextTransportError(ctx, err))
	}

	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("kerberos: UDP recv: %w", contextTransportError(ctx, err))
	}
	return buf[:n], nil
}

// kdcSendTCP sends msg over TCP using the RFC 4120 4-byte big-endian length prefix.
func kdcSendTCP(kdc_host string, kdc_port int, msg []byte) ([]byte, error) {
	return kdcSendTCPContext(context.Background(), kdc_host, kdc_port, msg, defaultTimeout)
}

func kdcSendTCPContext(ctx context.Context, kdc_host string, kdc_port int, msg []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := kdcOperationContext(ctx, timeout)
	defer cancel()
	addr := net.JoinHostPort(kdc_host, strconv.Itoa(kdc_port))
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("kerberos: connect to KDC %s: %w", addr, contextTransportError(ctx, err))
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	setContextDeadline(ctx, conn)

	if err := writeTCPFramed(conn, msg); err != nil {
		return nil, contextTransportError(ctx, err)
	}
	resp, err := readTCPFramed(conn)
	return resp, contextTransportError(ctx, err)
}

func kdcOperationContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

func setContextDeadline(ctx context.Context, conn net.Conn) {
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
}

func contextTransportError(ctx context.Context, err error) error {
	if err != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

// maxTCPResponse caps the length a TCP length-prefix may declare, guarding
// against a hostile or corrupt peer requesting a huge allocation.
const maxTCPResponse = 16 * 1024 * 1024

// writeTCPFramed writes msg with the RFC 4120 Section 7.2.2 framing: a 4-byte
// big-endian length prefix followed by the message bytes. It is factored out of
// kdcSendTCP so the framing can be unit-tested against an in-memory writer.
func writeTCPFramed(w io.Writer, msg []byte) error {
	packet := make([]byte, 4+len(msg))
	binary.BigEndian.PutUint32(packet[:4], uint32(len(msg)))
	copy(packet[4:], msg)
	if _, err := w.Write(packet); err != nil {
		return fmt.Errorf("kerberos: TCP send: %w", err)
	}
	return nil
}

// readTCPFramed reads one RFC 4120 Section 7.2.2 length-prefixed message: a
// 4-byte big-endian length, then exactly that many bytes. A zero length and an
// implausibly large length are both rejected. It is factored out of kdcSendTCP
// so the de-framing can be unit-tested against an in-memory reader.
func readTCPFramed(r io.Reader) ([]byte, error) {
	resp_len_buf := make([]byte, 4)
	if err := readFull(r, resp_len_buf); err != nil {
		return nil, fmt.Errorf("kerberos: TCP read length: %w", err)
	}
	resp_len := binary.BigEndian.Uint32(resp_len_buf)
	if resp_len == 0 {
		return nil, fmt.Errorf("kerberos: KDC returned empty TCP response")
	}
	if resp_len > maxTCPResponse {
		return nil, fmt.Errorf("kerberos: KDC response too large: %d bytes", resp_len)
	}

	resp_buf := make([]byte, resp_len)
	if err := readFull(r, resp_buf); err != nil {
		return nil, fmt.Errorf("kerberos: TCP read body: %w", err)
	}
	return resp_buf, nil
}

// readFull reads exactly len(buf) bytes from r.
func readFull(r io.Reader, buf []byte) error {
	_, err := io.ReadFull(r, buf)
	return err
}
