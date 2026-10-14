// Package qgy holds what the QGY open-list program family shares.
//
// Each program has its own program-qualified package (qgyolaus, qgygtle,
// qgyclst) that builds neutral rpc.Calls and validates their results; none
// of them imports a network package. The V5R4-era JTOpen 4.8 client calls
// these programs in library QGY; current JTOpen uses QSYS, so the library is
// a parameter of every call contract.
package qgy

import (
	"encoding/binary"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
)

// DefaultLibrary returns QGY, the V5R4-compatible library of the QGY
// programs.
func DefaultLibrary() rpc.ObjectName {
	n, err := rpc.NewObjectName("QGY")
	if err != nil {
		panic("qgy: " + err.Error()) // unreachable: constant valid name
	}
	return n
}

// LibraryOr returns lib, or DefaultLibrary when lib is the zero name.
func LibraryOr(lib rpc.ObjectName) rpc.ObjectName {
	if lib.IsZero() {
		return DefaultLibrary()
	}
	return lib
}

// Binary4 encodes an IBM BINARY(4) value (big-endian, signed).
func Binary4(v int32) []byte {
	return binary.BigEndian.AppendUint32(nil, uint32(v))
}
