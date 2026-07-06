package functions

// IDL source: [MS-CMPO] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cmpo/3a4677f0-9aef-41f9-9bca-f9c2469cefa6
// A fetched copy is kept at ms-cmpo.idl in the interface directory.

import (
	"fmt"

	IXnRemote "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/906b0ce0-c70b-1067-b317-00dd010662da/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mscmpo "github.com/TheManticoreProject/Manticore/windows/protocols/ms-cmpo"
)

// pokeWRequest carries the [in] parameters of PokeW ([MS-CMPO] 3.4.4.7), the wide-string
// (UTF-16) counterpart of Poke. The GUID/host strings are [string] wchar_t arrays.
type pokeWRequest struct {
	SRank          mscmpo.SESSION_RANK `ndr:"enum"`
	PwszCalleeUuid ndr.WSTR
	PwszHostName   ndr.WSTR
	PwszUuidString ndr.WSTR
	DwcbSizeOfBlob ndr.DWORD
	RguchBlob      []uint8 `ndr:"ref,size_is=DwcbSizeOfBlob"`
}

func (*pokeWRequest) Opnum() uint16 { return IXnRemote.OpnumPokeW }

// pokeWResponse carries the HRESULT return value of PokeW.
type pokeWResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// PokeW calls PokeW (opnum 6) ([MS-CMPO] 3.4.4.7): the wide-string form of the initial
// contact a caller makes to establish a session. blob is the marshaled BIND_INFO_BLOB.
func PokeW(rpc ndr.Invoker, sRank mscmpo.SESSION_RANK, calleeUuid, hostName, uuidString string, blob []byte) error {
	req := &pokeWRequest{
		SRank:          sRank,
		PwszCalleeUuid: ndr.WSTR(calleeUuid),
		PwszHostName:   ndr.WSTR(hostName),
		PwszUuidString: ndr.WSTR(uuidString),
		DwcbSizeOfBlob: ndr.DWORD(len(blob)),
		RguchBlob:      blob,
	}
	var resp pokeWResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("PokeW: %w", err)
	}
	if uint32(resp.Status) != IXnRemote.StatusSuccess {
		return fmt.Errorf("PokeW failed: %s", IXnRemote.StatusString(uint32(resp.Status)))
	}
	return nil
}
