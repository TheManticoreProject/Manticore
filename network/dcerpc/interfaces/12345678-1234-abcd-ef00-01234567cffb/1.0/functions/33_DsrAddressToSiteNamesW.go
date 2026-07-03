package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// dsrAddressToSiteNamesWRequest carries the [in] parameters of DsrAddressToSiteNamesW.
type dsrAddressToSiteNamesWRequest struct {
	ComputerName    *ndr.WSTR `ndr:"unique"`
	EntryCount      ndr.DWORD
	SocketAddresses []msnrpc.NL_SOCKET_ADDRESS `ndr:"ref,size_is=EntryCount"`
}

func (*dsrAddressToSiteNamesWRequest) Opnum() uint16 { return logon.OpnumDsrAddressToSiteNamesW }

// dsrAddressToSiteNamesWResponse carries the [out] parameters and return value of DsrAddressToSiteNamesW.
type dsrAddressToSiteNamesWResponse struct {
	SiteNames *msnrpc.NL_SITE_NAME_ARRAY `ndr:"unique"`
	Status    ndr.DWORD                  `ndr:"retval"`
}

// DsrAddressToSiteNamesW calls DsrAddressToSiteNamesW (opnum 33) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrAddressToSiteNamesW(rpc ndr.Invoker, computerName *ndr.WSTR, entryCount ndr.DWORD, socketAddresses []msnrpc.NL_SOCKET_ADDRESS) (SiteNames *msnrpc.NL_SITE_NAME_ARRAY, err error) {
	req := &dsrAddressToSiteNamesWRequest{
		ComputerName:    computerName,
		EntryCount:      entryCount,
		SocketAddresses: socketAddresses,
	}
	var resp dsrAddressToSiteNamesWResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrAddressToSiteNamesW: %w", err)
		return
	}
	SiteNames = resp.SiteNames
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrAddressToSiteNamesW failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
