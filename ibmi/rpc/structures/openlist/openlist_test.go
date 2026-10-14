package openlist_test

import (
	"encoding/binary"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

// sample builds synthetic list information (IBM "Open list information
// format" offsets).
func sample() []byte {
	b := make([]byte, openlist.Length)
	binary.BigEndian.PutUint32(b[0:], 63)                                          // total records
	binary.BigEndian.PutUint32(b[4:], 10)                                          // records returned
	copy(b[8:], []byte{0xDE, 0xAD, 0xBE, 0xEF})                                    // handle
	binary.BigEndian.PutUint32(b[12:], 62)                                         // record length
	b[16] = 0xC3                                                                   // C
	copy(b[17:30], []byte("\xF1\xF2\xF6\xF1\xF0\xF1\xF4\xF1\xF2\xF0\xF0\xF0\xF0")) // 1261014120000
	b[30] = 0xF2                                                                   // complete
	binary.BigEndian.PutUint32(b[32:], 620)                                        // info length
	binary.BigEndian.PutUint32(b[36:], 1)                                          // first record
	return b
}

func TestUnmarshal(t *testing.T) {
	info, err := openlist.Unmarshal(sample())
	if err != nil {
		t.Fatal(err)
	}
	want := openlist.Info{
		TotalRecords: 63, RecordsReturned: 10, Handle: openlist.Handle{0xDE, 0xAD, 0xBE, 0xEF}, RecordLength: 62,
		Complete: openlist.CompleteAccurate, Status: openlist.StatusComplete, InfoLength: 620, FirstRecord: 1,
	}
	copy(want.Created[:], sample()[17:30])
	if info != want {
		t.Fatalf("Unmarshal = %+v\nwant %+v", info, want)
	}
}

func TestUnmarshalEnums(t *testing.T) {
	for b, want := range map[byte]openlist.Complete{0xC3: openlist.CompleteAccurate, 0xC9: openlist.Incomplete, 0xD7: openlist.PartialAccurate} {
		s := sample()
		s[16] = b
		if info, err := openlist.Unmarshal(s); err != nil || info.Complete != want {
			t.Errorf("complete 0x%02X = %s, %v", b, info.Complete, err)
		}
	}
	for i := 0; i <= 5; i++ {
		s := sample()
		s[30] = byte(0xF0 + i)
		if info, err := openlist.Unmarshal(s); err != nil || info.Status != openlist.Status(i) {
			t.Errorf("status %d = %s, %v", i, info.Status, err)
		}
	}
}

func TestUnmarshalRejects(t *testing.T) {
	for name, mutate := range map[string]func([]byte) []byte{
		"79 bytes":          func(b []byte) []byte { return b[:79] },
		"81 bytes":          func(b []byte) []byte { return append(b, 0) },
		"negative total":    func(b []byte) []byte { binary.BigEndian.PutUint32(b[0:], 0xFFFFFFFF); return b },
		"negative returned": func(b []byte) []byte { binary.BigEndian.PutUint32(b[4:], 0x80000000); return b },
		"negative length":   func(b []byte) []byte { binary.BigEndian.PutUint32(b[12:], 0xFFFFFFFE); return b },
		"negative info":     func(b []byte) []byte { binary.BigEndian.PutUint32(b[32:], 0xFFFFFFFF); return b },
		"negative first":    func(b []byte) []byte { binary.BigEndian.PutUint32(b[36:], 0xFFFFFFFF); return b },
		"complete ASCII C":  func(b []byte) []byte { b[16] = 'C'; return b },
		"complete blank":    func(b []byte) []byte { b[16] = 0x40; return b },
		"status 6":          func(b []byte) []byte { b[30] = 0xF6; return b },
		"status ASCII 2":    func(b []byte) []byte { b[30] = '2'; return b },
	} {
		if _, err := openlist.Unmarshal(mutate(sample())); !errors.Is(err, openlist.ErrMalformed) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
}

func TestRecoverHandle(t *testing.T) {
	bad := sample()
	bad[30] = 0x00 // invalid status, handle still recoverable
	if h, ok := openlist.RecoverHandle(bad); !ok || h != (openlist.Handle{0xDE, 0xAD, 0xBE, 0xEF}) {
		t.Errorf("RecoverHandle = %x, %t", h, ok)
	}
	if _, ok := openlist.RecoverHandle(make([]byte, 11)); ok {
		t.Error("handle recovered from 11 bytes")
	}
}

func FuzzUnmarshal(f *testing.F) {
	f.Add(sample())
	f.Add(make([]byte, 80))
	f.Fuzz(func(t *testing.T, b []byte) {
		info, err := openlist.Unmarshal(b)
		if err != nil {
			return
		}
		if len(b) != 80 || info.TotalRecords < 0 || info.RecordsReturned < 0 || info.RecordLength < 0 || info.Status > openlist.StatusTooMuchData {
			t.Fatalf("accepted %+v from %d bytes", info, len(b))
		}
	})
}
