package commands

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// The tests in this file declare lengths and offsets whose sum overflows a 32-bit
// int. They are only meaningful under GOARCH=386, where an unguarded start+count
// wraps negative, passes the bounds check, and panics on the slice; on a 64-bit
// build they still pin the error.

// mustNotPanic runs decode and fails the test, rather than the process, if it panics.
func mustNotPanic(t *testing.T, name string, decode func() error) error {
	t.Helper()
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("%s panicked: %v", name, r)
			}
		}()
		err = decode()
	}()
	return err
}

func TestWriteAndxRequestDataLengthHighOverflowRejected(t *testing.T) {
	for _, high := range []uint16{0x7FFF, 0x8000, 0xFFFF} {
		request := NewWriteAndxRequest()
		request.FID = 1
		request.Data = []types.UCHAR(countingPayload(64))
		raw, err := request.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		binary.LittleEndian.PutUint16(raw[writeAndxDataLengthHighAt:], high)

		err = mustNotPanic(t, "WriteAndxRequest.Unmarshal", func() error {
			_, err := NewWriteAndxRequest().Unmarshal(raw)
			return err
		})
		if err == nil {
			t.Errorf("DataLengthHigh 0x%04X: Unmarshal accepted a length past the end of the message", high)
		}
	}
}

func TestReadAndxResponseDataLengthHighOverflowRejected(t *testing.T) {
	for _, high := range []uint16{0x7FFF, 0x8000, 0xFFFF} {
		response := NewReadAndxResponse()
		response.Data = []types.UCHAR(countingPayload(64))
		raw, err := response.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		// WordCount(1) AndX(4) Available(2) DataCompactionMode(2) Reserved1(2)
		// DataLength(2) DataOffset(2) DataLengthHigh(2).
		if binary.LittleEndian.Uint16(raw[11:13]) != 64 {
			t.Fatalf("DataLength is not where expected: % x", raw[:20])
		}
		binary.LittleEndian.PutUint16(raw[15:17], high)

		err = mustNotPanic(t, "ReadAndxResponse.Unmarshal", func() error {
			_, err := NewReadAndxResponse().Unmarshal(raw)
			return err
		})
		if err == nil {
			t.Errorf("DataLengthHigh 0x%04X: Unmarshal accepted a length past the end of the message", high)
		}
	}
}

func TestNtTransactResponseCountOverflowRejected(t *testing.T) {
	// ParameterCount sits at 12 and DataCount at 24: WordCount(1) Reserved1(3)
	// TotalParameterCount(4) TotalDataCount(4) ParameterCount(4)
	// ParameterOffset(4) ParameterDisplacement(4) DataCount(4).
	for _, field := range []struct {
		name string
		at   int
	}{{"ParameterCount", 12}, {"DataCount", 24}} {
		for _, count := range []uint32{0x7FFFFFF0, 0x80000000, 0xFFFFFFF0} {
			r := NewNtTransactResponse()
			r.TotalParameterCount, r.ParameterCount = 4, 4
			r.Parameters = []types.UCHAR{1, 2, 3, 4}
			r.ParameterOffset = types.ULONG(32 + 37 + 2)
			r.Pad1 = []types.UCHAR{}
			r.Pad2 = []types.UCHAR{}
			r.Data = []types.UCHAR{}
			raw, err := r.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if binary.LittleEndian.Uint32(raw[12:16]) != 4 {
				t.Fatalf("ParameterCount is not where expected: % x", raw[:28])
			}
			binary.LittleEndian.PutUint32(raw[field.at:], count)

			err = mustNotPanic(t, "NtTransactResponse.Unmarshal", func() error {
				_, err := NewNtTransactResponse().Unmarshal(raw)
				return err
			})
			if err == nil {
				t.Errorf("%s 0x%08X: Unmarshal accepted a block past the end of the message", field.name, count)
			}
		}
	}
}

func TestNtTransactRequestOffsetCountOverflowRejected(t *testing.T) {
	r := NewNtTransactRequest()
	r.TotalParameterCount, r.ParameterCount = 4, 4
	r.NT_Trans_Parameters = []types.UCHAR{1, 2, 3, 4}
	raw, err := r.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	// ParameterCount at 20 and ParameterOffset at 24: WordCount(1)
	// MaxSetupCount(1) Reserved1(2) TotalParameterCount(4) TotalDataCount(4)
	// MaxParameterCount(4) MaxDataCount(4).
	if binary.LittleEndian.Uint32(raw[20:24]) != 4 {
		t.Fatalf("ParameterCount is not where expected: % x", raw[:28])
	}
	binary.LittleEndian.PutUint32(raw[20:24], 0x7FFFFFF0)
	binary.LittleEndian.PutUint32(raw[24:28], 0x7FFFFFF0)

	err = mustNotPanic(t, "NtTransactRequest.Unmarshal", func() error {
		_, err := NewNtTransactRequest().Unmarshal(raw)
		return err
	})
	if err == nil {
		t.Error("Unmarshal accepted a parameter block past the end of the message")
	}
}

func TestLocateTransactionBlockOffsetCountOverflowRejected(t *testing.T) {
	for _, tc := range []struct{ offset, count int }{
		{0x7FFFFFF0, 0x7FFFFFF0},
		{80 + 8, 0x7FFFFFFC},
		{80 + 17, 1},
	} {
		var err error
		mustNotPanic(t, "locateTransactionBlock", func() error {
			_, _, err = locateTransactionBlock(make([]byte, 16), 80, tc.offset, tc.count, "the parameter block")
			return err
		})
		if err == nil {
			t.Errorf("offset %d count %d: accepted a block outside the 16-byte data block", tc.offset, tc.count)
		}
	}
}
