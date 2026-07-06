package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmrp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmrp"
)

// apiGetClusterVersion2Request carries the [in] parameters of ApiGetClusterVersion2.
type apiGetClusterVersion2Request struct {
}

func (*apiGetClusterVersion2Request) Opnum() uint16 { return clusapi.OpnumApiGetClusterVersion2 }

// apiGetClusterVersion2Response carries the [out] parameters and return value of ApiGetClusterVersion2.
type apiGetClusterVersion2Response struct {
	LpwMajorVersion    uint16
	LpwMinorVersion    uint16
	LpwBuildNumber     uint16
	LpszVendorId       ndr.WSTR
	LpszCSDVersion     ndr.WSTR
	PpClusterOpVerInfo *mscmrp.CLUSTER_OPERATIONAL_VERSION_INFO `ndr:"unique"`
	Rpc_status         ndr.DWORD
	Status             ndr.DWORD `ndr:"retval"`
}

// ApiGetClusterVersion2 calls ApiGetClusterVersion2 (opnum 102) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetClusterVersion2(rpc ndr.Invoker) (LpwMajorVersion uint16, LpwMinorVersion uint16, LpwBuildNumber uint16, LpszVendorId ndr.WSTR, LpszCSDVersion ndr.WSTR, PpClusterOpVerInfo *mscmrp.CLUSTER_OPERATIONAL_VERSION_INFO, Rpc_status ndr.DWORD, err error) {
	req := &apiGetClusterVersion2Request{}
	var resp apiGetClusterVersion2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetClusterVersion2: %w", err)
		return
	}
	LpwMajorVersion = resp.LpwMajorVersion
	LpwMinorVersion = resp.LpwMinorVersion
	LpwBuildNumber = resp.LpwBuildNumber
	LpszVendorId = resp.LpszVendorId
	LpszCSDVersion = resp.LpszCSDVersion
	PpClusterOpVerInfo = resp.PpClusterOpVerInfo
	Rpc_status = resp.Rpc_status
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetClusterVersion2 failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
