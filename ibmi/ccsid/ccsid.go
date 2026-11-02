// Package ccsid maps IBM i coded character set identifiers (CCSIDs) to text
// codecs.
//
// IBM i negotiates CCSIDs as numbers: Sign-on reports the system CCSID and
// Remote Command reports the job CCSID. They can be equal but are separate
// facts, and both are resolved here rather than in network/ibmi because
// receiver structures under ibmi/rpc also need them after transport decoding.
//
// The single-byte EBCDIC CCSIDs listed by Supported are available: the
// national code pages 37 and 273–871, 1047, and their euro variants 1140–1149.
// Lookup of any other value returns an error matching ErrUnsupported; it never
// falls back to a similar code page.
package ccsid

import (
	"errors"
	"fmt"
	"sort"

	"github.com/TheManticoreProject/Manticore/encoding/ebcdic"
)

// CCSID is an IBM coded character set identifier.
type CCSID uint32

const (
	// CCSID37 is EBCDIC US/Canada, code page 037.
	CCSID37 CCSID = 37
	// CCSID273 is EBCDIC Germany/Austria.
	CCSID273 CCSID = 273
	// CCSID277 is EBCDIC Denmark/Norway.
	CCSID277 CCSID = 277
	// CCSID278 is EBCDIC Finland/Sweden.
	CCSID278 CCSID = 278
	// CCSID280 is EBCDIC Italy.
	CCSID280 CCSID = 280
	// CCSID284 is EBCDIC Spain/Latin America.
	CCSID284 CCSID = 284
	// CCSID285 is EBCDIC United Kingdom.
	CCSID285 CCSID = 285
	// CCSID297 is EBCDIC France.
	CCSID297 CCSID = 297
	// CCSID500 is EBCDIC International (Belgium, Switzerland).
	CCSID500 CCSID = 500
	// CCSID871 is EBCDIC Iceland.
	CCSID871 CCSID = 871
	// CCSID1047 is EBCDIC Latin-1 for open systems.
	CCSID1047 CCSID = 1047
	// CCSID1140 is CCSID 37 with the euro sign, code page 1140.
	CCSID1140 CCSID = 1140
	// CCSID1141 to CCSID1149 are CCSIDs 273, 277, 278, 280, 284, 285, 297, 500
	// and 871 with the euro sign.
	CCSID1141 CCSID = 1141
	CCSID1142 CCSID = 1142
	CCSID1143 CCSID = 1143
	CCSID1144 CCSID = 1144
	CCSID1145 CCSID = 1145
	CCSID1146 CCSID = 1146
	CCSID1147 CCSID = 1147
	CCSID1148 CCSID = 1148
	CCSID1149 CCSID = 1149
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

// supported lists every CCSID Lookup resolves, ascending.
var supported = []CCSID{
	CCSID37, CCSID273, CCSID277, CCSID278, CCSID280, CCSID284, CCSID285, CCSID297,
	CCSID500, CCSID871, CCSID1047, CCSID1140, CCSID1141, CCSID1142, CCSID1143,
	CCSID1144, CCSID1145, CCSID1146, CCSID1147, CCSID1148, CCSID1149,
}

var registry = func() map[CCSID]Codec {
	r := make(map[CCSID]Codec, len(supported))
	for _, id := range supported {
		inner, ok := ebcdic.IBMCodePage(int(id))
		if !ok {
			panic(fmt.Sprintf("ccsid: no EBCDIC table for CCSID %d", id))
		}
		r[id] = codec{id: id, inner: inner}
	}
	return r
}()

// Lookup returns the codec registered for id.
func Lookup(id CCSID) (Codec, error) {
	if c, ok := registry[id]; ok {
		return c, nil
	}
	return nil, &UnsupportedError{CCSID: id}
}

// Supported returns the registered CCSIDs in ascending order.
func Supported() []CCSID {
	out := append([]CCSID(nil), supported...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
