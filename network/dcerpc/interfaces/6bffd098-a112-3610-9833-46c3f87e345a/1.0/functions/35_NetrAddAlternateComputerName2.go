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

// netrAddAlternateComputerName2Request carries the [in] parameters of NetrAddAlternateComputerName2.
type netrAddAlternateComputerName2Request struct {
	ServerName        *ndr.WSTR                                  `ndr:"unique"`
	AlternateName     *ndr.WSTR                                  `ndr:"unique"`
	DomainAccount     *ndr.WSTR                                  `ndr:"unique"`
	EncryptedPassword *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES `ndr:"unique"`
	Reserved          ndr.DWORD
}

func (*netrAddAlternateComputerName2Request) Opnum() uint16 {
	return wkssvc.OpnumNetrAddAlternateComputerName2
}

// netrAddAlternateComputerName2Response carries the [out] parameters and return value of NetrAddAlternateComputerName2.
type netrAddAlternateComputerName2Response struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrAddAlternateComputerName2 calls NetrAddAlternateComputerName2 (opnum 35) ([MS-WKST] 3.2.4).
func NetrAddAlternateComputerName2(rpc ndr.Invoker, serverName *ndr.WSTR, alternateName *ndr.WSTR, domainAccount *ndr.WSTR, encryptedPassword *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD_AES, reserved ndr.DWORD) (err error) {
	req := &netrAddAlternateComputerName2Request{
		ServerName:        serverName,
		AlternateName:     alternateName,
		DomainAccount:     domainAccount,
		EncryptedPassword: encryptedPassword,
		Reserved:          reserved,
	}
	var resp netrAddAlternateComputerName2Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrAddAlternateComputerName2: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrAddAlternateComputerName2 failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
