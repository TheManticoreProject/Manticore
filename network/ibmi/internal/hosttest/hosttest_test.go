package hosttest_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
)

// The scripted servers are exercised end to end by the signon, rmtcmd and
// client tests; these tests pin the helpers they share.

func TestDialRoutesByPort(t *testing.T) {
	srv := &hosttest.Server{Profile: "USER", Password: "PW"}
	if _, err := srv.Dial(context.Background(), "tcp", "ibmi.test:23"); err == nil {
		t.Error("an unscripted port was dialed")
	}
	if _, err := srv.Dial(context.Background(), "tcp", "no-port"); err == nil {
		t.Error("an address without a port was dialed")
	}
	c, err := srv.Dial(context.Background(), "tcp", "ibmi.test:8475")
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	if err := srv.WaitClosed(time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestMessageRecordLayout(t *testing.T) {
	r := hosttest.MessageRecord("CPF9801", 15, 40, "QCPFMSG", "QSYS", []byte{1, 2}, "A")
	if len(r) != 41+2+1 || binary.BigEndian.Uint32(r) != 44 {
		t.Fatalf("record of %d bytes, length field %d", len(r), binary.BigEndian.Uint32(r))
	}
	if r[13] != 0xF1 || r[14] != 0xF5 || binary.BigEndian.Uint16(r[15:]) != 40 {
		t.Errorf("type/severity bytes % x", r[13:17])
	}
	if binary.BigEndian.Uint16(r[37:]) != 2 || binary.BigEndian.Uint16(r[39:]) != 1 || r[43] != 0xC1 {
		t.Errorf("substitution/text fields % x", r[37:])
	}
}

func TestEnvelopeAndReply(t *testing.T) {
	e := hosttest.OutputEnvelope(12, 8, []byte{9})
	if !bytes.Equal(e, []byte{0, 0, 0, 13, 0x11, 0x03, 0, 0, 0, 8, 0, 12, 9}) {
		t.Errorf("envelope = % x", e)
	}
	f := hosttest.OperationReply(0x8003, 5, 0x0500, 2, []byte{7})
	if f.Header.ServerID != hostserver.ServerIDRemoteCommand || f.Header.CorrelationID != 5 ||
		!bytes.Equal(f.Template, []byte{0x05, 0x00, 0x00, 0x02}) || !bytes.Equal(f.Variable, []byte{7}) {
		t.Errorf("reply = %+v", f)
	}
}
