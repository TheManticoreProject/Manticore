package attacks

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// TestFormatASREPHashJohnWithSalt checks the KDC-reported salt replaces the
// default UPPER(realm)+user salt in the John AES layout, that an empty salt keeps
// the default, and that RC4 ignores the salt.
func TestFormatASREPHashJohnWithSalt(t *testing.T) {
	edata := "42e34732112be6cec1532177a6c93af5"
	cksum := "420973360c2e907b9053f1db"
	cipher := append(mustHex(t, edata), mustHex(t, cksum)...)

	for _, tc := range []struct {
		name, user, salt, want string
	}{
		{"computer account", "HOST$", "EXAMPLE.COMhosthost.example.com", "$krb5asrep$18$EXAMPLE.COMhosthost.example.com$" + edata + "$" + cksum},
		{"renamed user", "newname", "EXAMPLE.COMoldname", "$krb5asrep$18$EXAMPLE.COMoldname$" + edata + "$" + cksum},
		{"default salt", "luser", "", "$krb5asrep$18$EXAMPLE.COMluser$" + edata + "$" + cksum},
	} {
		got, err := FormatASREPHashJohnWithSalt(tc.user, "example.com", tc.salt, iana.ETypeAES256CTSHMACSHA196, cipher)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got != tc.want {
			t.Errorf("%s:\n got  %s\n want %s", tc.name, got, tc.want)
		}
	}

	rc4, err := FormatASREPHashJohnWithSalt("victim", "corp.local", "ignored", iana.ETypeRC4HMAC, append(mustHex(t, cksum+"00000000"), mustHex(t, edata)...))
	if err != nil {
		t.Fatal(err)
	}
	if want := "$krb5asrep$23$" + cksum + "00000000$" + edata; rc4 != want {
		t.Errorf("RC4:\n got  %s\n want %s", rc4, want)
	}
}

// TestFormatJohnRefusesDollarSalt checks an AES salt containing John's '$' field
// separator is refused rather than producing a hash John splits wrongly, whether
// it is the default salt of a '$'-terminated name or an explicit one.
func TestFormatJohnRefusesDollarSalt(t *testing.T) {
	cipher := make([]byte, 48)
	if _, err := FormatASREPHashJohn("HOST$", "example.com", iana.ETypeAES256CTSHMACSHA196, cipher); err == nil {
		t.Error("FormatASREPHashJohn accepted the default salt of a computer account")
	}
	if _, err := FormatASREPHashJohnWithSalt("user", "example.com", "EXAMPLE.COMa$b", iana.ETypeAES128CTSHMACSHA196, cipher); err == nil {
		t.Error("FormatASREPHashJohnWithSalt accepted a salt containing '$'")
	}
	if _, err := FormatTGSHashJohn("svc$", "example.com", "http/web", iana.ETypeAES256CTSHMACSHA196, cipher); err == nil {
		t.Error("FormatTGSHashJohn accepted a default salt containing '$'")
	}
	if _, err := FormatASREPHashJohn("HOST$", "example.com", iana.ETypeRC4HMAC, cipher); err != nil {
		t.Errorf("RC4 carries no salt, so a '$' user name must not be refused: %v", err)
	}
}

