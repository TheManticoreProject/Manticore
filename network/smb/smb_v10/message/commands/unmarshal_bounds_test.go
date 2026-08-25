package commands

import (
	"encoding/binary"
	"testing"
)

// A decoder reachable from the network must refuse a declared length that runs
// past the buffer rather than slicing past its end. These three sites guarded only
// the first byte of the slice while its length came from a client-controlled
// 16-bit field, so any peer could crash the process with one short message.
//
// The class matters more than the individual sites: a guard that is present but
// bounds the wrong expression looks correct in review and in a grep for guards.

// TestWriteMpxRequestUnmarshalRejectsOversizePad asserts a DataOffset larger than
// the data block is refused.
func TestWriteMpxRequestUnmarshalRejectsOversizePad(t *testing.T) {
	// WordCount 12, then the twelve parameter words. DataOffset is the last one
	// and claims 0x3030 bytes of padding.
	parameters := make([]byte, 1+12*2)
	parameters[0] = 0x0C
	binary.LittleEndian.PutUint16(parameters[1+10*2:], 0x0010) // DataLength
	binary.LittleEndian.PutUint16(parameters[1+11*2:], 0x3030) // DataOffset

	// A data block far shorter than the padding the parameters claim.
	data := []byte{0x04, 0x00, 0x01, 0x02, 0x03, 0x04}

	request := NewWriteMpxRequest()
	if _, err := request.Unmarshal(append(parameters, data...)); err == nil {
		t.Fatal("Unmarshal() accepted a DataOffset larger than the data block")
	}
}

// TestSessionSetupAndxRequestUnmarshalRejectsOversizePasswords asserts a password
// length larger than the data block is refused, for both the OEM and the Unicode
// field.
//
// These are on the non-extended-security path, which a client can select simply by
// sending that layout — so a server decoding a request reaches them.
func TestSessionSetupAndxRequestUnmarshalRejectsOversizePasswords(t *testing.T) {
	// WordCount 10 selects the non-extended-security layout, whose parameter
	// words end with the two password lengths.
	build := func(oemLen, unicodeLen uint16) []byte {
		parameters := make([]byte, 1+10*2)
		parameters[0] = 0x0A
		binary.LittleEndian.PutUint16(parameters[1+0*2:], 0x4100)   // MaxBufferSize
		binary.LittleEndian.PutUint16(parameters[1+1*2:], 0x0032)   // MaxMpxCount
		binary.LittleEndian.PutUint16(parameters[1+2*2:], 0x0000)   // VcNumber
		binary.LittleEndian.PutUint32(parameters[1+3*2:], 0x000000) // SessionKey
		binary.LittleEndian.PutUint16(parameters[1+5*2:], oemLen)
		binary.LittleEndian.PutUint16(parameters[1+6*2:], unicodeLen)

		// A data block far shorter than the passwords the parameters claim.
		data := []byte{0x06, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
		return append(parameters, data...)
	}

	t.Run("OEM password", func(t *testing.T) {
		request := NewSessionSetupAndxRequest()
		if _, err := request.Unmarshal(build(0x3030, 0)); err == nil {
			t.Fatal("Unmarshal() accepted an OEMPasswordLen larger than the data block")
		}
	})

	t.Run("Unicode password", func(t *testing.T) {
		request := NewSessionSetupAndxRequest()
		if _, err := request.Unmarshal(build(0, 0x3030)); err == nil {
			t.Fatal("Unmarshal() accepted a UnicodePasswordLen larger than the data block")
		}
	})
}
