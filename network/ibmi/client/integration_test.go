//go:build integration

package client_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/api/users"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/client"
	"github.com/TheManticoreProject/Manticore/network/ibmi/dataqueue"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/ifs"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
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
	async, err := u.List(ctx, users.ListOptions{PageSize: 25, Asynchronous: true})
	if err != nil {
		t.Fatalf("list users asynchronously: %v", err)
	}
	t.Logf("asynchronous build listed %d profiles of %d reported, status %s", len(async.Users), async.TotalReported, async.SnapshotStatus)
	if len(async.Users) != len(res.Users) {
		t.Errorf("asynchronous list has %d profiles, synchronous %d", len(async.Users), len(res.Users))
	}
	for i := range async.Users {
		if i < len(res.Users) && async.Users[i] != res.Users[i] {
			t.Errorf("profile %d differs between build modes", i)
			break
		}
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

// TestLiveProfileToken issues profile tokens with the configured password
// and authenticates with them: a multiple-use token for the full Connect
// sequence, a single-use token for one Sign-on validation. Read-only apart
// from issuing the tokens, which IBM i may record in its audit journal.
func TestLiveProfileToken(t *testing.T) {
	host := os.Getenv("MANTICORE_IBMI_HOST")
	profile := os.Getenv("MANTICORE_IBMI_USER")
	password := os.Getenv("MANTICORE_IBMI_PASSWORD")
	if host == "" || profile == "" || password == "" || os.Getenv("MANTICORE_IBMI_ALLOW_CLEARTEXT") != "1" {
		t.Skip("needs MANTICORE_IBMI_HOST, _USER, _PASSWORD and _ALLOW_CLEARTEXT=1")
	}
	cfg := client.Config{Host: host, Transport: ibminet.TransportCleartextExplicit, OperationTimeout: time.Minute}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cred, err := ibminet.NewPasswordCredential(profile, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Destroy()
	issuer, err := client.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	multi, err := issuer.GenerateProfileToken(ctx, cred, signon.TokenRequest{Type: ibminet.TokenMultipleUse, Timeout: time.Minute})
	if err != nil {
		t.Fatalf("generate multiple-use token: %v", err)
	}
	defer multi.Destroy()
	c, err := client.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Connect(ctx, multi); err != nil {
		t.Fatalf("connect with token: %v", err)
	}
	info := c.SessionInfo()
	if _, err := c.RunCommand(ctx, "DSPJOB OUTPUT(*PRINT)", rmtcmd.RunCommandOptions{}); err != nil && !errors.Is(err, rmtcmd.ErrNative) {
		t.Fatalf("DSPJOB with token: %v", err)
	}
	u, _ := users.New(c, users.Options{SystemCCSID: info.SystemCCSID, JobCCSID: info.JobCCSID})
	res, err := u.List(ctx, users.ListOptions{})
	if err != nil {
		t.Fatalf("list users with token: %v", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
	t.Logf("multiple-use token: connected (VRM %s, system CCSID %d), listed %d profiles", info.VRM, info.SystemCCSID, len(res.Users))

	single, err := issuer.GenerateProfileToken(ctx, cred, signon.TokenRequest{Type: ibminet.TokenSingleUse, Timeout: time.Minute})
	if err != nil {
		t.Fatalf("generate single-use token: %v", err)
	}
	defer single.Destroy()
	catalog, err := issuer.DiscoverServices(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := signon.Validate(ctx, catalog.Signon, single, signon.Options{}); err != nil {
		t.Fatalf("single-use token: %v", err)
	}
	if _, err := signon.Validate(ctx, catalog.Signon, single, signon.Options{}); !errors.Is(err, ibminet.ErrTokenConsumed) {
		t.Errorf("second use of a single-use token = %v", err)
	}
	t.Log("single-use token: one Sign-on validation succeeded; reuse refused locally")
}

// TestLiveDataQueue exercises every Data Queue operation on two queues it
// creates for the run (a FIFO queue and a keyed queue with sender
// information, named MNTDQ plus random hex) in MANTICORE_IBMI_DTAQ_LIBRARY
// (default QGPL), and always deletes them. It never touches other objects.
func TestLiveDataQueue(t *testing.T) {
	host := os.Getenv("MANTICORE_IBMI_HOST")
	profile := os.Getenv("MANTICORE_IBMI_USER")
	password := os.Getenv("MANTICORE_IBMI_PASSWORD")
	if host == "" || profile == "" || password == "" || os.Getenv("MANTICORE_IBMI_ALLOW_CLEARTEXT") != "1" {
		t.Skip("needs MANTICORE_IBMI_HOST, _USER, _PASSWORD and _ALLOW_CLEARTEXT=1")
	}
	library := os.Getenv("MANTICORE_IBMI_DTAQ_LIBRARY")
	if library == "" {
		library = "QGPL"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cred, err := ibminet.NewPasswordCredential(profile, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Destroy()
	c, err := client.New(client.Config{Host: host, Transport: ibminet.TransportCleartextExplicit, OperationTimeout: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Connect(ctx, cred); err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()
	dq, err := c.OpenDataQueue(ctx, cred)
	if err != nil {
		t.Fatalf("open data queue: %v", err)
	}
	// A cleanup, not a defer: it must run after the queue deletions below.
	t.Cleanup(func() { dq.Close() })
	t.Logf("data queue session: system CCSID %d, server job %q", dq.Info().SystemCCSID, dq.Info().ServerJob)

	suffix := make([]byte, 2)
	_, _ = rand.Read(suffix)
	fifo := dataqueue.Queue{Library: library, Name: "MNTDQ" + strings.ToUpper(hex.EncodeToString(suffix)) + "F"}
	keyed := dataqueue.Queue{Library: library, Name: "MNTDQ" + strings.ToUpper(hex.EncodeToString(suffix)) + "K"}

	if _, err := dq.Attributes(ctx, fifo); !errors.Is(err, dataqueue.ErrQueueNotFound) {
		t.Fatalf("attributes of a missing queue = %v", err)
	}
	var e *dataqueue.Error
	if errors.As(err, &e) {
		t.Logf("missing queue: return code 0x%04X, message %s", e.ReturnCode, e.MessageID)
	}

	if err := dq.Create(ctx, fifo, dataqueue.CreateOptions{MaxEntryLength: 128, Kind: dataqueue.FIFO, Authority: dataqueue.AuthorityExclude, Description: "Manticore live test"}); err != nil {
		t.Fatalf("create FIFO: %v", err)
	}
	t.Cleanup(func() {
		if err := dq.Delete(context.Background(), fifo); err != nil {
			t.Errorf("delete FIFO: %v", err)
		}
	})
	if err := dq.Create(ctx, keyed, dataqueue.CreateOptions{MaxEntryLength: 64, Kind: dataqueue.Keyed, KeyLength: 4, SaveSenderInfo: true, Authority: dataqueue.AuthorityExclude}); err != nil {
		t.Fatalf("create keyed: %v", err)
	}
	t.Cleanup(func() {
		if err := dq.Delete(context.Background(), keyed); err != nil {
			t.Errorf("delete keyed: %v", err)
		}
	})
	if err := dq.Create(ctx, fifo, dataqueue.CreateOptions{MaxEntryLength: 128, Kind: dataqueue.FIFO}); !errors.Is(err, dataqueue.ErrAlreadyExists) {
		t.Errorf("second create = %v", err)
	}

	a, err := dq.Attributes(ctx, fifo)
	if err != nil {
		t.Fatalf("attributes: %v", err)
	}
	if a.MaxEntryLength != 128 || a.Kind != dataqueue.FIFO || a.KeyLength != 0 || a.SaveSenderInfo || a.Description != "Manticore live test" {
		t.Errorf("FIFO attributes = %+v", a)
	}
	if a, err = dq.Attributes(ctx, keyed); err != nil || a.Kind != dataqueue.Keyed || a.KeyLength != 4 || !a.SaveSenderInfo {
		t.Errorf("keyed attributes = %+v, %v", a, err)
	}

	if e, err := dq.Read(ctx, fifo, 0); e != nil || err != nil {
		t.Fatalf("read of an empty queue = %+v, %v", e, err)
	}
	start := time.Now()
	if e, err := dq.Read(ctx, fifo, 2*time.Second); e != nil || err != nil || time.Since(start) < time.Second {
		t.Fatalf("2s read of an empty queue = %+v, %v after %s", e, err, time.Since(start))
	}
	for _, data := range [][]byte{{1, 2, 3}, {0}, bytes.Repeat([]byte{0xAB}, 128)} {
		if err := dq.Write(ctx, fifo, data); err != nil {
			t.Fatalf("write %d bytes: %v", len(data), err)
		}
	}
	// V5R4 refuses empty and oversized entries with a command check
	// (CPF2498 for an empty one) and keeps the session.
	for _, n := range []int{0, 129} {
		if err := dq.Write(ctx, fifo, make([]byte, n)); !errors.As(err, &e) {
			t.Errorf("write of %d bytes = %v", n, err)
		} else {
			t.Logf("%d-byte entry: return code 0x%04X, message %s", n, e.ReturnCode, e.MessageID)
		}
	}
	if e, err := dq.Peek(ctx, fifo, 0); err != nil || e == nil || !bytes.Equal(e.Data, []byte{1, 2, 3}) {
		t.Fatalf("peek = %+v, %v", e, err)
	}
	for i, want := range [][]byte{{1, 2, 3}, {0}, bytes.Repeat([]byte{0xAB}, 128)} {
		e, err := dq.Read(ctx, fifo, 0)
		if err != nil || e == nil || !bytes.Equal(e.Data, want) || e.Key != nil || e.Sender != nil {
			t.Fatalf("read %d = %+v, %v", i, e, err)
		}
	}
	if err := dq.Write(ctx, fifo, []byte{9}); err != nil {
		t.Fatal(err)
	}
	if err := dq.Clear(ctx, fifo); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if e, err := dq.Read(ctx, fifo, 0); e != nil || err != nil {
		t.Fatalf("read after clear = %+v, %v", e, err)
	}

	for _, k := range []string{"0002", "0001", "0003", "0001"} {
		if err := dq.WriteKey(ctx, keyed, []byte(k), []byte("entry "+k)); err != nil {
			t.Fatalf("write key %s: %v", k, err)
		}
	}
	e2, err := dq.PeekKey(ctx, keyed, dataqueue.SearchGreater, []byte("0001"), 0)
	if err != nil || e2 == nil || string(e2.Key) != "0002" || e2.Sender == nil {
		t.Fatalf("peek GT 0001 = %+v, %v", e2, err)
	}
	if !strings.EqualFold(e2.Sender.CurrentUser, profile) || !strings.HasSuffix(dq.Info().ServerJob, e2.Sender.JobName) {
		t.Errorf("sender = %+v; want the server job and the signed-on profile", *e2.Sender)
	}
	t.Logf("sender: job %s, current user %s", e2.Sender.Job(), e2.Sender.CurrentUser)
	if e, err := dq.ReadKey(ctx, keyed, dataqueue.SearchEqual, []byte("0001"), 0); err != nil || e == nil || string(e.Data) != "entry 0001" {
		t.Fatalf("read EQ 0001 = %+v, %v", e, err)
	}
	if err := dq.ClearKey(ctx, keyed, []byte("0001")); err != nil {
		t.Fatalf("clear key: %v", err)
	}
	if e, err := dq.ReadKey(ctx, keyed, dataqueue.SearchEqual, []byte("0001"), 0); e != nil || err != nil {
		t.Fatalf("read EQ 0001 after clear key = %+v, %v", e, err)
	}
	if e, err := dq.ReadKey(ctx, keyed, dataqueue.SearchLessOrEqual, []byte("9999"), 0); err != nil || e == nil || string(e.Key) != "0002" {
		t.Fatalf("read LE 9999 = %+v, %v", e, err)
	}
	if _, err := dq.Read(ctx, keyed, 0); !errors.Is(err, dataqueue.ErrWrongQueueKind) {
		t.Errorf("sequential read of a keyed queue = %v", err)
	} else if errors.As(err, &e) {
		t.Logf("kind mismatch: return code 0x%04X, message %s", e.ReturnCode, e.MessageID)
	}
	if err := dq.Err(); err != nil {
		t.Fatalf("session unusable: %v", err)
	}
}

// TestLiveIFS reads one stream file through the File server and checks the
// error forms. It is read-only. MANTICORE_IBMI_IFS_FILE names the file
// (default /QIBM/ProdData/OS400/jt400/lib/jt400Native.jar, a ZIP archive
// readable by *PUBLIC on V5R4).
func TestLiveIFS(t *testing.T) {
	host := os.Getenv("MANTICORE_IBMI_HOST")
	profile := os.Getenv("MANTICORE_IBMI_USER")
	password := os.Getenv("MANTICORE_IBMI_PASSWORD")
	if host == "" || profile == "" || password == "" || os.Getenv("MANTICORE_IBMI_ALLOW_CLEARTEXT") != "1" {
		t.Skip("needs MANTICORE_IBMI_HOST, _USER, _PASSWORD and _ALLOW_CLEARTEXT=1")
	}
	path := os.Getenv("MANTICORE_IBMI_IFS_FILE")
	isDefault := path == ""
	if isDefault {
		path = "/QIBM/ProdData/OS400/jt400/lib/jt400Native.jar"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cred, err := ibminet.NewPasswordCredential(profile, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Destroy()
	c, err := client.New(client.Config{Host: host, Transport: ibminet.TransportCleartextExplicit, OperationTimeout: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Connect(ctx, cred); err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer c.Close()
	s, err := c.OpenIFS(ctx, cred)
	if err != nil {
		t.Fatalf("open IFS: %v", err)
	}
	defer s.Close()
	info := s.Info()
	t.Logf("file server: level %d, max block %d, path CCSID %d, server job %q", info.DataStreamLevel, info.MaxBlock, info.PathCCSID, info.ServerJob)

	f, err := s.Open(ctx, path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	fi := f.Info()
	t.Logf("file: %d bytes, data CCSID %d, modified %s", fi.Size, fi.DataCCSID, fi.Modified)
	if fi.Size < 1 {
		t.Fatalf("size %d", fi.Size)
	}
	if data, err := f.ReadAt(ctx, fi.Size, 16); !errors.Is(err, io.EOF) || len(data) != 0 {
		t.Errorf("read at the end = %d bytes, %v", len(data), err)
	}
	if data, err := f.ReadAt(ctx, fi.Size-1, 16); err != nil || len(data) != 1 {
		t.Errorf("read of the last byte = %d bytes, %v", len(data), err)
	}
	if err := f.Close(ctx); err != nil {
		t.Fatalf("close: %v", err)
	}

	start := time.Now()
	data, err := s.ReadFile(ctx, path, int(fi.Size))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if int64(len(data)) != fi.Size || isDefault && !bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		t.Errorf("read %d bytes (want %d), prefix % x", len(data), fi.Size, data[:min(4, len(data))])
	}
	t.Logf("read %d bytes in %s", len(data), time.Since(start).Round(time.Millisecond))
	if _, err := s.ReadFile(ctx, path, int(fi.Size)-1); !errors.Is(err, ifs.ErrTooLarge) {
		t.Errorf("limit below the size = %v", err)
	}

	if _, err := s.Open(ctx, "/manticore-missing-file"); !errors.Is(err, ifs.ErrNotFound) {
		t.Errorf("missing file = %v", err)
	}
	var e *ifs.Error
	if _, err := s.Open(ctx, "/QIBM"); !errors.As(err, &e) {
		t.Errorf("open of a directory = %v", err)
	} else {
		t.Logf("open of a directory: return code %d", e.ReturnCode)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("session unusable: %v", err)
	}
}
