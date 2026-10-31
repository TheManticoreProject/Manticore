package epmclient_test

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/epmclient"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
)

// TestConnectOverTCPBindNakReleasesConnection checks that a TCP endpoint mapper whose
// bind is rejected closes the socket it dialed. The server answers the bind with a
// bind_nak and then waits for the client to hang up; a leaked connection never does.
func TestConnectOverTCPBindNakReleasesConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	serverErr := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

		// Consume the bind PDU (frag_length is at offset 8 of the common header).
		hdr := make([]byte, 16)
		if _, err := io.ReadFull(conn, hdr); err != nil {
			serverErr <- err
			return
		}
		rest := make([]byte, int(binary.LittleEndian.Uint16(hdr[8:10]))-16)
		if _, err := io.ReadFull(conn, rest); err != nil {
			serverErr <- err
			return
		}

		nak := &pdu.BindNak{Header: pdu.NewHeader(pdu.PacketTypeBindNak, pdu.PFCFirstFrag|pdu.PFCLastFrag, binary.LittleEndian.Uint32(hdr[12:16]))}
		b, err := nak.Marshal()
		if err != nil {
			serverErr <- err
			return
		}
		if _, err := conn.Write(b); err != nil {
			serverErr <- err
			return
		}

		// The client must close its end once the bind fails.
		_, err = conn.Read(make([]byte, 1))
		serverErr <- err
	}()

	addr := ln.Addr().(*net.TCPAddr)
	e := epmclient.NewOverTCP("127.0.0.1", addr.Port, 5*time.Second)
	if err := e.Connect(); err == nil {
		t.Fatal("Connect succeeded against a bind_nak")
	}
	if e.IsConnected() {
		t.Fatal("IsConnected = true after a failed Connect")
	}

	if err := <-serverErr; err != io.EOF {
		t.Fatalf("server read after bind_nak = %v, want io.EOF (client leaked the connection)", err)
	}
}
