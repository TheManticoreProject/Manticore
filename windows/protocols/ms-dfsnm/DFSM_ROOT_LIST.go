package msdfsnm

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// DFSM_ROOT_LIST models DFSM_ROOT_LIST ([MS-DFSNM] 2.2.3.5). Entry is an inline
// conformant array (IDL "DFSM_ROOT_LIST_ENTRY Entry[]" — embedded, not a pointer),
// so it is a plain conformant array with no referent id: its maximum_count is hoisted
// to the head of the structure.
type DFSM_ROOT_LIST struct {
	CEntries ndr.DWORD
	Entry    []DFSM_ROOT_LIST_ENTRY `ndr:"conformant,size_is=CEntries"`
}
