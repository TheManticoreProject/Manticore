package subcommands

import (
	"encoding/binary"
	"fmt"
)

// MS-SMB server-side data copy ([MS-SMB] section 2.2.7.2 and 3.2.4.3 / 3.3.5.11): the
// client opens the source and destination files, requests a copychunk resume key for the
// source via FSCTL_SRV_REQUEST_RESUME_KEY, then issues FSCTL_SRV_COPYCHUNK against the
// destination. Both FSCTLs are carried as the NT_TRANSACT_IOCTL FunctionCode of an
// SMB_COM_NT_TRANSACT request, with the structures below as the NT_Trans_Data payload.
const (
	// FSCTL_SRV_REQUEST_RESUME_KEY requests the 24-byte copychunk resume key that uniquely
	// identifies the open source file ([MS-SMB] section 2.2.7.2).
	FSCTL_SRV_REQUEST_RESUME_KEY uint32 = 0x00140078

	// FSCTL_SRV_COPYCHUNK performs a server-side copy of one or more chunks from the source
	// file (identified by the resume key) into the destination file ([MS-SMB] section 2.2.7.2).
	FSCTL_SRV_COPYCHUNK uint32 = 0x001440F2
)

// CopychunkResumeKeyLength is the fixed size, in bytes, of a copychunk resume key
// ([MS-SMB] section 2.2.7.2.2.2): an opaque server-generated value the client echoes
// back in the FSCTL_SRV_COPYCHUNK request.
const CopychunkResumeKeyLength = 24

const (
	srvCopychunkSize         = 24 // SourceOffset(8) + DestinationOffset(8) + Length(4) + Reserved(4)
	srvCopychunkResponseSize = 12 // ChunksWritten(4) + ChunkBytesWritten(4) + TotalBytesWritten(4)
)

// SrvCopychunk describes a single data range to copy server-side ([MS-SMB] section
// 2.2.7.2.1 SRV_COPYCHUNK). One or more are carried in the Chunks array of an
// SrvCopychunkCopy.
type SrvCopychunk struct {
	// SourceOffset (8 bytes): offset from the start of the source file to copy from.
	SourceOffset uint64
	// DestinationOffset (8 bytes): offset from the start of the destination file to copy to.
	DestinationOffset uint64
	// Length (4 bytes): number of bytes to copy.
	Length uint32
	// Reserved (4 bytes): MUST be zero and MUST be ignored on receipt.
	Reserved uint32
}

// Marshal serializes the SRV_COPYCHUNK into its 24-byte little-endian wire form.
func (c *SrvCopychunk) Marshal() ([]byte, error) {
	b := make([]byte, srvCopychunkSize)
	binary.LittleEndian.PutUint64(b[0:8], c.SourceOffset)
	binary.LittleEndian.PutUint64(b[8:16], c.DestinationOffset)
	binary.LittleEndian.PutUint32(b[16:20], c.Length)
	binary.LittleEndian.PutUint32(b[20:24], c.Reserved)
	return b, nil
}

// Unmarshal parses a 24-byte SRV_COPYCHUNK, returning the number of bytes consumed.
func (c *SrvCopychunk) Unmarshal(data []byte) (int, error) {
	if len(data) < srvCopychunkSize {
		return 0, fmt.Errorf("subcommands: SRV_COPYCHUNK requires %d bytes, got %d", srvCopychunkSize, len(data))
	}
	c.SourceOffset = binary.LittleEndian.Uint64(data[0:8])
	c.DestinationOffset = binary.LittleEndian.Uint64(data[8:16])
	c.Length = binary.LittleEndian.Uint32(data[16:20])
	c.Reserved = binary.LittleEndian.Uint32(data[20:24])
	return srvCopychunkSize, nil
}

// SrvCopychunkCopy is the NT_Trans_Data of an FSCTL_SRV_COPYCHUNK request ([MS-SMB]
// section 2.2.7.2.1 SRV_COPYCHUNK_COPY). The on-the-wire ChunkCount is derived from the
// length of Chunks so the count and the array cannot disagree.
type SrvCopychunkCopy struct {
	// CopychunkResumeKey (24 bytes): the resume key returned by FSCTL_SRV_REQUEST_RESUME_KEY
	// for the source file.
	CopychunkResumeKey [CopychunkResumeKeyLength]byte
	// Reserved (4 bytes): MUST be zero and MUST be ignored on receipt.
	Reserved uint32
	// Chunks: the data ranges to copy. ChunkCount on the wire equals len(Chunks).
	Chunks []SrvCopychunk
}

// Marshal serializes the SRV_COPYCHUNK_COPY: resume key, ChunkCount, Reserved, chunks.
func (c *SrvCopychunkCopy) Marshal() ([]byte, error) {
	b := make([]byte, 0, CopychunkResumeKeyLength+8+len(c.Chunks)*srvCopychunkSize)
	b = append(b, c.CopychunkResumeKey[:]...)

	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(len(c.Chunks))) // ChunkCount
	b = append(b, tmp...)
	binary.LittleEndian.PutUint32(tmp, c.Reserved)
	b = append(b, tmp...)

	for i := range c.Chunks {
		cb, err := c.Chunks[i].Marshal()
		if err != nil {
			return nil, err
		}
		b = append(b, cb...)
	}
	return b, nil
}

