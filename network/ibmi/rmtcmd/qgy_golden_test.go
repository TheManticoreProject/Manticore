package rmtcmd_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyclst"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgygtle"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyolaus"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

func goldenFrame(t *testing.T, dump string) []byte {
	t.Helper()
	b, err := hex.DecodeString(strings.Join(strings.Fields(dump), ""))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestQGYCallsMatchGoldenFrames runs the QGY call contracts through the
// Remote Command encoder at level 8 and compares the requests with the
// JTOpen 4.8 golden frames in iSeriesNavigatorReverse docs/10 ("User-list
// open, retrieve, and close"). The golden frames use correlation 0, so the
// captured correlation is cleared before comparing.
func TestQGYCallsMatchGoldenFrames(t *testing.T) {
	var captured [][]byte
	listInfo := make([]byte, openlist.Length)
	srv := &hosttest.Server{RemoteCommandLevel: 8, Handle: func(f hostserver.Frame) (hostserver.Frame, bool) {
		b, _ := f.Marshal(0)
		copy(b[12:16], []byte{0, 0, 0, 0})
		captured = append(captured, b)
		var rest []byte
		switch {
		case bytes.Contains(f.Template, []byte{0xd8, 0xc7, 0xe8, 0xd6, 0xd3, 0xc1, 0xe4, 0xe2}): // QGYOLAUS
			rest = hosttest.OutputEnvelope(12, 80, listInfo)
		case bytes.Contains(f.Template, []byte{0xd8, 0xc7, 0xe8, 0xc7, 0xe3, 0xd3, 0xc5}): // QGYGTLE
			rest = append(hosttest.OutputEnvelope(12, 62, nil), hosttest.OutputEnvelope(12, 80, listInfo)...)
		}
		return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, rest), true
	}}
	s := session(t, srv, rmtcmd.StartOptions{})
	codec, _ := ccsid.Lookup(ccsid.CCSID37)
	handle := openlist.Handle{1, 2, 3, 4}

	calls := []func() (rpc.Call, error){
		func() (rpc.Call, error) { return qgyolaus.NewCall(qgyolaus.AUTU0150(), qgyolaus.Request{}, codec) },
		func() (rpc.Call, error) {
			return qgygtle.NewCall(qgygtle.Request{Handle: handle, RecordLength: 62, Count: 1, Start: 1})
		},
		func() (rpc.Call, error) { return qgyclst.NewCall(qgyclst.Request{Handle: handle}) },
	}
	for _, build := range calls {
		call, err := build()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.Call(context.Background(), call); err != nil {
			t.Fatalf("%v: %v", call, err)
		}
	}

	want := []string{
		`00 00 00 c1 00 00 e0 08 00 00 00 00 00 00 00 00
		00 17 10 03 d8 c7 e8 d6 d3 c1 e4 e2 40 40 d8 c7
		e8 40 40 40 40 40 40 40 00 00 09 00 00 00 0c 11
		03 00 00 00 00 00 0b 00 00 00 0c 11 03 00 00 00
		04 00 0b 00 00 00 0c 11 03 00 00 00 50 00 16 00
		00 00 10 11 03 00 00 00 04 00 0b ff ff ff ff 00
		00 00 14 11 03 00 00 00 08 00 0b c1 e4 e3 e4 f0
		f1 f5 f0 00 00 00 16 11 03 00 00 00 0a 00 0b 5c
		c1 d3 d3 40 40 40 40 40 40 00 00 00 16 11 03 00
		00 00 0a 00 0b 5c d5 d6 d5 c5 40 40 40 40 40 00
		00 00 0c 11 03 00 00 00 08 00 0b 00 00 00 16 11
		03 00 00 00 0a 00 0b 5c c1 d3 d3 40 40 40 40 40
		40`,
		`00 00 00 8f 00 00 e0 08 00 00 00 00 00 00 00 00
		00 17 10 03 d8 c7 e8 c7 e3 d3 c5 40 40 40 d8 c7
		e8 40 40 40 40 40 40 40 00 00 07 00 00 00 0c 11
		03 00 00 00 3e 00 16 00 00 00 10 11 03 00 00 00
		04 00 0b 00 00 00 3e 00 00 00 10 11 03 00 00 00
		04 00 0b 01 02 03 04 00 00 00 0c 11 03 00 00 00
		50 00 16 00 00 00 10 11 03 00 00 00 04 00 0b 00
		00 00 01 00 00 00 10 11 03 00 00 00 04 00 0b 00
		00 00 01 00 00 00 0c 11 03 00 00 00 08 00 0b`,
		`00 00 00 47 00 00 e0 08 00 00 00 00 00 00 00 00
		00 17 10 03 d8 c7 e8 c3 d3 e2 e3 40 40 40 d8 c7
		e8 40 40 40 40 40 40 40 00 00 02 00 00 00 10 11
		03 00 00 00 04 00 0b 01 02 03 04 00 00 00 0c 11
		03 00 00 00 08 00 0b`,
	}
	if len(captured) != len(want) {
		t.Fatalf("captured %d requests", len(captured))
	}
	for i, w := range want {
		if g := goldenFrame(t, w); !bytes.Equal(captured[i], g) {
			t.Errorf("request %d:\n% x\nwant\n% x", i, captured[i], g)
		}
	}
}
