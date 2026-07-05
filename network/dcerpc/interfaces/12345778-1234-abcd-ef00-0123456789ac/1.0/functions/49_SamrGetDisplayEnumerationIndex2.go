package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	mssamr "github.com/TheManticoreProject/Manticore/windows/protocols/ms-samr"
)

// samrGetDisplayEnumerationIndex2Request is the [in] parameter set of
// SamrGetDisplayEnumerationIndex2 (identical shape to opnum 41): a domain handle, the display
// class, and the [ref] name prefix (inline, single pointer) to search for.
type samrGetDisplayEnumerationIndex2Request struct {
	DomainHandle            mssamr.SAMPR_HANDLE
	DisplayInformationClass mssamr.DOMAIN_DISPLAY_INFORMATION `ndr:"enum"`
	Prefix                  dtyp.RPC_UNICODE_STRING
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
		Prefix:                  dtyp.NewUnicodeString(prefix),
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
