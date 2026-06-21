package msdrsr

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// CrackedName is one resolved entry from CrackNames: the per-item status
// (structures.DS_NAME_* — DS_NAME_NO_ERROR means resolved), the crossRef domain, and the
// translated name in the requested format.
type CrackedName struct {
	Status uint32
	Domain string
	Name   string
}

// CrackNames translates names from one DS_NAME_FORMAT to another via IDL_DRSCrackNames
// (opnum 12), e.g. a "DOMAIN\user" account name (DS_NT4_ACCOUNT_NAME) to the object's
// GUID (DS_UNIQUE_ID_NAME). It returns one CrackedName per input name, in order. A
// per-name failure is reported in that entry's Status, not as a Go error; the error
// return is reserved for transport/RPC failures.
func (c *Client) CrackNames(formatOffered, formatDesired uint32, names ...string) ([]CrackedName, error) {
	if !c.bound {
		return nil, fmt.Errorf("msdrsr: not connected")
	}

	rpNames := make([]*ndr.WSTR, len(names))
	for i, n := range names {
		w := ndr.WSTR(n)
		rpNames[i] = &w
	}

	msgIn := structures.DRS_MSG_CRACKREQ{
		Tag: 1,
		V1: structures.DRS_MSG_CRACKREQ_V1{
			FormatOffered: ndr.DWORD(formatOffered),
			FormatDesired: ndr.DWORD(formatDesired),
			CNames:        ndr.DWORD(len(names)),
			RpNames:       rpNames,
		},
	}

	_, msgOut, err := functions.IDL_DRSCrackNames(c.rpc, c.handle, 1, msgIn)
	if err != nil {
		return nil, fmt.Errorf("msdrsr: CrackNames: %w", err)
	}

	res := msgOut.V1.PResult
	if res == nil {
		return nil, fmt.Errorf("msdrsr: CrackNames returned no result")
	}
	out := make([]CrackedName, 0, len(res.RItems))
	for _, item := range res.RItems {
		cn := CrackedName{Status: uint32(item.Status)}
		if item.PDomain != nil {
			cn.Domain = string(*item.PDomain)
		}
		if item.PName != nil {
			cn.Name = string(*item.PName)
		}
		out = append(out, cn)
	}
	return out, nil
}

// ResolveToGUID resolves a single account name (in the given offered format, e.g.
// structures.DS_NT4_ACCOUNT_NAME for "DOMAIN\\user") to its objectGUID by cracking it to
// DS_UNIQUE_ID_NAME and parsing the "{guid}" result.
func (c *Client) ResolveToGUID(accountName string, formatOffered uint32) (guid.GUID, error) {
	results, err := c.CrackNames(formatOffered, structures.DS_UNIQUE_ID_NAME, accountName)
	if err != nil {
		return guid.GUID{}, err
	}
	if len(results) != 1 {
		return guid.GUID{}, fmt.Errorf("msdrsr: ResolveToGUID: expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Status != structures.DS_NAME_NO_ERROR {
		return guid.GUID{}, fmt.Errorf("msdrsr: ResolveToGUID %q failed: DS_NAME error %d", accountName, r.Status)
	}
	// DS_UNIQUE_ID_NAME is the objectGUID wrapped in braces, e.g. "{1a2b...-...}".
	g, err := guid.FromFormatD(strings.Trim(r.Name, "{}"))
	if err != nil {
		return guid.GUID{}, fmt.Errorf("msdrsr: ResolveToGUID: parse %q: %w", r.Name, err)
	}
	return *g, nil
}
