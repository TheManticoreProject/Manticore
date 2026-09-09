package server

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/capabilities"
	smb1client "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// countingBytes is a block whose every byte identifies its own position, so a
// transfer that is right in length but wrong in content still fails.
func countingBytes(length int) []byte {
	block := make([]byte, length)
	for index := range block {
		block[index] = byte(index % 251)
	}
	return block
}

// largeTransferShare serves one file of the given size, filled so that every byte
// identifies its own position.
func largeTransferShare(t *testing.T, size int) *MemoryFileSystem {
	t.Helper()

	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("large.bin", countingBytes(size)); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	return fs
}

// sendReadAndx sends a hand-built SMB_COM_READ_ANDX and returns the raw reply.
//
// The client API caps its own reads well below the large-read range, so a request
// that exercises MaxCountHigh has to be built here.
func sendReadAndx(
	t *testing.T,
	client *smb1client.Client,
	fid smb1client.FID,
	maxCountLow uint16,
	timeoutOrMaxCountHigh uint32,
) []byte {
	t.Helper()

	request := newRequest(codes.SMB_COM_READ_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	cmd := commands.NewReadAndxRequest()
	cmd.FID = types.USHORT(fid)
	cmd.Offset = types.ULONG(0)
	cmd.MaxCountOfBytesToReturn = types.USHORT(maxCountLow)
	cmd.MinCountOfBytesToReturn = types.USHORT(0)
	cmd.Timeout = types.ULONG(timeoutOrMaxCountHigh)
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the read: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	return raw
}

// readReplyLength reads the length a read reply describes, across both of the
// words that carry it.
func readReplyLength(t *testing.T, raw []byte) int {
	t.Helper()

	decoded := readReplyOf(t, raw)
	return int(decoded.DataLengthHigh)<<16 | int(decoded.DataLength)
}

// readReplyOf decodes a read reply.
func readReplyOf(t *testing.T, raw []byte) *commands.ReadAndxResponse {
	t.Helper()

	if status := binary.LittleEndian.Uint32(raw[5:9]); status != 0 {
		t.Fatalf("the read reported 0x%08X, want success", status)
	}
	response := commands.NewReadAndxResponse()
	if _, err := response.Unmarshal(raw[header.SMB_HEADER_SIZE:]); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}
	return response
}

// openLargeFile connects a share over the given config and opens the test file.
func openLargeFile(t *testing.T, config Config, size int) (*smb1client.Client, smb1client.FID) {
	t.Helper()

	srv, transportEnd := pipedServer(t, config)
	if err := srv.AddShare(&Share{
		Name: fileShareName, Type: ShareTypeDisk, FS: largeTransferShare(t, size),
	}); err != nil {
		t.Fatalf("AddShare() error = %v", err)
	}

	client := smb1client.NewFromTransport(transportEnd, net.IPv4(127, 0, 0, 1), 445)
	if err := client.Negotiate(); err != nil {
		t.Fatalf("Negotiate() error = %v", err)
	}
	creds, err := credentials.NewCredentials(captureDomain, captureUsername, capturePassword, "")
	if err != nil {
		t.Fatalf("NewCredentials() error = %v", err)
	}
	if err := client.SessionSetup(creds); err != nil {
		t.Fatalf("SessionSetup() error = %v", err)
	}
	if err := client.TreeConnect(fileShareName); err != nil {
		t.Fatalf("TreeConnect() error = %v", err)
	}

	fid, err := client.OpenFile("large.bin",
		fileflags.GENERIC_READ|fileflags.GENERIC_WRITE,
		fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN,
		fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("opening the file failed: %v", err)
	}
	return client, fid
}

