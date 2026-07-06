package functions

// IDL source: [MS-NRPC] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nrpc/89f9b028-ee68-4fe2-afca-cc188f7079f7
// A fetched copy is kept at ms-nrpc.idl in the interface directory.

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// dsrDeregisterDnsHostRecordsRequest carries the [in] parameters of DsrDeregisterDnsHostRecords.
type dsrDeregisterDnsHostRecordsRequest struct {
	ServerName    *ndr.WSTR  `ndr:"unique"`
	DnsDomainName *ndr.WSTR  `ndr:"unique"`
	DomainGuid    *guid.GUID `ndr:"unique"`
	DsaGuid       *guid.GUID `ndr:"unique"`
	DnsHostName   *ndr.WSTR  `ndr:"unique"`
}

func (*dsrDeregisterDnsHostRecordsRequest) Opnum() uint16 {
	return logon.OpnumDsrDeregisterDnsHostRecords
}

// dsrDeregisterDnsHostRecordsResponse carries the [out] parameters and return value of DsrDeregisterDnsHostRecords.
type dsrDeregisterDnsHostRecordsResponse struct {
	Status ndr.DWORD `ndr:"retval"`
}

// DsrDeregisterDnsHostRecords calls DsrDeregisterDnsHostRecords (opnum 41) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrDeregisterDnsHostRecords(rpc ndr.Invoker, serverName *ndr.WSTR, dnsDomainName *ndr.WSTR, domainGuid *guid.GUID, dsaGuid *guid.GUID, dnsHostName *ndr.WSTR) (err error) {
	req := &dsrDeregisterDnsHostRecordsRequest{
		ServerName:    serverName,
		DnsDomainName: dnsDomainName,
		DomainGuid:    domainGuid,
		DsaGuid:       dsaGuid,
		DnsHostName:   dnsHostName,
	}
	var resp dsrDeregisterDnsHostRecordsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrDeregisterDnsHostRecords: %w", err)
		return
	}
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrDeregisterDnsHostRecords failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
