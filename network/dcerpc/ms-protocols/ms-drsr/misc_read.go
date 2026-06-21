package msdrsr

import (
	"fmt"

	drsuapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// readNgcKeyRequest / readNgcKeyResponse drive IDL_DRSReadNgcKey directly (rather than via
// functions.IDL_DRSReadNgcKey) so a non-success RetVal — notably "account has no NGC key"
// — is returned as a code rather than collapsed into a transport error.
type readNgcKeyRequest struct {
	HDrs        structures.DRS_HANDLE
	DwInVersion ndr.DWORD
	PmsgIn      structures.DRS_MSG_READNGCKEYREQ
}

func (*readNgcKeyRequest) Opnum() uint16 { return drsuapi.OpnumIDL_DRSReadNgcKey }

type readNgcKeyResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DRS_MSG_READNGCKEYREPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// ReadNgcKey reads the NGC (Windows Hello for Business) protector key of an account via
// IDL_DRSReadNgcKey (opnum 30). account is the account's distinguished name. It returns
// the raw key blob (empty when the account has no key) and the server's RetVal: 0 means a
// key was returned; a non-zero code (e.g. 0x200A ERROR_DS_NO_ATTRIBUTE_OR_VALUE) means no
// key is present. err is only set for transport/RPC failures. Read-only.
func (c *Client) ReadNgcKey(accountDN string) (key []byte, retVal uint32, err error) {
	if !c.bound {
		return nil, 0, fmt.Errorf("msdrsr: not connected")
	}
	acct := ndr.WSTR(accountDN)
	req := &readNgcKeyRequest{
		DwInVersion: 1,
		PmsgIn:      structures.DRS_MSG_READNGCKEYREQ{Tag: 1, V1: structures.DRS_MSG_READNGCKEYREQ_V1{PwszAccount: &acct}},
		HDrs:        c.handle,
	}
	var resp readNgcKeyResponse
	if err := c.rpc.Invoke(req, &resp); err != nil {
		return nil, 0, fmt.Errorf("msdrsr: ReadNgcKey: %w", err)
	}
	return append([]byte(nil), resp.PmsgOut.V1.PNgcKey...), uint32(resp.Status), nil
}

// NT4ChangeLog is the result of IDL_DRSGetNT4ChangeLog: the opaque change-log and restart
// blobs (legacy NT4 BDC synchronization, [MS-DRSR] 4.1.9) and the server's NTSTATUS. The
// blobs are returned verbatim (legacy NT4 SAM delta stream). Read-only. Modern DCs return
// ERROR_NOT_SUPPORTED.
type NT4ChangeLog struct {
	Restart        []byte
	Log            []byte
	ActualNTStatus uint32
}

// GetNT4ChangeLog calls IDL_DRSGetNT4ChangeLog (opnum 11). restart is the opaque cursor
// from a previous call (nil to start); preferredMaxLength bounds the returned log size.
func (c *Client) GetNT4ChangeLog(restart []byte, preferredMaxLength uint32) (*NT4ChangeLog, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}
	v1 := structures.DRS_MSG_NT4_CHGLOG_REQ_V1{
		PreferredMaximumLength: ndr.DWORD(preferredMaxLength),
		CbRestart:              ndr.DWORD(len(restart)),
		PRestart:               restart,
	}
	msgIn := structures.DRS_MSG_NT4_CHGLOG_REQ{Tag: 1, V1: v1}
	_, msgOut, err := functions.IDL_DRSGetNT4ChangeLog(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: GetNT4ChangeLog: %w", err)
	}
	r := msgOut.V1
	return &NT4ChangeLog{
		Restart:        append([]byte(nil), r.PRestart...),
		Log:            append([]byte(nil), r.PLog...),
		ActualNTStatus: uint32(r.ActualNtStatus),
	}, nil
}
