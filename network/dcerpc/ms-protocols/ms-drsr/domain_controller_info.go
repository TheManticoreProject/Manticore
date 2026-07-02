package msdrsr

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// DCInfo is one domain-controller entry from IDL_DRSDomainControllerInfo (InfoLevel 2).
// NtdsDsaObjectGuid is the DC's NTDS DSA objectGUID — the value used as the source DSA
// in replication requests.
type DCInfo struct {
	NetbiosName       string
	DNSHostName       string
	SiteName          string
	NtdsDsaObjectGuid guid.GUID
	ServerObjectGuid  guid.GUID
	IsPDC             bool
	IsGC              bool
	DSEnabled         bool
}

// DomainControllerInfo calls IDL_DRSDomainControllerInfo (opnum 16) at InfoLevel 2 for
// the given domain (its FQDN, e.g. "lab.local"), returning one DCInfo per DC. It is the
// source of a DC's NtdsDsaObjectGuid, which a strict server or full-NC replication wants
// as the source DSA in IDL_DRSGetNCChanges.
func (c *Client) DomainControllerInfo(domainFQDN string) ([]DCInfo, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}
	dom := ndr.WSTR(domainFQDN)
	msgIn := drsrtypes.DRS_MSG_DCINFOREQ{
		Tag: 1,
		V1:  drsrtypes.DRS_MSG_DCINFOREQ_V1{Domain: &dom, InfoLevel: 2},
	}
	outVersion, msgOut, err := functions.IDL_DRSDomainControllerInfo(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: DomainControllerInfo: %w", err)
	}
	if uint32(outVersion) != 2 {
		return nil, fmt.Errorf("msdrsr: DomainControllerInfo: server replied version %d, expected 2", outVersion)
	}
	out := make([]DCInfo, 0, len(msgOut.V2.RItems))
	for _, it := range msgOut.V2.RItems {
		out = append(out, DCInfo{
			NetbiosName:       wstr(it.NetbiosName),
			DNSHostName:       wstr(it.DnsHostName),
			SiteName:          wstr(it.SiteName),
			NtdsDsaObjectGuid: it.NtdsDsaObjectGuid.GUID(),
			ServerObjectGuid:  it.ServerObjectGuid.GUID(),
			IsPDC:             it.FIsPdc != 0,
			IsGC:              it.FIsGc != 0,
			DSEnabled:         it.FDsEnabled != 0,
		})
	}
	return out, nil
}

// SetSourceDSA sets the source DSA objectGUID sent as uuidDsaObjDest/uuidInvocIdSrc in
// subsequent IDL_DRSGetNCChanges calls. Single-object EXOP_REPL_OBJ replication works
// with the NULL GUID (the default), but a strict server or full-NC replication wants the
// real source DSA GUID, obtained from DomainControllerInfo.
func (c *Client) SetSourceDSA(g guid.GUID) { c.sourceDSA = drsrtypes.UUIDFromGUID(g) }

// wstr dereferences an optional NDR wide string to a Go string ("" when nil).
func wstr(p *ndr.WSTR) string {
	if p == nil {
		return ""
	}
	return string(*p)
}
