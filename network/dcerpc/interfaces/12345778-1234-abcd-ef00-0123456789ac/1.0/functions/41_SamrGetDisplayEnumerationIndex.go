package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// samrGetDisplayEnumerationIndexRequest is the [in] parameter set of
// SamrGetDisplayEnumerationIndex: a domain handle, the display class, and the [ref] name
// prefix (inline, single pointer) to search for.
type samrGetDisplayEnumerationIndexRequest struct {
	DomainHandle            structures.SAMPR_HANDLE
	DisplayInformationClass structures.DOMAIN_DISPLAY_INFORMATION
	Prefix                  dtyp.RPC_UNICODE_STRING
}

func (*samrGetDisplayEnumerationIndexRequest) Opnum() uint16 {
	return samr.OpnumSamrGetDisplayEnumerationIndex
}

// samrGetDisplayEnumerationIndexResponse is the reply: the [out] index and the NTSTATUS.
type samrGetDisplayEnumerationIndexResponse struct {
	Index  ndr.DWORD
	Status ndr.DWORD `ndr:"retval"`
}

// SamrGetDisplayEnumerationIndex calls SamrGetDisplayEnumerationIndex (opnum 41), returning
// the index of the first account whose name is greater than or equal to the given prefix
// ([MS-SAMR] 3.1.5.3.3).
func SamrGetDisplayEnumerationIndex(rpc ndr.Invoker, domainHandle structures.SAMPR_HANDLE, class uint16, prefix string) (uint32, error) {
	req := &samrGetDisplayEnumerationIndexRequest{
		DomainHandle:            domainHandle,
		DisplayInformationClass: structures.DOMAIN_DISPLAY_INFORMATION(class),
		Prefix:                  dtyp.NewUnicodeString(prefix),
	}
	var resp samrGetDisplayEnumerationIndexResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return 0, fmt.Errorf("SamrGetDisplayEnumerationIndex: %w", err)
	}
	status := uint32(resp.Status)
	if status != samr.StatusSuccess && status != samr.StatusMoreEntries && status != samr.StatusNoMoreEntries {
		return uint32(resp.Index), fmt.Errorf("SamrGetDisplayEnumerationIndex failed: %s", samr.StatusString(status))
	}
	return uint32(resp.Index), nil
}
