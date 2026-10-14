//go:build integration

package client_test

import (
	"context"
	"crypto/x509"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/api/users"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/client"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
)

// TestLiveReadOnlySequence runs the first-slice acceptance sequence against
// a real IBM i system:
//
//	discover -> Sign-on -> Remote Command -> DSPJOB -> list users (AUTU0150)
//	-> close the list -> end the server job
//
// It is read-only and skips unless configured:
//
//	MANTICORE_IBMI_HOST            target host
//	MANTICORE_IBMI_USER            profile (use a least-privileged one)
//	MANTICORE_IBMI_PASSWORD        password
//	MANTICORE_IBMI_CA_FILE         optional PEM roots for TLS
//	MANTICORE_IBMI_ALLOW_CLEARTEXT "1" to use cleartext (isolated labs only)
//
// It never logs the profile, password, seeds, substitutes or raw frames.
func TestLiveReadOnlySequence(t *testing.T) {
	host := os.Getenv("MANTICORE_IBMI_HOST")
	profile := os.Getenv("MANTICORE_IBMI_USER")
	password := os.Getenv("MANTICORE_IBMI_PASSWORD")
	if host == "" || profile == "" || password == "" {
		t.Skip("MANTICORE_IBMI_HOST, MANTICORE_IBMI_USER and MANTICORE_IBMI_PASSWORD are not set")
	}
	cfg := client.Config{Host: host, Transport: ibminet.TransportTLSRequired, OperationTimeout: time.Minute}
	if os.Getenv("MANTICORE_IBMI_ALLOW_CLEARTEXT") == "1" {
		cfg.Transport = ibminet.TransportCleartextExplicit
	}
	if path := os.Getenv("MANTICORE_IBMI_CA_FILE"); path != "" {
		pem, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			t.Fatal("no certificate in MANTICORE_IBMI_CA_FILE")
		}
		cfg.TLS = hostserver.TLSOptions{RootCAs: pool}
	}
	var events eventCounter
	cfg.Tracer = &events

	c, err := client.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	catalog, err := c.DiscoverServices(ctx)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	t.Logf("sign-on %s:%d (%s), remote command %s:%d (%s), transport %s",
		catalog.Signon.Name, catalog.Signon.Port, catalog.Signon.Source,
		catalog.RemoteCommand.Name, catalog.RemoteCommand.Port, catalog.RemoteCommand.Source, cfg.Transport)

	cred, err := ibminet.NewPasswordCredential(profile, []byte(password))
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	err = c.Connect(ctx, cred)
	cred.Destroy()
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	}()
	info := c.SessionInfo()
	t.Logf("VRM %s (raw 0x%08X), password level %d (%s), system CCSID %d, job CCSID %d, remote command level %d, capabilities %+v, server job %q, warnings %q",
		info.VRM, info.ServerVersion, info.PasswordLevel, info.PasswordFamily, info.SystemCCSID, info.JobCCSID,
		info.RemoteCommandLevel, info.Capabilities, info.ServerJob, info.Warnings)

	cmd, err := c.RunCommand(ctx, "DSPJOB OUTPUT(*PRINT) OPTION(*DFNA)", rmtcmd.RunCommandOptions{})
	if err != nil && !errors.Is(err, rmtcmd.ErrNative) {
		t.Fatalf("DSPJOB: %v", err)
	}
	ids := []string{}
	for _, m := range cmd.Messages {
		ids = append(ids, m.ID)
	}
	t.Logf("DSPJOB return code 0x%04X, effective option %s, message IDs %v", cmd.ReturnCode, cmd.EffectiveMessageOption, ids)

	u, err := users.New(c, users.Options{SystemCCSID: info.SystemCCSID, JobCCSID: info.JobCCSID})
	if err != nil {
		t.Fatal(err)
	}
	res, err := u.List(ctx, users.ListOptions{PageSize: 25})
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	groups := 0
	for _, p := range res.Users {
		if p.GroupProfile {
			groups++
		}
		if p.Name == "" || strings.TrimSpace(p.Name) != p.Name {
			t.Errorf("malformed profile name %q", p.Name)
		}
	}
	t.Logf("listed %d profiles (%d groups) of %d reported, status %s, partial %t",
		len(res.Users), groups, res.TotalReported, res.SnapshotStatus, res.Partial)
	if uint32(len(res.Users)) != res.TotalReported {
		t.Errorf("listed %d of %d reported profiles", len(res.Users), res.TotalReported)
	}
	t.Logf("before close: sent %04X, received %04X", events.sent, events.received)
}

// eventCounter records request and reply IDs only (tracer events carry no
// payload).
type eventCounter struct{ sent, received []uint16 }

func (e *eventCounter) Trace(ev hostserver.Event) {
	switch ev.Kind {
	case hostserver.EventFrameSent:
		e.sent = append(e.sent, ev.RequestID)
	case hostserver.EventFrameReceived:
		e.received = append(e.received, ev.RequestID)
	}
}
