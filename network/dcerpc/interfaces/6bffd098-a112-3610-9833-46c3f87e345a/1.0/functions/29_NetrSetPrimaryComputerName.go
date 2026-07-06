package functions

// IDL source: [MS-WKST] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-wkst/9fdbc753-0397-4236-bbfc-a380f9d23789
// A fetched copy is kept at ms-wkst.idl in the interface directory.

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrSetPrimaryComputerNameRequest carries the [in] parameters of NetrSetPrimaryComputerName.
type netrSetPrimaryComputerNameRequest struct {
	ServerName        *ndr.WSTR                              `ndr:"unique"`
	PrimaryName       *ndr.WSTR                              `ndr:"unique"`
	DomainAccount     *ndr.WSTR                              `ndr:"unique"`
	EncryptedPassword *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	Reserved          ndr.DWORD
}

func (*netrSetPrimaryComputerNameRequest) Opnum() uint16 {
	return wkssvc.OpnumNetrSetPrimaryComputerName
}

// netrSetPrimaryComputerNameResponse carries the [out] parameters and return value of NetrSetPrimaryComputerName.
type netrSetPrimaryComputerNameResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrSetPrimaryComputerName calls NetrSetPrimaryComputerName (opnum 29) ([MS-WKST] 3.2.4).
func NetrSetPrimaryComputerName(rpc ndr.Invoker, serverName *ndr.WSTR, primaryName *ndr.WSTR, domainAccount *ndr.WSTR, encryptedPassword *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD, reserved ndr.DWORD) (err error) {
	req := &netrSetPrimaryComputerNameRequest{
		ServerName:        serverName,
		PrimaryName:       primaryName,
		DomainAccount:     domainAccount,
		EncryptedPassword: encryptedPassword,
		Reserved:          reserved,
	}
	var resp netrSetPrimaryComputerNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrSetPrimaryComputerName: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrSetPrimaryComputerName failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
