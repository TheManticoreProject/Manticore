package functions

// IDL source: [MS-DHCPM] — this interface is translated from and verified
// against the protocol's authoritative IDL. Full IDL (Appendix A):
//   https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-dhcpm/d1932d08-3249-44cb-90f1-8661f8fb5b90
// A fetched copy is kept at ms-dhcpm.idl in the interface directory.

import (
	"fmt"

	dhcpsrv2 "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/5b821720-f63b-11d0-aad2-00c04fc324db/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	msdhcpm "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dhcpm"
)

// r_DhcpV4FailoverEnumRelationshipRequest carries the [in] parameters of R_DhcpV4FailoverEnumRelationship.
type r_DhcpV4FailoverEnumRelationshipRequest struct {
	ServerIpAddress  *ndr.WSTR `ndr:"unique"`
	ResumeHandle     ndr.DWORD
	PreferredMaximum ndr.DWORD
}

func (*r_DhcpV4FailoverEnumRelationshipRequest) Opnum() uint16 {
	return dhcpsrv2.OpnumR_DhcpV4FailoverEnumRelationship
}

// r_DhcpV4FailoverEnumRelationshipResponse carries the [out] parameters and return value of R_DhcpV4FailoverEnumRelationship.
type r_DhcpV4FailoverEnumRelationshipResponse struct {
	ResumeHandle      ndr.DWORD
	PRelationship     *msdhcpm.DHCP_FAILOVER_RELATIONSHIP_ARRAY `ndr:"unique"`
	RelationshipRead  ndr.DWORD
	RelationshipTotal ndr.DWORD
	Status            ndr.DWORD `ndr:"retval"`
}

// R_DhcpV4FailoverEnumRelationship calls R_DhcpV4FailoverEnumRelationship (opnum 93) ([MS-DHCPM] — verify the parameter
// modeling and status handling).
func R_DhcpV4FailoverEnumRelationship(rpc ndr.Invoker, serverIpAddress *ndr.WSTR, resumeHandle ndr.DWORD, preferredMaximum ndr.DWORD) (ResumeHandle ndr.DWORD, PRelationship *msdhcpm.DHCP_FAILOVER_RELATIONSHIP_ARRAY, RelationshipRead ndr.DWORD, RelationshipTotal ndr.DWORD, err error) {
	req := &r_DhcpV4FailoverEnumRelationshipRequest{
		ServerIpAddress:  serverIpAddress,
		ResumeHandle:     resumeHandle,
		PreferredMaximum: preferredMaximum,
	}
	var resp r_DhcpV4FailoverEnumRelationshipResponse
	if err = rpc.Invoke(req, &resp); err != nil {
		err = fmt.Errorf("R_DhcpV4FailoverEnumRelationship: %w", err)
		return
	}
	ResumeHandle = resp.ResumeHandle
	PRelationship = resp.PRelationship
	RelationshipRead = resp.RelationshipRead
	RelationshipTotal = resp.RelationshipTotal
	if uint32(resp.Status) != dhcpsrv2.StatusSuccess && !dhcpsrv2.StatusIsPagination(uint32(resp.Status)) {
		err = fmt.Errorf("R_DhcpV4FailoverEnumRelationship failed: %s", dhcpsrv2.StatusString(uint32(resp.Status)))
	}
	return
}
