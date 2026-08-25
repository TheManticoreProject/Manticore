package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// The Name field of an SMB_COM_TRANSACTION request is, per [MS-CIFS] section
// 2.2.4.33.1, "a null-terminated array of OEM characters" — or of 16-bit Unicode
// characters when SMB_FLAGS2_UNICODE is set — and "MUST be the first field in this
// section". It carries no SMB_STRING buffer-format byte.
//
// These tests pin that down in both directions, because getting it wrong is
// invisible in a round trip that shares the mistake: a marshaller that prepends a
// format byte and an unmarshaller that consumes one agree with each other and
// disagree with every real client.

// TestTransactionNameCarriesNoBufferFormatByte asserts the marshalled data block
// begins with the name itself.
func TestTransactionNameCarriesNoBufferFormatByte(t *testing.T) {
	request := commands.NewTransactionRequest()
	if err := request.Name.SetString(`\PIPE\srvsvc`); err != nil {
		t.Fatalf("Name.SetString() error = %v", err)
	}
	request.Setup = []types.USHORT{0x0026, 0x0001}
	request.SetupCount = types.UCHAR(2)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// The data block is WordCount(1) + 2*WordCount words + ByteCount(2), then the
	// bytes. WordCount is 14 plus the setup words.
	wordCount := int(marshalled[0])
	if want := 14 + len(request.Setup); wordCount != want {
		t.Fatalf("WordCount = %d, want %d", wordCount, want)
	}
	dataStart := 1 + 2*wordCount + 2
	if dataStart >= len(marshalled) {
		t.Fatalf("the message is %d bytes, too short to hold a data block", len(marshalled))
	}

	block := marshalled[dataStart:]
	if !bytes.HasPrefix(block, []byte(`\PIPE\srvsvc`+"\x00")) {
		t.Fatalf("the data block begins % x, want the name at its start with no format byte", block[:min(8, len(block))])
	}
	// Spelled out, because this is the exact byte the bug added: 0x04 is
	// SMB_STRING's null-terminated-ASCII format code.
	if block[0] == 0x04 {
		t.Fatal("the data block begins with an SMB_STRING buffer-format byte, which the wire format has no room for")
	}
}

// TestTransactionNameRoundTrips asserts a name survives marshal and unmarshal in
// both encodings, terminator width included.
func TestTransactionNameRoundTrips(t *testing.T) {
	cases := []struct {
		name    string
		unicode bool
		value   string
	}{
		{"OEM pipe path", false, `\PIPE\srvsvc`},
		{"OEM bare prefix", false, `\PIPE\`},
		{"OEM empty", false, ""},
		{"Unicode pipe path", true, `\PIPE\srvsvc`},
		{"Unicode empty", true, ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			request := commands.NewTransactionRequest()
			request.SetUnicode(tc.unicode)
			if tc.unicode {
				request.Name.Buffer = []types.UCHAR(utf16.EncodeUTF16LE(tc.value))
			} else {
				request.Name.Buffer = []types.UCHAR(tc.value)
			}
			request.Name.Length = types.USHORT(len(request.Name.Buffer))
			request.Trans_Data = []types.UCHAR("payload")
			request.TotalDataCount = types.USHORT(len(request.Trans_Data))
			request.DataCount = types.USHORT(len(request.Trans_Data))

			marshalled, err := request.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			decoded := commands.NewTransactionRequest()
			decoded.Init()
			decoded.SetUnicode(tc.unicode)
			if _, err := decoded.Unmarshal(marshalled); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			got := string(decoded.Name.Buffer)
			if tc.unicode {
				got = utf16.DecodeUTF16LE(decoded.Name.Buffer)
			}
			if got != tc.value {
				t.Fatalf("the name round-tripped to %q, want %q", got, tc.value)
			}
			if string(decoded.Trans_Data) != "payload" {
				t.Fatalf("the data round-tripped to %q, want \"payload\"", decoded.Trans_Data)
			}
		})
	}
}

// TestTransactionNameFromTheWire asserts a name as a real client sends it decodes,
// which is the case the buffer-format byte broke: a name beginning with a
// backslash was read as a format code of 0x5C and refused outright.
func TestTransactionNameFromTheWire(t *testing.T) {
	// Built by hand rather than by this package's own marshaller, so the
	// assertion is against the wire format and not against a shared assumption.
	name := []byte(`\PIPE\LANMAN` + "\x00")
	setup := []byte{0x26, 0x00, 0x01, 0x00}
	words := make([]byte, 28)
	// SetupCount at word 14, byte offset 26.
	words[26] = 0x02

	message := []byte{}
	message = append(message, byte(14+2))
	message = append(message, words...)
	message = append(message, setup...)
	message = append(message, byte(len(name)), 0x00)
	message = append(message, name...)

	decoded := commands.NewTransactionRequest()
	decoded.Init()
	read, err := decoded.Unmarshal(message)
	if err != nil {
		t.Fatalf("a request as a client sends it did not decode: %v", err)
	}
	if read == 0 {
		t.Fatal("Unmarshal reported reading nothing")
	}
	if got := string(decoded.Name.Buffer); got != `\PIPE\LANMAN` {
		t.Fatalf("the name decoded as %q, want %q", got, `\PIPE\LANMAN`)
	}
	if decoded.SetupCount != 2 || len(decoded.Setup) != 2 || decoded.Setup[0] != 0x0026 {
		t.Fatalf("the setup words decoded as %v (count %d), want [0x0026 0x0001]", decoded.Setup, decoded.SetupCount)
	}
}
