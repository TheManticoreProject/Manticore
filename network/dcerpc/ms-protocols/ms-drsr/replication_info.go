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
	v1 := structures.DRS_MSG_GETREPLINFO_REQ_V1{
		InfoType:             ndr.DWORD(infoType),
		UuidSourceDsaObjGuid: structures.UUIDFromGUID(sourceDSA),
	}
	// pszObjectDN is a [unique] (nullable) pointer: send NULL for an empty DN. Info types
	// that ignore it (pending ops, KCC failures) reject a non-NULL empty string with
	// ERROR_DS_DRA_INVALID_PARAMETER.
	if objectDN != "" {
		dn := ndr.WSTR(objectDN)
		v1.PszObjectDN = &dn
	}
	msgIn := structures.DRS_MSG_GETREPLINFO_REQ{Tag: 1, V1: v1}
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

// ReplNeighbor is one replication source (a repsFrom) of a naming context.
type ReplNeighbor struct {
	NamingContext         string
	SourceDsaDN           string
	SourceDsaAddress      string
	SourceDsaObjGuid      guid.GUID
	SourceDsaInvocationID guid.GUID
	ReplicaFlags          uint32
	LastSyncResult        uint32
	ConsecutiveFailures   uint32
}

// ReplicationNeighbors returns the replication neighbors (repsFrom) of a naming context
// via IDL_DRSGetReplInfo (DS_REPL_INFO_NEIGHBORS). ncDN is the NC distinguished name.
func (c *Client) ReplicationNeighbors(ncDN string) ([]ReplNeighbor, error) {
	reply, err := c.GetReplInfo(structures.DS_REPL_INFO_NEIGHBORS, ncDN, guid.GUID{})
	if err != nil {
		return nil, err
	}
	if reply.PNeighbors == nil {
		return nil, nil
	}
	out := make([]ReplNeighbor, 0, len(reply.PNeighbors.RgNeighbor))
	for _, n := range reply.PNeighbors.RgNeighbor {
		out = append(out, ReplNeighbor{
			NamingContext:         wstr(n.PszNamingContext),
			SourceDsaDN:           wstr(n.PszSourceDsaDN),
			SourceDsaAddress:      wstr(n.PszSourceDsaAddress),
			SourceDsaObjGuid:      n.UuidSourceDsaObjGuid.GUID(),
			SourceDsaInvocationID: n.UuidSourceDsaInvocationID.GUID(),
			ReplicaFlags:          uint32(n.DwReplicaFlags),
			LastSyncResult:        uint32(n.DwLastSyncResult),
			ConsecutiveFailures:   uint32(n.CNumConsecutiveSyncFailures),
		})
	}
	return out, nil
}

// ReplPendingOp is one queued replication operation on the DC.
type ReplPendingOp struct {
	SerialNumber  uint32
	Priority      uint32
	OpType        uint32
	NamingContext string
	DsaDN         string
	DsaAddress    string
}

// ReplicationPendingOps returns the DC's pending replication operations via
// IDL_DRSGetReplInfo (DS_REPL_INFO_PENDING_OPS).
func (c *Client) ReplicationPendingOps() ([]ReplPendingOp, error) {
	reply, err := c.GetReplInfo(structures.DS_REPL_INFO_PENDING_OPS, "", guid.GUID{})
	if err != nil {
		return nil, err
	}
	if reply.PPendingOps == nil {
		return nil, nil
	}
	out := make([]ReplPendingOp, 0, len(reply.PPendingOps.RgPendingOp))
	for _, o := range reply.PPendingOps.RgPendingOp {
		out = append(out, ReplPendingOp{
			SerialNumber:  uint32(o.UlSerialNumber),
			Priority:      uint32(o.UlPriority),
			OpType:        uint32(o.OpType),
			NamingContext: wstr(o.PszNamingContext),
			DsaDN:         wstr(o.PszDsaDN),
			DsaAddress:    wstr(o.PszDsaAddress),
		})
	}
	return out, nil
}

// ReplFailure is one KCC connect/link failure record.
type ReplFailure struct {
	DsaDN       string
	DsaObjGuid  guid.GUID
	NumFailures uint32
	LastResult  uint32
}

func projectFailures(f *structures.DS_REPL_KCC_DSA_FAILURESW) []ReplFailure {
	if f == nil {
		return nil
	}
	out := make([]ReplFailure, 0, len(f.RgDsaFailure))
	for _, e := range f.RgDsaFailure {
		out = append(out, ReplFailure{
			DsaDN:       wstr(e.PszDsaDN),
			DsaObjGuid:  e.UuidDsaObjGuid.GUID(),
			NumFailures: uint32(e.CNumFailures),
			LastResult:  uint32(e.DwLastResult),
		})
	}
	return out
}

// ReplicationConnectFailures returns KCC connection failures
// (DS_REPL_INFO_KCC_DSA_CONNECT_FAILURES).
func (c *Client) ReplicationConnectFailures() ([]ReplFailure, error) {
	reply, err := c.GetReplInfo(structures.DS_REPL_INFO_KCC_DSA_CONNECT_FAILURES, "", guid.GUID{})
	if err != nil {
		return nil, err
	}
	return projectFailures(reply.PConnectFailures), nil
}

// ReplicationLinkFailures returns KCC link failures (DS_REPL_INFO_KCC_DSA_LINK_FAILURES).
func (c *Client) ReplicationLinkFailures() ([]ReplFailure, error) {
	reply, err := c.GetReplInfo(structures.DS_REPL_INFO_KCC_DSA_LINK_FAILURES, "", guid.GUID{})
	if err != nil {
		return nil, err
	}
	return projectFailures(reply.PLinkFailures), nil
}
