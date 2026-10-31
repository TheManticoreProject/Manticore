package server_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"net"
	"testing"

	kerberos "github.com/TheManticoreProject/Manticore/network/kerberos/v5"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/gssapi"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/server"
	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// TestServerKerberosAES256Encryption runs an encrypted SMB 3.1.1 session with an
// AES-256 cipher over Kerberos AES256, whose GSS key is 32 bytes. Client and
// server must both derive the cipher keys from the full key (MS-SMB2 3.2.5.3.1,
// 3.3.5.5.3) rather than from its 16-byte truncation.
func TestServerKerberosAES256Encryption(t *testing.T) {
	const spn = "cifs/filesrv.lab.local"
	serviceKey := bytes.Repeat([]byte{0x5A}, 32)

	for _, cipher := range []uint16{commands.SMB2_ENCRYPTION_AES256_GCM, commands.SMB2_ENCRYPTION_AES256_CCM} {
		_, port := startServer(t, server.Config{
			RequireEncryption: true,
			Kerberos: &gssapi.AcceptOptions{
				ServiceName:  messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"cifs", "filesrv.lab.local"}},
				ServiceRealm: "LAB.LOCAL",
				Keys:         []gssapi.ServiceKey{{EType: messages.ETypeAES256CTSHMACSHA196, Key: serviceKey}},
			},
		})
		silver, err := kerberos.ForgeSilver(kerberos.ForgeOptions{
			Realm: "LAB.LOCAL", Username: "alice", DomainSID: "S-1-5-21-1-2-3", UserRID: 1104,
			Key: serviceKey, KeyEType: messages.ETypeAES256CTSHMACSHA196,
			SessionEType: messages.ETypeAES256CTSHMACSHA196,
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
		c.Profile.Dialects = []dialects.Dialect{dialects.SMB2_DIALECT_3_1_1}
		c.Profile.Ciphers = []uint16{cipher}
		defer c.Disconnect()
		if err := c.Connect(net.IPv4(127, 0, 0, 1), port); err != nil {
			t.Fatal(err)
		}
		if err := c.SessionSetupKerberosWithClient(kc, spn); err != nil {
			t.Fatalf("cipher 0x%04x: Kerberos session setup: %v", cipher, err)
		}
		if c.Connection.Cipher != cipher || !c.IsEncryptionActive() {
			t.Fatalf("cipher = 0x%04x, encrypted %v; want 0x%04x, encrypted", c.Connection.Cipher, c.IsEncryptionActive(), cipher)
		}

		// The encryption key is the SP800-108 HMAC-SHA256 derivation (L = 256)
		// over the full 32-byte GSS key.
		full := c.Session.FullSessionKey
		if len(full) != 32 {
			t.Fatalf("FullSessionKey is %d bytes, want the 32-byte AES256 key", len(full))
		}
		mac := hmac.New(sha256.New, full)
		mac.Write([]byte{0, 0, 0, 1})
		mac.Write([]byte("SMBC2SCipherKey\x00\x00"))
		mac.Write(c.Session.PreauthHash)
		mac.Write([]byte{0, 0, 1, 0})
		if want := mac.Sum(nil); !bytes.Equal(c.Session.EncryptionKey, want) {
			t.Fatalf("cipher 0x%04x: EncryptionKey = %X, want %X", cipher, c.Session.EncryptionKey, want)
		}

		// Both ends must agree for the encrypted session to carry requests.
		if err := c.TreeConnect("files"); err != nil {
			t.Fatalf("cipher 0x%04x: encrypted TreeConnect: %v", cipher, err)
		}
		if _, err := c.CreateFile("hello.txt", fileflags.GENERIC_READ, fileflags.FILE_SHARE_READ, fileflags.FILE_OPEN, 0); err != nil {
			t.Fatalf("cipher 0x%04x: encrypted Create: %v", cipher, err)
		}
	}
}
