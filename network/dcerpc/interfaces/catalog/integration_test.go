//go:build integration

package catalog_test

import (
	"encoding/binary"
	"os"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/catalog"
	epm "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/e1af8308-5d1f-11c9-91a4-08002b14a0fa/3.0/functions"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/transport/tcp"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	msrpce "github.com/TheManticoreProject/Manticore/windows/protocols/ms-rpce"
)

func TestIntegration_DocumentedInterfacesResolve(t *testing.T) {
	host := os.Getenv("DCERPC_TEST_HOST")
	if host == "" {
		t.Skip("DCERPC_TEST_HOST not set; skipping live catalog resolution test")
	}

	rpc := client.NewClient(tcp.New(host, tcp.EndpointMapperPort))
	if err := rpc.Bind(epm.SyntaxID()); err != nil {
		t.Fatalf("Bind(endpoint mapper): %v", err)
	}
	defer rpc.Close()

	entries, err := functions.Lookup(rpc)
	if err != nil {
		t.Fatalf("Lookup(endpoint mapper): %v", err)
	}

	want := map[string]string{
		"0b6edbfa-4a24-4fc6-8a23-942b1eca65d1": "IRPCAsyncNotify",
		"ae33069b-a2a8-46ee-a235-ddfd339be281": "IRPCRemoteObject",
		"d95afe70-a6d5-4259-822e-2c84da1ddb0d": "WindowsShutdown",
		"906b0ce0-c70b-1067-b317-00dd010662da": "IXnRemote",
	}
	found := make(map[string]bool)
	for _, endpoint := range entries {
		tower, err := endpoint.DecodeTower()
		if err != nil || len(tower.Floors) == 0 {
			continue
		}
		floor := tower.Floors[0]
		if floor.Protocol() != msrpce.FloorProtoUUID || len(floor.LHS) < 19 || len(floor.RHS) < 2 {
			continue
		}
		var id guid.GUID
		id.FromRawBytes(floor.LHS[1:17])
		uuid := id.ToFormatD()
		name, tracked := want[uuid]
		if !tracked {
			continue
		}
		major := binary.LittleEndian.Uint16(floor.LHS[17:19])
		minor := binary.LittleEndian.Uint16(floor.RHS[:2])
		entry, ok := catalog.Lookup(id, major, minor)
		if !ok || entry.Name != name {
			t.Errorf("catalog lookup for live %s v%d.%d = (%q, %v), want %q", uuid, major, minor, entry.Name, ok, name)
			continue
		}
		found[uuid] = true
		t.Logf("resolved live %s v%d.%d as %s", uuid, major, minor, entry.Name)
	}

	if !found["d95afe70-a6d5-4259-822e-2c84da1ddb0d"] {
		t.Error("live endpoint map did not contain the WindowsShutdown interface")
	}
}
