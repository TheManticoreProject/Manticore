package functions

// IDL source: [MS-CMRP] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmrp/e12b6d8f-c410-49d7-a27d-9992782a9027
// A fetched copy is kept at ms-cmrp.idl in the interface directory.

import (
	"fmt"

	clusapi "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/b97db8b2-4c63-11cf-bff6-08002be23f2f/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// apiGetClusterVersionRequest carries the [in] parameters of ApiGetClusterVersion.
type apiGetClusterVersionRequest struct {
}

func (*apiGetClusterVersionRequest) Opnum() uint16 { return clusapi.OpnumApiGetClusterVersion }

// apiGetClusterVersionResponse carries the [out] parameters and return value of ApiGetClusterVersion.
type apiGetClusterVersionResponse struct {
	LpwMajorVersion uint16
	LpwMinorVersion uint16
	LpwBuildNumber  uint16
	LpszVendorId    ndr.WSTR
	LpszCSDVersion  ndr.WSTR
	Status          ndr.DWORD `ndr:"retval"`
}

// ApiGetClusterVersion calls ApiGetClusterVersion (opnum 4) ([MS-CMRP] — verify the parameter
// modeling and status handling).
func ApiGetClusterVersion(rpc ndr.Invoker) (LpwMajorVersion uint16, LpwMinorVersion uint16, LpwBuildNumber uint16, LpszVendorId ndr.WSTR, LpszCSDVersion ndr.WSTR, err error) {
	req := &apiGetClusterVersionRequest{}
	var resp apiGetClusterVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("ApiGetClusterVersion: %w", err)
		return
	}
	LpwMajorVersion = resp.LpwMajorVersion
	LpwMinorVersion = resp.LpwMinorVersion
	LpwBuildNumber = resp.LpwBuildNumber
	LpszVendorId = resp.LpszVendorId
	LpszCSDVersion = resp.LpszCSDVersion
	if uint32(resp.Status) != clusapi.StatusSuccess {
		err = fmt.Errorf("ApiGetClusterVersion failed: %s", clusapi.StatusString(uint32(resp.Status)))
	}
	return
}
