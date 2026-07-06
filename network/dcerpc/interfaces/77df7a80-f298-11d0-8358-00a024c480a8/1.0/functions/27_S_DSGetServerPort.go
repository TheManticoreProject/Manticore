package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// s_DSGetServerPortRequest carries the [in] parameters of S_DSGetServerPort.
type s_DSGetServerPortRequest struct {
	FIP ndr.DWORD
}

func (*s_DSGetServerPortRequest) Opnum() uint16 { return dscomm.OpnumS_DSGetServerPort }

// s_DSGetServerPortResponse carries the return value of S_DSGetServerPort. Unlike the
// other dscomm methods this one does not return an HRESULT: the return value is the TCP/IP
// (fIP=1) or SPX (fIP=0) port the server listens on ([MS-MQDS] 3.1.4.24).
type s_DSGetServerPortResponse struct {
	Port ndr.DWORD `ndr:"retval"`
}

// S_DSGetServerPort calls S_DSGetServerPort (opnum 27) and returns the port on which the
// server accepts connections for the requested protocol ([MS-MQDS] 3.1.4.24). fIP selects
// TCP/IP (1) or SPX (0).
func S_DSGetServerPort(rpc ndr.Invoker, fIP ndr.DWORD) (Port ndr.DWORD, err error) {
	req := &s_DSGetServerPortRequest{
		FIP: fIP,
	}
	var resp s_DSGetServerPortResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSGetServerPort: %w", err)
		return
	}
	Port = resp.Port
	return
}
