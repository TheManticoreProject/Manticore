package msdrsr

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// GetReplInfo calls IDL_DRSGetReplInfo (opnum 19) for the given DS_REPL_INFO_TYPE and
// returns the raw reply union; the caller reads the arm matching infoType (e.g.
// reply.PCursors for DS_REPL_INFO_CURSORS_FOR_NC). objectDN is the target NC or object DN
// for the chosen info type; sourceDSA may be the zero GUID. This is a read-only recon
// call. For the common info types prefer the friendly helpers (ReplicationCursors, …).
func (c *Client) GetReplInfo(infoType uint32, objectDN string, sourceDSA guid.GUID) (*structures.DRS_MSG_GETREPLINFO_REPLY, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}
	dn := ndr.WSTR(objectDN)
	msgIn := structures.DRS_MSG_GETREPLINFO_REQ{
		Tag: 1,
		V1: structures.DRS_MSG_GETREPLINFO_REQ_V1{
			InfoType:             ndr.DWORD(infoType),
			PszObjectDN:          &dn,
			UuidSourceDsaObjGuid: structures.UUIDFromGUID(sourceDSA),
		},
	}
	_, msgOut, err := functions.IDL_DRSGetReplInfo(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: GetReplInfo(type=%d): %w", infoType, err)
	}
	return &msgOut, nil
}

// ReplCursor is one replication cursor: the up-to-dateness of a source DSA's changes as
// known to the queried DC.
type ReplCursor struct {
	SourceDSAInvocationID guid.GUID
	UpToDateUSN           int64
}

// ReplicationCursors returns the replication cursors for a naming context (the DC's
// up-to-dateness vector) via IDL_DRSGetReplInfo (DS_REPL_INFO_CURSORS_FOR_NC). ncDN is
// the NC distinguished name, e.g. "DC=lab,DC=local".
func (c *Client) ReplicationCursors(ncDN string) ([]ReplCursor, error) {
	reply, err := c.GetReplInfo(structures.DS_REPL_INFO_CURSORS_FOR_NC, ncDN, guid.GUID{})
	if err != nil {
		return nil, err
	}
	if reply.PCursors == nil {
		return nil, nil
	}
	out := make([]ReplCursor, 0, len(reply.PCursors.RgCursor))
	for _, cur := range reply.PCursors.RgCursor {
		out = append(out, ReplCursor{
			SourceDSAInvocationID: cur.UuidSourceDsaInvocationID.GUID(),
			UpToDateUSN:           cur.UsnAttributeFilter,
		})
	}
	return out, nil
}

// SiteCost is the replication cost from the source site to one target site (DwErrorCode
// is 0 on success; DwCost is the computed cost, 0xFFFFFFFF when unreachable).
type SiteCost struct {
	ToSite    string
	ErrorCode uint32
	Cost      uint32
}

// QuerySitesByCost calls IDL_DRSQuerySitesByCost (opnum 24) to compute the replication
// cost from fromSite to each of toSites (site names). It is read-only.
func (c *Client) QuerySitesByCost(fromSite string, toSites []string) ([]SiteCost, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}
	from := ndr.WSTR(fromSite)
	rg := make([]*ndr.WSTR, len(toSites))
	for i, s := range toSites {
		w := ndr.WSTR(s)
		rg[i] = &w
	}
	msgIn := structures.DRS_MSG_QUERYSITESREQ{
		Tag: 1,
		V1: structures.DRS_MSG_QUERYSITESREQ_V1{
			PwszFromSite: &from,
			CToSites:     ndr.DWORD(len(toSites)),
			RgszToSites:  rg,
		},
	}
	_, msgOut, err := functions.IDL_DRSQuerySitesByCost(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: QuerySitesByCost: %w", err)
	}
	out := make([]SiteCost, 0, len(msgOut.V1.RgCostInfo))
	for i, e := range msgOut.V1.RgCostInfo {
		sc := SiteCost{ErrorCode: uint32(e.DwErrorCode), Cost: uint32(e.DwCost)}
		if i < len(toSites) {
			sc.ToSite = toSites[i]
		}
		out = append(out, sc)
	}
	return out, nil
}
