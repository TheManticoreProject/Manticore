package functions

// IDL source: [MS-SAMR] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-samr/1cd138b9-cc1b-4706-b115-49e53189e32e
// A fetched copy is kept at ms-samr.idl in the interface directory.

import (
	"fmt"

	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrGetDisplayEnumerationIndex2Request is the [in] parameter set of
// SamrGetDisplayEnumerationIndex2 (identical shape to opnum 41): a domain handle, the display
// class, and the [ref] name prefix (inline, single pointer) to search for.
type samrGetDisplayEnumerationIndex2Request struct {
	DomainHandle            mssamr.SAMPR_HANDLE
	DisplayInformationClass mssamr.DOMAIN_DISPLAY_INFORMATION `ndr:"enum"`
	Prefix                  msdtyp.RPC_UNICODE_STRING
}

func (*samrGetDisplayEnumerationIndex2Request) Opnum() uint16 {
	return samr.OpnumSamrGetDisplayEnumerationIndex2
}

// samrGetDisplayEnumerationIndex2Response is the reply: the [out] index and the NTSTATUS.
type samrGetDisplayEnumerationIndex2Response struct {
	Index  ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// SamrGetDisplayEnumerationIndex2 calls SamrGetDisplayEnumerationIndex2 (opnum 49), returning
// the index of the first account whose name is greater than or equal to the given prefix
// ([MS-SAMR] 3.1.5.3.2).
func SamrGetDisplayEnumerationIndex2(rpc ndr.Invoker, domainHandle mssamr.SAMPR_HANDLE, class uint16, prefix string) (uint32, error) {
	req := &samrGetDisplayEnumerationIndex2Request{
		DomainHandle:            domainHandle,
		DisplayInformationClass: mssamr.DOMAIN_DISPLAY_INFORMATION(class),
		Prefix:                  msdtyp.NewUnicodeString(prefix),
	}
	var resp samrGetDisplayEnumerationIndex2Response
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("SamrGetDisplayEnumerationIndex2: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusMoreEntries && status != samr.StatusNoMoreEntries {
		return uint32(resp.Index), fmt.Errorf("SamrGetDisplayEnumerationIndex2 failed: %s", samr.StatusString(status))
	}
	return uint32(resp.Index), nil
}
