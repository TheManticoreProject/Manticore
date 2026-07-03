package functions

import (
	"fmt"

	qm2qm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1088a980-eae5-11d0-8d9b-00a02453c337/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// remoteQMGetQMQMServerPortRequest carries the [in] parameters of RemoteQMGetQMQMServerPort.
type remoteQMGetQMQMServerPortRequest struct {
	DwPortType ndr.DWORD
}

func (*remoteQMGetQMQMServerPortRequest) Opnum() uint16 { return qm2qm.OpnumRemoteQMGetQMQMServerPort }

// remoteQMGetQMQMServerPortResponse carries the return value of RemoteQMGetQMQMServerPort.
// Unlike the other methods, the DWORD return is NOT an HRESULT: it is the TCP (or, on legacy
// transports, SPX) port the client should bind to for subsequent calls ([MS-MQQP] 3.1.4.8).
type remoteQMGetQMQMServerPortResponse struct {
	Port ndr.DWORD `ndr:"retval"`
}

// RemoteQMGetQMQMServerPort calls RemoteQMGetQMQMServerPort (opnum 7) ([MS-MQQP] 3.1.4.8).
// dwPortType selects which server port to return (0..3, e.g. the qm2qm RPC port). The
// returned DWORD is the port number; the server returns 0x00000000 to indicate failure
// (for example, an unsupported dwPortType or transport), which is surfaced as an error.
func RemoteQMGetQMQMServerPort(rpc ndr.Invoker, dwPortType ndr.DWORD) (Port uint32, err error) {
	req := &remoteQMGetQMQMServerPortRequest{
		DwPortType: dwPortType,
	}
	var resp remoteQMGetQMQMServerPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RemoteQMGetQMQMServerPort: %w", err)
		return
	}
	Port = uint32(resp.Port)
	if Port == 0 {
		err = fmt.Errorf("RemoteQMGetQMQMServerPort: server returned port 0 (failure)")
	}
	return
}
