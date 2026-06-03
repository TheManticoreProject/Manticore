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
