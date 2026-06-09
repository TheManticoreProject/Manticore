package client

import (
	"fmt"

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

func (b *smb1Backend) Login(creds *credentials.Credentials) error {
	return b.engine.SessionSetup(creds)
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
	// SMB1 directory enumeration (TRANS2 FIND_FIRST2/FIND_NEXT2) is normalized
	// into []FileInfo in Phase 6.
	return nil, fmt.Errorf("ListDirectory is not yet implemented for SMB1 (Phase 6)")
}

func (b *smb1Backend) TreeDisconnect() error { return b.engine.TreeDisconnect() }
func (b *smb1Backend) Logoff() error         { return b.engine.Logoff() }
func (b *smb1Backend) Disconnect() error     { return b.engine.Disconnect() }

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
