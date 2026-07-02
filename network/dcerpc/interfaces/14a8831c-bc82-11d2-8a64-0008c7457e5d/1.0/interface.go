// Package rpcinterface_14a8831cbc8211d28a640008c7457e5d_1_0 is the descriptor for the
// ExtendedError RPC interface, abstract syntax 14a8831c-bc82-11d2-8a64-0008c7457e5d
// version 1.0 ([MS-EERR]).
//
// [MS-EERR] is a data-structure-only specification: the ExtendedError interface defines
// no on-the-wire methods (opnums). Its NDR types (ExtendedErrorInfo and friends, in the
// windows/protocols/ms-eerr package) are not carried by a dedicated RPC endpoint; they
// are marshalled with the [C706] Type Serialization ("pickling") rules and embedded in
// the response data of OTHER protocols to convey extended error information. There is
// therefore no named pipe and no functions subpackage for this interface — this
// descriptor exists only to record the abstract syntax identifier so callers that pickle
// or interpret these structures can name the interface unambiguously.
package rpcinterface_14a8831cbc8211d28a640008c7457e5d_1_0

import (
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is empty: the ExtendedError interface defines no on-the-wire methods and is
// not reached over a named pipe. Its structures are pickled ([C706] type serialization)
// and embedded in other protocols' responses (see the package doc).
const PipeName = ``

// SyntaxID returns the ExtendedError abstract syntax identifier:
// 14a8831c-bc82-11d2-8a64-0008c7457e5d, version 1.0.
func SyntaxID() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0x14a8831c, B: 0xbc82, C: 0x11d2, D: 0x8a64, E: 0x0008c7457e5d},
		MajorVersion: 1,
		MinorVersion: 0,
	}
}

// OpnumToName maps each on-the-wire opnum to its method name. [MS-EERR] defines no
// methods, so this map is empty; it is retained for symmetry with other interface
// descriptors and as the single source of truth for the (empty) opnum set.
var OpnumToName = map[uint16]string{}

// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.
var NameToOpnum = func() map[string]uint16 {
	m := make(map[string]uint16, len(OpnumToName))
	for op, name := range OpnumToName {
		m[name] = op
	}
	return m
}()
