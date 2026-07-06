package msrpce

import (
	"fmt"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// Annotation is the [string] char annotation[ept_max_annotation_size] field of an
// ept_entry_t ([C706] Appendix O, ept_max_annotation_size = 64): a human-readable label
// the endpoint mapper stores with each registration (e.g. "Messenger Service").
//
// A fixed-bound [string] array is encoded in NDR as a *varying* array — an offset and an
// actual_count (number of characters, including the NUL terminator) followed by that many
// octets, with no maximum_count. That two-count framing differs from the codec's general
// conformant-varying string handling, so Annotation marshals itself.
type Annotation string

// Compile-time assertion that Annotation encodes itself as NDR.
var _ ndr.Marshaler = (*Annotation)(nil)

// AnnotationMaxSize is ept_max_annotation_size, the fixed bound of the annotation array.
const AnnotationMaxSize = 64

// AlignmentNDR reports the 4-octet alignment of the leading offset/actual_count words.
func (*Annotation) AlignmentNDR() int { return 4 }

// MarshalNDR writes the annotation as a varying char array: offset 0, actual_count
// (string length plus the NUL terminator), then the bytes and the terminator.
func (a *Annotation) MarshalNDR(e *ndr.Encoder) error {
	e.Align(4)
	s := string(*a)
	e.WriteUint32(0)                  // offset
	e.WriteUint32(uint32(len(s) + 1)) // actual_count, including the NUL terminator
	e.WriteBytes([]byte(s))
	e.WriteUint8(0) // NUL terminator
	return nil
}

// UnmarshalNDR reads a varying char array (offset, actual_count, octets) and stores it
// with the trailing NUL removed.
func (a *Annotation) UnmarshalNDR(d *ndr.Decoder) error {
	d.Align(4)
	if _, err := d.ReadUint32(); err != nil { // offset
		return fmt.Errorf("epm: annotation offset: %w", err)
	}
	n, err := d.ReadUint32() // actual_count (characters, including any NUL terminator)
	if err != nil {
		return fmt.Errorf("epm: annotation length: %w", err)
	}
	b, err := d.ReadBytes(int(n))
	if err != nil {
		return fmt.Errorf("epm: annotation: %w", err)
	}
	*a = Annotation(strings.TrimRight(string(b), "\x00"))
	return nil
}

// EptEntry models ept_entry_t ([C706] Appendix O):
//
//	typedef struct {
//	    uuid_t  object;
//	    twr_p_t tower;
//	    [string] char annotation[ept_max_annotation_size];
//	} ept_entry_t;
//
// Object is the object UUID (often nil/zero), Tower is a [unique]/full pointer to the
// twr_t describing the binding, and Annotation is the registration's label. It is the
// element type of ept_lookup's [out] entries array; use DecodeTower to obtain the decoded
// protocol Tower (and Tower.Binding for a string binding).
type EptEntry struct {
	Object     EptUUID
	Tower      *Twr `ndr:"ptr"`
	Annotation Annotation
}

// DecodeTower decodes the entry's twr_t into a protocol Tower. It errors if the entry has
// no tower (a null pointer on the wire).
func (e EptEntry) DecodeTower() (Tower, error) {
	if e.Tower == nil {
		return Tower{}, fmt.Errorf("epm: entry has no tower")
	}
	return e.Tower.DecodeTower()
}
