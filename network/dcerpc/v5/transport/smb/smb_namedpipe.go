// Package smb implements the DCE/RPC ncacn_np protocol sequence: DCE/RPC directly
// over an SMB named pipe.
//
// Per [MS-RPCE] section 2.1.1.2 (SMB (NCACN_NP)), "All PDUs sent over SMB MUST be
// sent as named pipe writes ... and PDUs to be received MUST be received as named
// pipe reads." This transport therefore implements SendReceive as a pipe write
// followed by a pipe read, wrapping an already-connected and authenticated SMB v1.0
// client whose tree is connected to the IPC$ share. (The single-round-trip
// TransactNamedPipe optimization for synchronous calls is an explicit MAY in
// [MS-RPCE] and is not used here.)
package smb

import (
	"fmt"

	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// Default fragment sizes proposed at Bind time. 4280 (0x10B8) is the classic Windows
// default for RPC over named pipes; it is a safe, widely accepted value. Callers that
// want larger fragments can override via SetMaxFrag before Connect.
const (
	DefaultMaxXmitFrag uint16 = 4280
	DefaultMaxRecvFrag uint16 = 4280
)

// pipeConn is the subset of *client.Client that the named-pipe transport relies on.
// Expressing it as an interface keeps the transport unit-testable without a live SMB
// server: tests substitute a fake. *client.Client satisfies it.
type pipeConn interface {
	OpenFile(path string, desiredAccess, shareAccess, createDisp, createOptions uint32) (client.FID, error)
	ReadFile(fid client.FID, offset uint64, maxLen uint32) ([]byte, error)
	WriteFile(fid client.FID, offset uint64, data []byte) (int, error)
	CloseFile(fid client.FID) error
}

// SMBNamedPipe is a DCE/RPC transport over an SMB named pipe (ncacn_np).
type SMBNamedPipe struct {
	conn     pipeConn
	pipeName string
	fid      client.FID
	opened   bool
	maxXmit  uint16
	maxRecv  uint16
}

// Compile-time assertion that SMBNamedPipe implements the DCE/RPC transport contract.
var _ dcerpctransport.Transport = (*SMBNamedPipe)(nil)

// New creates a named-pipe transport over the supplied SMB client. The client MUST
// already be connected, have an authenticated session (SessionSetup), and have a tree
// connected to the IPC$ share (TreeConnect) before Connect is called.
//
// pipeName is the pipe endpoint, for example `\PIPE\lsarpc` or `\PIPE\epmapper`. A
// leading backslash is added if absent; pipe names are case-insensitive on Windows.
func New(c *client.Client, pipeName string) *SMBNamedPipe {
	return newWithConn(c, pipeName)
}

// newWithConn is the internal constructor used by New and by tests; it accepts any
// value satisfying pipeConn.
func newWithConn(conn pipeConn, pipeName string) *SMBNamedPipe {
	if len(pipeName) == 0 || pipeName[0] != '\\' {
		pipeName = "\\" + pipeName
	}
	return &SMBNamedPipe{
		conn:     conn,
		pipeName: pipeName,
		maxXmit:  DefaultMaxXmitFrag,
		maxRecv:  DefaultMaxRecvFrag,
	}
}

// SetMaxFrag overrides the transmit and receive fragment sizes proposed at Bind time.
// It must be called before Connect to take effect.
func (p *SMBNamedPipe) SetMaxFrag(xmit, recv uint16) {
	p.maxXmit = xmit
	p.maxRecv = recv
}

// Connect opens the named pipe on the SMB tree. It is idempotent.
func (p *SMBNamedPipe) Connect() error {
	if p.opened {
		return nil
	}

	fid, err := p.conn.OpenFile(
		p.pipeName,
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to open named pipe %q: %w", p.pipeName, err)
	}

	p.fid = fid
	p.opened = true
	return nil
}

// Send writes a complete PDU to the pipe. The pipe offset is meaningless, so a zero
// offset is used.
func (p *SMBNamedPipe) Send(pdu []byte) error {
	if !p.opened {
		return fmt.Errorf("named pipe %q is not open: call Connect first", p.pipeName)
	}
	if len(pdu) == 0 {
		return fmt.Errorf("refusing to send an empty PDU on named pipe %q", p.pipeName)
	}

	written, err := p.conn.WriteFile(p.fid, 0, pdu)
	if err != nil {
		return fmt.Errorf("failed to write PDU to named pipe %q: %w", p.pipeName, err)
	}
	if written != len(pdu) {
		return fmt.Errorf("short write to named pipe %q: wrote %d of %d bytes", p.pipeName, written, len(pdu))
	}
	return nil
}

// Recv reads up to MaxRecvFrag bytes from the pipe. The pipe offset is meaningless,
// so a zero offset is used.
func (p *SMBNamedPipe) Recv() ([]byte, error) {
	if !p.opened {
		return nil, fmt.Errorf("named pipe %q is not open: call Connect first", p.pipeName)
	}
	data, err := p.conn.ReadFile(p.fid, 0, uint32(p.maxRecv))
	if err != nil {
		return nil, fmt.Errorf("failed to read from named pipe %q: %w", p.pipeName, err)
	}
	return data, nil
}

// Close closes the pipe handle. The underlying SMB session and tree connect are left
// untouched; their lifecycle belongs to the caller that supplied the SMB client.
// Close is idempotent.
func (p *SMBNamedPipe) Close() error {
	if !p.opened {
		return nil
	}

	err := p.conn.CloseFile(p.fid)
	p.opened = false
	if err != nil {
		return fmt.Errorf("failed to close named pipe %q: %w", p.pipeName, err)
	}
	return nil
}

// MaxXmitFrag returns the proposed maximum transmit fragment size.
func (p *SMBNamedPipe) MaxXmitFrag() uint16 { return p.maxXmit }

// MaxRecvFrag returns the proposed maximum receive fragment size.
func (p *SMBNamedPipe) MaxRecvFrag() uint16 { return p.maxRecv }
