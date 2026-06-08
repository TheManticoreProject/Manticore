package subcommands

import (
	"encoding/binary"
	"fmt"
)

// MS-SMB per-user disk quota extension, carried as the NT_TRANSACT_QUERY_QUOTA and
// NT_TRANSACT_SET_QUOTA subcommands of SMB_COM_NT_TRANSACT ([MS-SMB] sections 2.2.7.5 and
// 2.2.7.6). The NT_Trans_Parameters of a query request are defined here, along with the
// MS-FSCC quota records used as the NT_Trans_Data: FILE_GET_QUOTA_INFORMATION ([MS-FSCC]
// section 2.4.41.1) names the SIDs queried, and FILE_QUOTA_INFORMATION ([MS-FSCC] section
// 2.4.41) carries a single user's quota — it is both the query response record and the
// record a client sends in an NT_TRANSACT_SET_QUOTA request.

const (
	ntTransQueryQuotaParametersSize  = 16 // FID(2)+ReturnSingleEntry(1)+RestartScan(1)+SidListLength(4)+StartSidLength(4)+StartSidOffset(4)
	fileQuotaInformationFixedSize    = 40 // NextEntryOffset(4)+SidLength(4)+ChangeTime(8)+QuotaUsed(8)+QuotaThreshold(8)+QuotaLimit(8)
	fileGetQuotaInformationFixedSize = 8  // NextEntryOffset(4)+SidLength(4)
)

// NtTransQueryQuotaRequestParameters is the NT_Trans_Parameters of an
// NT_TRANSACT_QUERY_QUOTA request ([MS-SMB] section 2.2.7.5.1). At least one of
// SidListLength or StartSidLength MUST be zero; if both are zero, all SIDs are enumerated.
type NtTransQueryQuotaRequestParameters struct {
	// FID (2 bytes): the open file/directory whose object store's quota is queried.
	FID uint16
	// ReturnSingleEntry (1 byte): if non-zero, return only a single SID's quota.
	ReturnSingleEntry bool
	// RestartScan (1 byte): if non-zero, restart the quota scan.
	RestartScan bool
	// SidListLength (4 bytes): length of the NT_Trans_Data SidList, or zero.
	SidListLength uint32
	// StartSidLength (4 bytes): length of the single start-SID entry, or zero.
	StartSidLength uint32
	// StartSidOffset (4 bytes): offset, from the start of NT_Trans_Data, of the start SID.
	StartSidOffset uint32
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// Marshal serializes the 16-octet NT_Trans_Parameters block.
func (p *NtTransQueryQuotaRequestParameters) Marshal() ([]byte, error) {
	b := make([]byte, ntTransQueryQuotaParametersSize)
	binary.LittleEndian.PutUint16(b[0:2], p.FID)
	b[2] = boolToByte(p.ReturnSingleEntry)
	b[3] = boolToByte(p.RestartScan)
	binary.LittleEndian.PutUint32(b[4:8], p.SidListLength)
	binary.LittleEndian.PutUint32(b[8:12], p.StartSidLength)
	binary.LittleEndian.PutUint32(b[12:16], p.StartSidOffset)
	return b, nil
}

// Unmarshal parses the 16-octet NT_Trans_Parameters block.
func (p *NtTransQueryQuotaRequestParameters) Unmarshal(data []byte) (int, error) {
	if len(data) < ntTransQueryQuotaParametersSize {
		return 0, fmt.Errorf("subcommands: NT_TRANSACT_QUERY_QUOTA parameters require %d bytes, got %d", ntTransQueryQuotaParametersSize, len(data))
	}
	p.FID = binary.LittleEndian.Uint16(data[0:2])
	p.ReturnSingleEntry = data[2] != 0
	p.RestartScan = data[3] != 0
	p.SidListLength = binary.LittleEndian.Uint32(data[4:8])
	p.StartSidLength = binary.LittleEndian.Uint32(data[8:12])
	p.StartSidOffset = binary.LittleEndian.Uint32(data[12:16])
	return ntTransQueryQuotaParametersSize, nil
}

// NtTransQuotaResponseParameters is the NT_Trans_Parameters of an NT_TRANSACT_QUERY_QUOTA
// response ([MS-SMB] section 2.2.7.5.2): the byte length of the returned quota data.
type NtTransQuotaResponseParameters struct {
	// DataLength (4 bytes): length of the returned quota information (equals TotalDataCount).
	DataLength uint32
}

// Marshal serializes the 4-octet response parameter block.
func (p *NtTransQuotaResponseParameters) Marshal() ([]byte, error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, p.DataLength)
	return b, nil
}

// Unmarshal parses the 4-octet response parameter block.
func (p *NtTransQuotaResponseParameters) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("subcommands: NT_TRANSACT_QUERY_QUOTA response parameters require 4 bytes, got %d", len(data))
	}
	p.DataLength = binary.LittleEndian.Uint32(data[0:4])
	return 4, nil
}

// FileGetQuotaInformation is a single SID entry in the SidList of an
// NT_TRANSACT_QUERY_QUOTA request ([MS-FSCC] section 2.4.41.1 FILE_GET_QUOTA_INFORMATION).
// SidLength is derived from len(Sid) on marshal.
type FileGetQuotaInformation struct {
	// NextEntryOffset (4 bytes): byte offset to the next entry, or zero if this is the last.
	NextEntryOffset uint32
	// Sid (variable): the SID whose quota is requested.
	Sid []byte
}

