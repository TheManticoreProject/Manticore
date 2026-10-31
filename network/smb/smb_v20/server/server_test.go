package server_test

import (
	"bytes"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/crypto/nt"
	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/smb/common/transport"
	smb1 "github.com/TheManticoreProject/Manticore/network/smb/smb_v10/server"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
	"github.com/TheManticoreProject/Manticore/windows/errors/nt_status"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
	"github.com/TheManticoreProject/Manticore/windows/filesystem"
	"github.com/TheManticoreProject/Manticore/windows/filesystem/infoclass"
)

// withTestIdentity fills the account and names the tests authenticate against.
func withTestIdentity(config *server.Config) server.Config {
	if config.Authenticator == nil {
		config.Authenticator = smb1.StaticAccounts(smb1.Account{Domain: "LAB", Username: "alice", NTHash: nt.NTHash("Passw0rd!")})
	}
	config.DomainName, config.ServerName = "LAB", "FILESRV"
	config.DNSComputerName, config.DNSDomainName = "filesrv.lab.local", "lab.local"
	return *config
}

var allDialects = []dialects.Dialect{
	dialects.SMB2_DIALECT_2_0_2, dialects.SMB2_DIALECT_2_1_0,
	dialects.SMB2_DIALECT_3_0_0, dialects.SMB2_DIALECT_3_0_2, dialects.SMB2_DIALECT_3_1_1,
}

// startServer serves a "files" share holding hello.txt on a loopback port.
func startServer(t *testing.T, config server.Config) (*smb1.MemoryFileSystem, int) {
	t.Helper()
	withTestIdentity(&config)

	srv, err := server.NewServer(config)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	fs := smb1.NewMemoryFileSystem("FILES")
	if err := fs.AddFile("hello.txt", []byte("hello world")); err != nil {
		t.Fatal(err)
	}
	if err := srv.AddShare(&server.Share{Name: "files", FS: fs}); err != nil {
		t.Fatal(err)
	}

	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(listener)
	t.Cleanup(func() { srv.Close() })
	return fs, listener.Addr().(*net.TCPAddr).Port
}

