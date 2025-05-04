package command_interface

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
)

func TestCommand_GetCommandCode(t *testing.T) {
	cmd := &Command{
		CommandCode: codes.SMB_COM_NEGOTIATE,
	}

	if cmd.GetCommandCode() != codes.SMB_COM_NEGOTIATE {
		t.Errorf("Expected command code %v, got %v", codes.SMB_COM_NEGOTIATE, cmd.GetCommandCode())
	}
}

func TestCommand_AndX(t *testing.T) {
	cmd := &Command{}
	andX := &andx.AndX{}

	// Test SetAndX and GetAndX
	cmd.SetAndX(andX)
	if cmd.GetAndX() != andX {
		t.Errorf("Expected AndX to be %v, got %v", andX, cmd.GetAndX())
	}

	// Test IsAndX
	if cmd.IsAndX() {
		t.Errorf("Expected IsAndX to be false for base Command")
	}
}

func TestCommand_Parameters(t *testing.T) {
	cmd := &Command{}
	params := parameters.NewParameters()
	params.AddWord(1234)

	// Test SetParameters and GetParameters
	cmd.SetParameters(params)
	if cmd.GetParameters() != params {
		t.Errorf("Expected Parameters to be %v, got %v", params, cmd.GetParameters())
	}
}

func TestCommand_Data(t *testing.T) {
	cmd := &Command{}
	d := data.NewData()
	d.Add([]byte{0x01, 0x02, 0x03, 0x04})

	// Test SetData and GetData
	cmd.SetData(d)
	if cmd.GetData() != d {
		t.Errorf("Expected Data to be %v, got %v", d, cmd.GetData())
	}
}

func TestCommand_ChainLength(t *testing.T) {
	cmd1 := &Command{
		CommandCode: codes.SMB_COM_NEGOTIATE,
	}

	// Test chain length for single command
	if cmd1.GetChainLength() != 1 {
		t.Errorf("Expected chain length 1 for single command, got %d", cmd1.GetChainLength())
	}

	// Create a chain of commands
	cmd2 := &Command{
		CommandCode: codes.SMB_COM_SESSION_SETUP_ANDX,
	}
	cmd3 := &Command{
		CommandCode: codes.SMB_COM_TREE_CONNECT_ANDX,
	}

	var cmd2Interface CommandInterface = cmd2
	var cmd3Interface CommandInterface = cmd3

	cmd1.SetNextCommand(cmd2Interface)
	cmd2.SetNextCommand(cmd3Interface)

	// Test chain length for command with next commands
	if cmd1.GetChainLength() != 3 {
		t.Errorf("Expected chain length 3 for command chain, got %d", cmd1.GetChainLength())
	}
}
