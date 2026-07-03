package functions

import (
	"fmt"

	qm2qm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1088a980-eae5-11d0-8d9b-00a02453c337/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// remoteQmGetVersionRequest carries the [in] parameters of RemoteQmGetVersion.
type remoteQmGetVersionRequest struct {
}

func (*remoteQmGetVersionRequest) Opnum() uint16 { return qm2qm.OpnumRemoteQmGetVersion }

// remoteQmGetVersionResponse carries the [out] parameters of RemoteQmGetVersion. The IDL
// method is declared `void` ([MS-MQQP] 3.1.4.9), so there is NO return value on the wire —
// only the three out parameters are marshaled back.
type remoteQmGetVersionResponse struct {
	PMajor       uint8
	PMinor       uint8
	PBuildNumber uint16
}

// RemoteQmGetVersion calls RemoteQmGetVersion (opnum 8) ([MS-MQQP] 3.1.4.9). The method
// returns void; the only error path is the underlying RPC transport.
func RemoteQmGetVersion(rpc ndr.Invoker) (PMajor uint8, PMinor uint8, PBuildNumber uint16, err error) {
	req := &remoteQmGetVersionRequest{}
	var resp remoteQmGetVersionResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQmGetVersion: %w", err)
		return
	}
	PMajor = resp.PMajor
	PMinor = resp.PMinor
	PBuildNumber = resp.PBuildNumber
	return
}
