package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// dsrGetDcNameExRequest carries the [in] parameters of DsrGetDcNameEx.
type dsrGetDcNameExRequest struct {
	ComputerName *ndr.WSTR  `ndr:"unique"`
	DomainName   *ndr.WSTR  `ndr:"unique"`
	DomainGuid   *guid.GUID `ndr:"unique"`
	SiteName     *ndr.WSTR  `ndr:"unique"`
	Flags        ndr.DWORD
}

func (*dsrGetDcNameExRequest) Opnum() uint16 { return logon.OpnumDsrGetDcNameEx }

// dsrGetDcNameExResponse carries the [out] parameters and return value of DsrGetDcNameEx.
type dsrGetDcNameExResponse struct {
	DomainControllerInfo *msnrpc.DOMAIN_CONTROLLER_INFOW `ndr:"unique"`
	Status               ndr.DWORD                       `ndr:"retval"`
}

// DsrGetDcNameEx calls DsrGetDcNameEx (opnum 27) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrGetDcNameEx(rpc ndr.Invoker, computerName *ndr.WSTR, domainName *ndr.WSTR, domainGuid *guid.GUID, siteName *ndr.WSTR, flags ndr.DWORD) (DomainControllerInfo *msnrpc.DOMAIN_CONTROLLER_INFOW, err error) {
	req := &dsrGetDcNameExRequest{
		ComputerName: computerName,
		DomainName:   domainName,
		DomainGuid:   domainGuid,
		SiteName:     siteName,
		Flags:        flags,
	}
	var resp dsrGetDcNameExResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrGetDcNameEx: %w", err)
		return
	}
	DomainControllerInfo = resp.DomainControllerInfo
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrGetDcNameEx failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
