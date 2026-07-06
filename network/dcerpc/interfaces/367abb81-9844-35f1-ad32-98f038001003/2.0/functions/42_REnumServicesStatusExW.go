package functions

// IDL source: [MS-SCMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-scmr/19168537-40b5-4d7a-99e0-d77f0f5e0241
// A fetched copy is kept at ms-scmr.idl in the interface directory.

import (
	"fmt"

	svcctl "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/367abb81-9844-35f1-ad32-98f038001003/2.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msscmr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-scmr"
)

// rEnumServicesStatusExWRequest carries the [in] parameters of REnumServicesStatusExW.
type rEnumServicesStatusExWRequest struct {
	HSCManager     msscmr.SC_RPC_HANDLE
	InfoLevel      msscmr.SC_ENUM_TYPE
	DwServiceType  ndr.DWORD
	DwServiceState ndr.DWORD
	CbBufSize      ndr.DWORD
	LpResumeIndex  *ndr.DWORD `ndr:"unique"`
	PszGroupName   *ndr.WSTR  `ndr:"unique"`
}

func (*rEnumServicesStatusExWRequest) Opnum() uint16 { return svcctl.OpnumREnumServicesStatusExW }

// rEnumServicesStatusExWResponse carries the [out] parameters and return value of REnumServicesStatusExW.
type rEnumServicesStatusExWResponse struct {
	LpBuffer           []uint8 `ndr:"ref,size_is=CbBufSize"`
	PcbBytesNeeded     msscmr.LPBOUNDED_DWORD_256K
	LpServicesReturned msscmr.LPBOUNDED_DWORD_256K
	LpResumeIndex      *ndr.DWORD `ndr:"unique"`
	Status             ndr.DWORD  `ndr:"retval"`
}

// REnumServicesStatusExW calls REnumServicesStatusExW (opnum 42) ([MS-SCMR] — verify the parameter
// modeling and status handling).
func REnumServicesStatusExW(rpc ndr.Invoker, hSCManager msscmr.SC_RPC_HANDLE, infoLevel msscmr.SC_ENUM_TYPE, dwServiceType ndr.DWORD, dwServiceState ndr.DWORD, cbBufSize ndr.DWORD, lpResumeIndex *ndr.DWORD, pszGroupName *ndr.WSTR) (LpBuffer []uint8, PcbBytesNeeded msscmr.LPBOUNDED_DWORD_256K, LpServicesReturned msscmr.LPBOUNDED_DWORD_256K, LpResumeIndex *ndr.DWORD, err error) {
	req := &rEnumServicesStatusExWRequest{
		HSCManager:     hSCManager,
		InfoLevel:      infoLevel,
		DwServiceType:  dwServiceType,
		DwServiceState: dwServiceState,
		CbBufSize:      cbBufSize,
		LpResumeIndex:  lpResumeIndex,
		PszGroupName:   pszGroupName,
	}
	var resp rEnumServicesStatusExWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("REnumServicesStatusExW: %w", err)
		return
	}
	LpBuffer = resp.LpBuffer
	PcbBytesNeeded = resp.PcbBytesNeeded
	LpServicesReturned = resp.LpServicesReturned
	LpResumeIndex = resp.LpResumeIndex
	if uint32(resp.Status) != svcctl.StatusSuccess {
		err = fmt.Errorf("REnumServicesStatusExW failed: %s", svcctl.StatusString(uint32(resp.Status)))
	}
	return
}
