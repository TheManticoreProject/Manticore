package nbdgm

import (
	"fmt"
	"sync"
	"time"
)

// Reassembly bounds. The datagram service fragments a datagram whose names plus
// USER_DATA do not fit in a single UDP packet (RFC 1002 4.4.1); a receiver
// reassembles the pieces keyed by (source, DGM_ID). These constants bound the
// memory and lifetime of the reassembly state so a hostile or lossy sender
// cannot exhaust memory with partial datagrams.
const (
	// MaxUserDataSize caps the reassembled USER_DATA of a single datagram. A
	// DGM_ID whose fragments would exceed it is dropped.
	MaxUserDataSize = 65535

	// DefaultReassemblyTimeout is how long an incomplete datagram is retained
	// before its fragments are discarded.
	DefaultReassemblyTimeout = 30 * time.Second

	// DefaultMaxPending bounds the number of distinct in-progress datagrams held
	// at once. Once reached, a new datagram is only admitted after expired
	// entries are evicted.
	DefaultMaxPending = 1024
)

// reassemblyKey identifies an in-progress datagram by its source (address
// string) and DGM_ID, the pair the RFC uses to associate fragments.
type reassemblyKey struct {
	source string
	dgmID  uint16
}

// partial accumulates the fragments of one datagram. USER_DATA is written into
// buf at each fragment's PACKET_OFFSET; covered tracks which bytes have arrived
// so gaps can be detected, and total becomes known once the last (MORE clear)
// fragment is seen.
type partial struct {
	msgType  uint8
	nodeType uint8
	src      Name
	dst      Name
	haveHdr  bool // true once the FIRST fragment (carrying the names) has arrived

	buf     []byte
	covered []bool
	filled  int // count of covered bytes
	total   int // final length once known, else -1
	created time.Time
}

// Reassembler reassembles fragmented NetBIOS datagrams. It is safe for
// concurrent use; a single Reassembler is shared by a Listener across the
// goroutines that handle inbound packets.
type Reassembler struct {
	mu      sync.Mutex
	pending map[reassemblyKey]*partial
	timeout time.Duration
}

// NewReassembler returns a Reassembler using the default timeout.
func NewReassembler() *Reassembler {
	return &Reassembler{
		pending: make(map[reassemblyKey]*partial),
		timeout: DefaultReassemblyTimeout,
	}
}

// SetTimeout overrides the retention time for incomplete datagrams. A
// non-positive value restores the default.
func (r *Reassembler) SetTimeout(d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d <= 0 {
		d = DefaultReassemblyTimeout
	}
	r.timeout = d
}

// Add feeds one received DIRECT/BROADCAST datagram fragment into the
// reassembler. source identifies the sender (typically its "ip:port"). When the
// datagram is complete it returns the fully reassembled Datagram and done=true;
// otherwise it returns done=false and retains the fragment. A datagram that is
// both FIRST and not MORE is complete in a single packet and is returned
// directly without any retained state.
func (r *Reassembler) Add(source string, d *Datagram) (assembled *Datagram, done bool, err error) {
	if !isDirect(d.MsgType) {
		return nil, false, fmt.Errorf("reassembly only applies to DIRECT/BROADCAST datagrams, got MSG_TYPE 0x%02x", d.MsgType)
	}

	// Fast path: an unfragmented datagram (FIRST set, MORE clear) needs no
	// state and is returned as-is.
	if d.IsFirst() && !d.HasMore() {
		return d, true, nil
	}

	end := int(d.PacketOffset) + len(d.UserData)
	if end > MaxUserDataSize {
		return nil, false, fmt.Errorf("fragment exceeds maximum USER_DATA size (%d > %d)", end, MaxUserDataSize)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.evictExpiredLocked()

	key := reassemblyKey{source: source, dgmID: d.DgmID}
	p := r.pending[key]
	if p == nil {
		if len(r.pending) >= DefaultMaxPending {
			return nil, false, fmt.Errorf("too many pending reassemblies (%d)", len(r.pending))
		}
		p = &partial{total: -1, created: time.Now()}
		r.pending[key] = p
	}

	// The FIRST fragment carries the source and destination names and the
	// message type used for the assembled datagram.
	if d.IsFirst() {
		p.haveHdr = true
		p.msgType = d.MsgType
		p.nodeType = d.NodeType()
		p.src = d.SourceName
		p.dst = d.DestinationName
	}

	if err := p.write(int(d.PacketOffset), d.UserData); err != nil {
		delete(r.pending, key)
		return nil, false, err
	}

	// The last fragment (MORE clear) fixes the total USER_DATA length.
	if !d.HasMore() {
		p.total = end
	}

	if p.total < 0 || p.filled < p.total || !p.haveHdr {
		return nil, false, nil
	}

	// Complete: hand back a synthesised datagram carrying the whole USER_DATA.
	delete(r.pending, key)
	return &Datagram{
		MsgType:         p.msgType,
		Flags:           FlagFirst | ((p.nodeType << sntShift) & sntMask),
		DgmID:           d.DgmID,
		SourceIP:        d.SourceIP,
		SourcePort:      d.SourcePort,
		SourceName:      p.src,
		DestinationName: p.dst,
		UserData:        p.buf[:p.total],
	}, true, nil
}

// write copies a fragment's USER_DATA into the partial buffer at offset,
// growing the buffer as needed and marking the newly covered bytes. Overlapping
// re-transmissions are tolerated (previously covered bytes are not recounted).
func (p *partial) write(offset int, data []byte) error {
	end := offset + len(data)
	if end > len(p.buf) {
		grown := make([]byte, end)
		copy(grown, p.buf)
		p.buf = grown
		coveredGrown := make([]bool, end)
		copy(coveredGrown, p.covered)
		p.covered = coveredGrown
	}
	copy(p.buf[offset:end], data)
	for i := offset; i < end; i++ {
		if !p.covered[i] {
			p.covered[i] = true
			p.filled++
		}
	}
	return nil
}

// evictExpiredLocked removes reassembly state older than the timeout. The caller
// must hold r.mu.
func (r *Reassembler) evictExpiredLocked() {
	now := time.Now()
	for key, p := range r.pending {
		if now.Sub(p.created) > r.timeout {
			delete(r.pending, key)
		}
	}
}
