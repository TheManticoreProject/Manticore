// Package smb2 implements the DCE/RPC ncacn_np protocol sequence over an SMB 2.x
// named pipe.
//
// Per [MS-RPCE] 2.1.1.2, PDUs are exchanged as named-pipe writes and reads. This
// transport uses the SMB2 FSCTL_PIPE_TRANSCEIVE optimization (the explicit MAY in
// [MS-RPCE]): the last request PDU before a read is written and the reply read in
// a single round-trip. Earlier fragments of a multi-fragment request are written
// with ordinary pipe writes, and response fragments beyond the first are read with
// ordinary pipe reads.
//
// It wraps an already-connected and authenticated SMB 2.x client whose tree is
// connected to IPC$; that session's lifecycle is owned by the caller.
package smb2

import (
	"fmt"
	"strings"

	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	smb2client "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// Default fragment sizes proposed at Bind time.
const (
	DefaultMaxXmitFrag uint16 = 4280
	DefaultMaxRecvFrag uint16 = 4280
)

// pipeConn is the subset of *smb2client.Client the transport relies on. Expressing
// it as an interface keeps the transport unit-testable with a fake.
type pipeConn interface {
	CreateFile(path string, desiredAccess, shareAccess, createDisposition, createOptions uint32) (types.SMB2_FILEID, error)
	TransactNamedPipe(fileId types.SMB2_FILEID, input []byte, maxOutputResponse uint32) ([]byte, error)
	WriteFile(fileId types.SMB2_FILEID, offset uint64, data []byte) (uint32, error)
	ReadFile(fileId types.SMB2_FILEID, offset uint64, length uint32) ([]byte, error)
	CloseFile(fileId types.SMB2_FILEID) error
}

// SMBNamedPipe is a DCE/RPC transport over an SMB 2.x named pipe (ncacn_np).
type SMBNamedPipe struct {
	conn     pipeConn
	pipeName string
	fileId   types.SMB2_FILEID
	opened   bool
	// pending holds the most recent Send; it is written (via transceive) on the
	// next Recv, or flushed with a plain write if another Send arrives first.
	pending []byte
	maxXmit uint16
	maxRecv uint16
}

var _ dcerpctransport.Transport = (*SMBNamedPipe)(nil)

// New creates an SMB2 named-pipe transport over the supplied client. The client
// MUST already be connected, authenticated, and tree-connected to IPC$.
//
// pipeName is the pipe endpoint (e.g. "srvsvc" or "\\PIPE\\lsarpc"); a leading
// backslash and a "PIPE\\" prefix are stripped, as SMB2 names are share-relative.
func New(c *smb2client.Client, pipeName string) *SMBNamedPipe {
	return newWithConn(c, pipeName)
}

func newWithConn(conn pipeConn, pipeName string) *SMBNamedPipe {
	pipeName = strings.TrimPrefix(pipeName, "\\")
	if len(pipeName) >= 5 && strings.EqualFold(pipeName[:5], "PIPE\\") {
		pipeName = pipeName[5:]
	}
	return &SMBNamedPipe{
		conn:     conn,
		pipeName: pipeName,
		maxXmit:  DefaultMaxXmitFrag,
		maxRecv:  DefaultMaxRecvFrag,
	}
}

// SetMaxFrag overrides the fragment sizes proposed at Bind time. Call before Connect.
func (p *SMBNamedPipe) SetMaxFrag(xmit, recv uint16) {
	p.maxXmit = xmit
	p.maxRecv = recv
}

// Connect opens the named pipe on the SMB tree. It is idempotent.
func (p *SMBNamedPipe) Connect() error {
	if p.opened {
		return nil
	}
	fid, err := p.conn.CreateFile(
		p.pipeName,
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to open named pipe %q: %w", p.pipeName, err)
	}
	p.fileId = fid
	p.opened = true
	return nil
}

// Send queues a PDU. The PDU is transmitted on the next Recv via a single
// transceive; if another Send arrives first (a multi-fragment request), the queued
// fragment is flushed with an ordinary pipe write.
func (p *SMBNamedPipe) Send(pdu []byte) error {
	if !p.opened {
		return fmt.Errorf("named pipe %q is not open: call Connect first", p.pipeName)
	}
	if len(pdu) == 0 {
		return fmt.Errorf("refusing to send an empty PDU on named pipe %q", p.pipeName)
	}
	if p.pending != nil {
		if err := p.writePending(); err != nil {
			return err
		}
	}
	p.pending = pdu
	return nil
}

func (p *SMBNamedPipe) writePending() error {
	written, err := p.conn.WriteFile(p.fileId, 0, p.pending)
	if err != nil {
		return fmt.Errorf("failed to write PDU to named pipe %q: %w", p.pipeName, err)
	}
	if int(written) != len(p.pending) {
		return fmt.Errorf("short write to named pipe %q: wrote %d of %d bytes", p.pipeName, written, len(p.pending))
	}
	p.pending = nil
	return nil
}

// Recv returns the next chunk of the reply. The first Recv after a Send writes the
// queued PDU and reads the reply in one transceive; subsequent Recv calls read
// further response fragments with ordinary pipe reads.
func (p *SMBNamedPipe) Recv() ([]byte, error) {
	if !p.opened {
		return nil, fmt.Errorf("named pipe %q is not open: call Connect first", p.pipeName)
	}
	if p.pending != nil {
		out, err := p.conn.TransactNamedPipe(p.fileId, p.pending, uint32(p.maxRecv))
		p.pending = nil
		if err != nil {
			return nil, fmt.Errorf("pipe transceive on %q failed: %w", p.pipeName, err)
		}
		return out, nil
	}
	data, err := p.conn.ReadFile(p.fileId, 0, uint32(p.maxRecv))
	if err != nil {
		return nil, fmt.Errorf("failed to read from named pipe %q: %w", p.pipeName, err)
	}
	return data, nil
}

// Close closes the pipe handle, leaving the SMB session and tree connect intact.
// It is idempotent.
func (p *SMBNamedPipe) Close() error {
	if !p.opened {
		return nil
	}
	err := p.conn.CloseFile(p.fileId)
	p.opened = false
	p.pending = nil
	if err != nil {
		return fmt.Errorf("failed to close named pipe %q: %w", p.pipeName, err)
	}
	return nil
}

// MaxXmitFrag returns the proposed maximum transmit fragment size.
func (p *SMBNamedPipe) MaxXmitFrag() uint16 { return p.maxXmit }

// MaxRecvFrag returns the proposed maximum receive fragment size.
func (p *SMBNamedPipe) MaxRecvFrag() uint16 { return p.maxRecv }
