package client

import (
	"fmt"
	"net"

	"github.com/TheManticoreProject/Manticore/network/smb"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	smb1dialects "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	smb1msg "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	smb1commands "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	smb1flags "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	smb1flags2 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	smb2 "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	smb2msg "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	smb2commands "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// engineSupportsSMB2 reports whether the SMB2 engine can drive version v. The
// engine negotiates SMB 2.0.2; the 2.0 family marker and 2.1 map onto it, while
// 3.x is not yet supported (a dedicated engine arrives in Phase 7).
func engineSupportsSMB2(v smb.SMBProtocolVersion) bool {
	switch v {
	case smb.SMB_VERSION_2_0, smb.SMB_VERSION_2_0_2, smb.SMB_VERSION_2_1:
		return true
	}
	return false
}

// dialStrictOrder tries the preferred versions in order, reconnecting between
// attempts, and returns the first the server accepts. Consecutive versions that
// resolve to the same engine attempt (all the SMB2 dialects) are collapsed so the
// same connection is not retried redundantly. This honors the caller's order
// exactly — a lower dialect listed first wins even when the server supports a
// higher one.
func dialStrictOrder(host string, ip net.IP, port int, opts Options, prefs []smb.SMBProtocolVersion) (*Client, error) {
	triedSMB1, triedSMB2 := false, false
	var lastErr error
	attempted := false

	for _, v := range prefs {
		var (
			c   *Client
			err error
		)
		switch {
		case v == smb.SMB_VERSION_1_0:
			if triedSMB1 {
				continue
			}
			triedSMB1 = true
			c, err = dialSMB1(ip, host, port, opts)
		case engineSupportsSMB2(v):
			if triedSMB2 {
				continue
			}
			triedSMB2 = true
			c, err = dialSMB2(ip, host, port, opts)
		default:
			// e.g. SMB 3.x — no engine yet; skip.
			continue
		}

		attempted = true
		if err == nil {
			return c, nil
		}
		lastErr = err
	}

	if !attempted {
		return nil, fmt.Errorf("no supported protocol version in preference list for %s", host)
	}
	return nil, fmt.Errorf("no preferred dialect accepted by %s: %w", host, lastErr)
}

// wantedFamilies reports which engine families a preference list calls for: SMB1
// when SMB_VERSION_1_0 is present, and SMB2 when any engine-supported SMB2 dialect
// is present. Unsupported versions (3.x) contribute nothing.
func wantedFamilies(prefs []smb.SMBProtocolVersion) (smb1, smb2 bool) {
	for _, v := range prefs {
		switch {
		case v == smb.SMB_VERSION_1_0:
			smb1 = true
		case engineSupportsSMB2(v):
			smb2 = true
		}
	}
	return smb1, smb2
}

// dialHighestInSet performs a single SMB1 multi-protocol negotiate — one request
// offering the requested SMB1 and SMB2 dialect markers — and binds the engine for
// whatever the server selects (its highest within the offered set). The reply is
// dispatched on its protocol marker; an SMB2 wildcard reply triggers a second,
// native SMB2 negotiate to pin the concrete dialect.
func dialHighestInSet(host string, ip net.IP, port int, opts Options, prefs []smb.SMBProtocolVersion) (*Client, error) {
	wantSMB1, wantSMB2 := wantedFamilies(prefs)
	if !wantSMB1 && !wantSMB2 {
		return nil, fmt.Errorf("no supported protocol version in preference list for %s", host)
	}

	t := transport.NewTransport("tcp")
	if err := t.Connect(ip, port); err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", host, err)
	}

	raw, offered, err := sendMultiProtocolNegotiate(t, wantSMB1, wantSMB2)
	if err != nil {
		t.Close()
		return nil, err
	}
	if len(raw) < 4 || string(raw[1:4]) != "SMB" {
		t.Close()
		return nil, fmt.Errorf("unrecognized negotiate response from %s (% x)", host, raw[:min(4, len(raw))])
	}

	switch raw[0] {
	case 0xFF: // SMB1
		if !wantSMB1 {
			t.Close()
			return nil, fmt.Errorf("server selected SMB1 but it was not requested")
		}
		return finishSMB1(t, ip, host, port, opts, raw, offered)
	case 0xFE: // SMB2
		if !wantSMB2 {
			t.Close()
			return nil, fmt.Errorf("server selected SMB2 but it was not requested")
		}
		return finishSMB2(t, ip, host, port, opts, raw)
	default:
		t.Close()
		return nil, fmt.Errorf("unrecognized negotiate response protocol marker from %s (% x)", host, raw[:4])
	}
}

