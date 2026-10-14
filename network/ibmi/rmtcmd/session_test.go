package rmtcmd_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/hosttest"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
	"github.com/TheManticoreProject/Manticore/network/ibmi/signon"
)

const (
	testProfile  = "CANARYUSR"
	testPassword = "CANARYPW7"
)

var rmtEndpoint = ibminet.Endpoint{Service: ibminet.ServiceRemoteCommand, Name: "as-rmtcmd", Host: "ibmi.test", Port: hosttest.RemoteCommandPort, Source: ibminet.EndpointExplicit}

func startSession(t *testing.T, srv *hosttest.Server, password string, family ibminet.PasswordFamily, opts rmtcmd.StartOptions) (*rmtcmd.Session, error) {
	t.Helper()
	return startSessionCtx(context.Background(), t, srv, password, family, opts)
}

func startSessionCtx(ctx context.Context, t *testing.T, srv *hosttest.Server, password string, family ibminet.PasswordFamily, opts rmtcmd.StartOptions) (*rmtcmd.Session, error) {
	t.Helper()
	if srv.Profile == "" {
		srv.Profile, srv.Password = testProfile, testPassword
	}
	conn, err := (&hostserver.Dialer{Dial: srv.Dial}).DialEndpoint(context.Background(), rmtEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	cred, err := ibminet.NewPasswordCredential(testProfile, []byte(password))
	if err != nil {
		t.Fatal(err)
	}
	defer cred.Destroy()
	return rmtcmd.Start(ctx, conn, cred, family, opts)
}

func TestStartAndClose(t *testing.T) {
	for _, tc := range []struct {
		level  byte
		family ibminet.PasswordFamily
	}{{0, ibminet.PasswordFamilyDES}, {2, ibminet.PasswordFamilySHA1}} {
		srv := &hosttest.Server{PasswordLevel: tc.level, RemoteCommandLevel: 7}
		s, err := startSession(t, srv, testPassword, tc.family, rmtcmd.StartOptions{})
		if err != nil {
			t.Fatalf("level %d: %v", tc.level, err)
		}
		info := s.Info()
		if info.JobCCSID != 37 || info.NegotiatedJobCCSID != 37 || info.ServerJob != "QZRCSRVS" || len(info.Warnings) != 0 {
			t.Errorf("info = %+v", info)
		}
		if c := s.Capabilities(); c != (rmtcmd.Capabilities{RemoteCommandLevel: 7, InputOutputUsage33: true, AllCommandMessages: true}) {
			t.Errorf("capabilities = %+v", c)
		}
		if s.Codec().CCSID() != ccsid.CCSID37 {
			t.Errorf("codec = %d", s.Codec().CCSID())
		}
		reqs := srv.Requests()
		want := []struct {
			id   uint16
			corr uint32
		}{{0x7001, 0}, {0x7002, 0}, {0x1001, 1}}
		for i, w := range want {
			if reqs[i].RequestID != w.id || reqs[i].CorrelationID != w.corr {
				t.Errorf("request %d = 0x%04X corr %d; want 0x%04X corr %d", i, reqs[i].RequestID, reqs[i].CorrelationID, w.id, w.corr)
			}
		}
		if err := s.Close(time.Second); err != nil {
			t.Fatal(err)
		}
		if err := s.Close(time.Second); err != nil {
			t.Errorf("second Close = %v", err)
		}
		if err := srv.WaitClosed(2 * time.Second); err != nil {
			t.Fatal(err)
		}
		if srv.Count(message.RequestEndServer) != 1 {
			t.Errorf("0x1004 sent %d times", srv.Count(message.RequestEndServer))
		}
		if !errors.Is(s.Err(), hostserver.ErrClosed) {
			t.Errorf("Err after Close = %v", s.Err())
		}
	}
}

func TestCapabilityThresholds(t *testing.T) {
	for level, want := range map[uint16]rmtcmd.Capabilities{
		0:  {RemoteCommandLevel: 0},
		4:  {RemoteCommandLevel: 4},
		5:  {RemoteCommandLevel: 5, InputOutputUsage33: true},
		6:  {RemoteCommandLevel: 6, InputOutputUsage33: true},
		7:  {RemoteCommandLevel: 7, InputOutputUsage33: true, AllCommandMessages: true},
		9:  {RemoteCommandLevel: 9, InputOutputUsage33: true, AllCommandMessages: true},
		10: {RemoteCommandLevel: 10, InputOutputUsage33: true, AllCommandMessages: true, ExplicitCommandCCSID: true},
	} {
		if got := rmtcmd.CapabilitiesForLevel(level); got != want {
			t.Errorf("level %d = %+v", level, got)
		}
	}
}

func TestStartUsesFreshSeedsAndSubstitute(t *testing.T) {
	srv := &hosttest.Server{}
	s, err := startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close(time.Second)
	s2, err := startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close(time.Second)
	var seeds [][8]byte
	var subs [][]byte
	for _, r := range srv.Requests() {
		switch r.RequestID {
		case 0x7001:
			seeds = append(seeds, r.ClientSeed)
		case 0x7002:
			subs = append(subs, r.Substitute)
		}
	}
	if len(seeds) != 2 || seeds[0] == seeds[1] || bytes.Equal(subs[0], subs[1]) {
		t.Errorf("sessions reused seeds %x or substitutes %x", seeds, subs)
	}
}

func TestStartRejectedCredentialIsNotRetried(t *testing.T) {
	srv := &hosttest.Server{}
	_, err := startSession(t, srv, "WRONGPW1", ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	var se *signon.Error
	if !errors.Is(err, signon.ErrAuthenticationRejected) || !errors.As(err, &se) || se.Reason() != signon.ReasonPasswordIncorrect {
		t.Fatalf("error = %v", err)
	}
	if srv.Count(0x7002) != 1 || srv.Count(0x1001) != 0 {
		t.Errorf("requests = %+v", srv.Requests())
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	for _, leak := range []string{"CANARY", "WRONGPW1"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("error leaks %q: %v", leak, err)
		}
	}
}

func TestAttributesReturnCodes(t *testing.T) {
	srv := &hosttest.Server{AttributesRC: 0x0106}
	s, err := startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	if err != nil {
		t.Fatalf("warning code rejected: %v", err)
	}
	if w := s.Info().Warnings; len(w) != 1 || !strings.Contains(w[0], "0x0106") {
		t.Errorf("warnings = %q", w)
	}
	s.Close(time.Second)

	srv = &hosttest.Server{AttributesRC: 0x0102}
	_, err = startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	var ae *rmtcmd.AttributesError
	if !errors.Is(err, rmtcmd.ErrAttributesRejected) || !errors.As(err, &ae) || ae.Code != 0x0102 {
		t.Fatalf("error = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestJobCCSID(t *testing.T) {
	srv := &hosttest.Server{JobCCSID: 273}
	if _, err := startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{}); !errors.Is(err, ccsid.ErrUnsupported) {
		t.Fatalf("unsupported job CCSID error = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}

	srv = &hosttest.Server{JobCCSID: 273}
	s, err := startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{JobCCSIDOverride: 1140})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close(time.Second)
	info := s.Info()
	if info.JobCCSID != 1140 || info.NegotiatedJobCCSID != 273 || len(info.Warnings) != 1 || s.Codec().CCSID() != ccsid.CCSID1140 {
		t.Errorf("override info = %+v", info)
	}
	info.Warnings[0] = "mutated"
	if s.Info().Warnings[0] == "mutated" {
		t.Error("Info shares its warnings slice")
	}
}

func TestStartPoisonsOnWrongCorrelation(t *testing.T) {
	srv := &hosttest.Server{AttributesCorrelationDelta: 1}
	_, err := startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	if !errors.Is(err, hostserver.ErrUnexpectedReply) || !errors.Is(err, hostserver.ErrPoisoned) {
		t.Fatalf("error = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	if srv.Count(message.RequestEndServer) != 0 {
		t.Error("0x1004 written to a poisoned stream")
	}
}

func TestStartCancellationClosesSocket(t *testing.T) {
	srv := &hosttest.Server{SilentOn: 0x7002}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := startSessionCtx(ctx, t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestInfoContainsNoSecrets(t *testing.T) {
	srv := &hosttest.Server{}
	s, err := startSession(t, srv, testPassword, ibminet.PasswordFamilyDES, rmtcmd.StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close(time.Second)
	out := fmt.Sprintf("%+v %#v", s.Info(), s.Info())
	for _, r := range srv.Requests() {
		for _, secret := range []string{"CANARY", fmt.Sprintf("%x", r.Substitute), fmt.Sprintf("%x", r.ClientSeed)} {
			if r.RequestID == 0x1001 || secret == "" || secret == "0000000000000000" {
				continue
			}
			if strings.Contains(out, secret) {
				t.Errorf("Info leaks %q", secret)
			}
		}
	}
}
