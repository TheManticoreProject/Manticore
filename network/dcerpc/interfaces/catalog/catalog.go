// Package catalog is a database of well-known DCE/RPC interfaces: it maps an
// interface UUID (and version) to human-readable metadata — a short name, a
// title, a description, the executable or DLL that implements the server, the
// hosting Windows service, the MS-* protocol document, and the well-known named
// pipe(s). It lets tooling label the raw UUIDs returned by ept_lookup / ept_map
// (e.g. 12345678-1234-abcd-ef00-0123456789ab v1.0 -> "spoolss", MS-RPRN,
// spoolsv.exe, Spooler service).
//
// The catalog is a pure data table, independent of which interfaces have a
// client implementation under network/dcerpc/interfaces. Look entries up through
// the package-level helpers (which use the built-in Default database) or build a
// custom Database with New.
package catalog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Version is a DCE/RPC interface version (major.minor).
type Version struct {
	Major uint16
	Minor uint16
}

// String renders the version as "major.minor".
func (v Version) String() string { return fmt.Sprintf("%d.%d", v.Major, v.Minor) }

// Interface is one catalog entry: a specific interface UUID + version and the
// metadata known about it. Executable is the image that implements the RPC
// server (an .exe or a .dll loaded into a host process), e.g. "spoolsv.exe" or
// "srvsvc.dll"; Service is the hosting Windows service name where applicable,
// e.g. "Spooler". Fields that are not known are left empty.
type Interface struct {
	UUID        guid.GUID
	Version     Version
	Name        string   // short handle, e.g. "spoolss"
	Title       string   // human title, e.g. "Print System Remote Protocol"
	Description string   // one-line summary
	Executable  string   // implementing image, e.g. "spoolsv.exe" / "srvsvc.dll"
	Service     string   // hosting Windows service, e.g. "Spooler"
	Protocol    string   // MS-* document id, e.g. "MS-RPRN"
	Pipes       []string // well-known named pipe(s), e.g. `\pipe\spoolss`
}

// String renders a compact one-line label, e.g.
// "spoolss v1.0 (MS-RPRN) [spoolsv.exe]".
func (i Interface) String() string {
	var b strings.Builder
	if i.Name != "" {
		b.WriteString(i.Name)
	} else {
		b.WriteString(i.UUID.ToFormatD())
	}
	fmt.Fprintf(&b, " v%s", i.Version)
	if i.Protocol != "" {
		fmt.Fprintf(&b, " (%s)", i.Protocol)
	}
	if i.Executable != "" {
		fmt.Fprintf(&b, " [%s]", i.Executable)
	}
	return b.String()
}

// verKey identifies one entry by UUID and version.
type verKey struct {
	uuid  guid.GUID
	major uint16
	minor uint16
}

// Database is an indexed, read-only collection of catalog entries.
type Database struct {
	entries    []Interface
	byUUID     map[guid.GUID][]Interface
	byKey      map[verKey]Interface
	byName     map[string][]Interface
	byExec     map[string][]Interface
	byService  map[string][]Interface
	byProtocol map[string][]Interface
	byPipe     map[string][]Interface
}

// New builds a Database from entries, constructing the lookup indexes. It errors
// on a malformed catalog: an entry with an empty Name, a zero UUID, or a
// duplicate (UUID, version).
func New(entries []Interface) (*Database, error) {
	db := &Database{
		byUUID:     map[guid.GUID][]Interface{},
		byKey:      map[verKey]Interface{},
		byName:     map[string][]Interface{},
		byExec:     map[string][]Interface{},
		byService:  map[string][]Interface{},
		byProtocol: map[string][]Interface{},
		byPipe:     map[string][]Interface{},
	}
	var zero guid.GUID
	for _, e := range entries {
		if e.Name == "" {
			return nil, fmt.Errorf("catalog: entry %s has an empty Name", e.UUID.ToFormatD())
		}
		if e.UUID == zero {
			return nil, fmt.Errorf("catalog: entry %q has a zero UUID", e.Name)
		}
		k := verKey{e.UUID, e.Version.Major, e.Version.Minor}
		if _, dup := db.byKey[k]; dup {
			return nil, fmt.Errorf("catalog: duplicate entry for %s v%s", e.UUID.ToFormatD(), e.Version)
		}
		db.entries = append(db.entries, e)
		db.byKey[k] = e
		db.byUUID[e.UUID] = append(db.byUUID[e.UUID], e)
		index(db.byName, e.Name, e)
		index(db.byExec, e.Executable, e)
		index(db.byService, e.Service, e)
		index(db.byProtocol, e.Protocol, e)
		for _, p := range e.Pipes {
			index(db.byPipe, p, e)
		}
	}
	// Keep each UUID's versions sorted so LookupUUID is deterministic and Resolve's
	// fallback is the lowest known version.
	for uuid := range db.byUUID {
		sortInterfaces(db.byUUID[uuid])
	}
	return db, nil
}

