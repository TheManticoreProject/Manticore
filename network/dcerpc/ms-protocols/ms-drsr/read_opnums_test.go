package msdrsr

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/guid"
	drsrtypes "github.com/TheManticoreProject/Manticore/windows/protocols/ms-drsr"
)

func TestBuildRevMembReqRequestsAttributes(t *testing.T) {
	sid := sidWithRID(500)
	req := (&Client{}).buildRevMembReq([][]byte{sid}, drsrtypes.RevMembGetGroupsForUser, "DC=lab,DC=local")
	if req.DwFlags != drsrtypes.DRS_REVMEMB_FLAG_GET_ATTRIBUTES {
		t.Fatalf("DwFlags = 0x%x, want DRS_REVMEMB_FLAG_GET_ATTRIBUTES", req.DwFlags)
	}
	if len(req.PpDsNames) != 1 || req.PpDsNames[0] == nil {
		t.Fatalf("PpDsNames = %#v, want one name", req.PpDsNames)
	}
	if got := req.PpDsNames[0].Sid.Data[:req.PpDsNames[0].SidLen]; string(got) != string(sid) {
		t.Errorf("request SID = %x, want %x", got, sid)
	}
}

func TestGetObjectExistenceRejectsInvalidVector(t *testing.T) {
	c := &Client{bound: true}
	start := guid.GUID{A: 1}
	tests := []struct {
		name   string
		vector *drsrtypes.UPTODATE_VECTOR_V1_EXT
		want   string
	}{
		{name: "nil", want: "vector is nil"},
		{name: "wrong version", vector: &drsrtypes.UPTODATE_VECTOR_V1_EXT{DwVersion: 2}, want: "version is 2"},
		{name: "wrong count", vector: &drsrtypes.UPTODATE_VECTOR_V1_EXT{DwVersion: 1, CNumCursors: 1}, want: "have 0 cursors"},
		{
			name: "unsorted",
			vector: &drsrtypes.UPTODATE_VECTOR_V1_EXT{
				DwVersion:   1,
				CNumCursors: 2,
				RgCursors: []drsrtypes.UPTODATE_CURSOR_V1{
					{UuidDsa: drsrtypes.UUID{Octets: [16]byte{2}}},
					{UuidDsa: drsrtypes.UUID{Octets: [16]byte{1}}},
				},
			},
			want: "not sorted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.GetObjectExistence("DC=lab,DC=local", start, 1, tt.vector, [16]byte{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}
