package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// The transaction-family requests lay out their SMB_Data block as
// [Name?] | Pad1 | <Parameters> | Pad2 | <Data>. ParameterOffset and DataOffset
// are header-relative offsets, not in-block lengths, so the Unmarshal must derive
// the Pad1/Pad2 lengths from the offset differences rather than using the offsets
// directly. These round-trip tests use a fixed, mutually consistent layout:
//
//	Pad1 = 1 byte, Parameters = 3 bytes, Pad2 = 2 bytes, Data = 4 bytes
//	ParameterOffset = 59, DataOffset = 64 (= 59 + ParameterCount(3) + len(Pad2)(2))
//
// Before the fix the Unmarshal sliced Pad1 with length ParameterOffset (59) and
// Pad2 with length DataOffset (64), which runs past the end of the small data
// block and fails to round-trip.
var (
	padTestPad1    = []types.UCHAR{0xAA}
	padTestParams  = []types.UCHAR{0x01, 0x02, 0x03}
	padTestPad2    = []types.UCHAR{0xBB, 0xCC}
	padTestData    = []types.UCHAR{0x10, 0x20, 0x30, 0x40}
	padTestParamOf = types.USHORT(59)
	padTestDataOf  = types.USHORT(64)
)

func assertPadRoundTrip(t *testing.T, name string, pad1, params, pad2, data []types.UCHAR) {
	t.Helper()
	if !bytes.Equal(pad1, padTestPad1) {
		t.Errorf("%s: Pad1 = % x, want % x", name, pad1, padTestPad1)
	}
	if !bytes.Equal(params, padTestParams) {
		t.Errorf("%s: Parameters = % x, want % x", name, params, padTestParams)
	}
	if !bytes.Equal(pad2, padTestPad2) {
		t.Errorf("%s: Pad2 = % x, want % x", name, pad2, padTestPad2)
	}
	if !bytes.Equal(data, padTestData) {
		t.Errorf("%s: Data = % x, want % x", name, data, padTestData)
	}
}

func marshalUnmarshal(t *testing.T, out, in interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) (int, error)
}) {
	t.Helper()
	raw, err := out.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if _, err := in.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}

func TestTransactionRequestPadRoundTrip(t *testing.T) {
	out := commands.NewTransactionRequest()
	out.Name = types.SMB_STRING{BufferFormat: types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING, Buffer: []types.UCHAR{}}
	out.Pad1 = padTestPad1
	out.Trans_Parameters = padTestParams
	out.Pad2 = padTestPad2
	out.Trans_Data = padTestData
	out.ParameterCount = types.USHORT(len(padTestParams))
	out.DataCount = types.USHORT(len(padTestData))
	out.ParameterOffset = padTestParamOf
	out.DataOffset = padTestDataOf

	in := commands.NewTransactionRequest()
	marshalUnmarshal(t, out, in)
	assertPadRoundTrip(t, "TransactionRequest", in.Pad1, in.Trans_Parameters, in.Pad2, in.Trans_Data)
}

func TestTransaction2RequestPadRoundTrip(t *testing.T) {
	out := commands.NewTransaction2Request()
	out.Name = types.UCHAR(0x00)
	out.Pad1 = padTestPad1
	out.Trans2_Parameters = padTestParams
	out.Pad2 = padTestPad2
	out.Trans2_Data = padTestData
	out.ParameterCount = types.USHORT(len(padTestParams))
	out.DataCount = types.USHORT(len(padTestData))
	out.ParameterOffset = padTestParamOf
	out.DataOffset = padTestDataOf

	in := commands.NewTransaction2Request()
	marshalUnmarshal(t, out, in)
	assertPadRoundTrip(t, "Transaction2Request", in.Pad1, in.Trans2_Parameters, in.Pad2, in.Trans2_Data)
}

func TestNtTransactRequestPadRoundTrip(t *testing.T) {
	out := commands.NewNtTransactRequest()
	out.Pad1 = padTestPad1
	out.NT_Trans_Parameters = padTestParams
	out.Pad2 = padTestPad2
	out.NT_Trans_Data = padTestData
	out.ParameterCount = types.ULONG(len(padTestParams))
	out.DataCount = types.ULONG(len(padTestData))
	out.ParameterOffset = types.ULONG(padTestParamOf)
	out.DataOffset = types.ULONG(padTestDataOf)

	in := commands.NewNtTransactRequest()
	marshalUnmarshal(t, out, in)
	assertPadRoundTrip(t, "NtTransactRequest", in.Pad1, in.NT_Trans_Parameters, in.Pad2, in.NT_Trans_Data)
}

func TestTransactionSecondaryRequestPadRoundTrip(t *testing.T) {
	out := commands.NewTransactionSecondaryRequest()
	out.Pad1 = padTestPad1
	out.Trans2_Parameters = padTestParams
	out.Pad2 = padTestPad2
	out.Trans2_Data = padTestData
	out.ParameterCount = types.USHORT(len(padTestParams))
	out.DataCount = types.USHORT(len(padTestData))
	out.ParameterOffset = padTestParamOf
	out.DataOffset = padTestDataOf

	in := commands.NewTransactionSecondaryRequest()
	marshalUnmarshal(t, out, in)
	assertPadRoundTrip(t, "TransactionSecondaryRequest", in.Pad1, in.Trans2_Parameters, in.Pad2, in.Trans2_Data)
}

func TestTransaction2SecondaryRequestPadRoundTrip(t *testing.T) {
	out := commands.NewTransaction2SecondaryRequest()
	out.Pad1 = padTestPad1
	out.Trans2_Parameters = padTestParams
	out.Pad2 = padTestPad2
	out.Trans2_Data = padTestData
	out.ParameterCount = types.USHORT(len(padTestParams))
	out.DataCount = types.USHORT(len(padTestData))
	out.ParameterOffset = padTestParamOf
	out.DataOffset = padTestDataOf

	in := commands.NewTransaction2SecondaryRequest()
	marshalUnmarshal(t, out, in)
	assertPadRoundTrip(t, "Transaction2SecondaryRequest", in.Pad1, in.Trans2_Parameters, in.Pad2, in.Trans2_Data)
}