// dial connects, authenticates as alice and connects the "files" share.
func dial(t *testing.T, port int, dialect dialects.Dialect, password string) (*client.Client, error) {
	t.Helper()
	c := client.NewClientUsingTCPTransport(net.IPv4(127, 0, 0, 1), port)
	c.Profile = c.Profile.Clone()
	c.Profile.Dialects = []dialects.Dialect{dialect}
	t.Cleanup(func() { c.Disconnect() })

	if err := c.Connect(net.IPv4(127, 0, 0, 1), port); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	creds, err := credentials.NewCredentials("LAB", "alice", password, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SessionSetup(creds); err != nil {
		return nil, err
	}
	if err := c.TreeConnect("files"); err != nil {
		t.Fatalf("TreeConnect: %v", err)
	}
	return c, nil
}

// dialAnonymous connects and sets up a null session.
func dialAnonymous(t *testing.T, port int) error {
	t.Helper()
	c := client.NewClientUsingTCPTransport(net.IPv4(127, 0, 0, 1), port)
	t.Cleanup(func() { c.Disconnect() })
	if err := c.Connect(net.IPv4(127, 0, 0, 1), port); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	creds, err := credentials.NewCredentials("", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	return c.SessionSetup(creds)
}

// exerciseShare runs the core file operations and checks their effects.
func exerciseShare(t *testing.T, c *client.Client, fs *smb1.MemoryFileSystem) {
	t.Helper()

	id, err := c.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatalf("open hello.txt: %v", err)
	}
	if data, err := c.ReadFile(id, 0, 64); err != nil || string(data) != "hello world" {
		t.Fatalf("ReadFile = %q, %v", data, err)
	}
	if info, err := c.QueryFileStandardInfo(id); err != nil || info.EndOfFile != 11 {
		t.Fatalf("QueryFileStandardInfo = %+v, %v", info, err)
	}
	if err := c.CloseFile(id); err != nil {
		t.Fatalf("CloseFile: %v", err)
	}

	payload := bytes.Repeat([]byte("0123456789"), 10000)
	id, err = c.CreateFile("new.bin", fileflags.GENERIC_READ|fileflags.GENERIC_WRITE|fileflags.DELETE, 0, fileflags.FILE_CREATE, fileflags.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("create new.bin: %v", err)
	}
	if n, err := c.WriteFile(id, 0, payload); err != nil || int(n) != len(payload) {
		t.Fatalf("WriteFile = %d, %v", n, err)
	}
	if err := c.Flush(id); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if data, err := c.ReadFile(id, 0, uint32(len(payload))); err != nil || !bytes.Equal(data, payload) {
		t.Fatalf("read back %d bytes, %v", len(data), err)
	}
	if err := c.RenameByHandle(id, "renamed.bin", false); err != nil {
		t.Fatalf("RenameByHandle: %v", err)
	}
	if err := c.CloseFile(id); err != nil {
		t.Fatalf("CloseFile: %v", err)
	}
	if _, err := fs.Stat("renamed.bin"); err != nil {
		t.Fatalf("renamed.bin missing: %v", err)
	}

	if err := c.CreateDirectory("dir"); err != nil {
		t.Fatalf("CreateDirectory: %v", err)
	}
	root, err := c.CreateFile("", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		t.Fatalf("open root: %v", err)
	}
	listing, err := c.QueryDirectory(root, uint8(infoclass.FileIdBothDirectoryInformation), "*", 0)
	if err != nil {
		t.Fatalf("QueryDirectory: %v", err)
	}
	var names []string
	for _, entry := range filesystem.ParseFileIdBothDirectoryInformationList(listing) {
		names = append(names, entry.FileName)
	}
	if got := strings.Join(names, ","); got != "dir,hello.txt,renamed.bin" {
		t.Errorf("listing = %s", got)
	}
	if more, err := c.QueryDirectory(root, uint8(infoclass.FileIdBothDirectoryInformation), "*", 0); err != nil || len(more) != 0 {
		t.Errorf("second QueryDirectory = %d bytes, %v; want end of listing", len(more), err)
	}
	if size, err := c.QueryFsSizeInfo(root); err != nil || size.BytesPerSector == 0 {
		t.Errorf("QueryFsSizeInfo = %+v, %v", size, err)
	}
	c.CloseFile(root)

	if err := c.DeleteFile("renamed.bin"); err != nil {
		t.Fatalf("DeleteFile: %v", err)
	}
	if _, err := fs.Stat("renamed.bin"); err == nil {
		t.Error("renamed.bin still exists after DeleteFile")
	}
	if err := c.Logoff(); err != nil {
		t.Errorf("Logoff: %v", err)
	}
}

func TestServerServesEveryDialect(t *testing.T) {
	for _, dialect := range allDialects {
		t.Run(dialect.String(), func(t *testing.T) {
			fs, port := startServer(t, server.Config{})
			c, err := dial(t, port, dialect, "Passw0rd!")
			if err != nil {
				t.Fatalf("SessionSetup: %v", err)
			}
			if c.Connection.Dialect != dialect {
				t.Fatalf("negotiated %s, want %s", c.Connection.Dialect, dialect)
			}
			exerciseShare(t, c, fs)
		})
	}
}

func TestServerRequiredSigning(t *testing.T) {
	for _, dialect := range allDialects {
		t.Run(dialect.String(), func(t *testing.T) {
			fs, port := startServer(t, server.Config{RequireSigning: true})
			c, err := dial(t, port, dialect, "Passw0rd!")
			if err != nil {
				t.Fatalf("SessionSetup: %v", err)
			}
			if !c.Session.SigningActive {
				t.Fatal("signing is not active on the client")
			}
			exerciseShare(t, c, fs)
		})
	}
}

func TestServerRequiredEncryption(t *testing.T) {
	for _, dialect := range allDialects[2:] {
		t.Run(dialect.String(), func(t *testing.T) {
			fs, port := startServer(t, server.Config{RequireEncryption: true})
			c, err := dial(t, port, dialect, "Passw0rd!")
			if err != nil {
				t.Fatalf("SessionSetup: %v", err)
			}
			if !c.IsEncryptionActive() {
				t.Fatal("the session is not encrypted")
			}
			exerciseShare(t, c, fs)
		})
	}

	t.Run("refuses SMB 2.1", func(t *testing.T) {
		_, port := startServer(t, server.Config{RequireEncryption: true})
		if _, err := dial(t, port, dialects.SMB2_DIALECT_2_1_0, "Passw0rd!"); err == nil {
			t.Fatal("an SMB 2.1 session was admitted although encryption is required")
		}
	})
}

func TestServerAuthenticationPolicy(t *testing.T) {
	t.Run("wrong password", func(t *testing.T) {
		_, port := startServer(t, server.Config{AllowGuest: true})
		if _, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "wrong"); err == nil {
			t.Fatal("a wrong password was accepted")
		}
	})
	t.Run("unknown user refused", func(t *testing.T) {
		_, port := startServer(t, server.Config{Authenticator: smb1.StaticAccounts()})
		if _, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!"); err == nil {
			t.Fatal("an unknown identity was accepted without AllowGuest")
		}
	})
	t.Run("unknown user as guest", func(t *testing.T) {
		_, port := startServer(t, server.Config{Authenticator: smb1.StaticAccounts(), AllowGuest: true})
		c, err := dial(t, port, dialects.SMB2_DIALECT_2_1_0, "Passw0rd!")
		if err != nil {
			t.Fatalf("guest session refused: %v", err)
		}
		if !c.Session.IsGuest {
			t.Error("the session is not flagged as guest")
		}
	})
	t.Run("anonymous without AllowGuest", func(t *testing.T) {
		_, port := startServer(t, server.Config{AllowAnonymous: true})
		if err := dialAnonymous(t, port); err != nil {
			t.Fatalf("null session refused with AllowAnonymous: %v", err)
		}
	})
	t.Run("anonymous refused", func(t *testing.T) {
		_, port := startServer(t, server.Config{AllowGuest: true})
		if err := dialAnonymous(t, port); err == nil {
			t.Fatal("a null session was accepted without AllowAnonymous")
		}
	})
}