// TestLargeReadExceedsTheNegotiatedBuffer asserts a read under CAP_LARGE_READX may
// return more than MaxBufferSize, which is the whole point of the capability.
func TestLargeReadExceedsTheNegotiatedBuffer(t *testing.T) {
	config := conformanceConfig(SigningDisabled)
	client, fid := openLargeFile(t, config, 200*1024)

	// 0x18000 = 98304 bytes, described across both words, and past the ceiling so
	// the answer is the ceiling rather than the request.
	raw := sendReadAndx(t, client, fid, 0x8000, 0x0001)

	length := readReplyLength(t, raw)
	if length <= int(config.MaxBufferSize) {
		t.Fatalf("the read returned %d bytes, which does not exceed the negotiated buffer of %d",
			length, config.MaxBufferSize)
	}
	if length != DefaultMaxLargeTransfer {
		t.Fatalf("the read returned %d bytes, want the %d-byte ceiling", length, DefaultMaxLargeTransfer)
	}

	response := readReplyOf(t, raw)

	// And the bytes are the file's, at the offset asked for.
	at := int(response.DataOffset)
	if at+length > len(raw) {
		t.Fatalf("DataOffset %d plus %d bytes runs past the %d-byte reply", at, length, len(raw))
	}
	if !bytes.Equal(raw[at:at+length], countingBytes(200 * 1024)[:length]) {
		t.Error("the bytes returned are not the file's")
	}
}

// TestLargeReadIsCappedWhenSigningIsActive asserts signing overrides the
// capability.
//
// [MS-SMB] section 2.2.4.5.2.1: "When signing is active on a connection, then
// clients MUST limit read lengths to the MaxBufferSize value negotiated by the
// server irrespective of the value of the CAP_LARGE_READX flag." A signed response
// is verified over the bytes as sent, so a client that sized its buffer to
// MaxBufferSize and received more would fail the signature rather than truncate.
//
// This is asserted against readLimit rather than over the wire because signing a
// hand-built request needs the connection's MAC key, which the client keeps to
// itself — and the rule is about the ceiling, which is what readLimit is.
func TestLargeReadIsCappedWhenSigningIsActive(t *testing.T) {
	conn := &Connection{
		Server: &Server{config: Config{
			MaxBufferSize:    DefaultMaxBufferSize,
			MaxLargeTransfer: DefaultMaxLargeTransfer,
		}},
		ClientCapabilities: capabilities.CAP_LARGE_READX,
	}

	unsigned := conn.readLimit()
	if unsigned != DefaultMaxLargeTransfer {
		t.Fatalf("an unsigned connection allows %d bytes, want the %d-byte ceiling",
			unsigned, DefaultMaxLargeTransfer)
	}

	conn.SigningActive = true
	signed := conn.readLimit()
	if signed > int(DefaultMaxBufferSize) {
		t.Errorf("a signing connection allows %d bytes, more than the negotiated buffer of %d",
			signed, DefaultMaxBufferSize)
	}
	if signed >= unsigned {
		t.Errorf("signing did not reduce the ceiling: %d then %d", unsigned, signed)
	}
}

// TestLargeReadIgnoresTheReservedHalf asserts the second half of the overloaded
// field is ignored rather than treated as a refusal.
//
// [MS-SMB] section 2.2.4.2.1 tells the client to set Reserved to 0xFFFF when
// MaxCountHigh is 0xFFFF, and says "For all values, this field MUST be ignored by
// the server". A server that rejected a non-zero Reserved would refuse the largest
// read a client is told how to ask for.
func TestLargeReadIgnoresTheReservedHalf(t *testing.T) {
	config := conformanceConfig(SigningDisabled)
	client, fid := openLargeFile(t, config, 200*1024)

	// MaxCountHigh 0x0001 with the Reserved half set to 0xFFFF.
	raw := sendReadAndx(t, client, fid, 0x8000, 0xFFFF0001)

	if length := readReplyLength(t, raw); length != DefaultMaxLargeTransfer {
		t.Fatalf("the read returned %d bytes, want %d — the Reserved half was not ignored",
			length, DefaultMaxLargeTransfer)
	}
}

