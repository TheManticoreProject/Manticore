package message

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// The three records below were captured from a V7R5M0 server (2026-10-19)
// answering CHKOBJ OBJ(QSYS/NOSUCHX) OBJTYPE(*MSGQ) with message options 0
// (classic 0x1102), 3 (0x1106) and 5 (0x1107). The expected fields are the
// ones current JTOpen decodes from the same bytes.

const liveClassic = `
	0000006d1102c3d7c6f9f8f0f1f1f50028d8c3d7c6d4e2c7404040d8e2e8e2404040404040001b0029d5d6e2e4c3c8e7
	404040d8e2e8e2404040404040d4e2c7d8404040d682918583a340d5d6e2e4c3c8e740899540938982998199a840d8e2
	e8e2409596a3408696a495844b`

const liveExtended = `
	0000023411060000ffff0000ffff002800000002f1f500000007c3d7c6f9f8f0f10000000ad8c3d7c6d4e2c740404000
	00000ad8e2e8e240404040404000000029d682918583a340d5d6e2e4c3c8e740899540938982998199a840d8e2e8e240
	9596a3408696a495844b0000001bd5d6e2e4c3c8e7404040d8e2e8e2404040404040d4e2c7d8404040000001a7c381a4
	a285404b404b404b404b404b407a404040e38885409682918583a340d5d6e2e4c3c8e740899540938982998199a840d8
	e2e8e240a3a89785405cd4e2c7d8409596a3408696a495844b4040e38885409682918583a340958194856b4093898299
	8199a840958194856b40969940a38885409682918583a340a3a897854089a2409596a340839699998583a34b4040c986
	40a3888540938982998199a840958194854089a2409596a340a297858389868985846b40a38885409682918583a34094
	81a8408285408995408140938982998199a840a38881a34089a2409596a340839695a3818995858440899540a3888540
	938982998199a8409389a2a34b40d9858396a58599a840404b404b404b407a404040c39699998583a340a38885409682
	918583a340958194856b40938982998199a840958194856b409699409682918583a340a3a897854b4040c98640a38885
	40938982998199a8409581948540a681a2409596a340a297858389868985846b40a29785838986a840a3888540938982
	998199a840958194854081958440a399a840a3888540998598a485a2a34081878189954b`

const liveRcvm0300 = `
	000002e81107000000040000002800000007c3d7c6f9f8f0f100000002f1f500000004000000f20000000ad8c3d7c6d4
	e2c74040400000000a5cd3c9c2d340404040400000000ad8e2e8e24040404040400000000a4040404040404040404000
	00000a40404040404040404040000000064040404040400000000cd8d3c9c3d2d6c2d14040404000000004f0f2f5c200
	000007f1f2f6f0f8f2f500000006f1f2f4f8f0f80000000ad8e9d9c3e2d9e5e2404000000004f0f0f0f000000001f000
	000001f100000004000000010000000400000002000000095cd5d6404040404040000000040000ffff000000040000ff
	ff0000001bd5d6e2e4c3c8e7404040d8e2e8e2404040404040d4e2c7d840404000000029d682918583a340d5d6e2e4c3
	c8e740899540938982998199a840d8e2e8e2409596a3408696a495844b000001a7c381a4a285404b404b404b404b404b
	407a404040e38885409682918583a340d5d6e2e4c3c8e740899540938982998199a840d8e2e8e240a3a89785405cd4e2
	c7d8409596a3408696a495844b4040e38885409682918583a340958194856b40938982998199a840958194856b409699
	40a38885409682918583a340a3a897854089a2409596a340839699998583a34b4040c98640a3888540938982998199a8
	40958194854089a2409596a340a297858389868985846b40a38885409682918583a3409481a840828540899540814093
	8982998199a840a38881a34089a2409596a340839695a3818995858440899540a3888540938982998199a8409389a2a3
	4b40d9858396a58599a840404b404b404b407a404040c39699998583a340a38885409682918583a340958194856b4093
	8982998199a840958194856b409699409682918583a340a3a897854b4040c98640a3888540938982998199a840958194
	8540a681a2409596a340a297858389868985846b40a29785838986a840a3888540938982998199a84095819485408195
	8440a399a840a3888540998598a485a2a34081878189954b`

