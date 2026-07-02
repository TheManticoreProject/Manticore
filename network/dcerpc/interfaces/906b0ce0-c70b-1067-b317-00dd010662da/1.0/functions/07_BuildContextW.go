package functions

import (
	"fmt"

	IXnRemote "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/906b0ce0-c70b-1067-b317-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmpo "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmpo"
)

// buildContextWRequest carries the [in]/[in,out] parameters of BuildContextW
// ([MS-CMPO] 3.4.4.8), the wide-string counterpart of BuildContext. pwszGuidOut and
// pBoundVersionSet are [in,out] and so also appear in the response.
type buildContextWRequest struct {
	SRank            mscmpo.SESSION_RANK `ndr:"enum"`
	BindVersionSet   mscmpo.BIND_VERSION_SET
	PwszCalleeUuid   ndr.WSTR
	PwszHostName     ndr.WSTR
	PwszUuidString   ndr.WSTR
	PwszGuidIn       ndr.WSTR
	PwszGuidOut      ndr.WSTR
	PBoundVersionSet mscmpo.BOUND_VERSION_SET
	DwcbSizeOfBlob   ndr.DWORD
	RguchBlob        []uint8 `ndr:"ref,size_is=DwcbSizeOfBlob"`
}

func (*buildContextWRequest) Opnum() uint16 { return IXnRemote.OpnumBuildContextW }

// buildContextWResponse carries the [in,out]/[out] parameters and HRESULT return value of
// BuildContextW. ppHandle is the freshly created 20-byte context handle.
type buildContextWResponse struct {
	PwszGuidOut      ndr.WSTR
	PBoundVersionSet mscmpo.BOUND_VERSION_SET
	PpHandle         mscmpo.PPCONTEXT_HANDLE
	Status           ndr.DWORD `ndr:"retval"`
}

// BuildContextW calls BuildContextW (opnum 7) ([MS-CMPO] 3.4.4.8): the wide-string form
// that negotiates a version set and establishes a bound session, returning the context
// handle the partner uses for subsequent calls. It returns the negotiated GUID-out string,
// the accepted BOUND_VERSION_SET, and the new context handle.
func BuildContextW(rpc ndr.Invoker, sRank mscmpo.SESSION_RANK, bindVersionSet mscmpo.BIND_VERSION_SET, calleeUuid, hostName, uuidString, guidIn, guidOut string, boundVersionSet mscmpo.BOUND_VERSION_SET, blob []byte) (string, mscmpo.BOUND_VERSION_SET, mscmpo.PPCONTEXT_HANDLE, error) {
	req := &buildContextWRequest{
		SRank:            sRank,
		BindVersionSet:   bindVersionSet,
		PwszCalleeUuid:   ndr.WSTR(calleeUuid),
		PwszHostName:     ndr.WSTR(hostName),
		PwszUuidString:   ndr.WSTR(uuidString),
		PwszGuidIn:       ndr.WSTR(guidIn),
		PwszGuidOut:      ndr.WSTR(guidOut),
		PBoundVersionSet: boundVersionSet,
		DwcbSizeOfBlob:   ndr.DWORD(len(blob)),
		RguchBlob:        blob,
	}
	var resp buildContextWResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return "", mscmpo.BOUND_VERSION_SET{}, mscmpo.PPCONTEXT_HANDLE{}, fmt.Errorf("BuildContextW: %w", err)
	}
	if uint32(resp.Status) != IXnRemote.StatusSuccess {
		return string(resp.PwszGuidOut), resp.PBoundVersionSet, resp.PpHandle, fmt.Errorf("BuildContextW failed: %s", IXnRemote.StatusString(uint32(resp.Status)))
	}
	return string(resp.PwszGuidOut), resp.PBoundVersionSet, resp.PpHandle, nil
}