func TestServerOplockBreak(t *testing.T) {
	_, port := startServer(t, server.Config{})
	holder, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	id, level, err := holder.CreateFileWithOplock("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE,
		fileflags.FILE_OPEN, 0, commands.SMB2_OPLOCK_LEVEL_BATCH)
	if err != nil || level != commands.SMB2_OPLOCK_LEVEL_II {
		t.Fatalf("open with oplock = level %d, %v; want level II", level, err)
	}

	writer, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.CreateFile("hello.txt", fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, 0); err != nil {
		t.Fatalf("open for writing: %v", err)
	}

	// The break is unsigned, as MS-SMB2 3.3.4.6 has it, so it is read raw.
	raw, err := holder.Transport.Receive()
	if err != nil {
		t.Fatalf("no break received: %v", err)
	}
	var notification commands.OplockBreakRequest
	if len(raw) < 64 || raw[12] != 0x12 {
		t.Fatalf("expected an OPLOCK_BREAK, got % x", raw[:min(len(raw), 16)])
	}
	if _, err := notification.Unmarshal(raw[64:]); err != nil || notification.FileId != id || notification.OplockLevel != commands.SMB2_OPLOCK_LEVEL_NONE {
		t.Fatalf("break = %+v, %v", notification, err)
	}
}

func TestServerKerberos(t *testing.T) {
	const spn = "cifs/filesrv.lab.local"
	serviceKey := bytes.Repeat([]byte{0x5A}, 32)

	for _, dialect := range []dialects.Dialect{dialects.SMB2_DIALECT_2_1_0, dialects.SMB2_DIALECT_3_1_1} {
		t.Run(dialect.String(), func(t *testing.T) {
			_, port := startServer(t, server.Config{
				RequireSigning: true,
				Kerberos: &gssapi.AcceptOptions{
					ServiceName:  messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "filesrv.lab.local"}},
					ServiceRealm: "LAB.LOCAL",
					Keys:         []gssapi.ServiceKey{{EType: messages.ETypeAES256CTSHMACSHA196, Key: serviceKey}},
				},
			})

			silver, err := kerberos.ForgeSilver(kerberos.ForgeOptions{
				Realm: "LAB.LOCAL", Username: "alice", DomainSID: "S-1-5-21-1-2-3", UserRID: 1104,
				Key: serviceKey, KeyEType: messages.ETypeAES256CTSHMACSHA196,
			}, spn)
			if err != nil {
				t.Fatal(err)
			}
			kc := kerberos.NewClient("", "", "127.0.0.1")
			if err := kc.LoadForgedServiceTicket(silver); err != nil {
				t.Fatal(err)
			}

			c := client.NewClientUsingTCPTransport(net.IPv4(127, 0, 0, 1), port)
			c.Profile = c.Profile.Clone()
			c.Profile.Dialects = []dialects.Dialect{dialect}
			defer c.Disconnect()
			if err := c.Connect(net.IPv4(127, 0, 0, 1), port); err != nil {
				t.Fatal(err)
			}
			if err := c.SessionSetupKerberosWithClient(kc, spn); err != nil {
				t.Fatalf("Kerberos session setup: %v", err)
			}
			if err := c.TreeConnect("files"); err != nil {
				t.Fatalf("signed TreeConnect after Kerberos: %v", err)
			}
		})
	}
}

func TestServerRelatedCompound(t *testing.T) {
	for _, config := range []server.Config{{}, {RequireSigning: true}, {RequireEncryption: true}} {
		_, port := startServer(t, config)
		c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		info, err := c.CreateQueryInfoClose("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0,
			1, uint8(infoclass.FileStandardInformation), 0)
		if err != nil {
			t.Fatalf("CREATE+QUERY_INFO+CLOSE (%+v): %v", config, err)
		}
		var standard filesystem.FileStandardInformation
		if err := standard.Unmarshal(info); err != nil || standard.EndOfFile != 11 {
			t.Fatalf("FileStandardInformation = %+v, %v", standard, err)
		}
	}
}