func liveRecord(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.Join(strings.Fields(s), ""))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func ebcdic(s string) []byte {
	// Upper-case letters, digits and blank in CCSID 37.
	var b []byte
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'I':
			b = append(b, 0xC1+byte(r-'A'))
		case r >= 'J' && r <= 'R':
			b = append(b, 0xD1+byte(r-'J'))
		case r >= 'S' && r <= 'Z':
			b = append(b, 0xE2+byte(r-'S'))
		case r >= '0' && r <= '9':
			b = append(b, 0xF0+byte(r-'0'))
		default:
			b = append(b, 0x40)
		}
	}
	return b
}

func TestParseLiveMessageFormats(t *testing.T) {
	limits := MessageLimits{MaxMessages: 4, MaxText: 4096}
	text := []byte{0xD6, 0x82, 0x91, 0x85, 0x83, 0xA3, 0x40} // "Object "
	subst := append(append(ebcdic("NOSUCHX   "), ebcdic("QSYS      ")...), ebcdic("MSGQ   ")...)
	for name, tc := range map[string]struct {
		hex      string
		tag      uint16
		hasHelp  bool
		textCCSI uint32
	}{
		"classic 0x1102":  {liveClassic, TagClassic, false, 0},
		"0x1106":          {liveExtended, TagExtended, true, 0xFFFF},
		"RCVM0300 0x1107": {liveRcvm0300, TagRCVM0300, true, 0xFFFF},
	} {
		recs, err := ParseMessages(1, liveRecord(t, tc.hex), limits)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		r := recs[0]
		if r.Tag != tc.tag || string(r.ID[:]) != string(ebcdic("CPF9801")) || r.Type != 15 || r.Severity != 40 ||
			!bytes.Equal(r.File[:], ebcdic("QCPFMSG   ")) || !bytes.Equal(r.Library[:], ebcdic("QSYS      ")) {
			t.Errorf("%s: identity = tag 0x%04X id % x type %d severity %d file % x library % x", name, r.Tag, r.ID, r.Type, r.Severity, r.File, r.Library)
		}
		if !bytes.HasPrefix(r.Text, text) || len(r.Text) != 41 || !bytes.Equal(r.Substitution, subst) {
			t.Errorf("%s: text % x (%d bytes), substitution % x", name, r.Text, len(r.Text), r.Substitution)
		}
		if tc.hasHelp != (len(r.Help) == 423) || r.TextCCSID != tc.textCCSI || r.DataCCSID != tc.textCCSI || len(r.Extension) != 0 {
			t.Errorf("%s: help %d bytes, CCSIDs %d/%d, extension %d bytes", name, len(r.Help), r.TextCCSID, r.DataCCSID, len(r.Extension))
		}
	}
}

func TestParseFieldRecordsRejectMalformed(t *testing.T) {
	limits := MessageLimits{MaxMessages: 4, MaxText: 4096}
	for _, src := range []string{liveExtended, liveRcvm0300} {
		good := liveRecord(t, src)
		cases := map[string][]byte{}
		// A field length that runs past the record.
		b := append([]byte{}, good...)
		b[len(b)-427] = 0x7F // high byte of the help length
		cases["field past end"] = b
		// The record cut short (its LL adjusted to match).
		short := append([]byte{}, good[:200]...)
		short[0], short[1], short[2], short[3] = 0, 0, 0, 200
		cases["truncated"] = short
		// Text longer than the limit.
		cases["text limit"] = good
		for name, rec := range cases {
			lim := limits
			if name == "text limit" {
				lim.MaxText = 40
			}
			if _, err := ParseMessages(1, rec, lim); !errors.Is(err, hostserver.ErrMalformed) {
				t.Errorf("tag %x %s: %v", good[4:6], name, err)
			}
		}
	}
	// A non-decimal message type.
	b := liveRecord(t, liveExtended)
	b[22] = 0xC1 // first type digit of the 0x1106 record
	if _, err := ParseMessages(1, b, limits); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("bad type: %v", err)
	}
}

func FuzzParseLiveFormats(f *testing.F) {
	for _, s := range []string{liveClassic, liveExtended, liveRcvm0300} {
		b, _ := hex.DecodeString(strings.Join(strings.Fields(s), ""))
		f.Add(b)
	}
	limits := MessageLimits{MaxMessages: 1, MaxText: 1024}
	f.Fuzz(func(t *testing.T, b []byte) {
		msgs, err := ParseMessages(1, b, limits)
		if err != nil {
			return
		}
		m := msgs[0]
		if m.Type > 99 || len(m.Text) > 1024 || len(m.Help) > 1024 || len(m.Text)+len(m.Help)+len(m.Substitution) > len(b) {
			t.Fatalf("record outside bounds: type %d, text %d, help %d", m.Type, len(m.Text), len(m.Help))
		}
	})
}
