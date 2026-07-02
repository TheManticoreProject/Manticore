package functions

import (
	"fmt"

	IXnRemote "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/906b0ce0-c70b-1067-b317-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmpo "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmpo"
)

// buildContextRequest carries the [in] and [in,out] parameters of BuildContext
// ([MS-CMPO] 3.4.4.2). pszGuidOut and pBoundVersionSet are [in,out] and so also appear in
// the response.
type buildContextRequest struct {
	SRank            mscmpo.SESSION_RANK `ndr:"enum"`
	BindVersionSet   mscmpo.BIND_VERSION_SET
	PszCalleeUuid    ndr.STR
	PszHostName      ndr.STR
	PszUuidString    ndr.STR
	PszGuidIn        ndr.STR
	PszGuidOut       ndr.STR
	PBoundVersionSet mscmpo.BOUND_VERSION_SET
	DwcbSizeOfBlob   ndr.DWORD
	RguchBlob        []uint8 `ndr:"ref,size_is=DwcbSizeOfBlob"`
}

func (*buildContextRequest) Opnum() uint16 { return IXnRemote.OpnumBuildContext }

// buildContextResponse carries the [in,out]/[out] parameters and HRESULT return value of
// BuildContext. ppHandle is the freshly created 20-byte context handle.
type buildContextResponse struct {
	PszGuidOut       ndr.STR
	PBoundVersionSet mscmpo.BOUND_VERSION_SET
	PpHandle         mscmpo.PPCONTEXT_HANDLE
	Status           ndr.DWORD `ndr:"retval"`
}

// BuildContext calls BuildContext (opnum 1) ([MS-CMPO] 3.4.4.2): the ASCII form that
// negotiates a version set and establishes a bound session, returning the context handle
// the partner uses for subsequent calls. It returns the negotiated GUID-out string, the
// accepted BOUND_VERSION_SET, and the new context handle.
func BuildContext(rpc ndr.Invoker, sRank mscmpo.SESSION_RANK, bindVersionSet mscmpo.BIND_VERSION_SET, calleeUuid, hostName, uuidString, guidIn, guidOut string, boundVersionSet mscmpo.BOUND_VERSION_SET, blob []byte) (string, mscmpo.BOUND_VERSION_SET, mscmpo.PPCONTEXT_HANDLE, error) {
	req := &buildContextRequest{
		SRank:            sRank,
		BindVersionSet:   bindVersionSet,
		PszCalleeUuid:    ndr.STR(calleeUuid),
		PszHostName:      ndr.STR(hostName),
		PszUuidString:    ndr.STR(uuidString),
		PszGuidIn:        ndr.STR(guidIn),
		PszGuidOut:       ndr.STR(guidOut),
		PBoundVersionSet: boundVersionSet,
		DwcbSizeOfBlob:   ndr.DWORD(len(blob)),
		RguchBlob:        blob,
	}
	var resp buildContextResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return "", mscmpo.BOUND_VERSION_SET{}, mscmpo.PPCONTEXT_HANDLE{}, fmt.Errorf("BuildContext: %w", err)
	}
	if uint32(resp.Status) != IXnRemote.StatusSuccess {
		return string(resp.PszGuidOut), resp.PBoundVersionSet, resp.PpHandle, fmt.Errorf("BuildContext failed: %s", IXnRemote.StatusString(uint32(resp.Status)))
	}
	return string(resp.PszGuidOut), resp.PBoundVersionSet, resp.PpHandle, nil
}