// TestLargeTransfersRequireBothSidesToAgree asserts the high word is ignored for a
// client that did not set the capability in its session setup.
//
// [MS-SMB] section 2.2.4.5.2.1 makes the capability two-sided. A client that never
// agreed cannot receive a response above MaxBufferSize, so acting on the server's
// own advertisement alone would send it one it cannot read.
func TestLargeTransfersRequireBothSidesToAgree(t *testing.T) {
	conn := &Connection{
		Server:             &Server{config: Config{MaxBufferSize: DefaultMaxBufferSize}},
		ClientCapabilities: capabilities.CAP_UNICODE,
	}
	open := &Open{}

	request := commands.NewReadAndxRequest()
	request.MaxCountOfBytesToReturn = types.USHORT(0x8000)
	request.Timeout = types.ULONG(0x0001)

	if got := readLengthOf(conn, open, request); got != 0x8000 {
		t.Errorf("a client that did not agree asked for %d, want the low word alone (%d)", got, 0x8000)
	}

	write := commands.NewWriteAndxRequest()
	write.DataLength = types.USHORT(0x0100)
	write.Reserved = types.USHORT(0x0001)
	if got := writeLengthOf(conn, write); got != 0x0100 {
		t.Errorf("a client that did not agree wrote %d, want the low word alone (%d)", got, 0x0100)
	}

	// With the capability agreed, both words count.
	conn.ClientCapabilities |= capabilities.CAP_LARGE_READX | capabilities.CAP_LARGE_WRITEX
	if got := readLengthOf(conn, open, request); got != 0x18000 {
		t.Errorf("an agreed client asked for %d, want 0x18000", got)
	}
	if got := writeLengthOf(conn, write); got != 0x10100 {
		t.Errorf("an agreed client wrote %d, want 0x10100", got)
	}
}

// TestLargeReadOnAPipeReadsTheFieldAsATimeout asserts the overloaded field is not
// treated as a length for a pipe handle.
//
// [MS-SMB] section 2.2.4.2.1: the field is MaxCountHigh "When reading from a
// regular file" and Timeout "When reading from a name pipe or I/O device".
// Treating a pipe's timeout as a length would have the server try to return
// gigabytes because the client asked it to wait a while.
func TestLargeReadOnAPipeReadsTheFieldAsATimeout(t *testing.T) {
	conn := &Connection{
		Server:             &Server{config: Config{MaxBufferSize: DefaultMaxBufferSize}},
		ClientCapabilities: capabilities.CAP_LARGE_READX,
	}

	request := commands.NewReadAndxRequest()
	request.MaxCountOfBytesToReturn = types.USHORT(0x0100)
	// A five-second wait, which as a length would be 327680 bytes.
	request.Timeout = types.ULONG(5000)

	if got := readLengthOf(conn, &Open{IsPipe: true}, request); got != 0x0100 {
		t.Errorf("a pipe read asked for %d, want the low word alone (%d)", got, 0x0100)
	}
	if got := readLengthOf(conn, &Open{}, request); got == 0x0100 {
		t.Error("a regular-file read ignored MaxCountHigh")
	}
}

// TestLargeWriteExceedsTheNegotiatedBuffer asserts one write may carry more than
// MaxBufferSize, which is what CAP_LARGE_WRITEX is for.
//
// The size is above the negotiated 16644 bytes and below 0x10000, so the whole
// write is described by DataLength alone. The range at and above that is covered
// by TestWriteAtSixtyFourKiBLandsInFull.
func TestLargeWriteExceedsTheNegotiatedBuffer(t *testing.T) {
	const size = 40000

	config := conformanceConfig(SigningDisabled)
	client, fid := openLargeFile(t, config, size)

	payload := countingBytes(size)
	for index := range payload {
		// Distinguish what is written from what the file already held.
		payload[index] ^= 0xFF
	}

	request := newRequest(codes.SMB_COM_WRITE_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	cmd := commands.NewWriteAndxRequest()
	cmd.FID = types.USHORT(fid)
	cmd.Offset = types.ULONG(0)
	cmd.Pad = types.UCHAR(0)
	// DataLength, DataLengthHigh and DataOffset are derived from the data when the
	// request is marshalled, so the test does not restate this command's layout.
	cmd.Data = []types.UCHAR(payload)
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the write: %v", err)
	}
	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != 0 {
		t.Fatalf("the write reported 0x%08X, want success", status)
	}

	response := commands.NewWriteAndxResponse()
	if _, err := response.Unmarshal(raw[header.SMB_HEADER_SIZE:]); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}
	if written := int(response.CountHigh)<<16 | int(response.Count); written != size {
		t.Fatalf("the reply reports %d bytes written, want %d", written, size)
	}
	if size <= int(config.MaxBufferSize) {
		t.Fatalf("the test writes %d bytes, which does not exceed the negotiated buffer of %d",
			size, config.MaxBufferSize)
	}

	// And the bytes actually landed, all of them.
	readBack, err := client.ReadFile(fid, 0, size)
	if err != nil {
		t.Fatalf("reading the file back failed: %v", err)
	}
	if !bytes.Equal(readBack, payload) {
		t.Fatalf("the file holds %d bytes that do not match the %d written", len(readBack), len(payload))
	}
}

