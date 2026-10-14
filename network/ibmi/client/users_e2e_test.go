package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/api/users"
	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

// qgyWireServer answers QGYOLAUS/QGYGTLE/QGYCLST at the wire level for an
// AUTU0150 list of n profiles, logging each program called.
type qgyWireServer struct {
	n   int
	mu  sync.Mutex
	log []string
}

func (q *qgyWireServer) programs() []string {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]string(nil), q.log...)
}

func (q *qgyWireServer) handle(f hostserver.Frame) (hostserver.Frame, bool) {
	codec, _ := ccsid.Lookup(ccsid.CCSID37)
	if f.Header.RequestID != message.RequestCallProgram {
		return hosttest.OperationReply(message.ReplyRunCommand, f.Header.CorrelationID, 0, 0, nil), true
	}
	program, _ := ccsid.DecodeFixed(codec, f.Template[0:10])
	q.mu.Lock()
	q.log = append(q.log, program)
	q.mu.Unlock()
	var inputs [][]byte
	for off := 0; off < len(f.Variable); {
		p, n, _ := message.ReadParameter(f.Variable[off:])
		data := make([]byte, p.Length) // restore trailing zeros elided on the wire
		copy(data, p.Payload)
		inputs = append(inputs, data)
		off += n
	}
	info := func(returned, first int) []byte {
		b := make([]byte, 80)
		binary.BigEndian.PutUint32(b[0:], uint32(q.n))
		binary.BigEndian.PutUint32(b[4:], uint32(returned))
		copy(b[8:], []byte{0xCA, 0xFE, 0xBA, 0xBE})
		binary.BigEndian.PutUint32(b[12:], 62)
		b[16], b[30] = 0xC3, 0xF2
		binary.BigEndian.PutUint32(b[32:], uint32(returned*62))
		binary.BigEndian.PutUint32(b[36:], uint32(first))
		return b
	}
	var rest []byte
	switch program {
	case "QGYOLAUS":
		rest = hosttest.OutputEnvelope(12, 80, info(0, 0))
	case "QGYGTLE":
		count := int(binary.BigEndian.Uint32(inputs[4]))
		start := int(binary.BigEndian.Uint32(inputs[5]))
		n := min(count, q.n-start+1)
		var recs []byte
		for i := 0; i < n; i++ {
			name, _ := ccsid.EncodeFixed(codec, fmt.Sprintf("USER%02d", start+i), 10)
			desc, _ := ccsid.EncodeFixed(codec, "Synthetic profile", 50)
			recs = append(append(append(recs, name...), 0xF0, 0xF0), desc...)
		}
		rest = append(hosttest.OutputEnvelope(12, uint32(count*62), recs), hosttest.OutputEnvelope(12, 80, info(n, start))...)
	}
	return hosttest.OperationReply(message.ReplyCallProgram, f.Header.CorrelationID, 0, 0, rest), true
}

// TestUsersListEndToEnd is the Phase 7 exit gate: Connect, QGYOLAUS,
// QGYGTLE twice, QGYCLST, Close, with the whole list under one reservation.
func TestUsersListEndToEnd(t *testing.T) {
	srv := newServer()
	q := &qgyWireServer{n: 7}
	srv.Handle = q.handle
	c := newClient(t, srv)
	if err := c.Connect(context.Background(), credential(t, testPassword)); err != nil {
		t.Fatal(err)
	}
	info := c.SessionInfo()
	u, err := users.New(c, users.Options{SystemCCSID: info.SystemCCSID, JobCCSID: info.JobCCSID})
	if err != nil {
		t.Fatal(err)
	}

	// A concurrent command must not land between open and close.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 5; i++ {
			_, _ = c.RunCommand(context.Background(), "DSPJOB", rmtcmd.RunCommandOptions{})
		}
	}()
	res, err := u.List(context.Background(), users.ListOptions{PageSize: 4})
	<-done
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Users) != 7 || res.Users[0].Name != "USER01" || res.Users[6].Name != "USER07" || res.Users[6].Description != "Synthetic profile" {
		t.Fatalf("users = %+v", res.Users)
	}
	if got := fmt.Sprint(q.programs()); got != "[QGYOLAUS QGYGTLE QGYGTLE QGYCLST]" {
		t.Errorf("program calls = %s", got)
	}
	// The wire log shows no other request between the open and the close.
	var seq []uint16
	for _, r := range srv.Requests() {
		if r.RequestID == message.RequestCallProgram || r.RequestID == message.RequestRunCommand {
			seq = append(seq, r.RequestID)
		}
	}
	first, last := -1, -1
	for i, id := range seq {
		if id == message.RequestCallProgram {
			if first < 0 {
				first = i
			}
			last = i
		}
	}
	for _, id := range seq[first : last+1] {
		if id != message.RequestCallProgram {
			t.Fatalf("a command interleaved with the list: %04x", seq)
		}
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	if srv.Count(0x1004) != 1 {
		t.Error("Close did not end the server job")
	}
}
