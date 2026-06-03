package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/data"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/parameters"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestTreeConnectAndxRequestUnmarshalSplitsPathAndService verifies that
// Unmarshal parses the null-terminated Path string up to and including its
// terminator and parses the trailing bytes into the Service field, rather than
// collapsing both null-terminated strings into Path.
//
// Per MS-CIFS 2.2.4.55.1, SMB_Data.Bytes is Password[PasswordLength], then
// Pad[], then SMB_STRING Path (null-terminated), then OEM_STRING Service
// (null-terminated).
func TestTreeConnectAndxRequestUnmarshalSplitsPathAndService(t *testing.T) {
	path := append([]types.UCHAR(`\\server\share`), 0x00)
	service := append([]types.UCHAR("A:"), 0x00)

	// Build a request and marshal it to obtain a valid wire representation.
	src := commands.NewTreeConnectAndxRequest()
	src.Password = []types.UCHAR{0x00} // single null padding byte, PasswordLength = 1
	src.Path = path
	src.Service = service

	marshalled, err := src.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal into a fresh request. Initialize Parameters/Data explicitly so
	// this test does not depend on the nil-guard fix from a separate change.
	dst := commands.NewTreeConnectAndxRequest()
	dst.SetParameters(parameters.NewParameters())
	dst.SetData(data.NewData())

	if _, err := dst.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !bytes.Equal([]byte(dst.Path), []byte(path)) {
		t.Errorf("Path mismatch:\n got  %q\n want %q", []byte(dst.Path), []byte(path))
	}
	if !bytes.Equal([]byte(dst.Service), []byte(service)) {
		t.Errorf("Service mismatch:\n got  %q\n want %q", []byte(dst.Service), []byte(service))
	}
}

// TestTreeConnectAndxRequestUnmarshalFreshNoPanic verifies that calling
// Unmarshal on a request constructed with NewTreeConnectAndxRequest does not
// panic with a nil-pointer dereference. The constructor does not initialize the
// Parameters/Data structures, so Unmarshal must create them when nil (mirroring
// Marshal) before using them.
func TestTreeConnectAndxRequestUnmarshalFreshNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unmarshal panicked on a freshly constructed request: %v", r)
		}
	}()

	c := commands.NewTreeConnectAndxRequest()

	// A minimal SMB_Parameters/SMB_Data byte stream: WordCount=0 then ByteCount=0.
	// The exact decoding is irrelevant here; the point is that Unmarshal must not
	// panic on the nil Parameters/Data structures of a fresh request.
	input := []byte{0x00, 0x00, 0x00}

	if _, err := c.Unmarshal(input); err != nil {
		// An error is acceptable for this minimal input; a panic is not.
		t.Logf("Unmarshal returned an error (acceptable, no panic): %v", err)
	}
}
