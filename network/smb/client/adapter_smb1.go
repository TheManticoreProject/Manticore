package client

import (
	"fmt"
	"strings"

	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	dcerpcsmb "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/smb"
	"github.com/TheManticoreProject/Manticore/network/smb"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// smb1Backend adapts the SMB 1.0 engine (network/smb/smb_v10/client) to the
// version-agnostic Backend interface.
type smb1Backend struct {
	engine *smb1.Client
}

func newSMB1Backend(engine *smb1.Client) *smb1Backend {
	return &smb1Backend{engine: engine}
}

func (b *smb1Backend) Dialect() smb.SMBProtocolVersion { return smb.SMB_VERSION_1_0 }

func (b *smb1Backend) ConnectionInfo() ConnectionInfo {
	s := b.engine.Connection.Server
	// SMB1 negotiates a single MaxBufferSize, not separate read/write maxima, so it
	// is reported for both.
	return ConnectionInfo{
		SigningRequired: s.SecurityMode.IsSecuritySignatureRequired(),
		MaxReadSize:     s.MaxBufferSize,
		MaxWriteSize:    s.MaxBufferSize,
	}
}

func (b *smb1Backend) Login(creds *credentials.Credentials) error {
	return b.engine.SessionSetup(creds)
}

func (b *smb1Backend) ServerIdentity() ServerIdentity {
	s := b.engine.Connection.Server
	return ServerIdentity{
		NetBIOSComputerName: s.NetBIOSComputerName,
		NetBIOSDomainName:   s.NetBIOSDomainName,
		DNSComputerName:     s.DNSComputerName,
		DNSDomainName:       s.DNSDomainName,
		OSVersionMajor:      s.OSVersionMajor,
		OSVersionMinor:      s.OSVersionMinor,
		OSVersionBuild:      s.OSVersionBuild,
	}
}

func (b *smb1Backend) TreeConnect(share string) error {
	return b.engine.TreeConnect(share)
}

func (b *smb1Backend) OpenFile(path string, opts OpenOptions) (FileHandle, error) {
	fid, err := b.engine.OpenFile(path, opts.DesiredAccess, opts.ShareAccess, opts.CreateDisposition, opts.CreateOptions)
	if err != nil {
		return FileHandle{}, err
	}
	return newFileHandle(smb.SMB_VERSION_1_0, fid), nil
}

func (b *smb1Backend) ReadFile(h FileHandle, off uint64, n uint32) ([]byte, error) {
	fid, err := b.fid(h)
	if err != nil {
		return nil, err
	}
	return b.engine.ReadFile(fid, off, n)
}

func (b *smb1Backend) WriteFile(h FileHandle, off uint64, data []byte) (uint32, error) {
	fid, err := b.fid(h)
	if err != nil {
		return 0, err
	}
	written, err := b.engine.WriteFile(fid, off, data)
	if err != nil {
		return 0, err
	}
	return uint32(written), nil
}

func (b *smb1Backend) CloseFile(h FileHandle) error {
	fid, err := b.fid(h)
	if err != nil {
		return err
	}
	return b.engine.CloseFile(fid)
}

func (b *smb1Backend) ListDirectory(path, pattern string) ([]FileInfo, error) {
	// The SMB1 engine enumerates by a single share-relative pattern (TRANS2
	// FIND_FIRST2/FIND_NEXT2); join the directory path and the match pattern into
	// it, e.g. ("\\Windows", "*.ini") -> "\\Windows\\*.ini".
	if pattern == "" {
		pattern = "*"
	}
	p := path
	if p == "" {
		p = "\\"
	}
	if p[0] != '\\' {
		p = "\\" + p
	}
	if p[len(p)-1] != '\\' {
		p += "\\"
	}
	p += pattern

	entries, err := b.engine.ListEntries(p)
	if err != nil {
		return nil, err
	}

	out := make([]FileInfo, 0, len(entries))
	for _, e := range entries {
		// The SMB1 engine includes the OEM name's trailing NUL terminator in
		// LongName; normalize it away (SMB2 names are already NUL-trimmed).
		name := strings.TrimRight(e.LongName, "\x00")
		if name == "." || name == ".." {
			continue
		}
		out = append(out, FileInfo{
			Name:           name,
			FileAttributes: e.Attributes,
			Size:           e.Size,
			CreationTime:   e.CreatedAt,
			LastAccessTime: e.AccessedAt,
			LastWriteTime:  e.ModifiedAt,
			ChangeTime:     e.ChangedAt,
		})
	}
	return out, nil
}

func (b *smb1Backend) DeleteFile(path string) error { return b.engine.DeleteFile(smb1WirePath(path)) }

func (b *smb1Backend) CreateDirectory(path string) error {
	return b.engine.CreateDirectory(smb1WirePath(path))
}

func (b *smb1Backend) DeleteDirectory(path string) error {
	return b.engine.DeleteDirectory(smb1WirePath(path))
}

func (b *smb1Backend) RenameFile(oldPath, newPath string) error {
	return b.engine.RenameFile(smb1WirePath(oldPath), smb1WirePath(newPath))
}

func (b *smb1Backend) CheckDirectory(path string) error {
	return b.engine.CheckDirectory(smb1WirePath(path))
}

func (b *smb1Backend) RPCTransport(pipeName string) (dcerpctransport.Transport, error) {
	return dcerpcsmb.New(b.engine, pipeName), nil
}

func (b *smb1Backend) TreeDisconnect() error { return b.engine.TreeDisconnect() }
func (b *smb1Backend) Logoff() error         { return b.engine.Logoff() }
func (b *smb1Backend) Disconnect() error     { return b.engine.Disconnect() }

// smb1WirePath converts a generic share-relative path (no leading backslash; ""
// for the root) into the leading-backslash form the SMB1 file-management commands
// expect. Unlike the engine's OpenFile, those commands do not prepend it.
func smb1WirePath(path string) string {
	if path == "" {
		return "\\"
	}
	if path[0] != '\\' {
		return "\\" + path
	}
	return path
}

// fid unboxes an opaque FileHandle back into the SMB1 FID that produced it,
// guarding against a handle from another backend.
func (b *smb1Backend) fid(h FileHandle) (smb1.FID, error) {
	fid, ok := h.raw.(smb1.FID)
	if !ok {
		return 0, fmt.Errorf("file handle is not an SMB1 FID (got %T)", h.raw)
	}
	return fid, nil
}

// Compile-time assurance that the adapter satisfies the Backend contract.
var _ Backend = (*smb1Backend)(nil)
