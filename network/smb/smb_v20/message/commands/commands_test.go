package commands_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
)

// No concrete SMB2 command structures exist yet (they are added per phase), so
// the dispatcher reports every code as unsupported. As commands are wired in,
// their cases are added here and exercised by their own tests.
func TestCreateRequestCommand_Unsupported(t *testing.T) {
	if _, err := commands.CreateRequestCommand(codes.SMB2_NEGOTIATE); err == nil {
		t.Errorf("expected unsupported error for not-yet-implemented request command")
	}
	if _, err := commands.CreateRequestCommand(codes.CommandCode(0x00FF)); err == nil {
		t.Errorf("expected unsupported error for unknown request command code")
	}
}

func TestCreateResponseCommand_Unsupported(t *testing.T) {
	if _, err := commands.CreateResponseCommand(codes.SMB2_NEGOTIATE); err == nil {
		t.Errorf("expected unsupported error for not-yet-implemented response command")
	}
	if _, err := commands.CreateResponseCommand(codes.CommandCode(0x00FF)); err == nil {
		t.Errorf("expected unsupported error for unknown response command code")
	}
}
