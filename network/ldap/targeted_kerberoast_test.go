package ldap

import (
	"errors"
	"reflect"
	"testing"
)

// fakeSPNStore is an in-memory spnEditor for exercising the targeted-roast
// set -> roast -> restore logic without a live directory.
type fakeSPNStore struct {
	spns    map[string][]string
	addErr  error
	delErr  error
	getErr  error
	addCall int
	delCall int
}

func (f *fakeSPNStore) GetServicePrincipalNames(dn string) ([]string, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return append([]string(nil), f.spns[dn]...), nil
}

func (f *fakeSPNStore) AddServicePrincipalName(dn, spn string) error {
	f.addCall++
	if f.addErr != nil {
		return f.addErr
	}
	f.spns[dn] = append(f.spns[dn], spn)
	return nil
}

func (f *fakeSPNStore) RemoveServicePrincipalName(dn, spn string) error {
	f.delCall++
	if f.delErr != nil {
		return f.delErr
	}
	kept := f.spns[dn][:0]
	for _, s := range f.spns[dn] {
		if s != spn {
			kept = append(kept, s)
		}
	}
	f.spns[dn] = kept
	return nil
}

// TestTargetedKerberoastSetRestore verifies the SPN is added, the roast runs
// while it is present, and the attribute is restored to its exact prior state.
func TestTargetedKerberoastSetRestore(t *testing.T) {
	dn := "CN=svc,DC=corp,DC=local"
	store := &fakeSPNStore{spns: map[string][]string{dn: {"MSSQLSvc/db.corp.local"}}}
	spn := "HTTP/roast.corp.local"

	sawSPN := false
	err := targetedKerberoast(store, dn, spn, func() error {
		sawSPN = containsSPN(store.spns[dn], spn)
		return nil
	})
	if err != nil {
		t.Fatalf("targetedKerberoast: %v", err)
	}
	if !sawSPN {
		t.Error("temp SPN was not present during the roast callback")
	}
	if store.addCall != 1 || store.delCall != 1 {
		t.Errorf("expected one add and one remove, got add=%d del=%d", store.addCall, store.delCall)
	}
	if want := []string{"MSSQLSvc/db.corp.local"}; !reflect.DeepEqual(store.spns[dn], want) {
		t.Errorf("SPNs not restored: got %v want %v", store.spns[dn], want)
	}
}

// TestTargetedKerberoastRestoreOnError verifies the SPN is still removed when the
// roast callback fails, and the roast error is surfaced.
func TestTargetedKerberoastRestoreOnError(t *testing.T) {
	dn := "CN=svc,DC=corp,DC=local"
	store := &fakeSPNStore{spns: map[string][]string{dn: {}}}
	spn := "HTTP/roast.corp.local"
	roastErr := errors.New("kdc said no")

	err := targetedKerberoast(store, dn, spn, func() error { return roastErr })
	if !errors.Is(err, roastErr) {
		t.Fatalf("expected roast error, got %v", err)
	}
	if store.delCall != 1 {
		t.Errorf("expected the temp SPN to be removed on error, del=%d", store.delCall)
	}
	if len(store.spns[dn]) != 0 {
		t.Errorf("SPNs not restored after error: %v", store.spns[dn])
	}
}

// TestTargetedKerberoastAlreadyPresent verifies that if the SPN already exists it
// is neither added nor removed (the account is left exactly as found).
func TestTargetedKerberoastAlreadyPresent(t *testing.T) {
	dn := "CN=svc,DC=corp,DC=local"
	spn := "HTTP/roast.corp.local"
	store := &fakeSPNStore{spns: map[string][]string{dn: {"http/ROAST.corp.local"}}} // case-insensitive match

	err := targetedKerberoast(store, dn, spn, func() error { return nil })
	if err != nil {
		t.Fatalf("targetedKerberoast: %v", err)
	}
	if store.addCall != 0 || store.delCall != 0 {
		t.Errorf("expected no add/remove when SPN already present, add=%d del=%d", store.addCall, store.delCall)
	}
}

// TestTargetedKerberoastAddFails verifies that a failure to set the SPN aborts
// before the roast and does not attempt a restore.
func TestTargetedKerberoastAddFails(t *testing.T) {
	dn := "CN=svc,DC=corp,DC=local"
	store := &fakeSPNStore{spns: map[string][]string{dn: {}}, addErr: errors.New("access denied")}

	ran := false
	err := targetedKerberoast(store, dn, "HTTP/x.corp.local", func() error { ran = true; return nil })
	if err == nil {
		t.Fatal("expected an error when the SPN add fails")
	}
	if ran {
		t.Error("roast callback should not run when the SPN could not be set")
	}
	if store.delCall != 0 {
		t.Error("no restore should be attempted when nothing was added")
	}
}
