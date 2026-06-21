//go:build integration

package msdrsr_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e3514235-4b06-11d1-ab04-00c04fc2dcd2/4.0/structures"
	msdrsr "github.com/TheManticoreProject/Manticore/network/dcerpc/ms-protocols/ms-drsr"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// TestDRSBindUnbind exercises the full Phase 2 path against a live Domain Controller:
// endpoint-mapper resolution, ncacn_ip_tcp dial, NTLM packet-privacy auth, drsuapi bind,
// and the IDL_DRSBind / IDL_DRSUnbind handshake. It is skipped unless DRSUAPI_TEST_HOST
// is set.
//
//	DRSUAPI_TEST_HOST=10.0.0.10 DRSUAPI_TEST_DOMAIN=lab.local \
//	DRSUAPI_TEST_USER=Administrator DRSUAPI_TEST_PASS=... \
//	go test -tags integration ./network/dcerpc/ms-protocols/ms-drsr/
//
// Pass-the-hash: set DRSUAPI_TEST_HASHES=LM:NT (or :NT) instead of DRSUAPI_TEST_PASS.
// Skip endpoint-mapper resolution with DRSUAPI_TEST_PORT.
func TestDRSBindUnbind(t *testing.T) {
	host := os.Getenv("DRSUAPI_TEST_HOST")
	if host == "" {
		t.Skip("set DRSUAPI_TEST_HOST to run the drsuapi bind integration test")
	}
	creds, err := credentials.NewCredentials(
		os.Getenv("DRSUAPI_TEST_DOMAIN"),
		os.Getenv("DRSUAPI_TEST_USER"),
		os.Getenv("DRSUAPI_TEST_PASS"),
		os.Getenv("DRSUAPI_TEST_HASHES"),
	)
	if err != nil {
		t.Fatalf("build credentials: %v", err)
	}

	c := msdrsr.New(host, creds)
	c.SetTimeout(15 * time.Second)
	if p := os.Getenv("DRSUAPI_TEST_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			t.Fatalf("invalid DRSUAPI_TEST_PORT %q: %v", p, err)
		}
		c.SetPort(n)
	}

	if err := c.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	if c.Handle().IsNull() {
		t.Error("IDL_DRSBind returned a null context handle")
	}
	if len(c.SessionKey()) == 0 {
		t.Error("no NTLM session key after authenticated bind")
	}
	ext := c.ServerExtensions()
	if ext == nil {
		t.Fatal("server returned no DRS_EXTENSIONS")
	}
	t.Logf("server extensions: dwFlags=0x%08x dwFlagsExt=0x%08x dwReplEpoch=%d",
		ext.DwFlags, ext.DwFlagsExt, ext.DwReplEpoch)
	if ext.DwFlags&structures.DRS_EXT_STRONG_ENCRYPTION == 0 {
		t.Error("server did not negotiate STRONG_ENCRYPTION; secret replication would be refused")
	}
}
