package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrValidatePasswordRequest carries the [in] parameters. The handle_t binding
// handle is implicit (the RPC client) and is not marshalled; ValidationType selects
// the union arm and InputArg is the [ref, switch_is(ValidationType)] input union
// (modeled inline).
type samrValidatePasswordRequest struct {
	ValidationType mssamr.PASSWORD_POLICY_VALIDATION_TYPE `ndr:"enum"`
	InputArg       mssamr.SAM_VALIDATE_INPUT_ARG
}

func (*samrValidatePasswordRequest) Opnum() uint16 { return samr.OpnumSamrValidatePassword }

// samrValidatePasswordResponse carries the [out, switch_is(ValidationType)] double
// pointer to the output union and the NTSTATUS.
type samrValidatePasswordResponse struct {
	OutputArg *mssamr.SAM_VALIDATE_OUTPUT_ARG `ndr:"unique"`
	Status    ndr.DWORD                       `ndr:"retval"`
}

// SamrValidatePassword calls SamrValidatePassword (opnum 67), applying the account
// domain's password policy to the supplied state ([MS-SAMR] 3.1.5.13.7.1). The
// caller sets inputArg.Tag to match validationType before calling.
func SamrValidatePassword(rpc ndr.Invoker, validationType mssamr.PASSWORD_POLICY_VALIDATION_TYPE, inputArg mssamr.SAM_VALIDATE_INPUT_ARG) (*mssamr.SAM_VALIDATE_OUTPUT_ARG, error) {
	req := &samrValidatePasswordRequest{
		ValidationType: validationType,
		InputArg:       inputArg,
	}
	var resp samrValidatePasswordResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("SamrValidatePassword: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return resp.OutputArg, fmt.Errorf("SamrValidatePassword failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return resp.OutputArg, nil
}
