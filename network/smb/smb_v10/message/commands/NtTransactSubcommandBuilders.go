package commands

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// This file wires the NT_TRANSACT subcommand payload structures (network/smb/smb_v10/
// subcommands) into the SMB_COM_NT_TRANSACT message: request builders populate an
// NtTransactRequest (Function, Setup words, NT_Trans_Parameters, NT_Trans_Data and the
// associated counts) from a typed subcommand struct, and response parsers extract the
// typed result from an NtTransactResponse. See [MS-CIFS] section 2.2.4.62 and the
// per-subcommand sections referenced on each helper.

// setupWordsFromBytes packs an even-length little-endian byte block (a subcommand's NT_Trans
// setup) into the USHORT words carried by NtTransactRequest.Setup. An odd trailing byte, if
// any, is zero-padded.
func setupWordsFromBytes(b []byte) []types.USHORT {
	words := make([]types.USHORT, (len(b)+1)/2)
	for i := range words {
		var lo, hi byte
		lo = b[i*2]
		if i*2+1 < len(b) {
			hi = b[i*2+1]
		}
		words[i] = types.USHORT(uint16(lo) | uint16(hi)<<8)
	}
	return words
}

// uchars copies a byte slice into a []types.UCHAR (identical underlying element type).
func uchars(b []byte) []types.UCHAR { return append([]types.UCHAR{}, b...) }

// NewNtTransactIoctlSetup builds the 4-word (8-octet) NT_TRANSACT_IOCTL setup ([MS-CIFS]
// section 2.2.7.2.1): FunctionCode(4) + FID(2) + IsFsctl(1) + IsFlags(1). For the SRV_*
// FSCTLs, isFsctl is true and isFlags is zero.
func NewNtTransactIoctlSetup(functionCode uint32, fid uint16, isFsctl bool, isFlags uint8) []types.USHORT {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b[0:4], functionCode)
	binary.LittleEndian.PutUint16(b[4:6], fid)
	if isFsctl {
		b[6] = 1
	}
	b[7] = isFlags
	return setupWordsFromBytes(b)
}

// ---- NT_TRANSACT_NOTIFY_CHANGE (0x0004) ------------------------------------------------

// NewNtTransactNotifyChangeRequest builds an NT_TRANSACT_NOTIFY_CHANGE request ([MS-CIFS]
// section 2.2.7.4.1). The arguments travel in the NT_Trans setup; the request carries no
// NT_Trans_Parameters or NT_Trans_Data. maxParameterCount bounds the FILE_NOTIFY_INFORMATION
// bytes the server may return.
func NewNtTransactNotifyChangeRequest(setup subcommands.NtTransactNotifyChangeSetup, maxParameterCount uint32) (*NtTransactRequest, error) {
	setupBytes, err := setup.Marshal()
	if err != nil {
		return nil, err
	}
	r := NewNtTransactRequest()
	r.Function = types.USHORT(subcommands.NT_TRANSACT_NOTIFY_CHANGE)
	r.Setup = setupWordsFromBytes(setupBytes)
	r.SetupCount = types.UCHAR(len(r.Setup))
	r.MaxParameterCount = types.ULONG(maxParameterCount)
	return r, nil
}

// NotifyChangeInformation parses the FILE_NOTIFY_INFORMATION list from an
// NT_TRANSACT_NOTIFY_CHANGE response NT_Trans_Parameters ([MS-CIFS] section 2.2.7.4.2).
func (c *NtTransactResponse) NotifyChangeInformation() ([]subcommands.FileNotifyInformation, error) {
	return subcommands.ParseFileNotifyInformationList(c.Parameters)
}

// ---- NT_TRANSACT_QUERY_SECURITY_DESC (0x0006) / SET_SECURITY_DESC (0x0003) --------------

// NewNtTransactQuerySecurityDescRequest builds an NT_TRANSACT_QUERY_SECURITY_DESC request
// ([MS-CIFS] section 2.2.7.6.1). The FID/SecurityInformation travel in NT_Trans_Parameters;
// maxDataCount bounds the returned SECURITY_DESCRIPTOR and MaxParameterCount is 4 (LengthNeeded).
func NewNtTransactQuerySecurityDescRequest(params subcommands.NtTransactSecurityDescParameters, maxDataCount uint32) (*NtTransactRequest, error) {
	pb, err := params.Marshal()
	if err != nil {
		return nil, err
	}
	r := NewNtTransactRequest()
	r.Function = types.USHORT(subcommands.NT_TRANSACT_QUERY_SECURITY_DESC)
	r.NT_Trans_Parameters = uchars(pb)
	r.TotalParameterCount = types.ULONG(len(pb))
	r.ParameterCount = types.ULONG(len(pb))
	r.MaxParameterCount = types.ULONG(4)
	r.MaxDataCount = types.ULONG(maxDataCount)
	return r, nil
}