// TestFormatTGSHashJohnWithSalt checks a caller-supplied service-account salt
// replaces the default in the John AES TGS layout.
func TestFormatTGSHashJohnWithSalt(t *testing.T) {
	edata := "00112233445566778899aabbccddeeff"
	cksum := "0102030405060708090a0b0c"
	cipher := append(mustHex(t, edata), mustHex(t, cksum)...)
	got, err := FormatTGSHashJohnWithSalt("svc_new", "corp.local", "http/web", "CORP.LOCALsvc_old", iana.ETypeAES256CTSHMACSHA196, cipher)
	if err != nil {
		t.Fatal(err)
	}
	if want := "$krb5tgs$18$CORP.LOCALsvc_old$" + cksum + "$" + edata; got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

// johnTGSAES18 is John Jumbo's own krb5tgs-sha1 test vector
// (src/krb5_tgsrep_common_plug.c), for the password "openwall@123".
const johnTGSAES18 = "$krb5tgs$18$EXAMPLE.NETlulu$f72d514cff17d1a7f6213816$95613ea2ea4ef8080d9b5957b09a71a1852454c71a1e52bcd246bc118ae2a539f7f919decba1fccad911c85772f4be450f68804e7efe64c72b78a0fbdad4b6fdce837c0d3acd56a723c702994dbf5a51a38237c263e6c0ca5d691de478a868e0372902d069358accf2f84b1f4f3ca2ffe15c40af4e654292aa52759261e87089ac0a9daaf13dd02c0cd751a8303befecf05add1196b55eea7e86cfa95e4c1f92e5640610c44013ce5e7587c9d3c697523b9f2d3a54ad0dda841e7d07a7691ad369657bc4cc12ec2f514a91ab67b5ec0d2f70a15241cc7d988bb8e515da283fbc493a028a31ace67ba610e43788a611d2281a008b217eb935ae7f36dd75fa43dd8d5f9dd657f1073365076b81ae59faf4189284b7bd883b64be34e163904815a43122d175a52647205f9a031d8722608106fecdb7f495ecaa580e6e40063bfd8e94b1240ffb19f55c710ecf4fef135477672c2e389ff6a29b08de45d14e66f321a7d77106c431882c4b5a444640454166d2d7631aa24f7afefb69180b2e51d32f536aebbea3481dcf0329c9618916a823d4aed8b909b573c9efe0847d5f772abe7790a5a57d9ac846da162a198190f60a27407bfbe8ea38988c9f4b258ed7374d1a7dd75aad51797a6d92e1045b85cba72294466a377a91e4271538bcbef67961691671be979b3cc4702743d3ce2839c03108fca779f6576e058eb3e67429817ae21574665a4aad1988e374d7142998502a7ea00db98d0bb102fe56ccd3741e0cddff5d0ce193facf6916995741ebbc77e4034b083d900dd5aa4ed7fc1713579ffb23563372703e8f2dedffca2d0553e1082379c7612809f4c5ec0729e5bc7fea7bb223c43249f98f0c59d1f3bc4d71460568ed4d6ebea7582bcd0a073bea283276917665a49468e9211e7a780c9809dc67d50ded82f23cef606f6110428ac43361dbd79afb608ed1407e5ca7014ccaa060461d180c73c25259fac23e4af825d868a6f5813da9fdbcf94f3d63b9726b92ff6d15d3f00389bd6d0b035d1b1ca868b96f842fa4850187c76a4e5828c7951ab9a2bb9bb4f0594e3233fa09b755e8e5da73133f3bd00d0a7562359395ab7f3aeb569dbf20a7cd458813754fbd5438049896aac9b75f7e361c64a232ac4b54d95b5d153f86c8e3b9250e5c218797e3bd73e83d3cec0030475d6ffc253e365eb0f2c1920e4b3c276b2291615aeb196ed45f0d89a0db02434a4e219c665f6fa92631576322e0770debc2fee762c76646a654c0e4f77afdc0d714f2412610abd936ecc80f9aff2ed4f736f4e9b4c0dd1750b50301be533475a0144b1aa78967fbded21b1021b56d3692212080356e6c668b8098b776dd24fa886a3cc70826254ecd39923d1bceeb4686aba928337fb15d1e76a033166d648f30843efd25e629612f9a5b0146b0f348c0bd9c"

// TestFormatTGSHashJohnMatchesJohnVector rebuilds John's AES TGS vector from its
// ciphertext and salt: the checksum precedes the encrypted data in this layout.
func TestFormatTGSHashJohnMatchesJohnVector(t *testing.T) {
	f := strings.Split(johnTGSAES18, "$")
	// "", "krb5tgs", "18", salt, checksum, edata
	cipher := append(mustHex(t, f[5]), mustHex(t, f[4])...)
	got, err := FormatTGSHashJohnWithSalt("lulu", "example.net", "", "EXAMPLE.NETlulu", iana.ETypeAES256CTSHMACSHA196, cipher)
	if err != nil {
		t.Fatal(err)
	}
	if got != johnTGSAES18 {
		t.Errorf("AES TGS John hash differs from John's vector:\n got  %.90s...\n want %.90s...", got, johnTGSAES18)
	}
	if def, err := FormatTGSHashJohn("lulu", "example.net", "", iana.ETypeAES256CTSHMACSHA196, cipher); err != nil || def != johnTGSAES18 {
		t.Errorf("default salt: %v, hash matches vector: %v", err, def == johnTGSAES18)
	}
}