func TestServerUpgradesSMB1Negotiate(t *testing.T) {
	for _, tc := range []struct {
		offered []string
		want    dialects.Dialect
	}{
		{[]string{"NT LM 0.12", "SMB 2.002", "SMB 2.???"}, dialects.SMB2_DIALECT_WILDCARD},
		{[]string{"NT LM 0.12", "SMB 2.002"}, dialects.SMB2_DIALECT_2_0_2},
	} {
		_, port := startServer(t, server.Config{})
		conn := transport.NewTransport("tcp")
		if err := conn.Connect(net.IPv4(127, 0, 0, 1), port); err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		var body []byte
		for _, d := range tc.offered {
			body = append(append(append(body, 0x02), d...), 0)
		}
		request := make([]byte, 32, 35+len(body))
		copy(request, []byte{0xFF, 'S', 'M', 'B', 0x72})
		request = append(request, 0, byte(len(body)), byte(len(body)>>8))
		if _, err := conn.Send(append(request, body...)); err != nil {
			t.Fatal(err)
		}

		raw, err := conn.Receive()
		if err != nil {
			t.Fatalf("no answer to the SMB1 NEGOTIATE: %v", err)
		}
		var answer commands.NegotiateResponse
		if len(raw) < 64 || !bytes.HasPrefix(raw, []byte{0xFE, 'S', 'M', 'B'}) {
			t.Fatalf("answer is not SMB2: % x", raw)
		}
		if _, err := answer.Unmarshal(raw[64:]); err != nil || answer.DialectRevision != tc.want {
			t.Errorf("offered %v: DialectRevision = %s, %v; want %s", tc.offered, answer.DialectRevision, err, tc.want)
		}
	}
}

func TestServerChangeNotify(t *testing.T) {
	for _, config := range []server.Config{{}, {RequireSigning: true}, {RequireEncryption: true}} {
		_, port := startServer(t, config)
		watcher, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		writer, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		root, err := watcher.CreateFile("", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
		if err != nil {
			t.Fatal(err)
		}

		changes := make(chan []filesystem.FileNotifyInformation, 1)
		errs := make(chan error, 1)
		go func() {
			got, err := watcher.ChangeNotify(root, filesystem.FILE_NOTIFY_CHANGE_FILE_NAME, false)
			changes <- got
			errs <- err
		}()
		time.Sleep(100 * time.Millisecond) // let the request reach the server first
		if _, err := writer.CreateFile("created.txt", fileflags.GENERIC_WRITE, 0, fileflags.FILE_CREATE, 0); err != nil {
			t.Fatal(err)
		}

		select {
		case got := <-changes:
			if err := <-errs; err != nil || len(got) == 0 || got[0].FileName != "created.txt" || got[0].Action != filesystem.FileActionAdded {
				t.Fatalf("ChangeNotify (%+v) = %+v, %v; want created.txt added", config, got, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("no change notification (%+v)", config)
		}

		go func() {
			_, err := watcher.ChangeNotify(root, filesystem.FILE_NOTIFY_CHANGE_FILE_NAME, false)
			errs <- err
		}()
		time.Sleep(100 * time.Millisecond)
		if err := watcher.Cancel(); err != nil {
			t.Fatalf("Cancel: %v", err)
		}
		select {
		case err := <-errs:
			if !client.IsChangeNotifyCancelled(err) {
				t.Fatalf("cancelled ChangeNotify returned %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Cancel did not complete the pending ChangeNotify")
		}
	}
}

func TestServerShareModes(t *testing.T) {
	_, port := startServer(t, server.Config{})
	a, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	b, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}

	// A exclusively opens hello.txt for writing (shares nothing).
	if _, err := a.CreateFile("hello.txt", fileflags.GENERIC_WRITE, fileflags.FILE_SHARE_NONE, fileflags.FILE_OPEN, 0); err != nil {
		t.Fatalf("exclusive open: %v", err)
	}
	// B cannot open it at all while A holds it.
	if _, err := b.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0); err == nil {
		t.Fatal("a conflicting open was admitted despite FILE_SHARE_NONE")
	}
}

func TestServerAccessRights(t *testing.T) {
	_, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}

	// Opened for reading only: a write must be refused.
	id, err := c.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatalf("open for reading: %v", err)
	}
	if _, err := c.WriteFile(id, 0, []byte("nope")); err == nil {
		t.Fatal("a write succeeded on a read-only handle")
	}
	c.CloseFile(id)
}

func TestServerReadOnlyShare(t *testing.T) {
	config := server.Config{}
	srv, err := server.NewServer(withTestIdentity(&config))
	if err != nil {
		t.Fatal(err)
	}
	fs := smb1.NewMemoryFileSystem("FILES")
	fs.AddFile("hello.txt", []byte("hello world"))
	if err := srv.AddShare(&server.Share{Name: "files", FS: fs, ReadOnly: true}); err != nil {
		t.Fatal(err)
	}
	listener, _ := transport.ListenTCP("127.0.0.1:0")
	go srv.Serve(listener)
	t.Cleanup(func() { srv.Close() })
	port := listener.Addr().(*net.TCPAddr).Port

	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.CreateFile("new.txt", fileflags.GENERIC_WRITE, 0, fileflags.FILE_CREATE, 0); err == nil {
		t.Fatal("a create succeeded on a read-only share")
	}
	if id, err := c.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0); err != nil {
		t.Fatalf("read on a read-only share: %v", err)
	} else {
		c.CloseFile(id)
	}
}

