package message_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
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

// TestMarshalWritesAndXChain asserts Message.Marshal emits a batched message for a
// chain built with AddCommand: each AndX block naming the command that follows it
// and the offset it begins at, and the last one terminating the chain.
//
// The assembled bytes are then decoded back, because a chain that marshals into
// something Unmarshal cannot follow is the failure this is guarding against.
func TestMarshalWritesAndXChain(t *testing.T) {
	// The length of the first command's block, from a message of its own. A fresh
	// command is used because Marshal appends to a command's parameter and data
	// blocks, so marshalling one twice does not produce the same bytes twice.
	sizing := message.NewMessage()
	sizing.Header.SetFlags(flags.FLAGS_REPLY)
	sizing.AddCommand(commands.NewSessionSetupAndxResponse())
	sized, err := sizing.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the sizing message: %v", err)
	}
	firstBlockLength := len(sized) - header.SMB_HEADER_SIZE

	batched := message.NewMessage()
	batched.Header.SetFlags(flags.FLAGS_REPLY)
	batched.AddCommand(commands.NewSessionSetupAndxResponse())
	batched.AddCommand(commands.NewTreeConnectAndxResponse())

	raw, err := batched.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	// The header's command code is the first command's, per [MS-CIFS] 2.2.3.4.
	if got := codes.CommandCode(raw[4]); got != codes.SMB_COM_SESSION_SETUP_ANDX {
		t.Errorf("the header names command 0x%02X, want SMB_COM_SESSION_SETUP_ANDX", uint8(got))
	}

	// The first command's AndX block: AndXCommand(1) AndXReserved(1)
	// AndXOffset(2), immediately after its WordCount.
	andxAt := header.SMB_HEADER_SIZE + 1
	if got := codes.CommandCode(raw[andxAt]); got != codes.SMB_COM_TREE_CONNECT_ANDX {
		t.Errorf("AndXCommand is 0x%02X, want SMB_COM_TREE_CONNECT_ANDX", uint8(got))
	}
	if raw[andxAt+1] != 0x00 {
		t.Errorf("AndXReserved is 0x%02X, want 0x00", raw[andxAt+1])
	}

	wantOffset := header.SMB_HEADER_SIZE + firstBlockLength
	gotOffset := int(binary.LittleEndian.Uint16(raw[andxAt+2 : andxAt+4]))
	if gotOffset != wantOffset {
		t.Fatalf("AndXOffset is %d, want %d (the header plus the first command's %d bytes)",
			gotOffset, wantOffset, firstBlockLength)
	}

	// The second command terminates the chain.
	secondAndxAt := gotOffset + 1
	if got := codes.CommandCode(raw[secondAndxAt]); got != codes.SMB_COM_NO_ANDX_COMMAND {
		t.Errorf("the last command names 0x%02X as its follow-on, want SMB_COM_NO_ANDX_COMMAND", uint8(got))
	}
	if got := binary.LittleEndian.Uint16(raw[secondAndxAt+2 : secondAndxAt+4]); got != 0 {
		t.Errorf("the last command's AndXOffset is %d, want 0", got)
	}

	// And the whole thing decodes back into the chain it was built from.
	out := message.NewMessage()
	if err := out.Unmarshal(raw); err != nil {
		t.Fatalf("the marshalled chain did not decode: %v", err)
	}
	if got := out.Command.GetChainLength(); got != 2 {
		t.Fatalf("the decoded chain is %d commands, want 2", got)
	}
	if _, ok := out.Command.(*commands.SessionSetupAndxResponse); !ok {
		t.Errorf("the first decoded command is %T, want *SessionSetupAndxResponse", out.Command)
	}
	if _, ok := out.Command.GetNextCommand().(*commands.TreeConnectAndxResponse); !ok {
		t.Errorf("the second decoded command is %T, want *TreeConnectAndxResponse",
			out.Command.GetNextCommand())
	}
}

// TestMarshalRejectsFollowingANonAndXCommand asserts a chain whose non-final
// command is not an AndX command is refused rather than emitted.
//
// A non-AndX command has no AndX block, so there is nowhere to record what follows
// it: [MS-CIFS] 2.2.3.4 makes such a command the end of a chain. Emitting the
// follow-on anyway would put a block pair in the message that no client could find.
func TestMarshalRejectsFollowingANonAndXCommand(t *testing.T) {
	msg := message.NewMessage()
	msg.Header.SetFlags(flags.FLAGS_REPLY)
	msg.AddCommand(commands.NewNegotiateResponse())
	msg.AddCommand(commands.NewTreeConnectAndxResponse())

	if _, err := msg.Marshal(); err == nil {
		t.Fatal("marshalling a chain behind a non-AndX command succeeded, want an error")
	}
}

// TestMarshalPositionsChainedReadOffsets asserts a read response batched behind
// another command reports a DataOffset that points at its data.
//
// DataOffset is measured from the start of the SMB header, so a response that
// assumed it sat at the front of the message would send the client to bytes
// belonging to the command ahead of it.
func TestMarshalPositionsChainedReadOffsets(t *testing.T) {
	payload := []byte("the bytes the client asked to read")

	read := commands.NewReadAndxResponse()
	read.Data = []types.UCHAR(payload)
	read.DataLength = types.USHORT(len(payload))

	batched := message.NewMessage()
	batched.Header.SetFlags(flags.FLAGS_REPLY)
	batched.AddCommand(commands.NewTreeConnectAndxResponse())
	batched.AddCommand(read)

	raw, err := batched.Marshal()
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	offset := int(read.DataOffset)
	if offset+len(payload) > len(raw) {
		t.Fatalf("DataOffset %d plus %d bytes runs past the %d-byte message",
			offset, len(payload), len(raw))
	}
	if got := raw[offset : offset+len(payload)]; string(got) != string(payload) {
		t.Fatalf("DataOffset %d points at %q, want %q", offset, got, payload)
	}
}