// Unmarshal parses an SRV_COPYCHUNK_COPY, returning the number of bytes consumed.
func (c *SrvCopychunkCopy) Unmarshal(data []byte) (int, error) {
	if len(data) < CopychunkResumeKeyLength+8 {
		return 0, fmt.Errorf("subcommands: SRV_COPYCHUNK_COPY header requires %d bytes, got %d", CopychunkResumeKeyLength+8, len(data))
	}
	offset := 0
	copy(c.CopychunkResumeKey[:], data[offset:offset+CopychunkResumeKeyLength])
	offset += CopychunkResumeKeyLength

	chunkCount := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	c.Reserved = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	c.Chunks = make([]SrvCopychunk, 0, chunkCount)
	for i := uint32(0); i < chunkCount; i++ {
		var chunk SrvCopychunk
		n, err := chunk.Unmarshal(data[offset:])
		if err != nil {
			return offset, fmt.Errorf("subcommands: SRV_COPYCHUNK_COPY chunk %d: %w", i, err)
		}
		offset += n
		c.Chunks = append(c.Chunks, chunk)
	}
	return offset, nil
}

// SrvCopychunkResponse is the NT_Trans_Data of an FSCTL_SRV_COPYCHUNK response ([MS-SMB]
// section 2.2.7.2.2.1 SRV_COPYCHUNK_RESPONSE).
type SrvCopychunkResponse struct {
	// ChunksWritten (4 bytes): number of chunks successfully written.
	ChunksWritten uint32
	// ChunkBytesWritten (4 bytes): number of bytes written in the last (partially written) chunk.
	ChunkBytesWritten uint32
	// TotalBytesWritten (4 bytes): total number of bytes written.
	TotalBytesWritten uint32
}

// Marshal serializes the SRV_COPYCHUNK_RESPONSE into its 12-byte little-endian wire form.
func (r *SrvCopychunkResponse) Marshal() ([]byte, error) {
	b := make([]byte, srvCopychunkResponseSize)
	binary.LittleEndian.PutUint32(b[0:4], r.ChunksWritten)
	binary.LittleEndian.PutUint32(b[4:8], r.ChunkBytesWritten)
	binary.LittleEndian.PutUint32(b[8:12], r.TotalBytesWritten)
	return b, nil
}

// Unmarshal parses a 12-byte SRV_COPYCHUNK_RESPONSE, returning the bytes consumed.
func (r *SrvCopychunkResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < srvCopychunkResponseSize {
		return 0, fmt.Errorf("subcommands: SRV_COPYCHUNK_RESPONSE requires %d bytes, got %d", srvCopychunkResponseSize, len(data))
	}
	r.ChunksWritten = binary.LittleEndian.Uint32(data[0:4])
	r.ChunkBytesWritten = binary.LittleEndian.Uint32(data[4:8])
	r.TotalBytesWritten = binary.LittleEndian.Uint32(data[8:12])
	return srvCopychunkResponseSize, nil
}

// SrvRequestResumeKeyResponse is the NT_Trans_Data of an FSCTL_SRV_REQUEST_RESUME_KEY
// response ([MS-SMB] section 2.2.7.2.2.2). The ContextLength field on the wire is derived
// from len(Context); per the spec the extended context feature is reserved and not used,
// so Context is normally empty.
type SrvRequestResumeKeyResponse struct {
	// CopychunkResumeKey (24 bytes): opaque resume key identifying the open source file.
	CopychunkResumeKey [CopychunkResumeKeyLength]byte
	// Context (variable): reserved/unused extended context; normally zero-length.
	Context []byte
}

// Marshal serializes the FSCTL_SRV_REQUEST_RESUME_KEY response: resume key, ContextLength,
// then the (normally empty) Context.
func (r *SrvRequestResumeKeyResponse) Marshal() ([]byte, error) {
	b := make([]byte, 0, CopychunkResumeKeyLength+4+len(r.Context))
	b = append(b, r.CopychunkResumeKey[:]...)
	tmp := make([]byte, 4)
	binary.LittleEndian.PutUint32(tmp, uint32(len(r.Context))) // ContextLength
	b = append(b, tmp...)
	b = append(b, r.Context...)
	return b, nil
}

// Unmarshal parses an FSCTL_SRV_REQUEST_RESUME_KEY response, returning the bytes consumed.
func (r *SrvRequestResumeKeyResponse) Unmarshal(data []byte) (int, error) {
	if len(data) < CopychunkResumeKeyLength+4 {
		return 0, fmt.Errorf("subcommands: FSCTL_SRV_REQUEST_RESUME_KEY response requires at least %d bytes, got %d", CopychunkResumeKeyLength+4, len(data))
	}
	offset := 0
	copy(r.CopychunkResumeKey[:], data[offset:offset+CopychunkResumeKeyLength])
	offset += CopychunkResumeKeyLength

	contextLength := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	if len(data) < offset+int(contextLength) {
		return offset, fmt.Errorf("subcommands: FSCTL_SRV_REQUEST_RESUME_KEY response Context truncated: need %d, have %d", contextLength, len(data)-offset)
	}
	r.Context = append([]byte{}, data[offset:offset+int(contextLength)]...)
	offset += int(contextLength)
	return offset, nil
}
