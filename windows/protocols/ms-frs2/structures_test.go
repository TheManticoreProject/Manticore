package msfrs2

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/dtyp"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// roundTrip marshals in, unmarshals into a fresh value of the same type, and asserts the
// result is deeply equal to in. This is the wire-shape acceptance gate for the MS-FRS2
// NDR structures in the absence of a live DFS-R server (Go round-trip is necessary but
// not sufficient for wire correctness — see the dcerpc-interface-structure skill).
func roundTrip[T any](t *testing.T, name string, in T) {
	t.Helper()
	raw, err := ndr.Marshal(&in)
	if err != nil {
		t.Fatalf("%s: Marshal: %v", name, err)
	}
	var out T
	if err := ndr.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: Unmarshal: %v", name, err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Errorf("%s: round trip mismatch:\n in:  %+v\n out: %+v", name, in, out)
	}
}

func testGUID(a uint32) dtyp.GUID {
	return dtyp.GUID{Data1: a, Data2: 0x1234, Data3: 0x5678, Data4: [8]byte{0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78}}
}

// TestScalarStructs covers the plain fixed-layout structures (no pointers or arrays).
func TestScalarStructs(t *testing.T) {
	roundTrip(t, "FRS_VERSION_VECTOR", FRS_VERSION_VECTOR{
		DbGuid: testGUID(0x01), Low: 0x1122334455667788, High: 0x8877665544332211,
	})
	roundTrip(t, "FRS_ID_GVSN", FRS_ID_GVSN{
		UidDbGuid: testGUID(0x02), UidVersion: 42, GvsnDbGuid: testGUID(0x03), GvsnVersion: 43,
	})
	roundTrip(t, "FRS_RDC_SOURCE_NEED", FRS_RDC_SOURCE_NEED{NeedOffset: 0x1000, NeedSize: 0x400})
	roundTrip(t, "FRS_EPOQUE_VECTOR", FRS_EPOQUE_VECTOR{
		Machine: testGUID(0x04),
		Epoque:  dtyp.SYSTEMTIME{WYear: 2026, WMonth: 7, WDayOfWeek: 4, WDay: 2, WHour: 13, WMinute: 30, WSecond: 15, WMilliseconds: 500},
	})
}

// TestFRS_UPDATE covers the fixed-size arrays ([20]/[16] byte hashes, the [261]uint16
// name) and the embedded dtyp.FILETIME members.
func TestFRS_UPDATE(t *testing.T) {
	u := FRS_UPDATE{
		Present:      1,
		NameConflict: 0,
		Attributes:   0x20,
		Fence:        dtyp.FILETIME{DwLowDateTime: 0xAABBCCDD, DwHighDateTime: 0x01D7F00D},
		Clock:        dtyp.FILETIME{DwLowDateTime: 1, DwHighDateTime: 2},
		CreateTime:   dtyp.FILETIME{DwLowDateTime: 3, DwHighDateTime: 4},
		ContentSetId: testGUID(0x10),
		UidDbGuid:    testGUID(0x11),
		UidVersion:   100,
		GvsnDbGuid:   testGUID(0x12),
		GvsnVersion:  101,
		ParentDbGuid: testGUID(0x13),
		Flags:        0x08,
	}
	u.Hash[0], u.Hash[19] = 0xDE, 0xAD
	u.RdcSimilarity[0], u.RdcSimilarity[15] = 0xBE, 0xEF
	for i, r := range "file.txt" {
		u.Name[i] = uint16(r)
	}
	roundTrip(t, "FRS_UPDATE", u)

	roundTrip(t, "FRS_UPDATE_CANCEL_DATA", FRS_UPDATE_CANCEL_DATA{
		BlockingUpdate:   u,
		ContentSetId:     testGUID(0x20),
		GvsnDatabaseId:   testGUID(0x21),
		UidDatabaseId:    testGUID(0x22),
		ParentDatabaseId: testGUID(0x23),
		GvsnVersion:      1, UidVersion: 2, ParentVersion: 3,
		CancelType: 1,
		IsUidValid: 1, IsParentUidValid: 0, IsBlockerValid: 1,
	})
}

