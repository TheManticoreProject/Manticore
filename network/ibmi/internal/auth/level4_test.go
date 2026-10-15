package auth

import (
	"encoding/hex"
	"errors"
	"testing"
)

// Password level 4 known-answer vectors, produced on 2026-10-15 by calling
// generateSaltForPasswordLevel4, generatePwdTokenForPasswordLevel4 and
// generateSha512Substitute of a JTOpen 21.0.7 jar (SHA-256
// d2c06bba78d20b4733a3fabc9fe6ccabac337f7298bea95c6766108fa0b420f8) through
// reflection, as a black box with synthetic inputs, client seed 0102..08,
// server seed 1112..18 and sequence 1. They are not from a live system.
var level4Vectors = []struct {
	profile, password  string
	salt, token, subst string
}{
	{"DEMOUSR", "Synthetic42Level4",
		"f65a497f79779730b11810b9ad55e78529183c1d71c24a11734471b8b3e812a4",
		"9841e93b407de0dda910b5f4e3504eb297554592a8e01876bfe9d46996d3b163ef9b5fb2963c6b0d1735592baa3112034adad8214dc8b49cdd286addf7512336",
		"313baddbe33440fe4d3ac142ff8399a8c2c9eb148de1e1fea5d9c522e17cfcec9f63a70b8a2163b06e1a16e87374b3eb36d150ea6d498d9048f0e2a26713fdc5"},
	// Shorter than four characters: the salt tail is space-padded.
	{"DEMOUSR", "Ab1",
		"a51bebc6df3db4ac85b166abb245a202108224f05f20c5514fadc2882096de4f",
		"c0108830e7d5510cb169d397c4fb82b09267b790f807d94ac376d48c1f2ca741a47f9b9b5171d964a4874d04e641f4392f78b5b548f0a89fe727f581c8787dd0",
		"fa66912f0e6d74f3ce24abb3b5a2fafb43c7f73160e3b5814cffb9c43a990fd3dcd72bdc61dbd7778f018a9939fa94e4ef60fe4fe3f9f9cc3c15a3ea6f9c6a73"},
	// Ten-character profile and non-ASCII password: PBKDF2 takes UTF-8.
	{"ABCDEFGHIJ", "pässwörd99",
		"b238ee30234caf73900b34f94a63cbadea63f29d6184e99324ec8cc191c52bef",
		"81b1206b0c79d8f8049bb0525f71567ac14e95fe308e8185abf7e987b674e18273fbb8f2fafea79ff072b763cf3779158bd8a633432a906013649560636f5420",
		"51d26892ff24c3608cc25c58021d2e2ef382f8cd655287ae5db246727464f19f07a64470afd6394cfc94e80086a32de633161ebd298e4e60cf0915218beb91cb"},
}

func TestLevel4KnownAnswers(t *testing.T) {
	for _, v := range level4Vectors {
		p, err := EncodeProfile(v.profile)
		if err != nil {
			t.Fatal(err)
		}
		pw8, pw16, err := level4Password([]byte(v.password))
		if err != nil {
			t.Fatal(err)
		}
		salt := level4Salt(p.UTF16BE(), pw16)
		if got := hex.EncodeToString(salt[:]); got != v.salt {
			t.Errorf("%s/%s salt = %s", v.profile, v.password, got)
		}
		token := level4Token(pw8, salt)
		if got := hex.EncodeToString(token); got != v.token {
			t.Errorf("%s/%s token = %s", v.profile, v.password, got)
		}
		sub := sha512Substitute(token, p.UTF16BE(), kaClientSeed, kaServerSeed)
		if got := hex.EncodeToString(sub[:]); got != v.subst {
			t.Errorf("%s/%s substitute = %s", v.profile, v.password, got)
		}

		c, _ := NewPasswordCredential(v.profile, []byte(v.password))
		s, err := Derive(c, FamilySHA512, kaClientSeed, kaServerSeed)
		if err != nil || s.CredentialType != CredentialTypeSHA512 || hex.EncodeToString(s.Value) != v.subst {
			t.Errorf("%s/%s Derive = type %02x %x, %v", v.profile, v.password, s.CredentialType, s.Value, err)
		}
	}
}

func TestLevel4PasswordRules(t *testing.T) {
	c, _ := NewPasswordCredential("DEMOUSR", []byte("Synthetic42Level4  　"))
	s, err := Derive(c, FamilySHA512, kaClientSeed, kaServerSeed)
	if err != nil || hex.EncodeToString(s.Value) != level4Vectors[0].subst {
		t.Errorf("trailing blanks must be trimmed as at levels 2/3: %x, %v", s.Value, err)
	}
	for _, pw := range []string{"*NONE", "   "} {
		c, _ := NewPasswordCredential("DEMOUSR", []byte(pw))
		if _, err := Derive(c, FamilySHA512, kaClientSeed, kaServerSeed); !errors.Is(err, ErrInvalidPassword) {
			t.Errorf("password %q: %v", pw, err)
		}
	}
}
