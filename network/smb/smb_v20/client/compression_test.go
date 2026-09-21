package client

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

func TestCompressionTransformHeaderRoundTrip(t *testing.T) {
	h := &CompressionTransformHeader{
		OriginalCompressedSegmentSize: 4096,
		CompressionAlgorithm:          commands.SMB2_COMPRESSION_LZ77_HUFFMAN,
		Flags:                         SMB2_COMPRESSION_FLAG_NONE,
		Offset:                        24,
	}

	wire := MarshalCompressionTransformHeader(h)
	if len(wire) != compressionTransformHeaderSize {
		t.Fatalf("wire length = %d, want %d", len(wire), compressionTransformHeaderSize)
	}
	if wire[0] != 0xFC || wire[1] != 'S' || wire[2] != 'M' || wire[3] != 'B' {
		t.Errorf("protocol ID = %02x %02x %02x %02x, want FC 53 4D 42", wire[0], wire[1], wire[2], wire[3])
	}

	parsed, err := ParseCompressionTransformHeader(wire)
	if err != nil {
		t.Fatalf("ParseCompressionTransformHeader: %v", err)
	}
	if parsed.OriginalCompressedSegmentSize != 4096 {
		t.Errorf("OriginalCompressedSegmentSize = %d, want 4096", parsed.OriginalCompressedSegmentSize)
	}
	if parsed.CompressionAlgorithm != commands.SMB2_COMPRESSION_LZ77_HUFFMAN {
		t.Errorf("CompressionAlgorithm = 0x%04x, want 0x%04x", parsed.CompressionAlgorithm, commands.SMB2_COMPRESSION_LZ77_HUFFMAN)
	}
	if parsed.Flags != SMB2_COMPRESSION_FLAG_NONE {
		t.Errorf("Flags = 0x%04x, want 0x%04x", parsed.Flags, SMB2_COMPRESSION_FLAG_NONE)
	}
	if parsed.Offset != 24 {
		t.Errorf("Offset = %d, want 24", parsed.Offset)
	}
}

func TestParseCompressionTransformHeaderTooShort(t *testing.T) {
	_, err := ParseCompressionTransformHeader(make([]byte, 10))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestParseCompressionTransformHeaderBadProtocolId(t *testing.T) {
	data := make([]byte, 16)
	data[0] = 0xFE // SMB2 header, not compression
	data[1] = 'S'
	data[2] = 'M'
	data[3] = 'B'
	_, err := ParseCompressionTransformHeader(data)
	if err == nil {
		t.Error("expected error for bad protocol ID")
	}
}

func TestIsCompressedMessage(t *testing.T) {
	if IsCompressedMessage([]byte{0xFE, 'S', 'M', 'B'}) {
		t.Error("0xFE SMB header should not be identified as compressed")
	}
	if !IsCompressedMessage([]byte{0xFC, 'S', 'M', 'B', 0x00}) {
		t.Error("0xFC SMB should be identified as compressed")
	}
	if IsCompressedMessage([]byte{0xFC}) {
		t.Error("short data should not be identified as compressed")
	}
}

func TestCompressPatternV1(t *testing.T) {
	data := bytes.Repeat([]byte{0xAA}, 256)
	compressed := CompressPatternV1(data)
	if compressed == nil {
		t.Fatal("expected compression result, got nil")
	}
	if len(compressed) != 4 {
		t.Fatalf("compressed length = %d, want 4", len(compressed))
	}
	if compressed[0] != 0xAA {
		t.Errorf("pattern byte = 0x%02x, want 0xAA", compressed[0])
	}

	decompressed, err := DecompressPatternV1(compressed)
	if err != nil {
		t.Fatalf("DecompressPatternV1: %v", err)
	}
	if !bytes.Equal(decompressed, data) {
		t.Errorf("decompressed length = %d, want %d", len(decompressed), len(data))
	}
}

func TestCompressPatternV1MixedData(t *testing.T) {
	data := []byte{0xAA, 0xBB, 0xCC}
	compressed := CompressPatternV1(data)
	if compressed != nil {
		t.Error("expected nil for non-uniform data")
	}
}

func TestCompressPatternV1Empty(t *testing.T) {
	compressed := CompressPatternV1(nil)
	if compressed != nil {
		t.Error("expected nil for empty data")
	}
}

func TestDecompressPatternV1TooShort(t *testing.T) {
	_, err := DecompressPatternV1([]byte{0xAA, 0x00})
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestDecompressPatternV1ReservedByteNonZero(t *testing.T) {
	// Pattern_V1: pattern(1) + reserved(1) + count(2). Reserved MUST be zero.
	_, err := DecompressPatternV1([]byte{0xAA, 0xFF, 0x03, 0x00})
	if err == nil {
		t.Error("expected error for non-zero reserved byte")
	}
}

func TestCompressionNegotiateContext(t *testing.T) {
	algs := []uint16{
		commands.SMB2_COMPRESSION_LZ77_HUFFMAN,
		commands.SMB2_COMPRESSION_LZ77,
		commands.SMB2_COMPRESSION_LZNT1,
		commands.SMB2_COMPRESSION_PATTERN_V1,
	}
	ctx := commands.NewCompressionCapabilitiesContext(algs, commands.SMB2_COMPRESSION_CAPABILITIES_FLAG_NONE)
	if ctx.ContextType != commands.SMB2_COMPRESSION_CAPABILITIES {
		t.Errorf("ContextType = 0x%04x, want 0x%04x", ctx.ContextType, commands.SMB2_COMPRESSION_CAPABILITIES)
	}
	if len(ctx.Data) != 8+2*len(algs) {
		t.Fatalf("Data length = %d, want %d", len(ctx.Data), 8+2*len(algs))
	}

	// Simulate server response with a single algorithm selected.
	respCtx := &commands.NegotiateContext{
		ContextType: commands.SMB2_COMPRESSION_CAPABILITIES,
		Data:        ctx.Data[:10], // count=1, padding, flags, one algorithm
	}
	respCtx.Data[0] = 1 // count = 1
	respCtx.Data[1] = 0
	selected := commands.SelectedCompressionAlgorithms([]*commands.NegotiateContext{respCtx})
	if len(selected) != 1 {
		t.Fatalf("selected %d algorithms, want 1", len(selected))
	}
	if selected[0] != commands.SMB2_COMPRESSION_LZ77_HUFFMAN {
		t.Errorf("selected = 0x%04x, want 0x%04x", selected[0], commands.SMB2_COMPRESSION_LZ77_HUFFMAN)
	}
}

func TestSelectedCompressionAlgorithmsEmpty(t *testing.T) {
	algs := commands.SelectedCompressionAlgorithms(nil)
	if algs != nil {
		t.Errorf("expected nil, got %v", algs)
	}
}
