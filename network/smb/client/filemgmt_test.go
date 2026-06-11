package client

import (
	"testing"

	dcerpctransport "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport"
	"github.com/TheManticoreProject/Manticore/network/smb"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// recordingBackend is a no-op Backend that records the file-management calls the
// generic Client delegates to it.
type recordingBackend struct {
	calls    []string
	args     []string
	connInfo ConnectionInfo
	identity ServerIdentity
}

func (r *recordingBackend) Dialect() smb.SMBProtocolVersion      { return smb.SMB_VERSION_2_0_2 }
func (r *recordingBackend) ConnectionInfo() ConnectionInfo       { return r.connInfo }
func (r *recordingBackend) ServerIdentity() ServerIdentity       { return r.identity }
func (r *recordingBackend) Login(*credentials.Credentials) error { return nil }
func (r *recordingBackend) TreeConnect(string) error             { return nil }
func (r *recordingBackend) OpenFile(string, OpenOptions) (FileHandle, error) {
	return FileHandle{}, nil
}
func (r *recordingBackend) ReadFile(FileHandle, uint64, uint32) ([]byte, error)  { return nil, nil }
func (r *recordingBackend) WriteFile(FileHandle, uint64, []byte) (uint32, error) { return 0, nil }
func (r *recordingBackend) CloseFile(FileHandle) error                           { return nil }
func (r *recordingBackend) ListDirectory(string, string) ([]FileInfo, error)     { return nil, nil }
func (r *recordingBackend) RPCTransport(string) (dcerpctransport.Transport, error) {
	return nil, nil
}
func (r *recordingBackend) TreeDisconnect() error { return nil }
func (r *recordingBackend) Logoff() error         { return nil }
func (r *recordingBackend) Disconnect() error     { return nil }

func (r *recordingBackend) DeleteFile(path string) error {
	r.calls, r.args = append(r.calls, "DeleteFile"), append(r.args, path)
	return nil
}
func (r *recordingBackend) CreateDirectory(path string) error {
	r.calls, r.args = append(r.calls, "CreateDirectory"), append(r.args, path)
	return nil
}
func (r *recordingBackend) DeleteDirectory(path string) error {
	r.calls, r.args = append(r.calls, "DeleteDirectory"), append(r.args, path)
	return nil
}
func (r *recordingBackend) RenameFile(oldPath, newPath string) error {
	r.calls, r.args = append(r.calls, "RenameFile"), append(r.args, oldPath+"->"+newPath)
	return nil
}
func (r *recordingBackend) CheckDirectory(path string) error {
	r.calls, r.args = append(r.calls, "CheckDirectory"), append(r.args, path)
	return nil
}

// TestClientFileManagementDelegation verifies the generic Client forwards each
// file-management call to its backend with the path(s) unchanged.
func TestClientFileManagementDelegation(t *testing.T) {
	rb := &recordingBackend{}
	c := &Client{backend: rb}

	if err := c.DeleteFile("dir\\f.txt"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
	if err := c.CreateDirectory("dir\\new"); err != nil {
		t.Fatalf("CreateDirectory: %v", err)
	}
	if err := c.DeleteDirectory("dir\\old"); err != nil {
		t.Fatalf("DeleteDirectory: %v", err)
	}
	if err := c.RenameFile("a.txt", "b.txt"); err != nil {
		t.Fatalf("RenameFile: %v", err)
	}
	if err := c.CheckDirectory("dir"); err != nil {
		t.Fatalf("CheckDirectory: %v", err)
	}

	wantCalls := []string{"DeleteFile", "CreateDirectory", "DeleteDirectory", "RenameFile", "CheckDirectory"}
	wantArgs := []string{"dir\\f.txt", "dir\\new", "dir\\old", "a.txt->b.txt", "dir"}
	for i := range wantCalls {
		if rb.calls[i] != wantCalls[i] || rb.args[i] != wantArgs[i] {
			t.Errorf("call %d = (%s,%q), want (%s,%q)", i, rb.calls[i], rb.args[i], wantCalls[i], wantArgs[i])
		}
	}
}

// TestSMB1WirePath checks the leading-backslash normalization the SMB1 adapter
// applies to file-management paths.
func TestSMB1WirePath(t *testing.T) {
	cases := map[string]string{"": "\\", "dir": "\\dir", "dir\\f": "\\dir\\f", "\\already": "\\already"}
	for in, want := range cases {
		if got := smb1WirePath(in); got != want {
			t.Errorf("smb1WirePath(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestClientConnectionInfoDelegation verifies Client.ConnectionInfo forwards the
// backend's negotiated capabilities unchanged.
func TestClientConnectionInfoDelegation(t *testing.T) {
	want := ConnectionInfo{SigningRequired: true, MaxReadSize: 0x100000, MaxWriteSize: 0x80000, SupportsNTLMv2: true}
	c := &Client{backend: &recordingBackend{connInfo: want}}
	if got := c.ConnectionInfo(); got != want {
		t.Errorf("ConnectionInfo() = %+v, want %+v", got, want)
	}
}

// TestClientServerIdentityDelegation verifies Client.ServerIdentity forwards the
// backend's captured identity unchanged.
func TestClientServerIdentityDelegation(t *testing.T) {
	want := ServerIdentity{
		NetBIOSComputerName: "FILESRV",
		NetBIOSDomainName:   "CORP",
		DNSComputerName:     "filesrv.corp.example",
		DNSDomainName:       "corp.example",
		OSName:              "Windows Server 2019",
		OSVersionMajor:      10,
		OSVersionMinor:      0,
		OSVersionBuild:      17763,
	}
	c := &Client{backend: &recordingBackend{identity: want}}
	if got := c.ServerIdentity(); got != want {
		t.Errorf("ServerIdentity() = %+v, want %+v", got, want)
	}
}
