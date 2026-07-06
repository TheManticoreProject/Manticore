package functions

// IDL source: [MS-TSTS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-tsts/c43addc7-eebc-491b-9b01-2587262675e8
// A fetched copy is kept at ms-tsts.idl in the interface directory.

import (
	"fmt"

	RCMPublic "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/bde95fdf-eee0-45de-9e12-e5a61cd0d4fe/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// rpcGetUserCertificatesRequest carries the [in] parameters of RpcGetUserCertificates.
type rpcGetUserCertificatesRequest struct {
	SessionId ndr.DWORD
}

func (*rpcGetUserCertificatesRequest) Opnum() uint16 { return RCMPublic.OpnumRpcGetUserCertificates }

// rpcGetUserCertificatesResponse carries the [out] parameters and return value of RpcGetUserCertificates.
type rpcGetUserCertificatesResponse struct {
	PcCerts  ndr.DWORD
	PpbCerts []uint8 `ndr:"unique,conformant"`
	PcbCerts ndr.DWORD
	Status   ndr.DWORD `ndr:"retval"`
}

// RpcGetUserCertificates calls RpcGetUserCertificates (opnum 10) ([MS-TSTS] — verify the parameter
// modeling and status handling).
func RpcGetUserCertificates(rpc ndr.Invoker, sessionId ndr.DWORD) (PcCerts ndr.DWORD, PpbCerts []uint8, PcbCerts ndr.DWORD, err error) {
	req := &rpcGetUserCertificatesRequest{
		SessionId: sessionId,
	}
	var resp rpcGetUserCertificatesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("RpcGetUserCertificates: %w", err)
		return
	}
	PcCerts = resp.PcCerts
	PpbCerts = resp.PpbCerts
	PcbCerts = resp.PcbCerts
	if uint32(resp.Status) != RCMPublic.StatusSuccess {
		err = fmt.Errorf("RpcGetUserCertificates failed: %s", RCMPublic.StatusString(uint32(resp.Status)))
	}
	return
}
