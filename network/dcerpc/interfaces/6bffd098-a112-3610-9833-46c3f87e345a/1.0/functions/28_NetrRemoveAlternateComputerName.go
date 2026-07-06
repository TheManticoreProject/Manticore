package functions

import (
	"fmt"

	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// netrRemoveAlternateComputerNameRequest carries the [in] parameters of NetrRemoveAlternateComputerName.
type netrRemoveAlternateComputerNameRequest struct {
	ServerName        *ndr.WSTR                              `ndr:"unique"`
	AlternateName     *ndr.WSTR                              `ndr:"unique"`
	DomainAccount     *ndr.WSTR                              `ndr:"unique"`
	EncryptedPassword *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD `ndr:"unique"`
	Reserved          ndr.DWORD
}

func (*netrRemoveAlternateComputerNameRequest) Opnum() uint16 {
	return wkssvc.OpnumNetrRemoveAlternateComputerName
}

// netrRemoveAlternateComputerNameResponse carries the [out] parameters and return value of NetrRemoveAlternateComputerName.
type netrRemoveAlternateComputerNameResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrRemoveAlternateComputerName calls NetrRemoveAlternateComputerName (opnum 28) ([MS-WKST] 3.2.4).
func NetrRemoveAlternateComputerName(rpc ndr.Invoker, serverName *ndr.WSTR, alternateName *ndr.WSTR, domainAccount *ndr.WSTR, encryptedPassword *mswkst.JOINPR_ENCRYPTED_USER_PASSWORD, reserved ndr.DWORD) (err error) {
	req := &netrRemoveAlternateComputerNameRequest{
		ServerName:        serverName,
		AlternateName:     alternateName,
		DomainAccount:     domainAccount,
		EncryptedPassword: encryptedPassword,
		Reserved:          reserved,
	}
	var resp netrRemoveAlternateComputerNameResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrRemoveAlternateComputerName: %w", err)
		return
	}
	if uint32(resp.Status) != wkssvc.StatusSuccess {
		err = fmt.Errorf("NetrRemoveAlternateComputerName failed: %s", wkssvc.StatusString(uint32(resp.Status)))
	}
	return
}
