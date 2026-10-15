package dataqueue

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/network/ibmi/dataqueue/message"
)

// Queue names a data queue by library and object name, as in
// /QSYS.LIB/<Library>.LIB/<Name>.DTAQ. Names are 1-10 characters from
// A-Z, 0-9, $, #, @, _ and ".", not starting with a digit, "_" or ".";
// lower-case letters are upper-cased. Quoted names and special values such
// as *LIBL are not supported.
type Queue struct {
	Library, Name string
}

func (q Queue) String() string { return q.Library + "/" + q.Name }

// ErrInvalidName is matched when a queue or library name is not a valid
// IBM i object name.
var ErrInvalidName = errors.New("dataqueue: invalid object name")

func checkName(what, name string) (string, error) {
	n := strings.ToUpper(name)
	if len(n) < 1 || len(n) > message.NameSize {
		return "", fmt.Errorf("%w: %s of %d characters", ErrInvalidName, what, len(name))
	}
	for i := 0; i < len(n); i++ {
		c := n[i]
		ok := c >= 'A' && c <= 'Z' || c == '$' || c == '#' || c == '@' ||
			i > 0 && (c >= '0' && c <= '9' || c == '_' || c == '.')
		if !ok {
			return "", fmt.Errorf("%w: %s character %d", ErrInvalidName, what, i+1)
		}
	}
	return n, nil
}

func (s *Session) object(q Queue) (message.Object, error) {
	var o message.Object
	name, err := checkName("queue name", q.Name)
	if err != nil {
		return o, err
	}
	lib, err := checkName("library name", q.Library)
	if err != nil {
		return o, err
	}
	nb, err := ccsid.EncodeFixed(s.codec, name, message.NameSize)
	if err != nil {
		return o, err
	}
	lb, err := ccsid.EncodeFixed(s.codec, lib, message.NameSize)
	if err != nil {
		return o, err
	}
	copy(o.Queue[:], nb)
	copy(o.Library[:], lb)
	return o, nil
}

// Kind is the order in which a queue returns its entries.
type Kind uint8

const (
	// FIFO returns the oldest entry first.
	FIFO Kind = iota + 1
	// LIFO returns the newest entry first.
	LIFO
	// Keyed selects entries by comparing their key; equal keys are FIFO.
	Keyed
)

func (k Kind) String() string {
	switch k {
	case FIFO:
		return "FIFO"
	case LIFO:
		return "LIFO"
	case Keyed:
		return "keyed"
	}
	return fmt.Sprintf("Kind(%d)", uint8(k))
}

func (k Kind) wire() (byte, bool) {
	switch k {
	case FIFO:
		return message.TypeFIFO, true
	case LIFO:
		return message.TypeLIFO, true
	case Keyed:
		return message.TypeKeyed, true
	}
	return 0, false
}

func kindFromWire(b byte) Kind { return Kind(b + 1) }

// Attributes describe a queue.
type Attributes struct {
	// MaxEntryLength is the largest entry the queue accepts, in bytes.
	MaxEntryLength int
	Kind           Kind
	// KeyLength is the key size of a keyed queue, 0 otherwise.
	KeyLength       int
	SaveSenderInfo  bool
	ForceAuxStorage bool
	Description     string
}

// Authority is the public authority given to a new queue.
type Authority uint8

const (
	// AuthorityLibCrtAut (the zero value) takes the library's CRTAUT.
	AuthorityLibCrtAut Authority = iota
	AuthorityAll
	AuthorityChange
	AuthorityExclude
	AuthorityUse
)

func (a Authority) String() string {
	switch a {
	case AuthorityLibCrtAut:
		return "*LIBCRTAUT"
	case AuthorityAll:
		return "*ALL"
	case AuthorityChange:
		return "*CHANGE"
	case AuthorityExclude:
		return "*EXCLUDE"
	case AuthorityUse:
		return "*USE"
	}
	return fmt.Sprintf("Authority(%d)", uint8(a))
}

func (a Authority) wire() (byte, bool) {
	switch a {
	case AuthorityLibCrtAut:
		return message.AuthorityLibCrtAut, true
	case AuthorityAll:
		return message.AuthorityAll, true
	case AuthorityChange:
		return message.AuthorityChange, true
	case AuthorityExclude:
		return message.AuthorityExclude, true
	case AuthorityUse:
		return message.AuthorityUse, true
	}
	return 0, false
}

// CreateOptions describe a new queue. Kind and MaxEntryLength are required;
// KeyLength is required for a keyed queue and must be 0 otherwise.
type CreateOptions struct {
	MaxEntryLength  int
	Kind            Kind
	KeyLength       int
	Authority       Authority
	SaveSenderInfo  bool
	ForceAuxStorage bool
	// Description is at most 50 bytes in the system CCSID.
	Description string
}

// Search compares an entry's key with the requested key.
type Search uint8

const (
	SearchEqual Search = iota + 1
	SearchNotEqual
	SearchLess
	SearchLessOrEqual
	SearchGreater
	SearchGreaterOrEqual
)

// String returns the operator's wire mnemonic, for example "EQ".
func (o Search) String() string {
	switch o {
	case SearchEqual:
		return "EQ"
	case SearchNotEqual:
		return "NE"
	case SearchLess:
		return "LT"
	case SearchLessOrEqual:
		return "LE"
	case SearchGreater:
		return "GT"
	case SearchGreaterOrEqual:
		return "GE"
	}
	return fmt.Sprintf("Search(%d)", uint8(o))
}

func (o Search) valid() bool { return o >= SearchEqual && o <= SearchGreaterOrEqual }

// WaitForever makes a read wait until an entry arrives or ctx is done.
const WaitForever time.Duration = -1

// waitSeconds converts a read wait: 0 returns at once, WaitForever waits
// without limit, anything else must be whole seconds.
func waitSeconds(wait time.Duration) (int32, error) {
	switch {
	case wait == WaitForever:
		return -1, nil
	case wait < 0:
		return 0, fmt.Errorf("dataqueue: negative wait %s (use WaitForever)", wait)
	case wait%time.Second != 0:
		return 0, fmt.Errorf("dataqueue: wait %s is not whole seconds", wait)
	case wait/time.Second > math.MaxInt32:
		return 0, fmt.Errorf("dataqueue: wait %s too long", wait)
	}
	return int32(wait / time.Second), nil
}

// Entry is one entry returned by a read or peek. Data and Key are copies.
type Entry struct {
	Data []byte
	// Key is the entry's key on a keyed queue, nil otherwise.
	Key []byte
	// Sender identifies the job that wrote the entry, on a queue that
	// saves sender information; nil otherwise.
	Sender *Sender
}

// Sender is the sender information of an entry: the qualified job that
// wrote it and the profile it ran under. Its layout (job name, job user,
// job number, current user; 10+10+6+10 bytes in the system CCSID) was
// confirmed on V5R4.
type Sender struct {
	JobName, JobUser, JobNumber string
	CurrentUser                 string
}

// Job returns the qualified job name, for example 019871/QUSER/QZHQSSRV.
func (s Sender) Job() string { return s.JobNumber + "/" + s.JobUser + "/" + s.JobName }

func decodeSender(c ccsid.Codec, b [36]byte) (*Sender, error) {
	var fields [4]string
	for i, r := range [4][2]int{{0, 10}, {10, 20}, {20, 26}, {26, 36}} {
		v, err := ccsid.DecodeFixed(c, b[r[0]:r[1]])
		if err != nil {
			return nil, err
		}
		fields[i] = v
	}
	return &Sender{JobName: fields[0], JobUser: fields[1], JobNumber: fields[2], CurrentUser: fields[3]}, nil
}