// interceptEcho answers ECHO itself, to prove the handler chain runs first.
type interceptEcho struct{ seen int }

func (h *interceptEcho) Run(_ *server.Server, _ *server.Connection, w server.ResponseWriter, req *message.Message) bool {
	if req.Header.Command != codes.SMB2_ECHO {
		return false
	}
	h.seen++
	w.WriteResponse(commands.NewEchoResponse())
	return true
}

func TestServerHandlerChain(t *testing.T) {
	config := server.Config{}
	srv, err := server.NewServer(withTestIdentity(&config))
	if err != nil {
		t.Fatal(err)
	}
	handler := &interceptEcho{}
	srv.RegisterHandler(handler)
	fs := smb1.NewMemoryFileSystem("FILES")
	if err := srv.AddShare(&server.Share{Name: "files", FS: fs}); err != nil {
		t.Fatal(err)
	}
	listener, _ := transport.ListenTCP("127.0.0.1:0")
	go srv.Serve(listener)
	t.Cleanup(func() { srv.Close() })
	port := listener.Addr().(*net.TCPAddr).Port

	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Echo(); err != nil {
		t.Fatalf("Echo: %v", err)
	}
	if handler.seen == 0 {
		t.Fatal("the handler did not intercept ECHO")
	}
}

func TestServerReadCreditCharge(t *testing.T) {
	for _, tc := range []struct {
		name    string
		dialect dialects.Dialect
		length  uint32
		charge  uint16
	}{
		{"SMB 3.1.1 one credit for 1 MiB", dialects.SMB2_DIALECT_3_1_1, 1 << 20, 1},
		{"SMB 3.1.1 one credit short", dialects.SMB2_DIALECT_3_1_1, 1<<16 + 1, 1},
		{"SMB 2.0.2 over 64 KiB", dialects.SMB2_DIALECT_2_0_2, 1<<16 + 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, port := startServer(t, server.Config{})
			c, err := dial(t, port, tc.dialect, "Passw0rd!")
			if err != nil {
				t.Fatal(err)
			}
			id, err := c.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0)
			if err != nil {
				t.Fatal(err)
			}

			var msgs []*message.Message
			for _, length := range []uint32{tc.length, 11} {
				read := commands.NewReadRequest()
				read.FileId, read.Length = id, length
				msg := c.NewRequest(read)
				if length == tc.length {
					msg.Header.CreditCharge = tc.charge
				}
				msgs = append(msgs, msg)
			}
			answers, err := c.SendUnrelatedCompound(msgs)
			if err != nil {
				t.Fatal(err)
			}
			if got := answers[0].Header.Status; got != uint32(nt_status.NT_STATUS_INVALID_PARAMETER) {
				t.Errorf("under-charged READ of %d bytes: status 0x%08x, want STATUS_INVALID_PARAMETER", tc.length, got)
			}
			if got := answers[1].Header.Status; got != uint32(nt_status.NT_STATUS_SUCCESS) {
				t.Errorf("correctly charged READ: status 0x%08x, want STATUS_SUCCESS", got)
			}
		})
	}
}

