package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/smb"
	smb2 "github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// smb2Backend adapts the SMB 2.x engine (network/smb/smb_v20/client) to the
// version-agnostic Backend interface.
type smb2Backend struct {
	engine  *smb2.Client
	dialect smb.SMBProtocolVersion
}

// newSMB2Backend wraps an already-negotiated SMB2 engine. The negotiated dialect
// is resolved once; an unrecognized revision falls back to SMB 2.0.2.
func newSMB2Backend(engine *smb2.Client) *smb2Backend {
	dialect, ok := versionForSMB2Dialect(engine.Connection.Dialect)
	if !ok {
		dialect = smb.SMB_VERSION_2_0_2
	}
	return &smb2Backend{engine: engine, dialect: dialect}
}

func (b *smb2Backend) Dialect() smb.SMBProtocolVersion { return b.dialect }

func (b *smb2Backend) Login(creds *credentials.Credentials) error {
	return b.engine.SessionSetup(creds)
}

func (b *smb2Backend) TreeConnect(share string) error {
	return b.engine.TreeConnect(share)
}

func (b *smb2Backend) OpenFile(path string, opts OpenOptions) (FileHandle, error) {
	fileId, err := b.engine.CreateFile(path, opts.DesiredAccess, opts.ShareAccess, opts.CreateDisposition, opts.CreateOptions)
	if err != nil {
		return FileHandle{}, err
	}
	return newFileHandle(b.dialect, fileId), nil
}

func (b *smb2Backend) ReadFile(h FileHandle, off uint64, n uint32) ([]byte, error) {
	fileId, err := b.fileId(h)
	if err != nil {
		return nil, err
	}
	return b.engine.ReadFile(fileId, off, n)
}

func (b *smb2Backend) WriteFile(h FileHandle, off uint64, data []byte) (uint32, error) {
	fileId, err := b.fileId(h)
	if err != nil {
		return 0, err
	}
	return b.engine.WriteFile(fileId, off, data)
}

func (b *smb2Backend) CloseFile(h FileHandle) error {
	fileId, err := b.fileId(h)
	if err != nil {
		return err
	}
	return b.engine.CloseFile(fileId)
}

func (b *smb2Backend) ListDirectory(path, pattern string) ([]FileInfo, error) {
	if pattern == "" {
		pattern = "*"
	}

	// Open the directory for enumeration.
	fileId, err := b.engine.CreateFile(path,
		fileflags.FILE_LIST_DIRECTORY|fileflags.FILE_READ_ATTRIBUTES,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		return nil, fmt.Errorf("open directory %q: %w", path, err)
	}
	defer b.engine.CloseFile(fileId)

	// Page through QUERY_DIRECTORY: the pattern applies to the first call, and
	// subsequent calls continue from the server's cursor until it reports no more
	// files (an empty buffer).
	var out []FileInfo
	search := pattern
	for {
		buf, err := b.engine.QueryDirectory(fileId, fileBothDirectoryInformation, search, 0)
		if err != nil {
			return nil, err
		}
		if len(buf) == 0 {
			break
		}
		out = append(out, parseBothDirectoryInfo(buf)...)
		search = "" // continuation
	}
	return out, nil
}

func (b *smb2Backend) DeleteFile(path string) error { return b.engine.DeleteFile(path) }

func (b *smb2Backend) CreateDirectory(path string) error { return b.engine.CreateDirectory(path) }

func (b *smb2Backend) DeleteDirectory(path string) error { return b.engine.DeleteDirectory(path) }

func (b *smb2Backend) RenameFile(oldPath, newPath string) error {
	// replaceIfExists=false: fail rather than clobber an existing target, matching
	// the generic contract.
	return b.engine.RenameFile(oldPath, newPath, false)
}

func (b *smb2Backend) CheckDirectory(path string) error {
	st, err := b.engine.Stat(path)
	if err != nil {
		return err
	}
	if !st.IsDirectory {
		return fmt.Errorf("%q is not a directory", path)
	}
	return nil
}

func (b *smb2Backend) TreeDisconnect() error { return b.engine.TreeDisconnect() }
func (b *smb2Backend) Logoff() error         { return b.engine.Logoff() }
func (b *smb2Backend) Disconnect() error     { return b.engine.Disconnect() }

// fileId unboxes an opaque FileHandle back into the SMB2 file id that produced
// it, guarding against a handle from another backend.
func (b *smb2Backend) fileId(h FileHandle) (types.SMB2_FILEID, error) {
	fileId, ok := h.raw.(types.SMB2_FILEID)
	if !ok {
		return types.SMB2_FILEID{}, fmt.Errorf("file handle is not an SMB2 file id (got %T)", h.raw)
	}
	return fileId, nil
}

// Compile-time assurance that the adapter satisfies the Backend contract.
var _ Backend = (*smb2Backend)(nil)
