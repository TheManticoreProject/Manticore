package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestSMB_NMPIPE_STATUS_SetICount(t *testing.T) {
	status := types.SMB_NMPIPE_STATUS{}
	status.SetICount(5)
	if status.ICount != 5 {
		t.Errorf("Expected ICount to be 5, got %d", status.ICount)
	}
}

func TestSMB_NMPIPE_STATUS_GetICount(t *testing.T) {
	status := types.SMB_NMPIPE_STATUS{ICount: 10}
	if status.GetICount() != 10 {
		t.Errorf("Expected GetICount() to return 10, got %d", status.GetICount())
	}
}

func TestSMB_NMPIPE_STATUS_SetNonBlockingStatus(t *testing.T) {
	tests := []struct {
		name        string
		initialFlag uint8
		setTo       bool
		expected    uint8
	}{
		{
			name:        "Set to true from 0",
			initialFlag: 0,
			setTo:       true,
			expected:    types.SMB_NMPIPE_STATUS_NONBLOCKING,
		},
		{
			name:        "Set to false from NONBLOCKING",
			initialFlag: types.SMB_NMPIPE_STATUS_NONBLOCKING,
			setTo:       false,
			expected:    0,
		},
		{
			name:        "Set to true with other flags",
			initialFlag: types.SMB_NMPIPE_STATUS_ENDPOINT_SERVER,
			setTo:       true,
			expected:    types.SMB_NMPIPE_STATUS_ENDPOINT_SERVER | types.SMB_NMPIPE_STATUS_NONBLOCKING,
		},
		{
			name:        "Set to false with other flags",
			initialFlag: types.SMB_NMPIPE_STATUS_ENDPOINT_SERVER | types.SMB_NMPIPE_STATUS_NONBLOCKING,
			setTo:       false,
			expected:    types.SMB_NMPIPE_STATUS_ENDPOINT_SERVER,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := types.SMB_NMPIPE_STATUS{Flags: tt.initialFlag}
			status.SetNonBlockingStatus(tt.setTo)
			if status.Flags != tt.expected {
				t.Errorf("Expected Flags to be %d, got %d", tt.expected, status.Flags)
			}
		})
	}
}

func TestSMB_NMPIPE_STATUS_IsNonBlocking(t *testing.T) {
	tests := []struct {
		name     string
		flags    uint8
		expected bool
	}{
		{
			name:     "NONBLOCKING flag set",
			flags:    types.SMB_NMPIPE_STATUS_NONBLOCKING,
			expected: true,
		},
		{
			name:     "NONBLOCKING flag not set",
			flags:    0,
			expected: false,
		},
		{
			name:     "NONBLOCKING flag set with other flags",
			flags:    types.SMB_NMPIPE_STATUS_ENDPOINT_SERVER | types.SMB_NMPIPE_STATUS_NONBLOCKING,
			expected: true,
		},
		{
			name:     "NONBLOCKING flag not set with other flags",
			flags:    types.SMB_NMPIPE_STATUS_ENDPOINT_SERVER,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := types.SMB_NMPIPE_STATUS{Flags: tt.flags}
			if status.IsNonBlocking() != tt.expected {
				t.Errorf("Expected IsNonBlocking() to return %v, got %v", tt.expected, status.IsNonBlocking())
			}
		})
	}
}

func TestSMB_NMPIPE_STATUS_GetReadMode(t *testing.T) {
	tests := []struct {
		name     string
		flags    uint8
		expected uint8
	}{
		{
			name:     "Read mode BYTE",
			flags:    types.SMB_NMPIPE_STATUS_READ_MODE_BYTE,
			expected: types.SMB_NMPIPE_STATUS_READ_MODE_BYTE,
		},
		{
			name:     "Read mode MESSAGE",
			flags:    types.SMB_NMPIPE_STATUS_READ_MODE_MESSAGE,
			expected: types.SMB_NMPIPE_STATUS_READ_MODE_MESSAGE,
		},
		{
			name:     "Read mode with other flags",
			flags:    types.SMB_NMPIPE_STATUS_READ_MODE_MESSAGE | types.SMB_NMPIPE_STATUS_ENDPOINT_SERVER,
			expected: types.SMB_NMPIPE_STATUS_READ_MODE_MESSAGE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := types.SMB_NMPIPE_STATUS{Flags: tt.flags}
			if status.GetReadMode() != tt.expected {
				t.Errorf("Expected GetReadMode() to return %d, got %d", tt.expected, status.GetReadMode())
			}
		})
	}
}

func TestSMB_NMPIPE_STATUS_String(t *testing.T) {
	status := types.SMB_NMPIPE_STATUS{ICount: 5, Flags: 0x81}
	expected := "ICount: 5, Flags: 129"
	if status.String() != expected {
		t.Errorf("Expected String() to return %q, got %q", expected, status.String())
	}
}

func TestSMB_NMPIPE_STATUS_Marshal(t *testing.T) {
	status := types.SMB_NMPIPE_STATUS{ICount: 5, Flags: 0x81}
	expected := []byte{5, 0x81}
	result, err := status.Marshal()
	if err != nil {
		t.Errorf("Expected Marshal() to return %v, got %v", expected, err)
	}

	if !bytes.Equal(result, expected) {
		t.Errorf("Expected Marshal() to return %v, got %v", expected, result)
	}
}

func TestSMB_NMPIPE_STATUS_Unmarshal(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expected    types.SMB_NMPIPE_STATUS
		expectedLen int
		expectError bool
	}{
		{
			name:        "Valid data",
			data:        []byte{5, 0x81},
			expected:    types.SMB_NMPIPE_STATUS{ICount: 5, Flags: 0x81},
			expectedLen: 2,
			expectError: false,
		},
		{
			name:        "Invalid data length",
			data:        []byte{5},
			expected:    types.SMB_NMPIPE_STATUS{},
			expectedLen: 0,
			expectError: true,
		},
		{
			name:        "Invalid data length (too long)",
			data:        []byte{5, 0x81, 0x00},
			expected:    types.SMB_NMPIPE_STATUS{},
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := types.SMB_NMPIPE_STATUS{}
			n, err := status.Unmarshal(tt.data)

			if tt.expectError && err == nil {
				t.Errorf("Expected an error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError {
				if status.ICount != tt.expected.ICount {
					t.Errorf("Expected ICount to be %d, got %d", tt.expected.ICount, status.ICount)
				}

				if status.Flags != tt.expected.Flags {
					t.Errorf("Expected Flags to be %d, got %d", tt.expected.Flags, status.Flags)
				}

				if n != tt.expectedLen {
					t.Errorf("Expected Unmarshal to return length %d, got %d", tt.expectedLen, n)
				}
			}
		})
	}
}
