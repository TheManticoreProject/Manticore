package types_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

func TestSMB_FILE_ATTRIBUTES_GetAttributes(t *testing.T) {
	testCases := []struct {
		name           string
		attributes     uint16
		expectedResult uint16
	}{
		{
			name:           "Zero attributes",
			attributes:     0x0000,
			expectedResult: 0x0000,
		},
		{
			name:           "Read-only attribute",
			attributes:     0x0001,
			expectedResult: 0x0001,
		},
		{
			name:           "Hidden attribute",
			attributes:     0x0002,
			expectedResult: 0x0002,
		},
		{
			name:           "Multiple attributes",
			attributes:     0x0013, // Read-only, System, Directory
			expectedResult: 0x0013,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileAttr := &types.SMB_FILE_ATTRIBUTES{
				Attributes: tc.attributes,
			}

			result := fileAttr.GetAttributes()
			if result != tc.expectedResult {
				t.Errorf("Expected attributes %#04x, got %#04x", tc.expectedResult, result)
			}
		})
	}
}

func TestSMB_FILE_ATTRIBUTES_SetAttributes(t *testing.T) {
	testCases := []struct {
		name               string
		initialAttributes  uint16
		attributesToSet    uint16
		expectedAttributes uint16
	}{
		{
			name:               "Set zero attributes",
			initialAttributes:  0x0001,
			attributesToSet:    0x0000,
			expectedAttributes: 0x0000,
		},
		{
			name:               "Set single attribute",
			initialAttributes:  0x0000,
			attributesToSet:    0x0004, // System
			expectedAttributes: 0x0004,
		},
		{
			name:               "Set multiple attributes",
			initialAttributes:  0x0001, // Read-only
			attributesToSet:    0x0022, // Hidden, Archive
			expectedAttributes: 0x0022,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileAttr := &types.SMB_FILE_ATTRIBUTES{
				Attributes: tc.initialAttributes,
			}

			fileAttr.SetAttributes(tc.attributesToSet)
			if fileAttr.Attributes != tc.expectedAttributes {
				t.Errorf("Expected attributes %#04x, got %#04x", tc.expectedAttributes, fileAttr.Attributes)
			}
		})
	}
}

func TestSMB_FILE_ATTRIBUTES_Marshal(t *testing.T) {
	testCases := []struct {
		name           string
		attributes     uint16
		expectedOutput []byte
	}{
		{
			name:           "Zero attributes",
			attributes:     0x0000,
			expectedOutput: []byte{0x00, 0x00},
		},
		{
			name:           "Read-only attribute",
			attributes:     0x0001,
			expectedOutput: []byte{0x00, 0x01},
		},
		{
			name:           "Multiple attributes",
			attributes:     0x0037, // Read-only, Hidden, System, Directory, Archive
			expectedOutput: []byte{0x00, 0x37},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileAttr := &types.SMB_FILE_ATTRIBUTES{
				Attributes: tc.attributes,
			}

			result, err := fileAttr.Marshal()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !bytes.Equal(result, tc.expectedOutput) {
				t.Errorf("Marshal output mismatch\nExpected: %v\nGot: %v", tc.expectedOutput, result)
			}
		})
	}
}

func TestSMB_FILE_ATTRIBUTES_Unmarshal(t *testing.T) {
	testCases := []struct {
		name              string
		input             []byte
		expectedAttr      uint16
		expectedBytesRead int
	}{
		{
			name:              "Zero attributes",
			input:             []byte{0x00, 0x00},
			expectedAttr:      0x0000,
			expectedBytesRead: 2,
		},
		{
			name:              "Read-only attribute",
			input:             []byte{0x00, 0x01},
			expectedAttr:      0x0001,
			expectedBytesRead: 2,
		},
		{
			name:              "Multiple attributes",
			input:             []byte{0x00, 0x37, 0xFF}, // Extra byte should be ignored
			expectedAttr:      0x0037,
			expectedBytesRead: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileAttr := &types.SMB_FILE_ATTRIBUTES{}

			bytesRead, err := fileAttr.Unmarshal(tc.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if bytesRead != tc.expectedBytesRead {
				t.Errorf("Expected to read %d bytes, got %d", tc.expectedBytesRead, bytesRead)
			}

			if fileAttr.Attributes != tc.expectedAttr {
				t.Errorf("Expected attributes %#04x, got %#04x", tc.expectedAttr, fileAttr.Attributes)
			}
		})
	}
}
