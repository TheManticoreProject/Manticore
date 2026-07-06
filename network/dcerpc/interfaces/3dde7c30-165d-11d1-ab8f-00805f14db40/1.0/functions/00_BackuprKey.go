package functions

import (
	"fmt"

	BackupKey "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/3dde7c30-165d-11d1-ab8f-00805f14db40/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

// backuprKeyRequest carries the [in] parameters of BackuprKey. PguidActionAgent is the
// NDR-marshallable [MS-DTYP] GUID (16 octets on the wire), not windows/guid.GUID whose
// uint64 tail would marshal to 24 octets.
type backuprKeyRequest struct {
	PguidActionAgent msdtyp.GUID
	PDataIn          []uint8 `ndr:"ref,size_is=CbDataIn"`
	CbDataIn         ndr.DWORD
	DwParam          ndr.DWORD
}

func (*backuprKeyRequest) Opnum() uint16 { return BackupKey.OpnumBackuprKey }

// backuprKeyResponse carries the [out] parameters and return value of BackuprKey.
// ppDataOut is [out, size_is(,*pcbDataOut)] byte** — a unique pointer to a conformant
// array of bytes (one referent id + max_count + octets), so it is modelled as a single
// []uint8 behind a unique pointer, NOT a []*uint8 (which would emit a per-byte referent
// id array and mismatch the wire).
type backuprKeyResponse struct {
	PpDataOut  []uint8 `ndr:"unique,size_is=PcbDataOut"`
	PcbDataOut ndr.DWORD
	Status     ndr.DWORD `ndr:"retval"`
}

// BackuprKey calls BackuprKey (opnum 0) ([MS-BKRP] section 3.1.4.1). pguidActionAgent
// selects the operation (one of the BackupKey* action GUIDs); pDataIn is the opaque
// input BLOB (its format depends on the action) and cbDataIn its length in bytes.
// dwParam is unused and MUST be zero. On success the server returns the output BLOB in
// ppDataOut with its length in pcbDataOut.
func BackuprKey(rpc ndr.Invoker, pguidActionAgent guid.GUID, pDataIn []uint8, cbDataIn ndr.DWORD, dwParam ndr.DWORD) (PpDataOut []uint8, PcbDataOut ndr.DWORD, err error) {
	req := &backuprKeyRequest{
		PguidActionAgent: msdtyp.NewGUID(pguidActionAgent),
		PDataIn:          pDataIn,
		CbDataIn:         cbDataIn,
		DwParam:          dwParam,
	}
	var resp backuprKeyResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("BackuprKey: %w", err)
		return
	}
	PpDataOut = resp.PpDataOut
	PcbDataOut = resp.PcbDataOut
	if uint32(resp.Status) != BackupKey.ErrorSuccess {
		err = fmt.Errorf("BackuprKey failed: %s", BackupKey.StatusString(uint32(resp.Status)))
	}
	return
}
