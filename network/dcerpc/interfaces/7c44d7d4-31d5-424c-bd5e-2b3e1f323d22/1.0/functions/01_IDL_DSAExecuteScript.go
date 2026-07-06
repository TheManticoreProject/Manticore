package functions

// IDL source: [MS-DRSR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-drsr/3f5d9495-9563-44de-876a-ce6f880e3fb2
// A fetched copy is kept at ms-drsr.idl in the interface directory.

import (
	"fmt"

	dsaop "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/7c44d7d4-31d5-424c-bd5e-2b3e1f323d22/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdrsr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// iDL_DSAExecuteScriptRequest carries the [in] parameters of IDL_DSAExecuteScript.
type iDL_DSAExecuteScriptRequest struct {
	DwInVersion ndr.DWORD
	PmsgIn      msdrsr.DSA_MSG_EXECUTE_SCRIPT_REQ
}

func (*iDL_DSAExecuteScriptRequest) Opnum() uint16 { return dsaop.OpnumIDL_DSAExecuteScript }

// iDL_DSAExecuteScriptResponse carries the [out] parameters and return value of IDL_DSAExecuteScript.
type iDL_DSAExecuteScriptResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       msdrsr.DSA_MSG_EXECUTE_SCRIPT_REPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DSAExecuteScript calls IDL_DSAExecuteScript (opnum 1) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DSAExecuteScript(rpc ndr.Invoker, dwInVersion ndr.DWORD, pmsgIn msdrsr.DSA_MSG_EXECUTE_SCRIPT_REQ) (PdwOutVersion ndr.DWORD, PmsgOut msdrsr.DSA_MSG_EXECUTE_SCRIPT_REPLY, err error) {
	req := &iDL_DSAExecuteScriptRequest{
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DSAExecuteScriptResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DSAExecuteScript: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != dsaop.StatusSuccess {
		err = fmt.Errorf("IDL_DSAExecuteScript failed: %s", dsaop.StatusString(uint32(resp.Status)))
	}
	return
}
