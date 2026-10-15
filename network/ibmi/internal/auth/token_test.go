package auth

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func tokenBytes() []byte { return bytes.Repeat([]byte{0xC7}, ProfileTokenLength) }

func TestNewProfileToken(t *testing.T) {
	tok, err := NewProfileToken(tokenBytes(), TokenMultipleUse, time.Time{})
	if err != nil || tok.Type() != TokenMultipleUse || !tok.ExpiresAt().IsZero() {
		t.Fatalf("NewProfileToken = %v, %v", tok, err)
	}
	for name, tc := range map[string]struct {
		value []byte
		typ   ProfileTokenType
	}{
		"short":  {make([]byte, 31), TokenSingleUse},
		"long":   {make([]byte, 33), TokenSingleUse},
		"type 0": {tokenBytes(), 0},
		"type 4": {tokenBytes(), 4},
	} {
		if _, err := NewProfileToken(tc.value, tc.typ, time.Time{}); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestProfileTokenMaterial(t *testing.T) {
	src := tokenBytes()
	tok, _ := NewProfileToken(src, TokenMultipleUse, time.Time{})
	src[0] = 0 // the token copied its input
	for i := 0; i < 2; i++ {
		s, err := Derive(tok, FamilyDES, [8]byte{}, [8]byte{})
		if err != nil || !s.IsToken() || s.CredentialType != CredentialTypeProfileToken || !bytes.Equal(s.Value, tokenBytes()) || s.Profile != ([10]byte{}) {
			t.Fatalf("use %d: %+v, %v", i, s, err)
		}
		s.Value[0] = 0 // material is a copy
	}
}

func TestSingleUseTokenIsConsumedLocally(t *testing.T) {
	tok, _ := NewProfileToken(tokenBytes(), TokenSingleUse, time.Time{})
	if _, err := Derive(tok, FamilyDES, [8]byte{}, [8]byte{}); err != nil {
		t.Fatal(err)
	}
	if _, err := Derive(tok, FamilyDES, [8]byte{}, [8]byte{}); !errors.Is(err, ErrTokenConsumed) {
		t.Errorf("second use = %v", err)
	}
}

func TestExpiredAndDestroyedTokens(t *testing.T) {
	expired, _ := NewProfileToken(tokenBytes(), TokenMultipleUse, time.Now().Add(-time.Second))
	if _, err := Derive(expired, FamilyDES, [8]byte{}, [8]byte{}); !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expired = %v", err)
	}
	tok, _ := NewProfileToken(tokenBytes(), TokenMultipleUse, time.Now().Add(time.Hour))
	tok.Destroy()
	tok.Destroy()
	if _, err := Derive(tok, FamilyDES, [8]byte{}, [8]byte{}); !errors.Is(err, ErrCredentialDestroyed) {
		t.Errorf("destroyed = %v", err)
	}
	if tok.value != ([ProfileTokenLength]byte{}) {
		t.Error("Destroy left token bytes")
	}
}

func TestProfileTokenRedaction(t *testing.T) {
	tok, _ := NewProfileToken(tokenBytes(), TokenMultipleUse, time.Time{})
	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X", "%d"} {
		out := fmt.Sprintf(verb, tok)
		if out != redactedToken || strings.Contains(strings.ToLower(out), "c7c7") {
			t.Errorf("%s = %q", verb, out)
		}
	}
}

func TestDeriveRejectsNilCredentials(t *testing.T) {
	if _, err := Derive(nil, FamilyDES, [8]byte{}, [8]byte{}); err == nil {
		t.Error("nil credential accepted")
	}
	var nilPassword *PasswordCredential
	if _, err := Derive(nilPassword, FamilyDES, [8]byte{}, [8]byte{}); err == nil {
		t.Error("nil *PasswordCredential accepted")
	}
}