// TestFRS_RDC_PARAMETERS exercises the discriminated union with each of its three arms
// selected (the discriminant is a 16-bit rdcChunkerAlgorithm).
func TestFRS_RDC_PARAMETERS(t *testing.T) {
	generic := FRS_RDC_PARAMETERS_GENERIC{ChunkerType: 7}
	generic.ChunkerParameters[0], generic.ChunkerParameters[63] = 1, 2
	cases := []struct {
		name string
		in   FRS_RDC_PARAMETERS
	}{
		{"generic", FRS_RDC_PARAMETERS{RdcChunkerAlgorithm: 0, U: FRS_RDC_PARAMETERS_U{Tag: 0, FilterGeneric: generic}}},
		{"max", FRS_RDC_PARAMETERS{RdcChunkerAlgorithm: 1, U: FRS_RDC_PARAMETERS_U{Tag: 1, FilterMax: FRS_RDC_PARAMETERS_FILTERMAX{HorizonSize: 1024, WindowSize: 32}}}},
		{"point", FRS_RDC_PARAMETERS{RdcChunkerAlgorithm: 2, U: FRS_RDC_PARAMETERS_U{Tag: 2, FilterPoint: FRS_RDC_PARAMETERS_FILTERPOINT{MinChunkSize: 128, MaxChunkSize: 4096}}}},
	}
	for _, c := range cases {
		roundTrip(t, "FRS_RDC_PARAMETERS/"+c.name, c.in)
	}
}

// TestFRS_RDC_FILEINFO covers the embedded conformant array rdcFilterParameters[*] whose
// size_is is the sibling RdcSignatureLevels count.
func TestFRS_RDC_FILEINFO(t *testing.T) {
	roundTrip(t, "FRS_RDC_FILEINFO", FRS_RDC_FILEINFO{
		OnDiskFileSize:              0x100000,
		FileSizeEstimate:            0x100000,
		RdcVersion:                  1,
		RdcMinimumCompatibleVersion: 1,
		RdcSignatureLevels:          2,
		CompressionAlgorithm:        RDC_XPRESS,
		RdcFilterParameters: []FRS_RDC_PARAMETERS{
			{RdcChunkerAlgorithm: 1, U: FRS_RDC_PARAMETERS_U{Tag: 1, FilterMax: FRS_RDC_PARAMETERS_FILTERMAX{HorizonSize: 512, WindowSize: 16}}},
			{RdcChunkerAlgorithm: 2, U: FRS_RDC_PARAMETERS_U{Tag: 2, FilterPoint: FRS_RDC_PARAMETERS_FILTERPOINT{MinChunkSize: 128, MaxChunkSize: 2048}}},
		},
	})
}

// TestFRS_ASYNC_VERSION_VECTOR_RESPONSE covers the two [unique] pointers to conformant
// arrays (versionVector / epoqueVector), both populated and both null.
func TestFRS_ASYNC_VERSION_VECTOR_RESPONSE(t *testing.T) {
	roundTrip(t, "FRS_ASYNC_VERSION_VECTOR_RESPONSE/full", FRS_ASYNC_VERSION_VECTOR_RESPONSE{
		VvGeneration:       0xABCD,
		VersionVectorCount: 2,
		VersionVector: []FRS_VERSION_VECTOR{
			{DbGuid: testGUID(0x30), Low: 1, High: 2},
			{DbGuid: testGUID(0x31), Low: 3, High: 4},
		},
		EpoqueVectorCount: 1,
		EpoqueVector: []FRS_EPOQUE_VECTOR{
			{Machine: testGUID(0x32), Epoque: dtyp.SYSTEMTIME{WYear: 2026, WMonth: 1, WDay: 1}},
		},
	})
	roundTrip(t, "FRS_ASYNC_VERSION_VECTOR_RESPONSE/empty", FRS_ASYNC_VERSION_VECTOR_RESPONSE{
		VvGeneration: 1,
	})

	roundTrip(t, "FRS_ASYNC_RESPONSE_CONTEXT", FRS_ASYNC_RESPONSE_CONTEXT{
		SequenceNumber: 5,
		Status:         0,
		Result: FRS_ASYNC_VERSION_VECTOR_RESPONSE{
			VvGeneration:       7,
			VersionVectorCount: 1,
			VersionVector:      []FRS_VERSION_VECTOR{{DbGuid: testGUID(0x40), Low: 9, High: 10}},
		},
	})
}
