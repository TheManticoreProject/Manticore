//go:build integration

package tcp_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/tcp"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// eptSyntax is the endpoint mapper (ept) abstract syntax,
// e1af8308-5d1f-11c9-91a4-08002b14a0fa version 3.0. The endpoint mapper always listens
// on TCP port 135, so it is a stable target for proving the ncacn_ip_tcp transport.
func eptSyntax() syntax.SyntaxID {
	return syntax.SyntaxID{
		UUID:         guid.GUID{A: 0xe1af8308, B: 0x5d1f, C: 0x11c9, D: 0x91a4, E: 0x08002b14a0fa},
		MajorVersion: 3,
		MinorVersion: 0,
	}
}

// TestTCPTransport_BindEndpointMapper binds the ept interface over ncacn_ip_tcp against
// a live host. It is skipped unless DCERPC_TCP_TEST_HOST is set; the port defaults to
// 135 and can be overridden with DCERPC_TCP_TEST_PORT.
//
//	DCERPC_TCP_TEST_HOST=10.0.0.30 go test -tags integration \
//	    ./network/dcerpc/v5/transport/tcp/
func TestTCPTransport_BindEndpointMapper(t *testing.T) {
	host := os.Getenv("DCERPC_TCP_TEST_HOST")
	if host == "" {
		t.Skip("set DCERPC_TCP_TEST_HOST to run the ncacn_ip_tcp integration test")
	}
	port := tcp.EndpointMapperPort
	if p := os.Getenv("DCERPC_TCP_TEST_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			t.Fatalf("invalid DCERPC_TCP_TEST_PORT %q: %v", p, err)
		}
		port = n
	}

	tr := tcp.New(host, port)
	tr.SetTimeout(10 * time.Second)

	c := client.NewClient(tr)
	defer c.Close()

	if err := c.Bind(eptSyntax()); err != nil {
		t.Fatalf("Bind(ept) over ncacn_ip_tcp to %s:%d failed: %v", host, port, err)
	}
}
