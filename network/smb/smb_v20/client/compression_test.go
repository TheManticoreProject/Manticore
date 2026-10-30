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
	if len(compressed) != 8 {
		t.Fatalf("compressed length = %d, want 8", len(compressed))
	}
	if compressed[0] != 0xAA {
		t.Errorf("pattern byte = 0x%02x, want 0xAA", compressed[0])
	}

	decompressed, err := DecompressPatternV1(compressed, uint32(len(data)))
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
	_, err := DecompressPatternV1([]byte{0xAA, 0x00, 0x00, 0x00}, 16)
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestDecompressPatternV1ReservedByteNonZero(t *testing.T) {
	// Pattern_V1: Pattern(1) + Reserved1(1) + Reserved2(2) + Repetitions(4).
	_, err := DecompressPatternV1([]byte{0xAA, 0xFF, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00}, 16)
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

func TestCompressionChainedPayloadHeaderRoundTrip(t *testing.T) {
	h := &CompressionChainedPayloadHeader{
		CompressionAlgorithm: commands.SMB2_COMPRESSION_PATTERN_V1,
		Flags:                SMB2_COMPRESSION_FLAG_CHAINED,
		Length:               8,
	}
	wire := MarshalCompressionChainedPayloadHeader(h)
	want := []byte{0x04, 0x00, 0x01, 0x00, 0x08, 0x00, 0x00, 0x00}
	if !bytes.Equal(wire, want) {
		t.Fatalf("wire = % x, want % x", wire, want)
	}

	parsed, err := ParseCompressionChainedPayloadHeader(wire)
	if err != nil {
		t.Fatalf("ParseCompressionChainedPayloadHeader: %v", err)
	}
	if *parsed != *h {
		t.Errorf("parsed = %+v, want %+v", *parsed, *h)
	}
}

func TestCompressionChainedPayloadHeaderOriginalPayloadSize(t *testing.T) {
	// MS-SMB2 2.2.42.2.1: OriginalPayloadSize is present for LZNT1, LZ77 and
	// LZ77+Huffman, and Length includes it.
	h := &CompressionChainedPayloadHeader{
		CompressionAlgorithm: commands.SMB2_COMPRESSION_LZ77,
		Flags:                SMB2_COMPRESSION_FLAG_NONE,
		Length:               9,
		OriginalPayloadSize:  256,
	}
	wire := MarshalCompressionChainedPayloadHeader(h)
	want := []byte{0x02, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00}
	if !bytes.Equal(wire, want) {
		t.Fatalf("wire = % x, want % x", wire, want)
	}
	parsed, err := ParseCompressionChainedPayloadHeader(wire)
	if err != nil {
		t.Fatalf("ParseCompressionChainedPayloadHeader: %v", err)
	}
	if *parsed != *h {
		t.Errorf("parsed = %+v, want %+v", *parsed, *h)
	}
	if _, err := ParseCompressionChainedPayloadHeader(want[:8]); err == nil {
		t.Error("expected error for LZ77 header missing OriginalPayloadSize")
	}
}

func TestParseCompressionChainedPayloadHeaderTooShort(t *testing.T) {
	_, err := ParseCompressionChainedPayloadHeader(make([]byte, 4))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestWalkCompressionChainSingleNone(t *testing.T) {
	payload := []byte("hello world")
	hdr := MarshalCompressionChainedPayloadHeader(&CompressionChainedPayloadHeader{
		CompressionAlgorithm: commands.SMB2_COMPRESSION_NONE,
		Flags:                SMB2_COMPRESSION_FLAG_CHAINED,
		Length:               uint32(len(payload)),
	})
	data := append(hdr, payload...)

	segs, err := WalkCompressionChain(data)
	if err != nil {
		t.Fatalf("WalkCompressionChain: %v", err)
	}
	if len(segs) != 1 {
		t.Fatalf("got %d segments, want 1", len(segs))
	}
	if !bytes.Equal(segs[0].CompressedData, payload) {
		t.Errorf("data = %q, want %q", segs[0].CompressedData, payload)
	}
}

// TestWalkCompressionChainSpecLayout walks a chained payload built byte-for-byte
// from MS-SMB2 2.2.42.2.1 / 2.2.42.2.2: a Pattern_V1 segment, a NONE segment and
// an LZ77 segment carrying OriginalPayloadSize.
func TestWalkCompressionChainSpecLayout(t *testing.T) {
	data := []byte{
		// CompressionAlgorithm=Pattern_V1, Flags=CHAINED, Length=8
		0x04, 0x00, 0x01, 0x00, 0x08, 0x00, 0x00, 0x00,
		// Pattern=0xAA, Reserved1, Reserved2, Repetitions=128
		0xAA, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x00,
		// CompressionAlgorithm=NONE, Flags=0, Length=5
		0x00, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00,
		'h', 'e', 'l', 'l', 'o',
		// CompressionAlgorithm=LZ77, Flags=0, Length=9, OriginalPayloadSize=256
		0x02, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x01, 0x02, 0x03, 0x04, 0x05,
	}

	segs, err := WalkCompressionChain(data)
	if err != nil {
		t.Fatalf("WalkCompressionChain: %v", err)
	}
	if len(segs) != 3 {
		t.Fatalf("got %d segments, want 3", len(segs))
	}

	if segs[0].Header.CompressionAlgorithm != commands.SMB2_COMPRESSION_PATTERN_V1 || segs[0].Header.Flags != SMB2_COMPRESSION_FLAG_CHAINED {
		t.Errorf("seg[0] header = %+v, want Pattern_V1 / CHAINED", segs[0].Header)
	}
	pat, err := DecompressPatternV1(segs[0].CompressedData, 1024)
	if err != nil {
		t.Fatalf("DecompressPatternV1: %v", err)
	}
	if !bytes.Equal(pat, bytes.Repeat([]byte{0xAA}, 128)) {
		t.Errorf("seg[0] decompressed %d bytes, want 128 x 0xAA", len(pat))
	}

	if segs[1].Header.CompressionAlgorithm != commands.SMB2_COMPRESSION_NONE || !bytes.Equal(segs[1].CompressedData, []byte("hello")) {
		t.Errorf("seg[1] = %+v, want NONE \"hello\"", segs[1])
	}

	if segs[2].Header.CompressionAlgorithm != commands.SMB2_COMPRESSION_LZ77 || segs[2].Header.OriginalPayloadSize != 256 {
		t.Errorf("seg[2] header = %+v, want LZ77 OriginalPayloadSize=256", segs[2].Header)
	}
	if !bytes.Equal(segs[2].CompressedData, []byte{0x01, 0x02, 0x03, 0x04, 0x05}) {
		t.Errorf("seg[2] data = % x, want 01 02 03 04 05", segs[2].CompressedData)
	}
}

func TestDecompressPatternV1RepetitionsExceedMax(t *testing.T) {
	_, err := DecompressPatternV1([]byte{0xAA, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF}, 4096)
	if err == nil {
		t.Error("expected error for Repetitions greater than the original segment size")
	}
}

func TestWalkCompressionChainTruncated(t *testing.T) {
	hdr := MarshalCompressionChainedPayloadHeader(&CompressionChainedPayloadHeader{
		CompressionAlgorithm: commands.SMB2_COMPRESSION_NONE,
		Length:               1000,
	})
	data := append(hdr, []byte{0x01, 0x02}...)
	_, err := WalkCompressionChain(data)
	if err == nil {
		t.Error("expected error for truncated segment")
	}
}

func TestWalkCompressionChainOversizedNoneSegment(t *testing.T) {
	// Segment Length values that overflow or wrap an int on 32-bit builds
	// must be rejected, not panic or be taken as "rest of buffer".
	for _, size := range []uint32{0x7FFFFFF8, 0x7FFFFFFF, 0x80000000, 0xFFFFFFFF} {
		hdr := MarshalCompressionChainedPayloadHeader(&CompressionChainedPayloadHeader{
			CompressionAlgorithm: commands.SMB2_COMPRESSION_NONE,
			Length:               size,
		})
		data := append(hdr, make([]byte, 24)...)
		segs, err := WalkCompressionChain(data)
		if err == nil {
			t.Errorf("size=0x%08X: expected error, got %d segments", size, len(segs))
		}
	}
}

func TestWalkCompressionChainEmpty(t *testing.T) {
	segs, err := WalkCompressionChain(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segs) != 0 {
		t.Errorf("got %d segments for empty input, want 0", len(segs))
	}
}
