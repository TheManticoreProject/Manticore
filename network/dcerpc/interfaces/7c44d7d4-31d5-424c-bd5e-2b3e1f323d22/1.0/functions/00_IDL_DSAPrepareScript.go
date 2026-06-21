package functions

import (
	"fmt"

	dsaop "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/7c44d7d4-31d5-424c-bd5e-2b3e1f323d22/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/7c44d7d4-31d5-424c-bd5e-2b3e1f323d22/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// iDL_DSAPrepareScriptRequest carries the [in] parameters of IDL_DSAPrepareScript.
type iDL_DSAPrepareScriptRequest struct {
	DwInVersion ndr.DWORD
	PmsgIn      structures.DSA_MSG_PREPARE_SCRIPT_REQ
}

func (*iDL_DSAPrepareScriptRequest) Opnum() uint16 { return dsaop.OpnumIDL_DSAPrepareScript }

// iDL_DSAPrepareScriptResponse carries the [out] parameters and return value of IDL_DSAPrepareScript.
type iDL_DSAPrepareScriptResponse struct {
	PdwOutVersion ndr.DWORD
	PmsgOut       structures.DSA_MSG_PREPARE_SCRIPT_REPLY
	Status        ndr.DWORD `ndr:"retval"`
}

// IDL_DSAPrepareScript calls IDL_DSAPrepareScript (opnum 0) ([MS-DRSR] — verify the parameter
// modeling and status handling).
func IDL_DSAPrepareScript(rpc ndr.Invoker, dwInVersion ndr.DWORD, pmsgIn structures.DSA_MSG_PREPARE_SCRIPT_REQ) (PdwOutVersion ndr.DWORD, PmsgOut structures.DSA_MSG_PREPARE_SCRIPT_REPLY, err error) {
	req := &iDL_DSAPrepareScriptRequest{
		DwInVersion: dwInVersion,
		PmsgIn:      pmsgIn,
	}
	var resp iDL_DSAPrepareScriptResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("IDL_DSAPrepareScript: %w", err)
		return
	}
	PdwOutVersion = resp.PdwOutVersion
	PmsgOut = resp.PmsgOut
	if uint32(resp.Status) != dsaop.StatusSuccess {
		err = fmt.Errorf("IDL_DSAPrepareScript failed: %s", dsaop.StatusString(uint32(resp.Status)))
	}
	return
}
