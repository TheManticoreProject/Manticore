package kerberos

import (
	"encoding/asn1"
	"net"
	"sync"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// TestGetTGTUsesETypeInfo2FromETypeNoSupp drives GetTGT against a KDC that, like
// a Windows Server 2025 KDC, rejects the AES-SHA2 etypes with
// KDC_ERR_ETYPE_NOSUPP and names the AES256 etype and the account's salt in that
// error's ETYPE-INFO2. The account's salt is not realm+username, so a pre-auth
// with the default salt is a failed logon on the KDC. GetTGT must take the salt
// from the ETYPE_NOSUPP error and never send a pre-auth the KDC cannot decrypt.
func TestGetTGTUsesETypeInfo2FromETypeNoSupp(t *testing.T) {
	const kdcSalt = "WIN-UK5B7SSQIUDAdministrator"
	c := NewClient("Administrator", "TMP-W-2025.local", "127.0.0.1").WithPassword("Passw0rd!")
	if c.cred.DefaultSalt() == kdcSalt {
		t.Fatal("test precondition: KDC salt must differ from the default realm+username salt")
	}
	key, err := c.cred.Key(messages.ETypeAES256CTSHMACSHA196, kdcSalt, nil)
	if err != nil {
		t.Fatal(err)
	}

	info := messages.ETypeInfo2{{EType: messages.ETypeAES256CTSHMACSHA196, Salt: kdcSalt}}
	infoDER, err := info.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	eData, err := asn1.Marshal([]messages.PAData{{PADataType: messages.PAETypeInfo2, PADataValue: infoDER}})
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	c.kdcHost = conn.LocalAddr().String()

	var mu sync.Mutex
	var wrongKey, rightKey int
	go func() {
		buf := make([]byte, 65535)
		for {
			size, peer, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			var req messages.ASReq
			if _, err := req.Unmarshal(buf[:size]); err != nil {
				return
			}
			code := iana.ErrETypeNoSupp
			for _, pa := range req.PAData {
				if pa.PADataType != messages.PAEncTimestamp {
					continue
				}
				var enc messages.EncryptedData
				if _, err := asn1.Unmarshal(pa.PADataValue, &enc); err != nil {
					return
				}
				if enc.EType != messages.ETypeAES256CTSHMACSHA196 {
					continue
				}
				mu.Lock()
				if _, err := kerbcrypto.Decrypt(enc.EType, key, kerbcrypto.KeyUsageASReqPAEncTimestamp, enc.Cipher); err != nil {
					wrongKey++
					code = messages.ErrPreauthFailed
				} else {
					// Pre-auth verified: end the exchange with an unrelated error,
					// since only the pre-auth attempts are under test.
					rightKey++
					code = 60 // KRB_ERR_GENERIC
				}
				mu.Unlock()
			}
			krbErr := messages.KRBError{ErrorCode: code, Realm: c.realm, EData: eData}
			out, err := krbErr.Marshal()
			if err != nil {
				return
			}
			if _, err := conn.WriteToUDP(out, peer); err != nil {
				return
			}
		}
	}()

	_ = c.GetTGT()

	mu.Lock()
	defer mu.Unlock()
	if wrongKey != 0 {
		t.Errorf("GetTGT sent %d AES256 pre-auth(s) with the default salt; the KDC logs each as a failed logon", wrongKey)
	}
	if rightKey != 1 {
		t.Errorf("GetTGT sent %d AES256 pre-auth(s) with the KDC-advertised salt, want 1", rightKey)
	}
}