// TestLargeTransferCeilingFitsOneTransportFrame asserts the ceiling stays inside
// what a transport will carry.
//
// This is the constraint that actually bounds the ceiling. SMB_Data.ByteCount no
// longer does: a transfer of 0x10000 bytes or more carries its length in
// DataLength and DataLengthHigh and its position in DataOffset. But a message
// larger than one frame cannot be sent at all, and NetBIOS over TCP is the
// tighter of the two transports — RFC 1002 4.3.1 gives its session message a
// 17-bit length, so 0x1FFFF bytes including the SMB header and framing.
func TestLargeTransferCeilingFitsOneTransportFrame(t *testing.T) {
	// A read response is the larger of the two framings, at the SMB header plus
	// the parameter words and the byte count.
	const framing = header.SMB_HEADER_SIZE + 1 + 2*12 + 2

	if DefaultMaxLargeTransfer+framing > 0x1FFFF {
		t.Fatalf("DefaultMaxLargeTransfer is %d, which with %d bytes of framing exceeds the %d a NetBIOS session message carries",
			DefaultMaxLargeTransfer, framing, 0x1FFFF)
	}

	// And it is at least the 64 KiB the capability exists to reach, since a
	// ceiling below that leaves the extension buying nothing a plain
	// MaxBufferSize increase would not.
	if DefaultMaxLargeTransfer < 0x10000 {
		t.Errorf("DefaultMaxLargeTransfer is %d, below the 64 KiB CAP_LARGE_READX and CAP_LARGE_WRITEX are for",
			DefaultMaxLargeTransfer)
	}
}

// TestReadAtSixtyFourKiBIsFramedAcrossBothLengthWords asserts a read of exactly
// 0x10000 bytes is described and delivered.
//
// This is the size that cannot be stated in one USHORT: DataLength wraps to zero
// and the whole length lives in DataLengthHigh, so a client reading DataLength
// alone sees an empty reply. It is also where ByteCount stops being usable, which
// is why the data is located by DataOffset.
func TestReadAtSixtyFourKiBIsFramedAcrossBothLengthWords(t *testing.T) {
	const size = 0x10000

	config := conformanceConfig(SigningDisabled)
	client, fid := openLargeFile(t, config, 200*1024)

	// MaxCountHigh 0x0001 with a zero low word asks for exactly 0x10000 bytes.
	raw := sendReadAndx(t, client, fid, 0x0000, 0x0001)
	response := readReplyOf(t, raw)

	if response.DataLength != 0 {
		t.Errorf("DataLength is 0x%04X, want 0 — a 64 KiB length has nothing in its low word",
			response.DataLength)
	}
	if response.DataLengthHigh != 1 {
		t.Errorf("DataLengthHigh is 0x%04X, want 1", response.DataLengthHigh)
	}
	if length := readReplyLength(t, raw); length != size {
		t.Fatalf("the read described %d bytes, want %d", length, size)
	}

	// The bytes are where DataOffset says, and they are the file's.
	at := int(response.DataOffset)
	if at+size > len(raw) {
		t.Fatalf("DataOffset %d plus %d bytes runs past the %d-byte reply", at, size, len(raw))
	}
	if !bytes.Equal(raw[at:at+size], countingBytes(200 * 1024)[:size]) {
		t.Error("the bytes returned are not the file's")
	}

	// ByteCount holds the low word of the block, which at this size is zero. The
	// reply is still readable because nothing needs it: asserting the value keeps
	// it from looking like a defect to the next reader.
	byteCount := binary.LittleEndian.Uint16(raw[at-2 : at])
	if byteCount != uint16(size&0xFFFF) {
		t.Errorf("ByteCount is 0x%04X, want the low word of %d (0x%04X)",
			byteCount, size, uint16(size&0xFFFF))
	}

	// And the decoder recovers it, which is what a client does with the reply.
	if len(response.Data) != size {
		t.Errorf("the decoded reply carries %d bytes, want %d", len(response.Data), size)
	}
}

