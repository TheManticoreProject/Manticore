package functions

import (
	"fmt"

	atsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/1ff70682-0a51-30e8-076d-740be8cee98b/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mstsch "github.com/TheManticoreProject/Manticore/windows/protocols/ms-tsch"
)

// netrJobEnumRequest carries the [in] parameters of NetrJobEnum.
type netrJobEnumRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	PEnumContainer        mstsch.AT_ENUM_CONTAINER
	PreferedMaximumLength ndr.DWORD
	PResumeHandle         *ndr.DWORD `ndr:"unique"`
}

func (*netrJobEnumRequest) Opnum() uint16 { return atsvc.OpnumNetrJobEnum }

// netrJobEnumResponse carries the [out] parameters and return value of NetrJobEnum.
type netrJobEnumResponse struct {
	PEnumContainer mstsch.AT_ENUM_CONTAINER
	PTotalEntries  ndr.DWORD
	PResumeHandle  *ndr.DWORD `ndr:"unique"`
	Status         ndr.DWORD  `ndr:"retval"`
}

// NetrJobEnum calls NetrJobEnum (opnum 2) ([MS-TSCH] section 3.2.5.2.3).
func NetrJobEnum(rpc ndr.Invoker, serverName *ndr.WSTR, pEnumContainer mstsch.AT_ENUM_CONTAINER, preferedMaximumLength ndr.DWORD, pResumeHandle *ndr.DWORD) (PEnumContainer mstsch.AT_ENUM_CONTAINER, PTotalEntries ndr.DWORD, PResumeHandle *ndr.DWORD, err error) {
	req := &netrJobEnumRequest{
		ServerName:            serverName,
		PEnumContainer:        pEnumContainer,
		PreferedMaximumLength: preferedMaximumLength,
		PResumeHandle:         pResumeHandle,
	}
	var resp netrJobEnumResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrJobEnum: %w", err)
		return
	}
	PEnumContainer = resp.PEnumContainer
	PTotalEntries = resp.PTotalEntries
	PResumeHandle = resp.PResumeHandle
	// ERROR_MORE_DATA is not a failure: it signals more entries remain and the caller
	// should resume with the returned pResumeHandle ([MS-TSCH] 3.2.5.2.3).
	if s := uint32(resp.Status); s != atsvc.StatusSuccess && s != atsvc.ErrorMoreData {
		err = fmt.Errorf("NetrJobEnum failed: %s", atsvc.StatusString(uint32(resp.Status)))
	}
	return
}