// sendMultiProtocolNegotiate builds and sends an SMB1 SMB_COM_NEGOTIATE offering
// the requested dialect markers, and returns the raw response together with the
// negotiate request command (its offered dialect list is needed to resolve an
// SMB1 selection). The wildcard "SMB 2.???" marker advertises SMB2 support.
func sendMultiProtocolNegotiate(t transport.Transport, wantSMB1, wantSMB2 bool) ([]byte, *smb1commands.NegotiateRequest, error) {
	req := smb1msg.NewMessage()
	req.Header.SetFlags(smb1flags.FLAGS_CANONICALIZED_PATHS | smb1flags.FLAGS_CASE_INSENSITIVE)
	req.Header.SetFlags2(smb1flags2.FLAGS2_UNICODE | smb1flags2.FLAGS2_NT_STATUS_ERROR_CODES | smb1flags2.FLAGS2_EXTENDED_SECURITY | smb1flags2.FLAGS2_LONG_NAMES_ALLOWED)

	neg := smb1commands.NewNegotiateRequest()
	if wantSMB1 {
		neg.Dialects.AddDialect(smb1dialects.DIALECT_NT_LM_0_12)
	}
	if wantSMB2 {
		// The SMB2 engine's ceiling is SMB 2.0.2, so offer the "SMB 2.002" marker
		// rather than the "SMB 2.???" wildcard. A wildcard offer declares 2.1+
		// support, after which the server replies with the wildcard revision and
		// expects a follow-up negotiate offering 2.1+ — a 2.0.2-only follow-up is
		// reset by Windows Server. "SMB 2.002" makes the server return a concrete
		// SMB 2.0.2 negotiate response directly, needing no second leg.
		neg.Dialects.AddDialect(smb2DialectString2002)
	}
	req.AddCommand(neg)

	marshalled, err := req.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal multi-protocol negotiate: %w", err)
	}
	if _, err := t.Send(marshalled); err != nil {
		return nil, nil, fmt.Errorf("failed to send multi-protocol negotiate: %w", err)
	}
	raw, err := t.Receive()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to receive negotiate response: %w", err)
	}
	return raw, neg, nil
}

// finishSMB1 parses the SMB1 negotiate response, hands the live transport to the
// SMB1 engine, and applies the response (SMB1 cannot renegotiate).
func finishSMB1(t transport.Transport, ip net.IP, host string, port int, opts Options, raw []byte, offered *smb1commands.NegotiateRequest) (*Client, error) {
	respMsg := smb1msg.NewMessage()
	respMsg.AddCommand(offered) // attach the negotiate command so Unmarshal decodes the response
	if err := respMsg.Unmarshal(raw); err != nil {
		t.Close()
		return nil, fmt.Errorf("failed to parse SMB1 negotiate response: %w", err)
	}
	resp, ok := respMsg.Command.(*smb1commands.NegotiateResponse)
	if !ok {
		t.Close()
		return nil, fmt.Errorf("unexpected SMB1 negotiate response command: %T", respMsg.Command)
	}

	engine := smb1.NewFromTransport(t, ip, port)
	engine.NativeOS = "Manticore"
	engine.NativeLanMan = "Manticore"
	if opts.Workstation != "" {
		engine.Workstation = opts.Workstation
	}
	if err := engine.ApplyNegotiateResponse(resp, offered.Dialects); err != nil {
		t.Close()
		return nil, err
	}
	return &Client{backend: newSMB1Backend(engine), host: host, port: port, opts: opts}, nil
}

// finishSMB2 parses the SMB2 negotiate response and hands the live transport to
// the SMB2 engine. Because the multi-protocol negotiate offers the "SMB 2.002"
// marker, the server returns a concrete SMB 2.0.2 negotiate response, which is
// applied directly. A dialect the engine cannot drive is rejected rather than
// applied.
func finishSMB2(t transport.Transport, ip net.IP, host string, port int, opts Options, raw []byte) (*Client, error) {
	respMsg := smb2msg.NewMessage()
	if _, err := respMsg.Header.Unmarshal(raw); err != nil {
		t.Close()
		return nil, fmt.Errorf("failed to parse SMB2 negotiate response header: %w", err)
	}
	if _, err := respMsg.Unmarshal(raw); err != nil {
		t.Close()
		return nil, fmt.Errorf("failed to parse SMB2 negotiate response: %w", err)
	}
	resp, ok := respMsg.Command.(*smb2commands.NegotiateResponse)
	if !ok {
		t.Close()
		return nil, fmt.Errorf("unexpected SMB2 negotiate response command: %T", respMsg.Command)
	}

	if v, ok := versionForSMB2Dialect(resp.DialectRevision); !ok || !engineSupportsSMB2(v) {
		t.Close()
		return nil, fmt.Errorf("server selected SMB2 dialect 0x%04x, which the engine cannot drive", uint16(resp.DialectRevision))
	}

	engine := smb2.NewFromTransport(t, ip, port)
	if opts.Workstation != "" {
		engine.Workstation = opts.Workstation
	}
	engine.ApplyNegotiateResponse(resp)
	return &Client{backend: newSMB2Backend(engine), host: host, port: port, opts: opts}, nil
}
