package functions

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarOpenPolicyWithCredsRequest carries the [in] parameters of LsarOpenPolicyWithCreds.
type lsarOpenPolicyWithCredsRequest struct {
	ObjectAttributes mslsad.LSAPR_OBJECT_ATTRIBUTES
	DesiredAccess    ndr.DWORD
	InVersion        ndr.DWORD
	InRevisionInfo   mslsad.LSAPR_REVISION_INFO
}

func (*lsarOpenPolicyWithCredsRequest) Opnum() uint16 { return lsarpc.OpnumLsarOpenPolicyWithCreds }

// lsarOpenPolicyWithCredsResponse carries the [out] parameters and return value of LsarOpenPolicyWithCreds.
type lsarOpenPolicyWithCredsResponse struct {
	OutVersion      ndr.DWORD
	OutRevisionInfo mslsad.LSAPR_REVISION_INFO
	PolicyHandle    mslsad.LSAPR_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// LsarOpenPolicyWithCreds opens a policy handle under supplied credentials (opnum 135)
// ([MS-LSAD] 3.1.4.4.3). The [in] handle_t binding handle is not marshalled. As with
// LsarOpenPolicy3, InRevisionInfo is a switch_is(InVersion) union transmitted inline; its
// discriminant is forced to InVersion. Wire modeling is not yet live-validated.
func LsarOpenPolicyWithCreds(rpc ndr.Invoker, objectAttributes mslsad.LSAPR_OBJECT_ATTRIBUTES, desiredAccess ndr.DWORD, inVersion ndr.DWORD, inRevisionInfo mslsad.LSAPR_REVISION_INFO) (OutVersion ndr.DWORD, OutRevisionInfo mslsad.LSAPR_REVISION_INFO, PolicyHandle mslsad.LSAPR_HANDLE, err error) {
	inRevisionInfo.Tag = inVersion
	req := &lsarOpenPolicyWithCredsRequest{
		ObjectAttributes: objectAttributes,
		DesiredAccess:    desiredAccess,
		InVersion:        inVersion,
		InRevisionInfo:   inRevisionInfo,
	}
	var resp lsarOpenPolicyWithCredsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarOpenPolicyWithCreds: %w", err)
		return
	}
	OutVersion = resp.OutVersion
	OutRevisionInfo = resp.OutRevisionInfo
	PolicyHandle = resp.PolicyHandle
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarOpenPolicyWithCreds failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
