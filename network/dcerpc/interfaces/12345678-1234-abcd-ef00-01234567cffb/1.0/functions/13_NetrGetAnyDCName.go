package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrGetAnyDCNameRequest carries the [in] parameters of NetrGetAnyDCName.
type netrGetAnyDCNameRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	DomainName *ndr.WSTR `ndr:"unique"`
}

func (*netrGetAnyDCNameRequest) Opnum() uint16 { return logon.OpnumNetrGetAnyDCName }

// netrGetAnyDCNameResponse carries the [out] parameters and return value of NetrGetAnyDCName.
type netrGetAnyDCNameResponse struct {
	Buffer ndr.WSTR
	Status ndr.DWORD `ndr:"retval"`
}

// NetrGetAnyDCName calls NetrGetAnyDCName (opnum 13) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func NetrGetAnyDCName(rpc ndr.Invoker, serverName *ndr.WSTR, domainName *ndr.WSTR) (Buffer ndr.WSTR, err error) {
	req := &netrGetAnyDCNameRequest{
		ServerName: serverName,
		DomainName: domainName,
	}
	var resp netrGetAnyDCNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrGetAnyDCName: %w", err)
		return
	}
	Buffer = resp.Buffer
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("NetrGetAnyDCName failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
