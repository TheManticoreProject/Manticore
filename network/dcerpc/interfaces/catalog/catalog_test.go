package catalog

import (
	"testing"
)

// spoolssUUID / efsrUUID name a couple of seed entries used across tests.
var (
	spoolssUUID = mustGUID("12345678-1234-abcd-ef00-0123456789ab")
	unknownUUID = mustGUID("00000000-0000-0000-0000-0000000000ff")
)

func TestLocalInterfacesPresent(t *testing.T) {
	// A few of the local/host-service interfaces added from endpoint enumeration.
	cases := []struct {
		uuid          string
		major, minor  uint16
		name, service string
	}{
		{"b25a52bf-e5dd-4f4a-aea6-8ca7272a0e86", 2, 0, "KeyIso", "KeyIso"},
		{"0767a036-0d22-48aa-ba69-b619480f38cb", 1, 0, "PcaSvc", "PcaSvc"},
		{"0e3ae095-8a23-48f4-9782-03c1594a890e", 1, 0, "NgcKsp", "NgcSvc"},
		{"378e52b0-c0a9-11cf-822d-00aa0051e40f", 1, 0, "SASec", "Schedule"},
	}
	for _, c := range cases {
		e, ok := Lookup(mustGUID(c.uuid), c.major, c.minor)
		if !ok || e.Name != c.name || e.Service != c.service {
			t.Errorf("Lookup(%s) = (%q/%q, %v), want (%q/%q)", c.uuid, e.Name, e.Service, ok, c.name, c.service)
		}
	}
	// Service search reaches the new entries.
	if got := SearchByService("NgcSvc"); len(got) < 5 {
		t.Errorf("SearchByService(NgcSvc) = %d, want >= 5", len(got))
	}
}

func TestDefault_BuildsAndIsConsistent(t *testing.T) {
	db := Default() // panics if the seed table is malformed
	all := db.All()
	if len(all) == 0 {
		t.Fatal("Default().All() is empty")
	}
	// Indexes must agree with the entry list.
	for _, e := range all {
		if got, ok := db.Lookup(e.UUID, e.Version.Major, e.Version.Minor); !ok || got.Name != e.Name {
			t.Errorf("Lookup(%s v%s) = (%q, %v), want %q", e.UUID.ToFormatD(), e.Version, got.Name, ok, e.Name)
		}
	}
	// All() is sorted by name.
	for i := 1; i < len(all); i++ {
		if all[i-1].Name > all[i].Name {
			t.Errorf("All() not sorted: %q before %q", all[i-1].Name, all[i].Name)
		}
	}
}

func TestLookupExactAndMiss(t *testing.T) {
	e, ok := Lookup(spoolssUUID, 1, 0)
	if !ok {
		t.Fatal("Lookup(spoolss v1.0) not found")
	}
	if e.Name != "spoolss" || e.Executable != "spoolsv.exe" || e.Service != "Spooler" || e.Protocol != "MS-RPRN" {
		t.Errorf("spoolss entry = %+v", e)
	}
	if _, ok := Lookup(spoolssUUID, 9, 9); ok {
		t.Error("Lookup(spoolss v9.9) found, want miss")
	}
	if _, ok := Lookup(unknownUUID, 0, 0); ok {
		t.Error("Lookup(unknown) found, want miss")
	}
}

func TestResolveFallback(t *testing.T) {
	// Exact version present.
	if e, ok := Resolve(spoolssUUID, 1, 0); !ok || e.Name != "spoolss" {
		t.Errorf("Resolve exact = (%q, %v), want spoolss", e.Name, ok)
	}
	// Unknown version of a known UUID falls back to a known version.
	if e, ok := Resolve(spoolssUUID, 7, 3); !ok || e.Name != "spoolss" {
		t.Errorf("Resolve fallback = (%q, %v), want spoolss", e.Name, ok)
	}
	// Entirely unknown UUID.
	if _, ok := Resolve(unknownUUID, 0, 0); ok {
		t.Error("Resolve(unknown) ok = true, want false")
	}
}

func TestSearchByFields(t *testing.T) {
	if got := SearchByExecutable("SPOOLSV.EXE"); len(got) != 2 { // spoolss + IRemoteWinspool, case-insensitive
		t.Errorf("SearchByExecutable(spoolsv.exe) = %d entries, want 2: %v", len(got), names(got))
	}
	if got := SearchByService("Spooler"); len(got) != 2 {
		t.Errorf("SearchByService(Spooler) = %d, want 2: %v", len(got), names(got))
	}
	if got := SearchByProtocol("MS-EFSR"); len(got) != 2 { // two EFSR interface UUIDs
		t.Errorf("SearchByProtocol(MS-EFSR) = %d, want 2: %v", len(got), names(got))
	}
	if got := SearchByName("efsrpc"); len(got) != 2 { // both EFSR entries share the name
		t.Errorf("SearchByName(efsrpc) = %d, want 2: %v", len(got), names(got))
	}
	if got := SearchByPipe(`\pipe\spoolss`); len(got) != 2 {
		t.Errorf("SearchByPipe(spoolss) = %d, want 2: %v", len(got), names(got))
	}
	if got := SearchByExecutable("nope.exe"); got != nil {
		t.Errorf("SearchByExecutable(nope.exe) = %v, want nil", names(got))
	}
}

func TestSearchSubstring(t *testing.T) {
	got := Search("coercion") // spoolss, efsrpc (x2), FssagentRpc descriptions
	if len(got) < 3 {
		t.Errorf("Search(coercion) = %d entries, want >= 3: %v", len(got), names(got))
	}
	if Search("") != nil {
		t.Error("Search(\"\") should match nothing")
	}
}

func TestNewValidation(t *testing.T) {
	good := mustGUID("11111111-1111-1111-1111-111111111111")

	// Duplicate (UUID, version).
	if _, err := New([]Interface{
		{UUID: good, Version: v(1, 0), Name: "a"},
		{UUID: good, Version: v(1, 0), Name: "b"},
	}); err == nil {
		t.Error("New with duplicate (UUID, version): err = nil, want error")
	}
	// Empty name.
	if _, err := New([]Interface{{UUID: good, Version: v(1, 0)}}); err == nil {
		t.Error("New with empty Name: err = nil, want error")
	}
	// Zero UUID.
	if _, err := New([]Interface{{Version: v(1, 0), Name: "a"}}); err == nil {
		t.Error("New with zero UUID: err = nil, want error")
	}
}

func TestLookupUUIDMultiVersion(t *testing.T) {
	u := mustGUID("22222222-2222-2222-2222-222222222222")
	db, err := New([]Interface{
		{UUID: u, Version: v(2, 0), Name: "iface"},
		{UUID: u, Version: v(1, 0), Name: "iface"},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got := db.LookupUUID(u)
	if len(got) != 2 {
		t.Fatalf("LookupUUID = %d versions, want 2", len(got))
	}
	// Sorted by version: 1.0 before 2.0.
	if got[0].Version != (v(1, 0)) || got[1].Version != (v(2, 0)) {
		t.Errorf("LookupUUID versions = %v, %v; want 1.0, 2.0", got[0].Version, got[1].Version)
	}
	// Resolve of an unlisted version falls back to the lowest.
	if e, ok := db.Resolve(u, 5, 0); !ok || e.Version != v(1, 0) {
		t.Errorf("Resolve fallback version = %v (ok=%v), want 1.0", e.Version, ok)
	}
}

// names extracts entry names for readable failure messages.
func names(in []Interface) []string {
	out := make([]string, len(in))
	for i, e := range in {
		out[i] = e.Name + " v" + e.Version.String()
	}
	return out
}
