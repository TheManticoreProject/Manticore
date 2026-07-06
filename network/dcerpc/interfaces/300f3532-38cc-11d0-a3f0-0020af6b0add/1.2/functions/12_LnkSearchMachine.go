package functions

// IDL source: [MS-DLTW] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dltw/e415dc67-3969-45c8-ba63-731e78870dfc
// A fetched copy is kept at ms-dltw.idl in the interface directory.

import (
	"fmt"

	trkwks "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/300f3532-38cc-11d0-a3f0-0020af6b0add/1.2"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdltw "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dltw"
)

// lnkSearchMachineRequest carries the [in] parameters of LnkSearchMachine.
type lnkSearchMachineRequest struct {
	Restrictions    ndr.DWORD
	PdroidBirthLast msdltw.CDomainRelativeObjId
	PdroidLast      msdltw.CDomainRelativeObjId
}

func (*lnkSearchMachineRequest) Opnum() uint16 { return trkwks.OpnumLnkSearchMachine }

// lnkSearchMachineResponse carries the [out] parameters and return value of LnkSearchMachine.
type lnkSearchMachineResponse struct {
	PdroidBirthNext msdltw.CDomainRelativeObjId
	PdroidNext      msdltw.CDomainRelativeObjId
	PmcidNext       msdltw.CMachineId
	PtszPath        ndr.WSTR
	Status          ndr.DWORD `ndr:"retval"`
}

// LnkSearchMachine calls LnkSearchMachine (opnum 12, [MS-DLTW] 3.1.4.1): it searches the
// target computer for the file identified by pdroidBirthLast (FileID) and pdroidLast
// (last-known FileLocation), returning the file's new FileID (PdroidBirthNext),
// FileLocation (PdroidNext), the holding computer's MachineID (PmcidNext), and its UNC
// path (PtszPath). Restrictions is unused and MUST be zero.
//
// The method returns an HRESULT; the [out] parameters remain meaningful for the soft
// failures TRK_E_REFERRAL (a move referral) and TRK_E_POTENTIAL_FILE_FOUND (a candidate
// file), so they are always returned to the caller and the HRESULT is surfaced verbatim
// in err. Success is any value with the sign bit clear (see trkwks.StatusIsSuccess).
func LnkSearchMachine(rpc ndr.Invoker, restrictions ndr.DWORD, pdroidBirthLast msdltw.CDomainRelativeObjId, pdroidLast msdltw.CDomainRelativeObjId) (PdroidBirthNext msdltw.CDomainRelativeObjId, PdroidNext msdltw.CDomainRelativeObjId, PmcidNext msdltw.CMachineId, PtszPath ndr.WSTR, err error) {
	req := &lnkSearchMachineRequest{
		Restrictions:    restrictions,
		PdroidBirthLast: pdroidBirthLast,
		PdroidLast:      pdroidLast,
	}
	var resp lnkSearchMachineResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LnkSearchMachine: %w", err)
		return
	}
	PdroidBirthNext = resp.PdroidBirthNext
	PdroidNext = resp.PdroidNext
	PmcidNext = resp.PmcidNext
	PtszPath = resp.PtszPath
	if !trkwks.StatusIsSuccess(uint32(resp.Status)) {
		err = fmt.Errorf("LnkSearchMachine failed: %s", trkwks.StatusString(uint32(resp.Status)))
	}
	return
}
