package andx_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/andx"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
)

func TestAndX_NewAndX(t *testing.T) {
	andx := andx.NewAndX()
	if andx == nil {
		t.Fatal("NewAndX() returned nil")
	}
	if andx.AndXCommand != 0 {
		t.Errorf("Expected AndXCommand to be 0, got %d", andx.AndXCommand)
	}
	if andx.AndXReserved != 0 {
		t.Errorf("Expected AndXReserved to be 0, got %d", andx.AndXReserved)
	}
	if andx.AndXOffset != 0 {
		t.Errorf("Expected AndXOffset to be 0, got %d", andx.AndXOffset)
	}
}

func TestAndX_GetParameters(t *testing.T) {
	andx := andx.NewAndX()
	andx.AndXCommand = 0x2A
	andx.AndXReserved = 0x00
	andx.AndXOffset = 0x1234

	params := andx.GetParameters()
	if len(params) != 2 {
		t.Fatalf("Expected 2 parameters, got %d", len(params))
	}
	if params[0] != 0x2A00 {
		t.Errorf("Expected first parameter to be 0x2A00, got 0x%04X", params[0])
	}
	// The words carry wire bytes, not values: AndXOffset is little-endian, so
	// 0x1234 occupies the bytes 34 12 and packs high byte first as 0x3412. This
	// expectation was 0x1234 while the offset was going out byte-swapped.
	if params[1] != 0x3412 {
		t.Errorf("Expected second parameter to be 0x3412, got 0x%04X", params[1])
	}
}

func TestAndX_Marshal(t *testing.T) {
	testCases := []struct {
		name        string
		andx        *andx.AndX
		expected    []byte
		expectError bool
	}{
		{
			name: "Valid AndX",
			andx: &andx.AndX{
				AndXCommand:  0x2A,
				AndXReserved: 0x00,
				AndXOffset:   0x1234,
			},
			expected:    []byte{0x2A, 0x00, 0x34, 0x12},
			expectError: false,
		},
		{
			name: "Zero values",
			andx: &andx.AndX{
				AndXCommand:  0x00,
				AndXReserved: 0x00,
				AndXOffset:   0x0000,
			},
			expected:    []byte{0x00, 0x00, 0x00, 0x00},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.andx.Marshal()
			if tc.expectError && err == nil {
				t.Fatal("Expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(data) != len(tc.expected) {
				t.Fatalf("Expected %d bytes, got %d", len(tc.expected), len(data))
			}
			for i := 0; i < len(data); i++ {
				if data[i] != tc.expected[i] {
					t.Errorf("Byte %d: expected 0x%02X, got 0x%02X", i, tc.expected[i], data[i])
				}
			}
		})
	}
}

func TestAndX_Unmarshal(t *testing.T) {
	testCases := []struct {
		name        string
		data        []byte
		expected    *andx.AndX
		expectError bool
	}{
		{
			name: "Valid data",
			data: []byte{0x2A, 0x00, 0x34, 0x12},
			expected: &andx.AndX{
				AndXCommand:  0x2A,
				AndXReserved: 0x00,
				AndXOffset:   0x1234,
			},
			expectError: false,
		},
		{
			name: "Zero values",
			data: []byte{0x00, 0x00, 0x00, 0x00},
			expected: &andx.AndX{
				AndXCommand:  0x00,
				AndXReserved: 0x00,
				AndXOffset:   0x0000,
			},
			expectError: false,
		},
		{
			name:        "Too short data",
			data:        []byte{0x2A, 0x00},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			andx := andx.NewAndX()
			bytesRead, err := andx.Unmarshal(tc.data)
			if tc.expectError && err == nil {
				t.Fatal("Expected error but got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if bytesRead != len(tc.data) && tc.expected != nil {
				t.Errorf("Expected %d bytes read, got %d", len(tc.data), bytesRead)
			}
			if !tc.expectError {
				if andx.AndXCommand != tc.expected.AndXCommand {
					t.Errorf("Expected AndXCommand to be 0x%02X, got 0x%02X", tc.expected.AndXCommand, andx.AndXCommand)
				}
				if andx.AndXReserved != tc.expected.AndXReserved {
					t.Errorf("Expected AndXReserved to be 0x%02X, got 0x%02X", tc.expected.AndXReserved, andx.AndXReserved)
				}
				if andx.AndXOffset != tc.expected.AndXOffset {
					t.Errorf("Expected AndXOffset to be 0x%04X, got 0x%04X", tc.expected.AndXOffset, andx.AndXOffset)
				}
			}
		})
	}
}

// TestGetParametersMatchesMarshal is the regression guard for an AndX block that
// went out with its offset byte-swapped: GetParameters feeds a container that
// splits each word high byte first, so the words it returns must reassemble into
// exactly the bytes Marshal produces for the same structure.
func TestGetParametersMatchesMarshal(t *testing.T) {
	for _, offset := range []uint16{0x0000, 0x0001, 0x1234, 0x0100, 0xFF00, 0xFFFF} {
		block := &andx.AndX{
			AndXCommand:  codes.SMB_COM_NO_ANDX_COMMAND,
			AndXReserved: 0x00,
			AndXOffset:   offset,
		}

		want, err := block.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}

		// Reassemble the words the way Parameters.GetBytesStream does.
		got := []byte{}
		for _, word := range block.GetParameters() {
			got = append(got, uint8(word>>8), uint8(word&0xFF))
		}

		if !bytes.Equal(got, want) {
			t.Errorf("offset 0x%04x: GetParameters gives % x, Marshal gives % x", offset, got, want)
		}
	}
}

// TestGetParametersOffsetIsLittleEndian pins the byte order of the offset field
// itself, so the pairing above cannot be satisfied by both methods being wrong
// together.
func TestGetParametersOffsetIsLittleEndian(t *testing.T) {
	block := &andx.AndX{AndXCommand: codes.SMB_COM_NO_ANDX_COMMAND, AndXOffset: 0x1234}

	words := block.GetParameters()
	if len(words) != 2 {
		t.Fatalf("GetParameters returned %d words, want 2", len(words))
	}
	if got := words[0]; got != 0xFF00 {
		t.Errorf("first word = 0x%04x, want 0xff00 (AndXCommand then AndXReserved)", got)
	}
	// High byte first packing of the wire bytes 34 12.
	if got := words[1]; got != 0x3412 {
		t.Errorf("second word = 0x%04x, want 0x3412 (the wire bytes 34 12)", got)
	}
}
