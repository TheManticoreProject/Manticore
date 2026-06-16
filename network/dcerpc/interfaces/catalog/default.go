package catalog

import (
	"sync"

	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// defaultDB is the built-in catalog, constructed once from the builtin seed.
var (
	defaultOnce sync.Once
	defaultDB   *Database
)

// Default returns the built-in catalog of well-known interfaces. It panics if
// the seed table is malformed (a programmer error caught by the package tests).
func Default() *Database {
	defaultOnce.Do(func() {
		db, err := New(builtin)
		if err != nil {
			panic("catalog: invalid built-in data: " + err.Error())
		}
		defaultDB = db
	})
	return defaultDB
}

// Lookup is Default().Lookup.
func Lookup(uuid guid.GUID, major, minor uint16) (Interface, bool) {
	return Default().Lookup(uuid, major, minor)
}

// Resolve is Default().Resolve.
func Resolve(uuid guid.GUID, major, minor uint16) (Interface, bool) {
	return Default().Resolve(uuid, major, minor)
}

// LookupUUID is Default().LookupUUID.
func LookupUUID(uuid guid.GUID) []Interface { return Default().LookupUUID(uuid) }

// SearchByName is Default().SearchByName.
func SearchByName(name string) []Interface { return Default().SearchByName(name) }

// SearchByExecutable is Default().SearchByExecutable.
func SearchByExecutable(exe string) []Interface { return Default().SearchByExecutable(exe) }

// SearchByService is Default().SearchByService.
func SearchByService(service string) []Interface { return Default().SearchByService(service) }

// SearchByProtocol is Default().SearchByProtocol.
func SearchByProtocol(protocol string) []Interface { return Default().SearchByProtocol(protocol) }

// SearchByPipe is Default().SearchByPipe.
func SearchByPipe(pipe string) []Interface { return Default().SearchByPipe(pipe) }

// Search is Default().Search.
func Search(substr string) []Interface { return Default().Search(substr) }

// All is Default().All.
func All() []Interface { return Default().All() }
