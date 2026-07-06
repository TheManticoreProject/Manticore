package functions

// IDL source: [MS-LSAD] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-lsad/d86ca799-b122-4fb6-bfa0-5c99dd862b11
// A fetched copy is kept at ms-lsad.idl in the interface directory.

import (
	"fmt"

	lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mslsad "github.com/TheManticoreProject/Manticore/windows/protocols/ms-lsad"
)

// lsarOpenPolicy3Request carries the [in] parameters of LsarOpenPolicy3.
type lsarOpenPolicy3Request struct {
	SystemName       *ndr.WSTR `ndr:"unique"`
	ObjectAttributes mslsad.LSAPR_OBJECT_ATTRIBUTES
	DesiredAccess    ndr.DWORD
	InVersion        ndr.DWORD
	InRevisionInfo   mslsad.LSAPR_REVISION_INFO
}

func (*lsarOpenPolicy3Request) Opnum() uint16 { return lsarpc.OpnumLsarOpenPolicy3 }

// lsarOpenPolicy3Response carries the [out] parameters and return value of LsarOpenPolicy3.
type lsarOpenPolicy3Response struct {
	OutVersion      ndr.DWORD
	OutRevisionInfo mslsad.LSAPR_REVISION_INFO
	PolicyHandle    mslsad.LSAPR_HANDLE
	Status          ndr.DWORD `ndr:"retval"`
}

// LsarOpenPolicy3 opens a policy handle and negotiates a revision level (opnum 130)
// ([MS-LSAD] 3.1.4.4.1). InRevisionInfo is a switch_is(InVersion) union transmitted
// inline; its discriminant is forced to InVersion so the marshalled arm matches the
// negotiated version. Wire modeling is not yet live-validated against Windows.
func LsarOpenPolicy3(rpc ndr.Invoker, systemName *ndr.WSTR, objectAttributes mslsad.LSAPR_OBJECT_ATTRIBUTES, desiredAccess ndr.DWORD, inVersion ndr.DWORD, inRevisionInfo mslsad.LSAPR_REVISION_INFO) (OutVersion ndr.DWORD, OutRevisionInfo mslsad.LSAPR_REVISION_INFO, PolicyHandle mslsad.LSAPR_HANDLE, err error) {
	inRevisionInfo.Tag = inVersion
	req := &lsarOpenPolicy3Request{
		SystemName:       systemName,
		ObjectAttributes: objectAttributes,
		DesiredAccess:    desiredAccess,
		InVersion:        inVersion,
		InRevisionInfo:   inRevisionInfo,
	}
	var resp lsarOpenPolicy3Response
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("LsarOpenPolicy3: %w", err)
		return
	}
	OutVersion = resp.OutVersion
	OutRevisionInfo = resp.OutRevisionInfo
	PolicyHandle = resp.PolicyHandle
	if uint32(resp.Status) != lsarpc.StatusSuccess {
		err = fmt.Errorf("LsarOpenPolicy3 failed: %s", lsarpc.StatusString(uint32(resp.Status)))
	}
	return
}
