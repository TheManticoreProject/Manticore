package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// lsarQueryInformationPolicy2Request is the [in] parameter set of
// LsarQueryInformationPolicy2: an open policy handle and the information class selecting
// which union arm is returned.
type lsarQueryInformationPolicy2Request struct {
	PolicyHandle     structures.LSAPR_HANDLE
	InformationClass structures.POLICY_INFORMATION_CLASS
}

func (*lsarQueryInformationPolicy2Request) Opnum() uint16 {
	return lsarpc.OpnumLsarQueryInformationPolicy2
}

// lsarQueryInformationPolicy2Response is the reply: the [out, switch_is] union (a double
// pointer in the IDL, so a [unique] pointer on the wire) and the NTSTATUS return value.
type lsarQueryInformationPolicy2Response struct {
	PolicyInformation *structures.LSAPR_POLICY_INFORMATION `ndr:"unique"`
	Status            ndr.DWORD                            `ndr:"retval"`
}

// LsarQueryInformationPolicy2 calls LsarQueryInformationPolicy2 (opnum 46), returning the
// requested policy information class ([MS-LSAD] 3.1.4.4.3).
func LsarQueryInformationPolicy2(rpc ndr.Invoker, policyHandle structures.LSAPR_HANDLE, infoClass structures.POLICY_INFORMATION_CLASS) (*structures.LSAPR_POLICY_INFORMATION, error) {
	req := &lsarQueryInformationPolicy2Request{
		PolicyHandle:     policyHandle,
		InformationClass: infoClass,
	}
	var resp lsarQueryInformationPolicy2Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return nil, fmt.Errorf("LsarQueryInformationPolicy2: %w", err)
	}
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		return resp.PolicyInformation, fmt.Errorf("LsarQueryInformationPolicy2 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return resp.PolicyInformation, nil
}