// NewNtTransactSetSecurityDescRequest builds an NT_TRANSACT_SET_SECURITY_DESC request
// ([MS-CIFS] section 2.2.7.3.1). FID/SecurityInformation travel in NT_Trans_Parameters and
// the self-relative SECURITY_DESCRIPTOR travels in NT_Trans_Data.
func NewNtTransactSetSecurityDescRequest(params subcommands.NtTransactSecurityDescParameters, securityDescriptor []byte) (*NtTransactRequest, error) {
	pb, err := params.Marshal()
	if err != nil {
		return nil, err
	}
	r := NewNtTransactRequest()
	r.Function = types.USHORT(subcommands.NT_TRANSACT_SET_SECURITY_DESC)
	r.NT_Trans_Parameters = uchars(pb)
	r.TotalParameterCount = types.ULONG(len(pb))
	r.ParameterCount = types.ULONG(len(pb))
	r.NT_Trans_Data = uchars(securityDescriptor)
	r.TotalDataCount = types.ULONG(len(securityDescriptor))
	r.DataCount = types.ULONG(len(securityDescriptor))
	return r, nil
}

// QuerySecurityDescriptor parses an NT_TRANSACT_QUERY_SECURITY_DESC response ([MS-CIFS]
// section 2.2.7.6.2): the LengthNeeded parameter and the returned SECURITY_DESCRIPTOR bytes.
func (c *NtTransactResponse) QuerySecurityDescriptor() (subcommands.NtTransactQuerySecurityDescResponseParameters, []byte, error) {
	var params subcommands.NtTransactQuerySecurityDescResponseParameters
	if _, err := params.Unmarshal(c.Parameters); err != nil {
		return params, nil, err
	}
	return params, append([]byte{}, c.Data...), nil
}

// ---- NT_TRANSACT_QUERY_QUOTA (0x0007) --------------------------------------------------

// NewNtTransactQueryQuotaRequest builds an NT_TRANSACT_QUERY_QUOTA request ([MS-SMB] section
// 2.2.7.5.1). The query parameters travel in NT_Trans_Parameters and the optional SidList
// (FILE_GET_QUOTA_INFORMATION entries) travels in NT_Trans_Data. maxDataCount bounds the
// returned FILE_QUOTA_INFORMATION list.
func NewNtTransactQueryQuotaRequest(params subcommands.NtTransQueryQuotaRequestParameters, sidList []byte, maxDataCount uint32) (*NtTransactRequest, error) {
	pb, err := params.Marshal()
	if err != nil {
		return nil, err
	}
	r := NewNtTransactRequest()
	r.Function = types.USHORT(subcommands.NT_TRANSACT_QUERY_QUOTA)
	r.NT_Trans_Parameters = uchars(pb)
	r.TotalParameterCount = types.ULONG(len(pb))
	r.ParameterCount = types.ULONG(len(pb))
	r.NT_Trans_Data = uchars(sidList)
	r.TotalDataCount = types.ULONG(len(sidList))
	r.DataCount = types.ULONG(len(sidList))
	r.MaxParameterCount = types.ULONG(4) // DataLength
	r.MaxDataCount = types.ULONG(maxDataCount)
	return r, nil
}

// QuotaInformation parses an NT_TRANSACT_QUERY_QUOTA response ([MS-SMB] section 2.2.7.5.2):
// the DataLength parameter and the FILE_QUOTA_INFORMATION list in NT_Trans_Data.
func (c *NtTransactResponse) QuotaInformation() (subcommands.NtTransQuotaResponseParameters, []subcommands.FileQuotaInformation, error) {
	var params subcommands.NtTransQuotaResponseParameters
	if len(c.Parameters) >= 4 {
		if _, err := params.Unmarshal(c.Parameters); err != nil {
			return params, nil, err
		}
	}
	list, err := subcommands.ParseFileQuotaInformationList(c.Data)
	return params, list, err
}

