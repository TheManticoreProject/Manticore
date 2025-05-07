package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestNewSMB_RESUME_KEY(t *testing.T) {
	resumeKey := types.NewSMB_RESUME_KEY()

	if resumeKey == nil {
		t.Fatal("NewSMB_RESUME_KEY returned nil")
	}

	if resumeKey.BufferFormat != types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK {
		t.Errorf("Expected BufferFormat to be %d, got %d", types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK, resumeKey.BufferFormat)
	}

	if resumeKey.Length != 0 {
		t.Errorf("Expected Length to be 0, got %d", resumeKey.Length)
	}

	if len(resumeKey.Buffer) != 0 {
		t.Errorf("Expected Buffer to be empty, got length %d", len(resumeKey.Buffer))
	}

	// Check that Reserved is initialized to zero
	if resumeKey.Reserved != 0 {
		t.Errorf("Expected Reserved to be 0, got %d", resumeKey.Reserved)
	}

	// Check that ServerState is initialized to zero bytes
	for i, b := range resumeKey.ServerState {
		if b != 0 {
			t.Errorf("Expected ServerState[%d] to be 0, got %d", i, b)
		}
	}

	// Check that ClientState is initialized to zero bytes
	for i, b := range resumeKey.ClientState {
		if b != 0 {
			t.Errorf("Expected ClientState[%d] to be 0, got %d", i, b)
		}
	}
}

func TestSMB_RESUME_KEY_Marshal(t *testing.T) {
	resumeKey := types.NewSMB_RESUME_KEY()

	// Set some test data in the fields
	resumeKey.Reserved = 1
	for i := 0; i < 16; i++ {
		resumeKey.ServerState[i] = byte(i + 2)
	}
	for i := 0; i < 4; i++ {
		resumeKey.ClientState[i] = byte(i + 18)
	}

	data, err := resumeKey.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Expected format:
	// - BufferFormat (1 byte): 0x05
	// - Length (2 bytes): 0x0015 (21 in little endian)
	// - Buffer (21 bytes): [1,2,3,...,21]
	expected := []byte{0x05, 0x16, 0x00}
	expected = append(expected, byte(1)) // Reserved
	// ServerState (16 bytes)
	expected = append(expected, []byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}...)
	// ClientState (4 bytes)
	expected = append(expected, []byte{18, 19, 20, 21}...)
	expected = append(expected, 0x00)

	if !bytes.Equal(data, expected) {
		t.Errorf("Marshal produced incorrect output\nExpected: %v\nGot: %v", expected, data)
	}
}

func TestSMB_RESUME_KEY_Unmarshal(t *testing.T) {
	resumeKey := types.NewSMB_RESUME_KEY()

	// Create test data
	testData := []byte{0x05, 0x15, 0x00}
	testData = append(testData, byte(1)) // Reserved
	// ServerState (16 bytes)
	for i := 2; i <= 17; i++ {
		testData = append(testData, byte(i))
	}
	// ClientState (4 bytes)
	for i := 18; i <= 21; i++ {
		testData = append(testData, byte(i))
	}

	bytesRead, err := resumeKey.Unmarshal(testData)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if bytesRead != len(testData) {
		t.Errorf("Expected to read %d bytes, got %d", len(testData), bytesRead)
	}

	if resumeKey.BufferFormat != types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK {
		t.Errorf("Expected BufferFormat to be %d, got %d", types.SMB_STRING_BUFFER_FORMAT_VARIABLE_BLOCK, resumeKey.BufferFormat)
	}

	if resumeKey.Length != 21 {
		t.Errorf("Expected Length to be 21, got %d", resumeKey.Length)
	}

	// Verify Reserved field contains the expected value
	if resumeKey.Reserved != 1 {
		t.Errorf("Expected Reserved to be 1, got %d", resumeKey.Reserved)
	}

	// Verify ServerState field contains the expected values
	for i := 0; i < 16; i++ {
		expected := byte(i + 2)
		if resumeKey.ServerState[i] != expected {
			t.Errorf("Expected ServerState[%d] to be %d, got %d", i, expected, resumeKey.ServerState[i])
		}
	}

	// Verify ClientState field contains the expected values
	for i := 0; i < 4; i++ {
		expected := byte(i + 18)
		if resumeKey.ClientState[i] != expected {
			t.Errorf("Expected ClientState[%d] to be %d, got %d", i, expected, resumeKey.ClientState[i])
		}
	}
}

func TestSMB_RESUME_KEY_Unmarshal_InvalidLength(t *testing.T) {
	resumeKey := types.NewSMB_RESUME_KEY()

	// Create test data with incorrect length (20 bytes instead of 21)
	testData := []byte{0x05, 0x14, 0x00}
	for i := 1; i <= 20; i++ {
		testData = append(testData, byte(i))
	}

	_, err := resumeKey.Unmarshal(testData)
	if err == nil {
		t.Fatal("Expected error for invalid buffer length, got nil")
	}
}

func TestSMB_RESUME_KEY_Unmarshal_EmptyData(t *testing.T) {
	resumeKey := types.NewSMB_RESUME_KEY()

	// Empty data
	testData := []byte{}

	_, err := resumeKey.Unmarshal(testData)
	if err == nil {
		t.Fatal("Expected error for empty data, got nil")
	}
}
