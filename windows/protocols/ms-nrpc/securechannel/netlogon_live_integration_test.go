//go:build integration

// Live integration test for the Netlogon secure channel and RPC_C_AUTHN_NETLOGON transport
// sealing over the full DCE/RPC-over-SMB stack. Excluded from the default build by the
// "integration" tag. Requires a reachable DC and a machine account whose password is known.
//
//	DCERPC_TEST_HOST=192.168.1.39 DCERPC_TEST_USER=Administrator DCERPC_TEST_PASS='Admin123!' \
//	NETLOGON_COMPUTER=MANTICORE1 NETLOGON_PASS='Manticore1Pass!' \
//	NETLOGON_SERVER='\\TMP-W-2016' NETLOGON_DOMAIN=TMP-W-2016.local \
//	go test -tags integration -v -run Integration_Netlogon \
//	  ./windows/protocols/ms-nrpc/securechannel/
//
// Set NETLOGON_SUITE=rc4 to negotiate the legacy strong-key (RC4) suite instead of AES.
package securechannel_test

import (
	"os"
	"testing"

	netlogon "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345678-1234-abcd-ef00-01234567cffb/1.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	dcerpcclient "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	smbclient "github.com/TheManticoreProject/Manticore/network/smb/client"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	msnrpc "github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc"
	"github.com/TheManticoreProject/Manticore/windows/protocols/ms-nrpc/securechannel"
)

func TestIntegration_NetlogonSecureChannel(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	computer := os.Getenv("NETLOGON_COMPUTER") // machine account name without the trailing '$'
	machinePass := os.Getenv("NETLOGON_PASS")  // that account's cleartext password
	if host == "" || computer == "" || machinePass == "" {
		t.Skip("set DCERPC_TEST_HOST, NETLOGON_COMPUTER and NETLOGON_PASS to run the live Netlogon test")
	}
	serverName := os.Getenv("NETLOGON_SERVER") // DC name for the ServerName parameter, e.g. \\DC01
	domain := os.Getenv("NETLOGON_DOMAIN")     // domain in the NL_AUTH_MESSAGE bind token

	creds, err := credentials.NewCredentials(os.Getenv("DCERPC_TEST_DOMAIN"), os.Getenv("DCERPC_TEST_USER"), os.Getenv("DCERPC_TEST_PASS"), "")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}
	smb, err := smbclient.Dial(host, 445, smbclient.Options{})
	if err != nil {
		t.Fatalf("SMB Dial: %v", err)
	}
	defer smb.Disconnect()
	if err := smb.Login(creds); err != nil {
		t.Fatalf("SMB Login: %v", err)
	}
	defer smb.Logoff()
	if err := smb.TreeConnect("IPC$"); err != nil {
		t.Fatalf("SMB TreeConnect(IPC$): %v", err)
	}
	defer smb.TreeDisconnect()

	// ---- Binding #1: anonymous, for the secure-channel handshake (Stage A) ----
	tr1, err := smb.RPCTransport(netlogon.PipeName)
	if err != nil {
		t.Fatalf("RPCTransport #1: %v", err)
	}
	rpc1 := dcerpcclient.NewClient(tr1)
	if err := rpc1.Bind(netlogon.SyntaxID()); err != nil {
		t.Fatalf("bind #1 (anon netlogon): %v", err)
	}

	cfg := securechannel.SecureChannelConfig{
		ComputerName:      computer,
		AccountName:       computer + "$",
		SecureChannelType: msnrpc.WorkstationSecureChannel,
		Password:          machinePass,
	}
	if os.Getenv("NETLOGON_SUITE") == "rc4" {
		cfg.NegotiateFlags = securechannel.DefaultNegotiateFlags &^ netlogon.NegotiateAES
	}
	sc, err := securechannel.Establish(rpc1, cfg)
	if err != nil {
		t.Fatalf("STAGE A FAILED — SecureChannel.Establish: %v", err)
	}
	_ = rpc1.Close()
	t.Logf("[STAGE A ok] secure channel established (aes=%v); session key %x, negotiated flags %#08x",
		sc.UsesAES(), sc.SessionKey(), sc.NegotiateFlags())

	// The RPC-transport per-message token cipher tracks the session's AES negotiation. On an
	// AES-capable DC the RPC SSP uses NL_AUTH_SHA2_SIGNATURE tokens even if a strong-key
	// session was forced, so the legacy RC4 token path only applies to genuinely pre-AES DCs
	// and cannot be exercised here; the strong-key session itself is validated by Stage A.
	if !sc.UsesAES() {
		t.Logf("[RC4 ok] strong-key secure channel validated (Stage A); RPC-transport RC4 sealing " +
			"is for pre-AES DCs and is not exercised against this AES-capable DC")
		return
	}

	// ---- Binding #2: RPC_C_AUTHN_NETLOGON sealed (PR-D) ----
	tr2, err := smb.RPCTransport(netlogon.PipeName)
	if err != nil {
		t.Fatalf("RPCTransport #2: %v", err)
	}
	rpc2 := dcerpcclient.NewClient(tr2)
	defer rpc2.Close()

	provider := securechannel.NewNetlogonSecurityContext(sc)
	bindToken := msnrpc.BuildClientNlAuthMessage(computer, domain).Marshal()
	if err := rpc2.SetAuthProvider(pdu.AuthTypeNetlogon, pdu.AuthLevelPktPrivacy, provider, bindToken); err != nil {
		t.Fatalf("SetAuthProvider(netlogon): %v", err)
	}
	if err := rpc2.Bind(netlogon.SyntaxID()); err != nil {
		t.Fatalf("PR-D FAILED — sealed netlogon bind: %v", err)
	}
	t.Logf("[PR-D ok] RPC_C_AUTHN_NETLOGON sealed bind accepted")

	// Decisive end-to-end call: NetrLogonGetCapabilities validates the secure channel. A
	// STATUS_SUCCESS return means the server accepted our sealed request stub, verified our
	// application authenticator, and returned a sealed response we unsealed; verifying the
	// ReturnAuthenticator proves the rolling credential is in lockstep.
	comp := ndr.WSTR(computer)
	auth := sc.NextAuthenticator()
	ret, caps, err := functions.NetrLogonGetCapabilities(rpc2, ndr.WSTR(serverName), &comp, auth, msnrpc.NETLOGON_AUTHENTICATOR{}, 1)
	if err != nil {
		t.Fatalf("PR-D FAILED — NetrLogonGetCapabilities over sealed channel: %v", err)
	}
	if err := sc.VerifyResponseAuthenticator(ret); err != nil {
		t.Fatalf("PR-D FAILED — server ReturnAuthenticator did not verify: %v", err)
	}
	t.Logf("[PR-D ok] sealed NetrLogonGetCapabilities round-trip verified; capabilities %#08x", uint32(caps.ServerCapabilities))
}