// index appends e to m[lower(key)] when key is non-empty.
func index(m map[string][]Interface, key string, e Interface) {
	if key == "" {
		return
	}
	lk := strings.ToLower(key)
	m[lk] = append(m[lk], e)
}

// Lookup returns the entry for an exact (UUID, version), reporting whether it is
// in the catalog.
func (db *Database) Lookup(uuid guid.GUID, major, minor uint16) (Interface, bool) {
	e, ok := db.byKey[verKey{uuid, major, minor}]
	return e, ok
}

// Resolve labels a UUID/version seen on the wire: it returns the exact
// (UUID, version) entry when present, otherwise falls back to any known version
// of the same UUID (the lowest version, for determinism). ok is false only when
// the UUID is entirely unknown. Use it to annotate ept_lookup results, where the
// reported version may not be one the catalog enumerates explicitly.
func (db *Database) Resolve(uuid guid.GUID, major, minor uint16) (Interface, bool) {
	if e, ok := db.byKey[verKey{uuid, major, minor}]; ok {
		return e, true
	}
	versions := db.byUUID[uuid]
	if len(versions) == 0 {
		return Interface{}, false
	}
	return versions[0], true
}

// LookupUUID returns every known version of a UUID, or nil if the UUID is
// unknown. Results are sorted by version.
func (db *Database) LookupUUID(uuid guid.GUID) []Interface {
	return sortedCopy(db.byUUID[uuid])
}

// SearchByName returns the entries whose short Name matches (case-insensitive).
func (db *Database) SearchByName(name string) []Interface { return db.exact(db.byName, name) }

// SearchByExecutable returns the entries implemented by the given image
// (case-insensitive), e.g. "spoolsv.exe".
func (db *Database) SearchByExecutable(exe string) []Interface { return db.exact(db.byExec, exe) }

// SearchByService returns the entries hosted by the given Windows service
// (case-insensitive), e.g. "Spooler".
func (db *Database) SearchByService(service string) []Interface {
	return db.exact(db.byService, service)
}

// SearchByProtocol returns the entries for the given MS-* protocol document
// (case-insensitive), e.g. "MS-RPRN".
func (db *Database) SearchByProtocol(protocol string) []Interface {
	return db.exact(db.byProtocol, protocol)
}

// SearchByPipe returns the entries reachable over the given named pipe
// (case-insensitive), e.g. `\pipe\spoolss`.
func (db *Database) SearchByPipe(pipe string) []Interface { return db.exact(db.byPipe, pipe) }

// exact returns a sorted copy of the index bucket for the lower-cased key.
func (db *Database) exact(idx map[string][]Interface, key string) []Interface {
	return sortedCopy(idx[strings.ToLower(key)])
}

// Search returns every entry whose Name, Title, Description, Protocol,
// Executable, Service, or one of its Pipes contains the given substring
// (case-insensitive). An empty query matches nothing.
func (db *Database) Search(substr string) []Interface {
	if substr == "" {
		return nil
	}
	q := strings.ToLower(substr)
	var out []Interface
	for _, e := range db.entries {
		if entryContains(e, q) {
			out = append(out, e)
		}
	}
	return sortInterfaces(out)
}

// entryContains reports whether any text field of e contains the lower-cased q.
func entryContains(e Interface, q string) bool {
	fields := []string{e.Name, e.Title, e.Description, e.Protocol, e.Executable, e.Service}
	fields = append(fields, e.Pipes...)
	for _, f := range fields {
		if f != "" && strings.Contains(strings.ToLower(f), q) {
			return true
		}
	}
	return false
}

// All returns every entry, sorted by name then version.
func (db *Database) All() []Interface { return sortedCopy(db.entries) }

// sortedCopy returns a sorted copy of in (never aliasing the caller's slice).
func sortedCopy(in []Interface) []Interface {
	if len(in) == 0 {
		return nil
	}
	out := make([]Interface, len(in))
	copy(out, in)
	return sortInterfaces(out)
}

// sortInterfaces sorts in place by Name, then UUID, then version, and returns it.
func sortInterfaces(s []Interface) []Interface {
	sort.Slice(s, func(i, j int) bool {
		a, b := s[i], s[j]
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		if a.UUID != b.UUID {
			return a.UUID.ToFormatD() < b.UUID.ToFormatD()
		}
		if a.Version.Major != b.Version.Major {
			return a.Version.Major < b.Version.Major
		}
		return a.Version.Minor < b.Version.Minor
	})
	return s
}
