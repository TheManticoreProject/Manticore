package message_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
)

// TestUnmarshalFollowsAndXChain verifies that Message.Unmarshal decodes every
// command in a batched ("AndX") message, not just the first. It assembles the
// classic SESSION_SETUP_ANDX + TREE_CONNECT_ANDX response pair with a
// spec-compliant (little-endian) AndXOffset and asserts the chain is recovered.
func TestUnmarshalFollowsAndXChain(t *testing.T) {
	// Marshal a standalone TreeConnectAndxResponse to obtain its command block
	// (everything after the SMB header).
	bMsg := message.NewMessage()
	bMsg.Header.SetFlags(flags.FLAGS_REPLY)
	bMsg.AddCommand(commands.NewTreeConnectAndxResponse())
	bRaw, err := bMsg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal TreeConnectAndxResponse: %v", err)
	}
	bBlock := bRaw[header.SMB_HEADER_SIZE:]

	// Marshal a SessionSetupAndxResponse message: header + first command block.
	aMsg := message.NewMessage()
	aMsg.Header.SetFlags(flags.FLAGS_REPLY)
	aMsg.AddCommand(commands.NewSessionSetupAndxResponse())
	aRaw, err := aMsg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal SessionSetupAndxResponse: %v", err)
	}

	// Assemble the batched message and patch the first command's AndX block to
	// point at the second command. The AndX block follows the WordCount byte:
	// AndXCommand(1) AndXReserved(1) AndXOffset(2, little-endian). AndXOffset is
	// measured from the start of the SMB header, which is exactly len(aRaw) (the
	// index of the second command's WordCount within the assembled buffer).
	full := append(append([]byte{}, aRaw...), bBlock...)
	andxCommandPos := header.SMB_HEADER_SIZE + 1
	full[andxCommandPos] = byte(codes.SMB_COM_TREE_CONNECT_ANDX)
	full[andxCommandPos+1] = 0x00 // AndXReserved
	binary.LittleEndian.PutUint16(full[andxCommandPos+2:andxCommandPos+4], uint16(len(aRaw)))

	out := message.NewMessage()
	if err := out.Unmarshal(full); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}

	if _, ok := out.Command.(*commands.SessionSetupAndxResponse); !ok {
		t.Fatalf("expected first command *SessionSetupAndxResponse, got %T", out.Command)
	}
	if got := out.Command.GetChainLength(); got != 2 {
		t.Fatalf("expected chain length 2, got %d", got)
	}
	next := out.Command.GetNextCommand()
	if next == nil {
		t.Fatal("expected a chained command after the first, got nil")
	}
	if _, ok := next.(*commands.TreeConnectAndxResponse); !ok {
		t.Errorf("expected chained command *TreeConnectAndxResponse, got %T", next)
	}
}

// TestUnmarshalNonAndXLeavesNoChain verifies that a non-AndX command (here a
// NegotiateResponse) is decoded with no chained command, i.e. the AndX-following
// loop does not run for commands that are not batched.
func TestUnmarshalNonAndXLeavesNoChain(t *testing.T) {
	msg := message.NewMessage()
	msg.Header.SetFlags(flags.FLAGS_REPLY)
	msg.AddCommand(commands.NewNegotiateResponse())
	raw, err := msg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal NegotiateResponse: %v", err)
	}

	out := message.NewMessage()
	if err := out.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if got := out.Command.GetChainLength(); got != 1 {
		t.Errorf("expected chain length 1 for a non-AndX command, got %d", got)
	}
	if next := out.Command.GetNextCommand(); next != nil {
		t.Errorf("expected no chained command, got %T", next)
	}
}

// TestUnmarshalRejectsCyclicAndXChain verifies that a batched message whose two
// AndX commands reference each other (A->B, B->A) is rejected with an error
// rather than looping forever. Without a forward-progress guard, Unmarshal would
// never terminate.
func TestUnmarshalRejectsCyclicAndXChain(t *testing.T) {
	// Build a SessionSetupAndxResponse message (header + command block A).
	aMsg := message.NewMessage()
	aMsg.Header.SetFlags(flags.FLAGS_REPLY)
	aMsg.AddCommand(commands.NewSessionSetupAndxResponse())
	aRaw, err := aMsg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal SessionSetupAndxResponse: %v", err)
	}

	// A second SessionSetupAndxResponse command block (bytes after the header).
	bMsg := message.NewMessage()
	bMsg.Header.SetFlags(flags.FLAGS_REPLY)
	bMsg.AddCommand(commands.NewSessionSetupAndxResponse())
	bRaw, err := bMsg.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal second SessionSetupAndxResponse: %v", err)
	}
	bBlock := bRaw[header.SMB_HEADER_SIZE:]

	full := append(append([]byte{}, aRaw...), bBlock...)
	aOffset := header.SMB_HEADER_SIZE // command A's WordCount index
	bOffset := len(aRaw)              // command B's WordCount index

	// A's AndX block -> B.
	aAndx := aOffset + 1
	full[aAndx] = byte(codes.SMB_COM_SESSION_SETUP_ANDX)
	full[aAndx+1] = 0x00
	binary.LittleEndian.PutUint16(full[aAndx+2:aAndx+4], uint16(bOffset))

	// B's AndX block -> A (the cycle).
	bAndx := bOffset + 1
	full[bAndx] = byte(codes.SMB_COM_SESSION_SETUP_ANDX)
	full[bAndx+1] = 0x00
	binary.LittleEndian.PutUint16(full[bAndx+2:bAndx+4], uint16(aOffset))

	out := message.NewMessage()
	if err := out.Unmarshal(full); err == nil {
		t.Fatal("Unmarshal should reject a cyclic AndX chain, got nil error")
	}
}
