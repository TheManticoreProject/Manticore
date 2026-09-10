// Package wirediff compares the SMB messages this implementation emits against
// messages recorded from a real exchange.
//
// The problem it solves is structural. This repository contains both an SMB
// client and an SMB server, and pairing them proves very little: a wire detail
// both halves get wrong agrees with itself, and every round trip passes. That
// blind spot has produced real defects — names carrying their terminator, a
// Unicode path a byte out of phase, a credit charge no server would accept —
// each found only when a third party was put on the other end.
//
// A corpus records one exchange with a real peer: the messages in order, with
// the direction each travelled. Two things can then be checked offline, in CI,
// with no network and no lab:
//
//   - what this implementation emits, byte for byte, against what it emitted
//     when the exchange was recorded, ignoring the fields that legitimately
//     differ between runs; and
//   - that its parsers accept every message the real peer sent.
//
// The second is the stronger of the two, because those bytes were produced by
// Windows rather than by this code.
package wirediff

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Direction is which end of the exchange sent a message.
const (
	FromClient = "client"
	FromServer = "server"
)

// Frame is one SMB message and the direction it travelled. The Direct TCP
// framing is not part of it: a corpus holds messages, so a change in how they
// are split across TCP segments does not alter it.
type Frame struct {
	Direction string `json:"direction"`
	Message   string `json:"message"`
}

// Bytes decodes the recorded message.
func (f Frame) Bytes() ([]byte, error) {
	raw, err := hex.DecodeString(f.Message)
	if err != nil {
		return nil, fmt.Errorf("frame message is not hex: %w", err)
	}
	return raw, nil
}

// Corpus is one recorded exchange.
type Corpus struct {
	Name   string  `json:"name"`
	Peer   string  `json:"peer"`
	Frames []Frame `json:"frames"`
}

// Load reads a corpus from a JSON file.
func Load(path string) (*Corpus, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	corpus := &Corpus{}
	if err := json.Unmarshal(raw, corpus); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(corpus.Frames) == 0 {
		return nil, fmt.Errorf("%s: corpus records no frames", path)
	}
	return corpus, nil
}

// Messages returns the recorded messages travelling in one direction, in order.
func (c *Corpus) Messages(direction string) ([][]byte, error) {
	out := [][]byte{}
	for i, frame := range c.Frames {
		if frame.Direction != direction {
			continue
		}
		raw, err := frame.Bytes()
		if err != nil {
			return nil, fmt.Errorf("%s frame %d: %w", c.Name, i, err)
		}
		out = append(out, raw)
	}
	return out, nil
}

// Span is a run of bytes that may legitimately differ between two recordings of
// the same exchange, and the reason it may.
type Span struct {
	Name  string
	Start int
	Len   int
}

// Mask is the set of spans a comparison ignores.
//
// Every entry is a deliberate statement that a field is allowed to vary, which
// is why they carry names: an unexplained mask is indistinguishable from a bug
// being hidden.
type Mask []Span

// Covers reports whether offset falls in a masked span, and which.
func (m Mask) Covers(offset int) (string, bool) {
	for _, span := range m {
		if offset >= span.Start && offset < span.Start+span.Len {
			return span.Name, true
		}
	}
	return "", false
}

// Difference is one byte that differs outside the mask.
type Difference struct {
	Offset int
	Got    byte
	Want   byte
}

// Compare returns the differences between got and want that the mask does not
// excuse. A length mismatch is reported as a difference at the first offset
// beyond the shorter message, since the bytes after it cannot be aligned.
func Compare(got, want []byte, mask Mask) []Difference {
	diffs := []Difference{}

	shorter := len(got)
	if len(want) < shorter {
		shorter = len(want)
	}
	for offset := 0; offset < shorter; offset++ {
		if got[offset] == want[offset] {
			continue
		}
		if _, masked := mask.Covers(offset); masked {
			continue
		}
		diffs = append(diffs, Difference{Offset: offset, Got: got[offset], Want: want[offset]})
	}

	// Trailing bytes present in only one of the two.
	for offset := shorter; offset < len(got); offset++ {
		if _, masked := mask.Covers(offset); masked {
			continue
		}
		diffs = append(diffs, Difference{Offset: offset, Got: got[offset], Want: 0})
	}
	for offset := shorter; offset < len(want); offset++ {
		if _, masked := mask.Covers(offset); masked {
			continue
		}
		diffs = append(diffs, Difference{Offset: offset, Got: 0, Want: want[offset]})
	}

	sort.Slice(diffs, func(i, j int) bool { return diffs[i].Offset < diffs[j].Offset })
	return diffs
}

// Report renders differences for a test failure, with the surrounding bytes so
// the field can be recognised without decoding the message by hand.
func Report(got, want []byte, diffs []Difference) string {
	if len(diffs) == 0 {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%d byte(s) differ (emitted %d bytes, recorded %d bytes)\n", len(diffs), len(got), len(want))

	shown := diffs
	const maxShown = 16
	truncated := false
	if len(shown) > maxShown {
		shown, truncated = shown[:maxShown], true
	}
	for _, d := range shown {
		fmt.Fprintf(&b, "  offset %-5d got %02x want %02x\n", d.Offset, d.Got, d.Want)
	}
	if truncated {
		fmt.Fprintf(&b, "  ... and %d more\n", len(diffs)-maxShown)
	}
	return b.String()
}
