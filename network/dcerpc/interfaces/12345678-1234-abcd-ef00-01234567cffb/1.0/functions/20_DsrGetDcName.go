package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// dsrGetDcNameRequest carries the [in] parameters of DsrGetDcName.
type dsrGetDcNameRequest struct {
	ComputerName *ndr.WSTR  `ndr:"unique"`
	DomainName   *ndr.WSTR  `ndr:"unique"`
	DomainGuid   *guid.GUID `ndr:"unique"`
	SiteGuid     *guid.GUID `ndr:"unique"`
	Flags        ndr.DWORD
}

func (*dsrGetDcNameRequest) Opnum() uint16 { return logon.OpnumDsrGetDcName }

// dsrGetDcNameResponse carries the [out] parameters and return value of DsrGetDcName.
type dsrGetDcNameResponse struct {
	DomainControllerInfo *msnrpc.DOMAIN_CONTROLLER_INFOW `ndr:"unique"`
	Status               ndr.DWORD                       `ndr:"retval"`
}

// DsrGetDcName calls DsrGetDcName (opnum 20) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrGetDcName(rpc ndr.Invoker, computerName *ndr.WSTR, domainName *ndr.WSTR, domainGuid *guid.GUID, siteGuid *guid.GUID, flags ndr.DWORD) (DomainControllerInfo *msnrpc.DOMAIN_CONTROLLER_INFOW, err error) {
	req := &dsrGetDcNameRequest{
		ComputerName: computerName,
		DomainName:   domainName,
		DomainGuid:   domainGuid,
		SiteGuid:     siteGuid,
		Flags:        flags,
	}
	var resp dsrGetDcNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrGetDcName: %w", err)
		return
	}
	DomainControllerInfo = resp.DomainControllerInfo
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrGetDcName failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
