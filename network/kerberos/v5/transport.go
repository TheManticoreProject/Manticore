// Package kerberos provides a native Kerberos client implementation for
// Active Directory authentication, without external dependencies.
// It supports RC4-HMAC and AES-CTS-HMAC-SHA1-96 encryption types.
package kerberos

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
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

// kdcSend sends a Kerberos message to the KDC.
// It tries UDP first for small messages, then falls back to TCP.
// If UDP returns a KRB-ERROR, TCP is always attempted — Windows KDCs sometimes
// return stale or protocol-level errors over UDP even when TCP would succeed.
// UDP has no length prefix; TCP uses the RFC 4120 4-byte big-endian prefix.
func kdcSend(kdc_host string, kdc_port int, msg []byte) ([]byte, error) {
	if len(msg) <= udpMaxSize {
		resp, err := kdcSendUDP(kdc_host, kdc_port, msg)
		if err == nil && len(resp) > 0 && resp[0] != krbErrorTag {
			return resp, nil
		}
	}
	return kdcSendTCP(kdc_host, kdc_port, msg)
}

// kdcSendUDP sends msg over UDP and returns the raw response (no length prefix).
func kdcSendUDP(kdc_host string, kdc_port int, msg []byte) ([]byte, error) {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(kdc_host, fmt.Sprintf("%d", kdc_port)))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(defaultTimeout))

	if _, err := conn.Write(msg); err != nil {
		return nil, fmt.Errorf("kerberos: UDP send: %w", err)
	}

	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("kerberos: UDP recv: %w", err)
	}
	return buf[:n], nil
}

// kdcSendTCP sends msg over TCP using the RFC 4120 4-byte big-endian length prefix.
func kdcSendTCP(kdc_host string, kdc_port int, msg []byte) ([]byte, error) {
	addr := net.JoinHostPort(kdc_host, fmt.Sprintf("%d", kdc_port))
	conn, err := net.DialTimeout("tcp", addr, defaultTimeout)
	if err != nil {
		return nil, fmt.Errorf("kerberos: connect to KDC %s: %w", addr, err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(defaultTimeout))

	len_buf := make([]byte, 4)
	binary.BigEndian.PutUint32(len_buf, uint32(len(msg)))
	packet := make([]byte, 4+len(msg))
	copy(packet[:4], len_buf)
	copy(packet[4:], msg)

	if _, err := conn.Write(packet); err != nil {
		return nil, fmt.Errorf("kerberos: TCP send: %w", err)
	}

	resp_len_buf := make([]byte, 4)
	if err := readFull(conn, resp_len_buf); err != nil {
		return nil, fmt.Errorf("kerberos: TCP read length: %w", err)
	}
	resp_len := binary.BigEndian.Uint32(resp_len_buf)
	if resp_len == 0 {
		return nil, fmt.Errorf("kerberos: KDC returned empty TCP response")
	}
	if resp_len > 16*1024*1024 {
		return nil, fmt.Errorf("kerberos: KDC response too large: %d bytes", resp_len)
	}

	resp_buf := make([]byte, resp_len)
	if err := readFull(conn, resp_buf); err != nil {
		return nil, fmt.Errorf("kerberos: TCP read body: %w", err)
	}
	return resp_buf, nil
}

// readFull reads exactly len(buf) bytes from conn.
func readFull(conn net.Conn, buf []byte) error {
	_, err := io.ReadFull(conn, buf)
	return err
}
