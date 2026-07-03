package functions

import (
	"fmt"

	logon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
)

// dsrUpdateReadOnlyServerDnsRecordsRequest carries the [in] parameters of DsrUpdateReadOnlyServerDnsRecords.
type dsrUpdateReadOnlyServerDnsRecordsRequest struct {
	ServerName    *ndr.WSTR `ndr:"unique"`
	ComputerName  ndr.WSTR
	Authenticator msnrpc.NETLOGON_AUTHENTICATOR
	SiteName      *ndr.WSTR `ndr:"unique"`
	DnsTtl        ndr.DWORD
	DnsNames      msnrpc.NL_DNS_NAME_INFO_ARRAY
}

func (*dsrUpdateReadOnlyServerDnsRecordsRequest) Opnum() uint16 {
	return logon.OpnumDsrUpdateReadOnlyServerDnsRecords
}

// dsrUpdateReadOnlyServerDnsRecordsResponse carries the [out] parameters and return value of DsrUpdateReadOnlyServerDnsRecords.
type dsrUpdateReadOnlyServerDnsRecordsResponse struct {
	ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR
	DnsNames            msnrpc.NL_DNS_NAME_INFO_ARRAY
	Status              ndr.DWORD `ndr:"retval"`
}

// DsrUpdateReadOnlyServerDnsRecords calls DsrUpdateReadOnlyServerDnsRecords (opnum 48) ([MS-NRPC] — verify the parameter
// modeling and status handling).
func DsrUpdateReadOnlyServerDnsRecords(rpc ndr.Invoker, serverName *ndr.WSTR, computerName ndr.WSTR, authenticator msnrpc.NETLOGON_AUTHENTICATOR, siteName *ndr.WSTR, dnsTtl ndr.DWORD, dnsNames msnrpc.NL_DNS_NAME_INFO_ARRAY) (ReturnAuthenticator msnrpc.NETLOGON_AUTHENTICATOR, DnsNames msnrpc.NL_DNS_NAME_INFO_ARRAY, err error) {
	req := &dsrUpdateReadOnlyServerDnsRecordsRequest{
		ServerName:    serverName,
		ComputerName:  computerName,
		Authenticator: authenticator,
		SiteName:      siteName,
		DnsTtl:        dnsTtl,
		DnsNames:      dnsNames,
	}
	var resp dsrUpdateReadOnlyServerDnsRecordsResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("DsrUpdateReadOnlyServerDnsRecords: %w", err)
		return
	}
	ReturnAuthenticator = resp.ReturnAuthenticator
	DnsNames = resp.DnsNames
	if uint32(resp.Status) != logon.StatusSuccess {
		err = fmt.Errorf("DsrUpdateReadOnlyServerDnsRecords failed: %s", logon.StatusString(uint32(resp.Status)))
	}
	return
}
