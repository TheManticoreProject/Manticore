package functions

// IDL source: [MS-DFSNM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dfsnm/b471e023-618d-4c48-877f-f30c3005320c
// A fetched copy is kept at ms-dfsnm.idl in the interface directory.

import (
	"fmt"

	netdfs "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4fc742e0-4a10-11cf-8273-00aa004ae673/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// netrDfsSetDcAddressRequest carries the [in] parameters of NetrDfsSetDcAddress.
type netrDfsSetDcAddressRequest struct {
	ServerName ndr.WSTR
	DcName     ndr.WSTR
	Timeout    ndr.DWORD
	Flags      ndr.DWORD
}

func (*netrDfsSetDcAddressRequest) Opnum() uint16 { return netdfs.OpnumNetrDfsSetDcAddress }

// netrDfsSetDcAddressResponse carries the [out] parameters and return value of NetrDfsSetDcAddress.
type netrDfsSetDcAddressResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// NetrDfsSetDcAddress calls NetrDfsSetDcAddress (opnum 17) ([MS-DFSNM] — verify the parameter
// modeling and status handling).
func NetrDfsSetDcAddress(rpc ndr.Invoker, serverName ndr.WSTR, dcName ndr.WSTR, timeout ndr.DWORD, flags ndr.DWORD) (err error) {
	req := &netrDfsSetDcAddressRequest{
		ServerName: serverName,
		DcName:     dcName,
		Timeout:    timeout,
		Flags:      flags,
	}
	var resp netrDfsSetDcAddressResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("NetrDfsSetDcAddress: %w", err)
		return
	}
	if uint32(resp.Status) != netdfs.StatusSuccess {
		err = fmt.Errorf("NetrDfsSetDcAddress failed: %s", netdfs.StatusString(uint32(resp.Status)))
	}
	return
}
