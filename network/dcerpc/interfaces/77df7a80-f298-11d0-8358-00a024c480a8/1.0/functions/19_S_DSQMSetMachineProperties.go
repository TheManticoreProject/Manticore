package functions

// IDL source: [MS-MQDS] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-mqds/7907bc25-e4e6-40ef-b990-9172d1808e94
// A fetched copy is kept at ms-mqds.idl in the interface directory.

import (
	"fmt"

	dscomm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/77df7a80-f298-11d0-8358-00a024c480a8/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msmqmq "github.com/TheManticoreProject/Manticore/windows/protocols/ms-mqmq"
)

// s_DSQMSetMachinePropertiesRequest carries the [in] parameters of S_DSQMSetMachineProperties.
type s_DSQMSetMachinePropertiesRequest struct {
	PwcsPathName ndr.WSTR
	Cp           ndr.DWORD
	AProp        []ndr.DWORD          `ndr:"ref,size_is=Cp"`
	ApVar        []msmqmq.PROPVARIANT `ndr:"ref,size_is=Cp"`
	DwContext    ndr.DWORD
}

func (*s_DSQMSetMachinePropertiesRequest) Opnum() uint16 {
	return dscomm.OpnumS_DSQMSetMachineProperties
}

// s_DSQMSetMachinePropertiesResponse carries the [out] parameters and return value of S_DSQMSetMachineProperties.
type s_DSQMSetMachinePropertiesResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// S_DSQMSetMachineProperties calls S_DSQMSetMachineProperties (opnum 19) ([MS-MQDS] — verify the parameter
// modeling and status handling).
func S_DSQMSetMachineProperties(rpc ndr.Invoker, pwcsPathName ndr.WSTR, cp ndr.DWORD, aProp []ndr.DWORD, apVar []msmqmq.PROPVARIANT, dwContext ndr.DWORD) (err error) {
	req := &s_DSQMSetMachinePropertiesRequest{
		PwcsPathName: pwcsPathName,
		Cp:           cp,
		AProp:        aProp,
		ApVar:        apVar,
		DwContext:    dwContext,
	}
	var resp s_DSQMSetMachinePropertiesResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("S_DSQMSetMachineProperties: %w", err)
		return
	}
	if uint32(resp.Status) != dscomm.StatusSuccess {
		err = fmt.Errorf("S_DSQMSetMachineProperties failed: %s", dscomm.StatusString(uint32(resp.Status)))
	}
	return
}
