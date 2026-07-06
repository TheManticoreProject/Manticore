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

// pokeRequest carries the [in] parameters of Poke ([MS-CMPO] 3.4.4.1). The three GUID/
// host strings are [string] ASCII arrays (conformant+varying, null-terminated); rguchBlob
// is a conformant byte array whose count is dwcbSizeOfBlob.
type pokeRequest struct {
	SRank          mscmpo.SESSION_RANK `ndr:"enum"`
	PszCalleeUuid  ndr.STR
	PszHostName    ndr.STR
	PszUuidString  ndr.STR
	DwcbSizeOfBlob ndr.DWORD
	RguchBlob      []uint8 `ndr:"ref,size_is=DwcbSizeOfBlob"`
}

func (*pokeRequest) Opnum() uint16 { return IXnRemote.OpnumPoke }

// pokeResponse carries the HRESULT return value of Poke.
type pokeResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// Poke calls Poke (opnum 0) ([MS-CMPO] 3.4.4.1): the ASCII form of the initial contact a
// caller makes to establish a session. calleeUuid and uuidString are 37-character GUID
// strings (including the terminator); blob is the marshaled BIND_INFO_BLOB.
func Poke(rpc ndr.Invoker, sRank mscmpo.SESSION_RANK, calleeUuid, hostName, uuidString string, blob []byte) error {
	req := &pokeRequest{
		SRank:          sRank,
		PszCalleeUuid:  ndr.STR(calleeUuid),
		PszHostName:    ndr.STR(hostName),
		PszUuidString:  ndr.STR(uuidString),
		DwcbSizeOfBlob: ndr.DWORD(len(blob)),
		RguchBlob:      blob,
	}
	var resp pokeResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("Poke: %w", err)
	}
	if uint32(resp.Status) != IXnRemote.StatusSuccess {
		return fmt.Errorf("Poke failed: %s", IXnRemote.StatusString(uint32(resp.Status)))
	}
	return nil
}
