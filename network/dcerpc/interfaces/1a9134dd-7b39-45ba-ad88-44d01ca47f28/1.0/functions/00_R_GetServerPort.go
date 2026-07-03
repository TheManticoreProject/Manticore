package functions

import (
	"fmt"

	RemoteRead "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1a9134dd-7b39-45ba-ad88-44d01ca47f28/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// r_GetServerPortRequest carries the [in] parameters of R_GetServerPort.
type r_GetServerPortRequest struct {
}

func (*r_GetServerPortRequest) Opnum() uint16 { return RemoteRead.OpnumR_GetServerPort }

// r_GetServerPortResponse carries the return value of R_GetServerPort. The DWORD result is
// not an HRESULT: it is the TCP port on which the server receives new remote-read requests
// ([MS-MQRR] section 3.1.4.1). A returned value of 0 signals failure.
type r_GetServerPortResponse struct {
	Port ndr.DWORD `ndr:"retval"`
}

// R_GetServerPort calls R_GetServerPort (opnum 0) ([MS-MQRR] section 3.1.4.1). It returns the
// TCP port on which the server listens for subsequent RemoteRead calls; a return of 0 means
// the server has no port available for this client and the caller cannot proceed.
func R_GetServerPort(rpc ndr.Invoker) (port uint32, err error) {
	req := &r_GetServerPortRequest{}
	var resp r_GetServerPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_GetServerPort: %w", err)
		return
	}
	port = uint32(resp.Port)
	if port == 0 {
		err = fmt.Errorf("R_GetServerPort: server returned port 0 (no port available)")
	}
	return
}
