package header_test

import (
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header/flags2"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/securityfeatures"
)

func Test_HeaderMarshalUnmarshalInvolution(t *testing.T) {
	h := header.NewHeader()
	h.Command = 0x11
	h.Status = 0x22222222
	h.Flags = flags.Flags(0x33)
	h.Flags2 = flags2.Flags2(0x4444)
	h.PIDHigh = 0x5555
	h.SecurityFeatures = securityfeatures.NewSecurityFeaturesSecuritySignature()
	h.Reserved = 0x6666
	h.TID = 0x7777
	h.PIDLow = 0x5555
	h.UID = 0x8888
	h.MID = 0x9999

	marshalled, err := h.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal SMB header: %v", err)
	}

	t.Logf("Marshalled SMB header (%d bytes): %x", len(marshalled), marshalled)

	unmarshalled := header.NewHeader()
	bytesRead, err := unmarshalled.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Failed to unmarshal SMB header: %v", err)
	}

	if bytesRead != len(marshalled) {
		t.Errorf("Bytes read mismatch: %d != %d", bytesRead, len(marshalled))
	}

	if unmarshalled.Command != h.Command {
		t.Errorf("Command mismatch: got %s (0x%02x) expected %s (0x%02x)", unmarshalled.Command, uint8(unmarshalled.Command), h.Command, uint8(h.Command))
	}

	if unmarshalled.Status != h.Status {
		t.Errorf("Status mismatch: got %x expected %x", unmarshalled.Status, h.Status)
	}

	if unmarshalled.Flags != h.Flags {
		t.Errorf("Flags mismatch: got %x expected %x", unmarshalled.Flags, h.Flags)
	}

	if unmarshalled.Flags2 != h.Flags2 {
		t.Errorf("Flags2 mismatch: got %x expected %x", unmarshalled.Flags2, h.Flags2)
	}

	if unmarshalled.PIDHigh != h.PIDHigh {
		t.Errorf("PIDHigh mismatch: got %x expected %x", unmarshalled.PIDHigh, h.PIDHigh)
	}

	if unmarshalled.PIDLow != h.PIDLow {
		t.Errorf("PIDLow mismatch: got %x expected %x", unmarshalled.PIDLow, h.PIDLow)
	}

	if unmarshalled.Reserved != h.Reserved {
		t.Errorf("Reserved mismatch: got %x expected %x", unmarshalled.Reserved, h.Reserved)
	}

	if unmarshalled.TID != h.TID {
		t.Errorf("TID mismatch: got %x expected %x", unmarshalled.TID, h.TID)
	}

	if unmarshalled.UID != h.UID {
		t.Errorf("UID mismatch: got %x expected %x", unmarshalled.UID, h.UID)
	}

	if unmarshalled.MID != h.MID {
		t.Errorf("MID mismatch: got %x expected %x", unmarshalled.MID, h.MID)
	}
}

func TestUnmarshalFromHexHeader(t *testing.T) {
	// Test case with a hex-encoded SMB header
	hexHeader := "ff534d4272000000001805c0000000000000000000000000000000000000000006024e54204c4d20302e3132000000"

	// Convert hex string to byte array
	headerBytes, err := hex.DecodeString(hexHeader)
	if err != nil {
		t.Fatalf("Failed to decode hex header: %v", err)
	}

	// Create a new header to unmarshal into
	h := &header.Header{}

	// Unmarshal the header
	bytesRead, err := h.Unmarshal(headerBytes)
	if err != nil {
		t.Fatalf("Failed to unmarshal header from hex: %v", err)
	}

	// Verify bytes read matches expected length
	if bytesRead != header.SMB_HEADER_SIZE {
		t.Errorf("Bytes read mismatch: got %d, expected %d", bytesRead, header.SMB_HEADER_SIZE)
	}

	// Verify protocol is correct
	expectedProtocol := [4]byte{0xFF, 'S', 'M', 'B'}
	if h.Protocol != expectedProtocol {
		t.Errorf("Protocol mismatch: got %v, expected %v", h.Protocol, expectedProtocol)
	}

	// Verify other fields are zero as per the test hex string
	if h.Command != codes.SMB_COM_NEGOTIATE {
		t.Errorf("Command should be 0, got %s", h.Command)
	}

	if h.Status != 0 {
		t.Errorf("Status should be 0, got %x", h.Status)
	}

	if h.Flags != 0x18 {
		t.Errorf("Flags should be 0, got %x", h.Flags)
	}

	if h.Flags2 != 0xc005 {
		t.Errorf("Flags2 should be 0, got %x", h.Flags2)
	}
}
