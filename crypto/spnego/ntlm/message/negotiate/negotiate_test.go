package negotiate_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/message/negotiate/flags"
)

func TestMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		workstation string
		useUnicode  bool
		flags       flags.NegotiateFlags
	}{
		{
			name:        "Basic Unicode Message",
			domain:      "MANTICORE",
			workstation: "DC01",
			useUnicode:  true,
			flags:       flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY,
		},
		{
			name:        "Basic OEM Message",
			domain:      "MANTICORE",
			workstation: "DC01",
			useUnicode:  false,
			flags:       flags.NTLMSSP_NEGOTIATE_OEM | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY,
		},
		{
			name:        "Empty Domain",
			domain:      "",
			workstation: "DC01",
			useUnicode:  true,
			flags:       flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY,
		},
		{
			name:        "Empty Workstation",
			domain:      "MANTICORE",
			workstation: "",
			useUnicode:  true,
			flags:       flags.NTLMSSP_NEGOTIATE_UNICODE | flags.NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create message
			msg, err := negotiate.CreateNegotiateMessage(tt.domain, tt.workstation, tt.flags, nil)
			if err != nil {
				t.Fatalf("Failed to create negotiate message: %v", err)
			}

			// Marshal
			data, err := msg.Marshal()
			if err != nil {
				t.Fatalf("Failed to marshal: %v", err)
			}

			// Unmarshal
			unmarshaledMsg := &negotiate.NegotiateMessage{}
			_, err = unmarshaledMsg.Unmarshal(data)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}

			// Compare fields
			if !bytes.Equal(msg.DomainName, unmarshaledMsg.DomainName) && msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED) {
				t.Error("DomainName mismatch")
				t.Errorf("  | msg.DomainName: \"%s\"", string(msg.DomainName))
				t.Errorf("  | msg.DomainNameFields.BufferOffset: %d", msg.DomainNameFields.BufferOffset)
				t.Errorf("  | msg.DomainNameFields.Len: %d", msg.DomainNameFields.Len)
				t.Errorf("  | msg.DomainNameFields.MaxLen: %d", msg.DomainNameFields.MaxLen)
				t.Errorf("  | ")
				t.Errorf("  | unmarshaledMsg.DomainName: \"%s\"", string(unmarshaledMsg.DomainName))
				t.Errorf("  | unmarshaledMsg.DomainNameFields.BufferOffset: %d", unmarshaledMsg.DomainNameFields.BufferOffset)
				t.Errorf("  | unmarshaledMsg.DomainNameFields.Len: %d", unmarshaledMsg.DomainNameFields.Len)
				t.Errorf("  | unmarshaledMsg.DomainNameFields.MaxLen: %d", unmarshaledMsg.DomainNameFields.MaxLen)
			}
			if !bytes.Equal(msg.Workstation, unmarshaledMsg.Workstation) && msg.NegotiateFlags.HasFlag(flags.NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED) {
				t.Error("Workstation mismatch")
				t.Errorf("  | msg.Workstation: \"%s\"", string(msg.Workstation))
				t.Errorf("  | msg.WorkstationFields.BufferOffset: %d", msg.WorkstationFields.BufferOffset)
				t.Errorf("  | msg.WorkstationFields.Len: %d", msg.WorkstationFields.Len)
				t.Errorf("  | msg.WorkstationFields.MaxLen: %d", msg.WorkstationFields.MaxLen)
				t.Errorf("  | ")
				t.Errorf("  | unmarshaledMsg.Workstation: \"%s\"", string(unmarshaledMsg.Workstation))
				t.Errorf("  | unmarshaledMsg.WorkstationFields.BufferOffset: %d", unmarshaledMsg.WorkstationFields.BufferOffset)
				t.Errorf("  | unmarshaledMsg.WorkstationFields.Len: %d", unmarshaledMsg.WorkstationFields.Len)
				t.Errorf("  | unmarshaledMsg.WorkstationFields.MaxLen: %d", unmarshaledMsg.WorkstationFields.MaxLen)
			}
			if msg.NegotiateFlags != unmarshaledMsg.NegotiateFlags {
				t.Error("NegotiateFlags mismatch")
			}

			// Compare DataFields
			if !msg.DomainNameFields.Equal(&unmarshaledMsg.DomainNameFields) {
				t.Error("DomainNameFields mismatch")
			}
			if !msg.WorkstationFields.Equal(&unmarshaledMsg.WorkstationFields) {
				t.Error("WorkstationFields mismatch")
			}
		})
	}
}

func TestUnmarshallRealData(t *testing.T) {
	bytesBlob, err := hex.DecodeString("4e544c4d5353500001000000050288a000000000000000000000000000000000")
	if err != nil {
		t.Fatalf("failed to decode hex string: %v", err)
	}

	msg := &negotiate.NegotiateMessage{}
	_, err = msg.Unmarshal(bytesBlob)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

// TestNegotiateMessagePayloadOffsetOverflow asserts that a DomainName or
// Workstation BufferOffset near the top of the 32-bit range is refused rather
// than wrapping past the bounds check and panicking in the slice expression.
//
// BufferOffset is a 32-bit wire field and Len a 16-bit one, so computing the end
// of the payload as BufferOffset+Len in uint32 overflows: 0xFFFFFFFF plus a small
// length compares as inside the buffer.
func TestNegotiateMessagePayloadOffsetOverflow(t *testing.T) {
	// The fixed part of a NEGOTIATE: signature and type (12), NegotiateFlags (4),
	// DomainNameFields (8), WorkstationFields (8), Version (8).
	build := func(domainOffset, workstationOffset uint32) []byte {
		message := make([]byte, 40)
		copy(message, []byte{'N', 'T', 'L', 'M', 'S', 'S', 'P', 0})
		binary.LittleEndian.PutUint32(message[8:12], 1)
		binary.LittleEndian.PutUint32(message[12:16], 0)

		binary.LittleEndian.PutUint16(message[16:18], 8) // DomainName Len
		binary.LittleEndian.PutUint16(message[18:20], 8) // DomainName MaxLen
		binary.LittleEndian.PutUint32(message[20:24], domainOffset)

		binary.LittleEndian.PutUint16(message[24:26], 8) // Workstation Len
		binary.LittleEndian.PutUint16(message[26:28], 8) // Workstation MaxLen
		binary.LittleEndian.PutUint32(message[28:32], workstationOffset)

		return message
	}

	tests := []struct {
		name    string
		message []byte
	}{
		{name: "DomainName offset overflows", message: build(0xFFFFFFFF, 32)},
		{name: "Workstation offset overflows", message: build(32, 0xFFFFFFFF)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("Unmarshal() panicked instead of returning an error: %v", recovered)
				}
			}()

			parsed := &negotiate.NegotiateMessage{}
			if _, err := parsed.Unmarshal(test.message); err == nil {
				t.Error("Unmarshal() accepted a payload offset outside the message")
			}
		})
	}
}