// ---- NT_TRANSACT_IOCTL (0x0002): server-side copy and snapshots -------------------------

// NewNtTransactRequestResumeKeyRequest builds the FSCTL_SRV_REQUEST_RESUME_KEY request for
// the source file ([MS-SMB] section 2.2.7.2): an NT_TRANSACT_IOCTL whose setup names the
// FSCTL and the source FID, with no input data. maxDataCount bounds the returned resume-key
// data (a 24-byte key plus a 4-byte ContextLength).
func NewNtTransactRequestResumeKeyRequest(fid uint16, maxDataCount uint32) *NtTransactRequest {
	r := NewNtTransactRequest()
	r.Function = types.USHORT(subcommands.NT_TRANSACT_IOCTL)
	r.Setup = NewNtTransactIoctlSetup(subcommands.FSCTL_SRV_REQUEST_RESUME_KEY, fid, true, 0)
	r.SetupCount = types.UCHAR(len(r.Setup))
	r.MaxDataCount = types.ULONG(maxDataCount)
	return r
}

// NewNtTransactCopychunkRequest builds an FSCTL_SRV_COPYCHUNK request against the destination
// file ([MS-SMB] section 2.2.7.2): an NT_TRANSACT_IOCTL whose setup names the FSCTL and the
// destination FID, carrying the SRV_COPYCHUNK_COPY as input data. The response is a 12-byte
// SRV_COPYCHUNK_RESPONSE.
func NewNtTransactCopychunkRequest(fid uint16, copy subcommands.SrvCopychunkCopy) (*NtTransactRequest, error) {
	db, err := copy.Marshal()
	if err != nil {
		return nil, err
	}
	r := NewNtTransactRequest()
	r.Function = types.USHORT(subcommands.NT_TRANSACT_IOCTL)
	r.Setup = NewNtTransactIoctlSetup(subcommands.FSCTL_SRV_COPYCHUNK, fid, true, 0)
	r.SetupCount = types.UCHAR(len(r.Setup))
	r.NT_Trans_Data = uchars(db)
	r.TotalDataCount = types.ULONG(len(db))
	r.DataCount = types.ULONG(len(db))
	r.MaxDataCount = types.ULONG(12) // SRV_COPYCHUNK_RESPONSE
	return r, nil
}

// NewNtTransactEnumerateSnapshotsRequest builds an FSCTL_SRV_ENUMERATE_SNAPSHOTS request
// ([MS-SMB] section 2.2.7.3): an NT_TRANSACT_IOCTL whose setup names the FSCTL and the open
// FID, with no input data. maxDataCount bounds the returned SRV_SNAPSHOT_ARRAY.
func NewNtTransactEnumerateSnapshotsRequest(fid uint16, maxDataCount uint32) *NtTransactRequest {
	r := NewNtTransactRequest()
	r.Function = types.USHORT(subcommands.NT_TRANSACT_IOCTL)
	r.Setup = NewNtTransactIoctlSetup(subcommands.FSCTL_SRV_ENUMERATE_SNAPSHOTS, fid, true, 0)
	r.SetupCount = types.UCHAR(len(r.Setup))
	r.MaxDataCount = types.ULONG(maxDataCount)
	return r
}

// RequestResumeKey parses an FSCTL_SRV_REQUEST_RESUME_KEY response from the NT_Trans_Data
// ([MS-SMB] section 2.2.7.2.2.2).
func (c *NtTransactResponse) RequestResumeKey() (subcommands.SrvRequestResumeKeyResponse, error) {
	var out subcommands.SrvRequestResumeKeyResponse
	_, err := out.Unmarshal(c.Data)
	return out, err
}

// CopychunkResponse parses an FSCTL_SRV_COPYCHUNK response from the NT_Trans_Data ([MS-SMB]
// section 2.2.7.2.2.1).
func (c *NtTransactResponse) CopychunkResponse() (subcommands.SrvCopychunkResponse, error) {
	var out subcommands.SrvCopychunkResponse
	_, err := out.Unmarshal(c.Data)
	return out, err
}

// SnapshotArray parses an FSCTL_SRV_ENUMERATE_SNAPSHOTS response from the NT_Trans_Data
// ([MS-SMB] section 2.2.7.3.2).
func (c *NtTransactResponse) SnapshotArray() (subcommands.SrvSnapshotArray, error) {
	var out subcommands.SrvSnapshotArray
	_, err := out.Unmarshal(c.Data)
	return out, err
}
