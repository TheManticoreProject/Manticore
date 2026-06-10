package avpair_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
)

func TestAvPairString(t *testing.T) {
	tests := []struct {
		avpair   avpair.AvPair
		expected string
	}{
		{
			avpair: avpair.AvPair{
				AvID:   avpair.MsvAvNbComputerName,
				AvLen:  4,
				AvData: []byte{0x01, 0x02, 0x03, 0x04},
			},
			expected: "AvId: MsvAvNbComputerName, AvLen: 4, AvData: [1 2 3 4]",
		},
		{
			avpair: avpair.AvPair{
				AvID:   avpair.MsvAvNbDomainName,
				AvLen:  3,
				AvData: []byte{0x05, 0x06, 0x07},
			},
			expected: "AvId: MsvAvNbDomainName, AvLen: 3, AvData: [5 6 7]",
		},
	}

	for _, test := range tests {
		if test.avpair.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.avpair.String())
		}
	}
}

func TestAvPairMarshal(t *testing.T) {
	tests := []struct {
		avpair   avpair.AvPair
		expected []byte
	}{
		{
			avpair: avpair.AvPair{
				AvID:   avpair.MsvAvNbComputerName,
				AvLen:  4,
				AvData: []byte{0x01, 0x02, 0x03, 0x04},
			},
			expected: []byte{0x01, 0x00, 0x04, 0x00, 0x01, 0x02, 0x03, 0x04},
		},
		{
			avpair: avpair.AvPair{
				AvID:   avpair.MsvAvNbDomainName,
				AvLen:  3,
				AvData: []byte{0x05, 0x06, 0x07},
			},
			expected: []byte{0x02, 0x00, 0x03, 0x00, 0x05, 0x06, 0x07},
		},
	}

	for _, test := range tests {
		marshaled, err := test.avpair.Marshal()
		if err != nil {
			t.Fatalf("Unexpected error during marshal: %v", err)
		}

		if !bytes.Equal(marshaled, test.expected) {
			t.Errorf("Expected %v, got %v", test.expected, marshaled)
		}
	}
}

func TestAvPairUnmarshal(t *testing.T) {
	tests := []struct {
		data           []byte
		expectedAvID   avpair.AvId
		expectedAvLen  uint16
		expectedAvData []byte
	}{
		{
			data:           []byte{0x01, 0x00, 0x04, 0x00, 0x01, 0x02, 0x03, 0x04},
			expectedAvID:   avpair.MsvAvNbComputerName,
			expectedAvLen:  4,
			expectedAvData: []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			data:           []byte{0x02, 0x00, 0x03, 0x00, 0x05, 0x06, 0x07},
			expectedAvID:   avpair.MsvAvNbDomainName,
			expectedAvLen:  3,
			expectedAvData: []byte{0x05, 0x06, 0x07},
		},
	}

	for _, test := range tests {
		av := avpair.AvPair{}
		_, err := av.Unmarshal(test.data)
		if err != nil {
			t.Fatalf("Unexpected error during unmarshal: %v", err)
		}

		if av.AvID != test.expectedAvID {
			t.Errorf("Expected AvID to be %x, got %x", test.expectedAvID, av.AvID)
		}

		if av.AvLen != test.expectedAvLen {
			t.Errorf("Expected AvLen to be %d, got %d", test.expectedAvLen, av.AvLen)
		}

		if !bytes.Equal(av.AvData, test.expectedAvData) {
			t.Errorf("Expected AvData to be %v, got %v", test.expectedAvData, av.AvData)
		}
	}
}

func TestAvPairUnmarshalTooShort(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "1 byte",
			data: []byte{0x01},
		},
		{
			name: "2 bytes",
			data: []byte{0x01, 0x00},
		},
		{
			name: "3 bytes",
			data: []byte{0x01, 0x00, 0x04},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			av := avpair.AvPair{}
			_, err := av.Unmarshal(test.data)
			if err == nil {
				t.Error("Expected error for too short data but got nil")
			}
		})
	}
}

// TestAvPairUnmarshalReturnsConsumedCount verifies that Unmarshal returns the
// number of bytes consumed (4-byte header + AvLen value) and that AvData is
// limited to exactly AvLen bytes when the buffer carries trailing data.
func TestAvPairUnmarshalReturnsConsumedCount(t *testing.T) {
	// AvId=MsvAvNbComputerName, AvLen=2, value {0xAA,0xBB}, plus 3 trailing bytes
	// belonging to a following AV_PAIR.
	data := []byte{0x01, 0x00, 0x02, 0x00, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE}

	av := avpair.AvPair{}
	n, err := av.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if n != 6 {
		t.Errorf("Unmarshal consumed = %d bytes, want 6 (4 header + 2 value)", n)
	}
	if !bytes.Equal(av.AvData, []byte{0xAA, 0xBB}) {
		t.Errorf("AvData = %v, want [0xAA 0xBB] (trailing bytes must be excluded)", av.AvData)
	}
}

// TestAvPairUnmarshalRejectsTruncatedValue verifies that an AvLen extending past
// the buffer is rejected rather than silently over-reading.
func TestAvPairUnmarshalRejectsTruncatedValue(t *testing.T) {
	// AvLen=0x10 (16) but only 2 value bytes are present.
	data := []byte{0x01, 0x00, 0x10, 0x00, 0xAA, 0xBB}

	av := avpair.AvPair{}
	if _, err := av.Unmarshal(data); err == nil {
		t.Fatal("Unmarshal should reject an AvLen that exceeds the buffer")
	}
}