func TestServerCancelRequiresSessionProtection(t *testing.T) {
	for _, config := range []server.Config{{RequireSigning: true}, {RequireEncryption: true}} {
		_, port := startServer(t, config)
		watcher, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		writer, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		root, err := watcher.CreateFile("", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
		if err != nil {
			t.Fatal(err)
		}
		sessionID := watcher.Session.SessionId

		changes := make(chan []filesystem.FileNotifyInformation, 1)
		errs := make(chan error, 1)
		go func() {
			got, err := watcher.ChangeNotify(root, filesystem.FILE_NOTIFY_CHANGE_FILE_NAME, false)
			changes <- got
			errs <- err
		}()
		time.Sleep(100 * time.Millisecond) // let the request reach the server first

		// An unsigned, unencrypted CANCEL naming the pending request (the first
		// AsyncId the connection hands out) must be ignored.
		forged := message.NewMessage()
		forged.Header.SetFlags(flags.SMB2_FLAGS_ASYNC_COMMAND)
		forged.Header.SetAsyncId(1)
		forged.Header.SessionId = sessionID
		forged.SetCommand(commands.NewCancelRequest())
		wire, err := forged.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := watcher.Transport.Send(wire); err != nil {
			t.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)

		if _, err := writer.CreateFile("created.txt", fileflags.GENERIC_WRITE, 0, fileflags.FILE_CREATE, 0); err != nil {
			t.Fatal(err)
		}
		select {
		case got := <-changes:
			if err := <-errs; err != nil || len(got) == 0 || got[0].FileName != "created.txt" {
				t.Fatalf("ChangeNotify (%+v) = %+v, %v; want created.txt, not a cancellation", config, got, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("no change notification (%+v)", config)
		}
	}
}

// changedFileSystem reports a change as soon as a directory is watched.
type changedFileSystem struct {
	*smb1.MemoryFileSystem
}

func (fs changedFileSystem) Watch(path string, recursive bool) (<-chan smb1.ChangeNotification, func(), error) {
	changes := make(chan smb1.ChangeNotification, 1)
	changes <- smb1.ChangeNotification{Action: uint32(filesystem.FileActionAdded), Name: "created.txt"}
	return changes, func() {}, nil
}

func TestServerChangeNotifyInterimFirst(t *testing.T) {
	config := withTestIdentity(&server.Config{})
	srv, err := server.NewServer(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := srv.AddShare(&server.Share{Name: "files", FS: changedFileSystem{smb1.NewMemoryFileSystem("FILES")}}); err != nil {
		t.Fatal(err)
	}
	listener, err := transport.ListenTCP("127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(listener)
	t.Cleanup(func() { srv.Close() })

	c, err := dial(t, listener.Addr().(*net.TCPAddr).Port, dialects.SMB2_DIALECT_2_1_0, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	root, err := c.CreateFile("", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 500; i++ {
		notify := commands.NewChangeNotifyRequest()
		notify.FileId, notify.OutputBufferLength = root, 4096
		notify.CompletionFilter = filesystem.FILE_NOTIFY_CHANGE_FILE_NAME
		wire, err := c.NewRequest(notify).Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := c.Transport.Send(wire); err != nil {
			t.Fatal(err)
		}
		var statuses []uint32
		for len(statuses) < 2 {
			raw, err := c.Transport.Receive()
			if err != nil {
				t.Fatal(err)
			}
			h := header.NewHeader()
			if _, err := h.Unmarshal(raw); err != nil {
				t.Fatal(err)
			}
			statuses = append(statuses, h.Status)
		}
		if statuses[0] != uint32(nt_status.NT_STATUS_PENDING) || statuses[1] != uint32(nt_status.NT_STATUS_SUCCESS) {
			t.Fatalf("CHANGE_NOTIFY %d answered 0x%08x then 0x%08x; want STATUS_PENDING then STATUS_SUCCESS", i, statuses[0], statuses[1])
		}
	}
}

func TestServerChangeNotifyLimit(t *testing.T) {
	fs, port := startServer(t, server.Config{MaxNotifiesPerConnection: 2})
	c, err := dial(t, port, dialects.SMB2_DIALECT_2_1_0, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	root, err := c.CreateFile("", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ|fileflags.FILE_SHARE_WRITE, fileflags.FILE_OPEN, fileflags.FILE_DIRECTORY_FILE)
	if err != nil {
		t.Fatal(err)
	}

	// notify sends a CHANGE_NOTIFY and returns the status it is first answered with.
	notify := func() uint32 {
		t.Helper()
		req := commands.NewChangeNotifyRequest()
		req.FileId, req.OutputBufferLength = root, 4096
		req.CompletionFilter = filesystem.FILE_NOTIFY_CHANGE_FILE_NAME
		wire, err := c.NewRequest(req).Marshal()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := c.Transport.Send(wire); err != nil {
			t.Fatal(err)
		}
		raw, err := c.Transport.Receive()
		if err != nil {
			t.Fatal(err)
		}
		h := header.NewHeader()
		if _, err := h.Unmarshal(raw); err != nil {
			t.Fatal(err)
		}
		return h.Status
	}

	for i := 0; i < 2; i++ {
		if status := notify(); status != uint32(nt_status.NT_STATUS_PENDING) {
			t.Fatalf("CHANGE_NOTIFY %d answered 0x%08x, want STATUS_PENDING", i, status)
		}
	}
	if status := notify(); status != uint32(nt_status.NT_STATUS_INSUFFICIENT_RESOURCES) {
		t.Fatalf("CHANGE_NOTIFY over the limit answered 0x%08x, want STATUS_INSUFFICIENT_RESOURCES", status)
	}

	// Completing the pending requests frees their slots.
	if err := fs.AddFile("created.txt", nil); err != nil {
		t.Fatal(err)
	}
	for completed := 0; completed < 2; completed++ {
		if _, err := c.Transport.Receive(); err != nil {
			t.Fatal(err)
		}
	}
	if status := notify(); status != uint32(nt_status.NT_STATUS_PENDING) {
		t.Fatalf("CHANGE_NOTIFY after completions answered 0x%08x, want STATUS_PENDING", status)
	}
}

func TestServerRefusesReplayedMessageId(t *testing.T) {
	_, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	wire, err := c.NewRequest(commands.NewEchoRequest()).Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Transport.Send(wire); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Transport.Receive(); err != nil {
		t.Fatalf("the ECHO was not answered: %v", err)
	}

	// The same request again reuses its MessageId: the server drops the connection.
	if _, err := c.Transport.Send(wire); err != nil {
		t.Fatal(err)
	}
	c.Transport.SetTimeout(5 * time.Second)
	if raw, err := c.Transport.Receive(); err == nil {
		t.Fatalf("the replayed ECHO was answered: % x", raw[:min(len(raw), 16)])
	}
}

func TestServerDropsUnauthenticatedAndIdleConnections(t *testing.T) {
	const bound = 300 * time.Millisecond
	_, port := startServer(t, server.Config{HandshakeTimeout: bound, IdleTimeout: bound})

	// closedWithin reports whether the server closes conn within d.
	closedWithin := func(conn net.Conn, d time.Duration) bool {
		conn.SetReadDeadline(time.Now().Add(d))
		_, err := conn.Read(make([]byte, 1))
		var timeout net.Error
		return err != nil && !(errors.As(err, &timeout) && timeout.Timeout())
	}

	silent, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatal(err)
	}
	defer silent.Close()
	if !closedWithin(silent, 5*time.Second) {
		t.Error("a connection that never negotiated was kept")
	}

	idle, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * bound)
	if _, err := idle.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0); err == nil {
		t.Error("an authenticated connection holding no open was kept past IdleTimeout")
	}

	busy, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	id, err := busy.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * bound)
	if _, err := busy.ReadFile(id, 0, 5); err != nil {
		t.Errorf("a connection holding an open was dropped while quiet: %v", err)
	}
}

func TestServerEcho(t *testing.T) {
	for _, dialect := range allDialects {
		_, port := startServer(t, server.Config{})
		c, err := dial(t, port, dialect, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Echo(); err != nil {
			t.Errorf("%s: ECHO on a signed session: %v", dialect, err)
		}
	}
	for _, config := range []server.Config{{RequireSigning: true}, {RequireEncryption: true}} {
		_, port := startServer(t, config)
		c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
		if err != nil {
			t.Fatal(err)
		}
		if err := c.Echo(); err != nil {
			t.Errorf("ECHO (%+v): %v", config, err)
		}
	}

	// Before any session, an ECHO carries no SessionId and is answered unsigned.
	_, port := startServer(t, server.Config{})
	c := client.NewClientUsingTCPTransport(net.IPv4(127, 0, 0, 1), port)
	t.Cleanup(func() { c.Disconnect() })
	if err := c.Connect(net.IPv4(127, 0, 0, 1), port); err != nil {
		t.Fatal(err)
	}
	if err := c.Echo(); err != nil {
		t.Errorf("ECHO without a session: %v", err)
	}
}

// TestServerRenameDirectory checks that the client's RenameFile renames a
// directory as well as a file: the old path must not be opened as a non-directory.
func TestServerRenameDirectory(t *testing.T) {
	fs, port := startServer(t, server.Config{})
	if err := fs.AddDirectory("dir1"); err != nil {
		t.Fatal(err)
	}
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_0_2, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}

	if err := c.RenameFile("dir1", "dir2", false); err != nil {
		t.Fatalf("RenameFile(dir1, dir2) = %v, want nil", err)
	}
	if _, err := fs.Stat("dir1"); err == nil {
		t.Error("dir1 still exists after the rename")
	}
	if attr, err := fs.Stat("dir2"); err != nil || !attr.IsDir {
		t.Errorf("Stat(dir2) = %+v, %v, want a directory", attr, err)
	}
	// Files still rename through the same call.
	if err := c.RenameFile("hello.txt", "moved.txt", false); err != nil {
		t.Fatalf("RenameFile(hello.txt, moved.txt) = %v, want nil", err)
	}
	if attr, err := fs.Stat("moved.txt"); err != nil || attr.IsDir {
		t.Errorf("Stat(moved.txt) = %+v, %v, want a file", attr, err)
	}
}

// TestServerRefusesSizeBeyondVolume checks that a SET_INFO end of file or a
// WRITE far beyond what the in-memory volume holds is answered with
// STATUS_DISK_FULL rather than allocated, and that the connection carries on.
//
// Without the cap the server process dies of an out-of-memory fatal error here,
// taking the test binary with it; fs_mem_test.go is the clean fail-before
// signal for the backend itself.
func TestServerRefusesSizeBeyondVolume(t *testing.T) {
	fs, port := startServer(t, server.Config{})
	c, err := dial(t, port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
	if err != nil {
		t.Fatal(err)
	}
	id, err := c.CreateFile("hello.txt", fileflags.GENERIC_READ|fileflags.GENERIC_WRITE, 0, fileflags.FILE_OPEN, 0)
	if err != nil {
		t.Fatalf("open hello.txt: %v", err)
	}

	eof := filesystem.FileEndOfFileInformation{EndOfFile: 1 << 40}
	buffer, err := eof.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	setInfo := commands.NewSetInfoRequest()
	setInfo.InfoType = commands.SMB2_0_INFO_FILE
	setInfo.FileInfoClass = uint8(infoclass.FileEndOfFileInformation)
	setInfo.FileId, setInfo.Buffer = id, buffer
	write := commands.NewWriteRequest()
	write.FileId, write.Offset, write.Data = id, 1<<40, []byte("x")

	answers, err := c.SendUnrelatedCompound([]*message.Message{c.NewRequest(setInfo), c.NewRequest(write)})
	if err != nil {
		t.Fatal(err)
	}
	for i, what := range []string{"SET_INFO EndOfFile 1<<40", "WRITE at offset 1<<40"} {
		if got := answers[i].Header.Status; got != uint32(nt_status.NT_STATUS_DISK_FULL) {
			t.Errorf("%s: status 0x%08x, want STATUS_DISK_FULL", what, got)
		}
	}

	// The file is untouched and the connection still answers.
	if attr, err := fs.Stat("hello.txt"); err != nil || attr.Size != 11 {
		t.Errorf("Stat(hello.txt) = %+v, %v; want 11 bytes", attr, err)
	}
	if data, err := c.ReadFile(id, 0, 64); err != nil || string(data) != "hello world" {
		t.Errorf("ReadFile after the refusals = %q, %v", data, err)
	}
	if err := c.SetEndOfFile(id, 5); err != nil {
		t.Errorf("SetEndOfFile(5): %v", err)
	}
	c.CloseFile(id)
}

// TestServerReadOnlyFileCannotBeDeleted asserts a file carrying
// FILE_ATTRIBUTE_READONLY survives a delete-on-close open and a
// FileDispositionInformation set, each refused with STATUS_CANNOT_DELETE as
// [MS-FSA] sections 2.1.5.1.2.1 and 2.1.5.14.3 require. The local backend
// removed such a file when the handle closed, since unlinking a host file needs
// write access to its directory, not to the file.
func TestServerReadOnlyFileCannotBeDeleted(t *testing.T) {
	newLocal := func(t *testing.T) (smb1.FileSystem, func() bool) {
		root := t.TempDir()
		host := filepath.Join(root, "keep.txt")
		if err := os.WriteFile(host, []byte("precious"), 0o444); err != nil {
			t.Fatal(err)
		}
		fs, err := smb1.NewLocalFileSystem(root, "LOCAL")
		if err != nil {
			t.Fatal(err)
		}
		return fs, func() bool {
			_, err := os.Stat(host)
			return err == nil
		}
	}
	newMemory := func(t *testing.T) (smb1.FileSystem, func() bool) {
		fs := smb1.NewMemoryFileSystem("FILES")
		if err := fs.AddFile("keep.txt", []byte("precious")); err != nil {
			t.Fatal(err)
		}
		if err := fs.SetAttr("keep.txt", smb1.FileAttr{ReadOnly: true}, smb1.AttrMask{ReadOnly: true}); err != nil {
			t.Fatal(err)
		}
		return fs, func() bool {
			_, err := fs.Stat("keep.txt")
			return err == nil
		}
	}

	const shareAll = fileflags.FILE_SHARE_READ | fileflags.FILE_SHARE_WRITE | fileflags.FILE_SHARE_DELETE
	deletes := map[string]func(c *client.Client) error{
		"delete-on-close": func(c *client.Client) error {
			id, err := c.CreateFile("keep.txt", fileflags.GENERIC_READ|fileflags.DELETE, shareAll,
				fileflags.FILE_OPEN, fileflags.FILE_DELETE_ON_CLOSE)
			if err != nil {
				return err
			}
			return c.CloseFile(id)
		},
		"disposition": func(c *client.Client) error {
			id, err := c.CreateFile("keep.txt", fileflags.GENERIC_READ|fileflags.DELETE, shareAll, fileflags.FILE_OPEN, 0)
			if err != nil {
				return err
			}
			defer c.CloseFile(id)
			return c.SetDeleteOnClose(id, true)
		},
	}
	backends := map[string]func(t *testing.T) (smb1.FileSystem, func() bool){"local": newLocal, "memory": newMemory}
	for backendName, newBackend := range backends {
		for deleteName, remove := range deletes {
			t.Run(backendName+"/"+deleteName, func(t *testing.T) {
				fs, present := newBackend(t)
				config := server.Config{}
				srv, err := server.NewServer(withTestIdentity(&config))
				if err != nil {
					t.Fatal(err)
				}
				if err := srv.AddShare(&server.Share{Name: "files", FS: fs}); err != nil {
					t.Fatal(err)
				}
				listener, err := transport.ListenTCP("127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				go srv.Serve(listener)
				t.Cleanup(func() { srv.Close() })
				c, err := dial(t, listener.Addr().(*net.TCPAddr).Port, dialects.SMB2_DIALECT_3_1_1, "Passw0rd!")
				if err != nil {
					t.Fatal(err)
				}

				err = remove(c)
				if err == nil {
					t.Error("deleting a read-only file succeeded")
				} else if !strings.Contains(err.Error(), "0xc0000121") {
					t.Errorf("deleting a read-only file failed with %v, want STATUS_CANNOT_DELETE", err)
				}
				if !present() {
					t.Error("the read-only file was deleted")
				}
			})
		}
	}
}
