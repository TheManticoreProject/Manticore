package msdrsr

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// MembershipGroup is one group returned by GetMemberships: the group's distinguished
// name and objectGUID, plus its per-entry attribute value (e.g. SE_GROUP flags).
type MembershipGroup struct {
	DN        string
	GUID      guid.GUID
	Attribute uint32
}

// GetMemberships computes the reverse group memberships of the given objects (addressed
// by binary SID — the basis reverse membership operates on) via IDL_DRSGetMemberships
// (opnum 9). opType selects which memberships to compute (e.g.
// drsrtypes.RevMembGetGroupsForUser). limitingDomainNC, if non-empty, bounds the search
// to that domain NC. It is read-only.
func (c *Client) GetMemberships(sids [][]byte, opType drsrtypes.REVERSE_MEMBERSHIP_OPERATION_TYPE, limitingDomainNC string) ([]MembershipGroup, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}
	_, reply, err := c.getMembershipsReq(c.buildRevMembReq(sids, opType, limitingDomainNC))
	if err != nil {
		return nil, err
	}
	return projectRevMembReply(reply), nil
}

// GetMemberships2 issues a batch of reverse-membership requests in one call via
// IDL_DRSGetMemberships2 (opnum 21), one request per SID set, returning the groups from
// each reply concatenated.
func (c *Client) GetMemberships2(sidSets [][][]byte, opType drsrtypes.REVERSE_MEMBERSHIP_OPERATION_TYPE, limitingDomainNC string) ([]MembershipGroup, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}
	reqs := make([]drsrtypes.DRS_MSG_REVMEMB_REQ_V1, len(sidSets))
	for i, sids := range sidSets {
		reqs[i] = c.buildRevMembReq(sids, opType, limitingDomainNC)
	}
	msgIn := drsrtypes.DRS_MSG_GETMEMBERSHIPS2_REQ{
		Tag: 1,
		V1:  drsrtypes.DRS_MSG_GETMEMBERSHIPS2_REQ_V1{Count: ndr.DWORD(len(reqs)), Requests: reqs},
	}
	_, msgOut, err := functions.IDL_DRSGetMemberships2(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: GetMemberships2: %w", err)
	}
	var out []MembershipGroup
	for _, r := range msgOut.V1.Replies {
		out = append(out, projectRevMembReply(r)...)
	}
	return out, nil
}

// buildRevMembReq builds a DRS_MSG_REVMEMB_REQ_V1 from SID-addressed names.
func (c *Client) buildRevMembReq(sids [][]byte, opType drsrtypes.REVERSE_MEMBERSHIP_OPERATION_TYPE, limitingDomainNC string) drsrtypes.DRS_MSG_REVMEMB_REQ_V1 {
	names := make([]*drsrtypes.DSNAME, len(sids))
	for i, sid := range sids {
		d := drsrtypes.NewDSNameFromSID(sid)
		names[i] = &d
	}
	return c.buildRevMembReqFromNames(names, opType, limitingDomainNC)
}

func (c *Client) buildRevMembReqFromNames(names []*drsrtypes.DSNAME, opType drsrtypes.REVERSE_MEMBERSHIP_OPERATION_TYPE, limitingDomainNC string) drsrtypes.DRS_MSG_REVMEMB_REQ_V1 {
	v1 := drsrtypes.DRS_MSG_REVMEMB_REQ_V1{
		CDsNames:      ndr.DWORD(len(names)),
		PpDsNames:     names,
		DwFlags:       drsrtypes.DRS_REVMEMB_FLAG_GET_ATTRIBUTES,
		OperationType: opType,
	}
	if limitingDomainNC != "" {
		ld := drsrtypes.NewDSNameFromDN(limitingDomainNC)
		v1.PLimitingDomain = &ld
	}
	return v1
}

// getMembershipsReq issues a single IDL_DRSGetMemberships call.
func (c *Client) getMembershipsReq(v1 drsrtypes.DRS_MSG_REVMEMB_REQ_V1) (ndr.DWORD, drsrtypes.DRS_MSG_REVMEMB_REPLY_V1, error) {
	msgIn := drsrtypes.DRS_MSG_REVMEMB_REQ{Tag: 1, V1: v1}
	outV, msgOut, err := functions.IDL_DRSGetMemberships(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return 0, drsrtypes.DRS_MSG_REVMEMB_REPLY_V1{}, fmt.Errorf("msdrsr: GetMemberships: %w", err)
	}
	return outV, msgOut.V1, nil
}

// projectRevMembReply turns a REVMEMB reply into friendly group entries.
func projectRevMembReply(r drsrtypes.DRS_MSG_REVMEMB_REPLY_V1) []MembershipGroup {
	out := make([]MembershipGroup, 0, len(r.PpDsNames))
	for i, g := range r.PpDsNames {
		if g == nil {
			continue
		}
		mg := MembershipGroup{DN: decodeWChars(g.StringName), GUID: g.Guid.GUID()}
		if i < len(r.PAttributes) {
			mg.Attribute = uint32(r.PAttributes[i])
		}
		out = append(out, mg)
	}
	return out
}