// Marshal serializes the FILE_GET_QUOTA_INFORMATION entry.
func (e *FileGetQuotaInformation) Marshal() ([]byte, error) {
	b := make([]byte, fileGetQuotaInformationFixedSize+len(e.Sid))
	binary.LittleEndian.PutUint32(b[0:4], e.NextEntryOffset)
	binary.LittleEndian.PutUint32(b[4:8], uint32(len(e.Sid)))
	copy(b[fileGetQuotaInformationFixedSize:], e.Sid)
	return b, nil
}

// Unmarshal parses a FILE_GET_QUOTA_INFORMATION entry, returning the bytes consumed.
func (e *FileGetQuotaInformation) Unmarshal(data []byte) (int, error) {
	if len(data) < fileGetQuotaInformationFixedSize {
		return 0, fmt.Errorf("subcommands: FILE_GET_QUOTA_INFORMATION requires at least %d bytes, got %d", fileGetQuotaInformationFixedSize, len(data))
	}
	e.NextEntryOffset = binary.LittleEndian.Uint32(data[0:4])
	sidLength := int(binary.LittleEndian.Uint32(data[4:8]))
	if len(data) < fileGetQuotaInformationFixedSize+sidLength {
		return fileGetQuotaInformationFixedSize, fmt.Errorf("subcommands: FILE_GET_QUOTA_INFORMATION SID truncated: need %d, have %d", sidLength, len(data)-fileGetQuotaInformationFixedSize)
	}
	e.Sid = append([]byte{}, data[fileGetQuotaInformationFixedSize:fileGetQuotaInformationFixedSize+sidLength]...)
	return fileGetQuotaInformationFixedSize + sidLength, nil
}

// FileQuotaInformation is a single user's quota record ([MS-FSCC] section 2.4.41
// FILE_QUOTA_INFORMATION). It is the NT_Trans_Data record of an NT_TRANSACT_QUERY_QUOTA
// response and the record a client sends in an NT_TRANSACT_SET_QUOTA request. SidLength is
// derived from len(Sid) on marshal. QuotaThreshold/QuotaLimit may be -1 (no limit) and
// QuotaLimit may be -2 (delete the entry), hence the signed types.
type FileQuotaInformation struct {
	// NextEntryOffset (4 bytes): byte offset to the next entry, or zero if this is the last.
	NextEntryOffset uint32
	// ChangeTime (8 bytes): last time the quota was changed (FILETIME). Ignored on set.
	ChangeTime uint64
	// QuotaUsed (8 bytes): bytes of quota used by this user.
	QuotaUsed int64
	// QuotaThreshold (8 bytes): warning threshold in bytes, or -1 for none.
	QuotaThreshold int64
	// QuotaLimit (8 bytes): quota limit in bytes, -1 for none, or -2 to delete the entry.
	QuotaLimit int64
	// Sid (variable): the SID this quota applies to.
	Sid []byte
}

// Marshal serializes the FILE_QUOTA_INFORMATION record.
func (q *FileQuotaInformation) Marshal() ([]byte, error) {
	b := make([]byte, fileQuotaInformationFixedSize+len(q.Sid))
	binary.LittleEndian.PutUint32(b[0:4], q.NextEntryOffset)
	binary.LittleEndian.PutUint32(b[4:8], uint32(len(q.Sid)))
	binary.LittleEndian.PutUint64(b[8:16], q.ChangeTime)
	binary.LittleEndian.PutUint64(b[16:24], uint64(q.QuotaUsed))
	binary.LittleEndian.PutUint64(b[24:32], uint64(q.QuotaThreshold))
	binary.LittleEndian.PutUint64(b[32:40], uint64(q.QuotaLimit))
	copy(b[fileQuotaInformationFixedSize:], q.Sid)
	return b, nil
}

// Unmarshal parses a FILE_QUOTA_INFORMATION record, returning the bytes consumed.
func (q *FileQuotaInformation) Unmarshal(data []byte) (int, error) {
	if len(data) < fileQuotaInformationFixedSize {
		return 0, fmt.Errorf("subcommands: FILE_QUOTA_INFORMATION requires at least %d bytes, got %d", fileQuotaInformationFixedSize, len(data))
	}
	q.NextEntryOffset = binary.LittleEndian.Uint32(data[0:4])
	sidLength := int(binary.LittleEndian.Uint32(data[4:8]))
	q.ChangeTime = binary.LittleEndian.Uint64(data[8:16])
	q.QuotaUsed = int64(binary.LittleEndian.Uint64(data[16:24]))
	q.QuotaThreshold = int64(binary.LittleEndian.Uint64(data[24:32]))
	q.QuotaLimit = int64(binary.LittleEndian.Uint64(data[32:40]))
	if len(data) < fileQuotaInformationFixedSize+sidLength {
		return fileQuotaInformationFixedSize, fmt.Errorf("subcommands: FILE_QUOTA_INFORMATION SID truncated: need %d, have %d", sidLength, len(data)-fileQuotaInformationFixedSize)
	}
	q.Sid = append([]byte{}, data[fileQuotaInformationFixedSize:fileQuotaInformationFixedSize+sidLength]...)
	return fileQuotaInformationFixedSize + sidLength, nil
}

// ParseFileQuotaInformationList parses the NT_TRANSACT_QUERY_QUOTA response NT_Trans_Data
// into its FILE_QUOTA_INFORMATION records, following NextEntryOffset until it is zero or
// the buffer is exhausted.
func ParseFileQuotaInformationList(data []byte) ([]FileQuotaInformation, error) {
	items := []FileQuotaInformation{}
	offset := 0
	for offset+fileQuotaInformationFixedSize <= len(data) {
		var rec FileQuotaInformation
		if _, err := rec.Unmarshal(data[offset:]); err != nil {
			return items, err
		}
		items = append(items, rec)
		if rec.NextEntryOffset == 0 {
			break
		}
		offset += int(rec.NextEntryOffset)
	}
	return items, nil
}
