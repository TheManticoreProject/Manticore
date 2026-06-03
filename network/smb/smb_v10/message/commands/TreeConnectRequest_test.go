package commands_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestTreeConnectRequestUnmarshalInitializesStructures verifies that Unmarshal
// does not panic on a freshly constructed TreeConnectRequest. NewTreeConnectRequest
// leaves the embedded Parameters/Data nil, and Unmarshal must initialize them
// before use (mirroring Marshal) rather than dereferencing a nil *Parameters.
func TestTreeConnectRequestUnmarshalInitializesStructures(t *testing.T) {
	out := commands.NewTreeConnectRequest()
	out.Path = *types.NewOEM_STRINGFromString(`\\server\share`)
	out.Password = *types.NewOEM_STRINGFromString("secret")
	out.Service = *types.NewOEM_STRINGFromString("A:")

	marshalled, err := out.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// A fresh request has nil Parameters/Data; Unmarshal must not panic.
	in := commands.NewTreeConnectRequest()
	if _, err := in.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}

// TestTreeConnectRequestUnmarshalDistinctFields verifies that Unmarshal decodes
// the three SMB_Data OEM strings (Path, Password, Service) from consecutive
// positions in the data block, rather than re-parsing Path for all three.
// Per [MS-CIFS] 2.2.4.50.1, the SMB_Data block is:
//
//	BufferFormat1(0x04) Path\0 BufferFormat2(0x04) Password\0 BufferFormat3(0x04) Service\0
func TestTreeConnectRequestUnmarshalDistinctFields(t *testing.T) {
	out := commands.NewTreeConnectRequest()
	out.Path = *types.NewOEM_STRINGFromString(`\\server\share`)
	out.Password = *types.NewOEM_STRINGFromString("secret")
	out.Service = *types.NewOEM_STRINGFromString("A:")

	marshalled, err := out.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	in := commands.NewTreeConnectRequest()
	if _, err := in.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got := in.Path.GetString(); got != `\\server\share` {
		t.Errorf("Path = %q, want %q", got, `\\server\share`)
	}
	if got := in.Password.GetString(); got != "secret" {
		t.Errorf("Password = %q, want %q", got, "secret")
	}
	if got := in.Service.GetString(); got != "A:" {
		t.Errorf("Service = %q, want %q", got, "A:")
	}
}
