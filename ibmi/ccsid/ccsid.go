// Package ccsid maps IBM i coded character set identifiers (CCSIDs) to text
// codecs.
//
// IBM i negotiates CCSIDs as numbers: Sign-on reports the system CCSID and
// Remote Command reports the job CCSID. They can be equal but are separate
// facts, and both are resolved here rather than in network/ibmi because
// receiver structures under ibmi/rpc also need them after transport decoding.
//
// Only CCSID 37 and 1140 are supported in the first milestone. Lookup of any
// other value returns an error matching ErrUnsupported; it never falls back
// to a similar code page.
package ccsid

import (
	"errors"
	"fmt"

	"github.com/TheManticoreProject/Manticore/encoding/ebcdic"
)

// CCSID is an IBM coded character set identifier.
type CCSID uint32

const (
	// CCSID37 is EBCDIC US/Canada, code page 037.
	CCSID37 CCSID = 37
	// CCSID1140 is CCSID 37 with the euro sign, code page 1140.
	CCSID1140 CCSID = 1140
)

// Codec converts between Unicode strings and one CCSID. Implementations never
// modify their input and reject text they cannot represent.
type Codec interface {
	CCSID() CCSID
	Encode(string) ([]byte, error)
	Decode([]byte) (string, error)
}

// ErrUnsupported is matched (via errors.Is) by Lookup failures.
var ErrUnsupported = errors.New("ccsid: unsupported CCSID")

// UnsupportedError reports a CCSID with no registered codec.
type UnsupportedError struct {
	CCSID CCSID
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("ccsid: CCSID %d is not supported", uint32(e.CCSID))
}

// Is makes errors.Is(err, ErrUnsupported) true.
func (e *UnsupportedError) Is(target error) bool { return target == ErrUnsupported }

type codec struct {
	id    CCSID
	inner *ebcdic.Codec
}

func (c codec) CCSID() CCSID                    { return c.id }
func (c codec) Encode(s string) ([]byte, error) { return c.inner.Encode(s) }
func (c codec) Decode(b []byte) (string, error) { return c.inner.Decode(b) }

var registry = map[CCSID]Codec{
	CCSID37:   codec{id: CCSID37, inner: ebcdic.CP037()},
	CCSID1140: codec{id: CCSID1140, inner: ebcdic.CP1140()},
}

// Lookup returns the codec registered for id.
func Lookup(id CCSID) (Codec, error) {
	if c, ok := registry[id]; ok {
		return c, nil
	}
	return nil, &UnsupportedError{CCSID: id}
}

// Supported returns the registered CCSIDs in ascending order.
func Supported() []CCSID {
	return []CCSID{CCSID37, CCSID1140}
}
