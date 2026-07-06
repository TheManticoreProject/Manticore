package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	msmqds "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqds"
)

// s_DSValidateServerRequest carries the [in] parameters of S_DSValidateServer.
type s_DSValidateServerRequest struct {
	PguidEnterpriseId   msdtyp.GUID
	FSetupMode          ndr.BOOL
	DwContext           ndr.DWORD
	DwClientBuffMaxSize ndr.DWORD
	PClientBuff         []uint8 `ndr:"ref,size_is=DwClientBuffMaxSize,varying,length_is=DwClientBuffSize"`
	DwClientBuffSize    ndr.DWORD
}

func (*s_DSValidateServerRequest) Opnum() uint16 { return dscomm.OpnumS_DSValidateServer }

// s_DSValidateServerResponse carries the [out] parameters and return value of S_DSValidateServer.
type s_DSValidateServerResponse struct {
	PphServerAuth msmqds.PPCONTEXT_HANDLE_SERVER_AUTH_TYPE
	Status        ndr.DWORD `ndr:"retval"`
}

// S_DSValidateServer calls S_DSValidateServer (opnum 22) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSValidateServer(rpc ndr.Invoker, pguidEnterpriseId msdtyp.GUID, fSetupMode ndr.BOOL, dwContext ndr.DWORD, dwClientBuffMaxSize ndr.DWORD, pClientBuff []uint8, dwClientBuffSize ndr.DWORD) (PphServerAuth msmqds.PPCONTEXT_HANDLE_SERVER_AUTH_TYPE, err error) {
	req := &s_DSValidateServerRequest{
		PguidEnterpriseId:   pguidEnterpriseId,
		FSetupMode:          fSetupMode,
		DwContext:           dwContext,
		DwClientBuffMaxSize: dwClientBuffMaxSize,
		PClientBuff:         pClientBuff,
		DwClientBuffSize:    dwClientBuffSize,
	}
	var resp s_DSValidateServerResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSValidateServer: %w", err)
		return
	}
	PphServerAuth = resp.PphServerAuth
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSValidateServer failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