// TestWriteAtSixtyFourKiBLandsInFull asserts a write of exactly 0x10000 bytes is
// accepted and every byte reaches the file.
//
// Before the data was located by DataOffset this was the request that could not
// be represented: ByteCount wrapped, the data block was truncated to the wrapped
// value before the command saw it, and the server answered STATUS_INVALID_SMB.
func TestWriteAtSixtyFourKiBLandsInFull(t *testing.T) {
	const size = 0x10000

	config := conformanceConfig(SigningDisabled)
	client, fid := openLargeFile(t, config, size)

	payload := countingBytes(size)
	for index := range payload {
		// Distinguish what is written from what the file already held.
		payload[index] ^= 0xFF
	}

	request := newRequest(codes.SMB_COM_WRITE_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	cmd := commands.NewWriteAndxRequest()
	cmd.FID = types.USHORT(fid)
	cmd.Offset = types.ULONG(0)
	cmd.Data = []types.UCHAR(payload)
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the write: %v", err)
	}
	// The length has to be described across both words for the request to say
	// what it carries at all.
	if cmd.DataLength != 0 || cmd.DataLengthHigh() != 1 {
		t.Fatalf("the request describes its %d bytes as DataLength 0x%04X and DataLengthHigh 0x%04X, want 0x0000 and 0x0001",
			size, cmd.DataLength, cmd.DataLengthHigh())
	}

	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if status := binary.LittleEndian.Uint32(raw[5:9]); status != 0 {
		t.Fatalf("the write reported 0x%08X, want success", status)
	}

	response := commands.NewWriteAndxResponse()
	if _, err := response.Unmarshal(raw[header.SMB_HEADER_SIZE:]); err != nil {
		t.Fatalf("the reply did not decode: %v", err)
	}
	if written := int(response.CountHigh)<<16 | int(response.Count); written != size {
		t.Fatalf("the reply reports %d bytes written, want %d", written, size)
	}

	// And every byte is in the file, read back in chunks the client can describe.
	readBack, err := client.ReadFile(fid, 0, size)
	if err != nil {
		t.Fatalf("reading the file back failed: %v", err)
	}
	if !bytes.Equal(readBack, payload) {
		t.Fatalf("the file holds %d bytes that do not match the %d written", len(readBack), len(payload))
	}
}

// TestWriteWithADataOffsetInsideItsParametersIsRefused asserts a request whose
// DataOffset does not describe where its data is gets STATUS_INVALID_SMB.
//
// [MS-CIFS] section 3.3.5.37: a DataOffset below the start of
// SMB_Data.Bytes.Data, or past the end of the data it claims, fails the request.
// Locating data by an offset the client supplies is only safe if the offset is
// checked.
func TestWriteWithADataOffsetInsideItsParametersIsRefused(t *testing.T) {
	config := conformanceConfig(SigningDisabled)
	client, fid := openLargeFile(t, config, 4096)

	request := newRequest(codes.SMB_COM_WRITE_ANDX)
	request.Header.UID = client.Session.SessionUID
	request.Header.TID = client.Session.TreeID

	cmd := commands.NewWriteAndxRequest()
	cmd.FID = types.USHORT(fid)
	cmd.Data = []types.UCHAR(countingBytes(512))
	request.AddCommand(cmd)

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal the write: %v", err)
	}

	// Point DataOffset at the FID, inside the parameter words. The offset of the
	// field itself: the header, WordCount(1), the AndX block(4), FID(2),
	// Offset(4), Timeout(4), WriteMode(2), Remaining(2), DataLengthHigh(2) and
	// DataLength(2).
	const dataOffsetField = header.SMB_HEADER_SIZE + 1 + 4 + 2 + 4 + 4 + 2 + 2 + 2 + 2
	binary.LittleEndian.PutUint16(marshalled[dataOffsetField:dataOffsetField+2],
		uint16(header.SMB_HEADER_SIZE+1+4))

	if _, err := client.Transport.Send(marshalled); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	raw, err := client.Transport.Receive()
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}

	status := binary.LittleEndian.Uint32(raw[5:9])
	if status != uint32(nt_status.NT_STATUS_INVALID_SMB) {
		t.Errorf("the write reported 0x%08X, want STATUS_INVALID_SMB (0x%08X)",
			status, uint32(nt_status.NT_STATUS_INVALID_SMB))
	}
}
