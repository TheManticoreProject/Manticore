package hostserver_test

import (
	"bytes"
	"errors"
	"math"
	"sync"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// sampleFrame is a synthetic Remote Command-shaped frame: 4-byte template,
// one LL/CP field.
func sampleFrame() hostserver.Frame {
	return hostserver.Frame{
		Header: hostserver.Header{
			Attributes: 0x0001, ServerID: hostserver.ServerIDRemoteCommand,
			Instance: 0x01020304, CorrelationID: 7, RequestID: 0x1003,
		},
		Template: []byte{0xAA, 0xBB, 0xCC, 0xDD},
		Variable: []byte{0x00, 0x00, 0x00, 0x08, 0x11, 0x03, 0xC1, 0xC2},
	}
}

var sampleBytes = []byte{
	0x00, 0x00, 0x00, 0x20, // length 32
	0x00, 0x01, // attributes
	0xE0, 0x08, // server ID
	0x01, 0x02, 0x03, 0x04, // instance
	0x00, 0x00, 0x00, 0x07, // correlation
	0x00, 0x04, // template length
	0x10, 0x03, // request ID
	0xAA, 0xBB, 0xCC, 0xDD,
	0x00, 0x00, 0x00, 0x08, 0x11, 0x03, 0xC1, 0xC2,
}

func TestFrameMarshalByteExact(t *testing.T) {
	got, err := sampleFrame().Marshal(0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, sampleBytes) {
		t.Fatalf("Marshal =\n% X\nwant\n% X", got, sampleBytes)
	}
	f, err := hostserver.UnmarshalFrame(sampleBytes, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := sampleFrame()
	want.Header.Length, want.Header.TemplateLength = 32, 4
	if f.Header != want.Header || !bytes.Equal(f.Template, want.Template) || !bytes.Equal(f.Variable, want.Variable) {
		t.Fatalf("UnmarshalFrame = %+v", f)
	}
	f.Template[0] = 0
	if sampleBytes[20] != 0xAA {
		t.Fatal("UnmarshalFrame result aliases its input")
	}
}

func TestHeaderOnlyFrame(t *testing.T) {
	b, err := hostserver.Frame{Header: hostserver.Header{ServerID: hostserver.ServerIDSignon, RequestID: 0x7006}}.Marshal(0)
	if err != nil || len(b) != hostserver.HeaderLength {
		t.Fatalf("header-only frame = % X, %v", b, err)
	}
	f, err := hostserver.UnmarshalFrame(b, hostserver.MinMaxFrame)
	if err != nil || len(f.Template) != 0 || len(f.Variable) != 0 {
		t.Fatalf("UnmarshalFrame = %+v, %v", f, err)
	}
}

func TestUnmarshalFrameRejects(t *testing.T) {
	mutate := func(f func([]byte)) []byte {
		b := append([]byte(nil), sampleBytes...)
		f(b)
		return b
	}
	for name, tc := range map[string]struct {
		b    []byte
		max  uint32
		want error
	}{
		"empty":             {nil, 0, hostserver.ErrMalformed},
		"short header":      {sampleBytes[:19], 0, hostserver.ErrMalformed},
		"length below 20":   {mutate(func(b []byte) { b[3] = 19 }), 0, hostserver.ErrMalformed},
		"truncated body":    {sampleBytes[:31], 0, hostserver.ErrMalformed},
		"trailing byte":     {append(append([]byte(nil), sampleBytes...), 0), 0, hostserver.ErrMalformed},
		"template overflow": {mutate(func(b []byte) { b[17] = 17 }), 0, hostserver.ErrMalformed},
		"over max":          {sampleBytes, 31, hostserver.ErrFrameTooLarge},
		"huge length":       {mutate(func(b []byte) { b[0], b[1], b[2], b[3] = 0xFF, 0xFF, 0xFF, 0xFF }), 0, hostserver.ErrFrameTooLarge},
	} {
		if _, err := hostserver.UnmarshalFrame(tc.b, tc.max); !errors.Is(err, tc.want) {
			t.Errorf("%s: error = %v; want %v", name, err, tc.want)
		}
	}
}

func TestResolveMaxFrame(t *testing.T) {
	for in, want := range map[uint32]uint32{0: hostserver.DefaultMaxFrame, 20: 20, hostserver.MaxMaxFrame: hostserver.MaxMaxFrame} {
		if got, err := hostserver.ResolveMaxFrame(in); err != nil || got != want {
			t.Errorf("ResolveMaxFrame(%d) = %d, %v", in, got, err)
		}
	}
	for _, in := range []uint32{1, 19, hostserver.MaxMaxFrame + 1, math.MaxUint32} {
		if _, err := hostserver.ResolveMaxFrame(in); err == nil {
			t.Errorf("ResolveMaxFrame(%d) accepted", in)
		}
	}
}

func TestFrameMarshalRejects(t *testing.T) {
	if _, err := (hostserver.Frame{Variable: make([]byte, 100)}).Marshal(64); !errors.Is(err, hostserver.ErrFrameTooLarge) {
		t.Errorf("over-max marshal error = %v", err)
	}
	if _, err := (hostserver.Frame{Template: make([]byte, math.MaxUint16+1)}).Marshal(hostserver.MaxMaxFrame); !errors.Is(err, hostserver.ErrFrameTooLarge) {
		t.Errorf("oversized template error = %v", err)
	}
}

func TestHeaderRejects(t *testing.T) {
	if _, err := hostserver.UnmarshalHeader(make([]byte, 21)); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("21-byte header error = %v", err)
	}
	if _, err := (hostserver.Header{Length: 19}).MarshalBinary(); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("length 19 marshal error = %v", err)
	}
	if _, err := (hostserver.Header{Length: 24, TemplateLength: 5}).MarshalBinary(); !errors.Is(err, hostserver.ErrMalformed) {
		t.Errorf("template overflow marshal error = %v", err)
	}
}

func TestFieldsRoundTrip(t *testing.T) {
	var b []byte
	var err error
	for _, f := range []hostserver.Field{{0x1101, []byte("one")}, {0x1103, nil}, {0x1103, []byte{9}}} {
		if b, err = hostserver.AppendField(b, f.CodePoint, f.Value); err != nil {
			t.Fatal(err)
		}
	}
	want := []byte{0, 0, 0, 9, 0x11, 0x01, 'o', 'n', 'e', 0, 0, 0, 6, 0x11, 0x03, 0, 0, 0, 7, 0x11, 0x03, 9}
	if !bytes.Equal(b, want) {
		t.Fatalf("AppendField = % X", b)
	}
	fs, err := hostserver.ParseFields(b, hostserver.FieldPolicy{Repeatable: map[uint16]bool{0x1103: true}})
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := fs.Get(0x1101); !ok || string(v) != "one" {
		t.Errorf("Get(0x1101) = %q, %t", v, ok)
	}
	if all := fs.All(0x1103); len(all) != 2 || len(all[0]) != 0 || all[1][0] != 9 {
		t.Errorf("All(0x1103) = %v", all)
	}
	if _, ok := fs.Get(0x9999); ok {
		t.Error("Get of an absent code point reported ok")
	}
	if fs, err := hostserver.ParseFields(nil, hostserver.FieldPolicy{}); err != nil || len(fs) != 0 {
		t.Errorf("empty ParseFields = %v, %v", fs, err)
	}
}

func TestParseFieldsRejects(t *testing.T) {
	field := func(ll uint32, cp uint16, value ...byte) []byte {
		return append([]byte{byte(ll >> 24), byte(ll >> 16), byte(ll >> 8), byte(ll), byte(cp >> 8), byte(cp)}, value...)
	}
	cat := func(parts ...[]byte) []byte { return bytes.Join(parts, nil) }
	many := []byte{}
	for i := 0; i < 3; i++ {
		many = append(many, field(6, uint16(i))...)
	}
	for name, tc := range map[string]struct {
		b      []byte
		policy hostserver.FieldPolicy
	}{
		"length below 6":      {field(5, 1), hostserver.FieldPolicy{}},
		"length zero":         {field(0, 1), hostserver.FieldPolicy{}},
		"length past end":     {field(9, 1, 'a'), hostserver.FieldPolicy{}},
		"trailing partial":    {cat(field(6, 1), []byte{0, 0, 0}), hostserver.FieldPolicy{}},
		"duplicate singleton": {cat(field(6, 1), field(6, 1)), hostserver.FieldPolicy{}},
		"too many fields":     {many, hostserver.FieldPolicy{MaxFields: 2}},
		"huge length":         {field(0xFFFFFFFF, 1), hostserver.FieldPolicy{}},
	} {
		_, err := hostserver.ParseFields(tc.b, tc.policy)
		var me *hostserver.MalformedError
		if !errors.Is(err, hostserver.ErrMalformed) || !errors.As(err, &me) {
			t.Errorf("%s: error = %v", name, err)
		}
	}
	// The offset points at the failing field.
	_, err := hostserver.ParseFields(cat(field(6, 1), field(4, 2)), hostserver.FieldPolicy{})
	var me *hostserver.MalformedError
	if !errors.As(err, &me) || me.Offset != 6 {
		t.Errorf("offset error = %v", err)
	}
}

func TestCorrelator(t *testing.T) {
	var c hostserver.Correlator
	for want := uint32(1); want <= 3; want++ {
		if got, err := c.Next(); err != nil || got != want {
			t.Fatalf("Next() = %d, %v; want %d", got, err, want)
		}
	}
	var wg sync.WaitGroup
	seen := sync.Map{}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := c.Next()
			if _, dup := seen.LoadOrStore(id, true); dup || err != nil || id == 0 {
				t.Errorf("concurrent Next() = %d, %v (duplicate %t)", id, err, dup)
			}
		}()
	}
	wg.Wait()
}

func TestCorrelatorNeverWraps(t *testing.T) {
	c := hostserver.NewCorrelatorAt(math.MaxUint32 - 1)
	if id, err := c.Next(); err != nil || id != math.MaxUint32 {
		t.Fatalf("Next() = %d, %v", id, err)
	}
	for i := 0; i < 2; i++ {
		if id, err := c.Next(); !errors.Is(err, hostserver.ErrCorrelationExhausted) || id != 0 {
			t.Fatalf("Next() after MaxUint32 = %d, %v", id, err)
		}
	}
}

func FuzzUnmarshalFrame(f *testing.F) {
	f.Add(sampleBytes)
	f.Add(sampleBytes[:20])
	f.Add([]byte{0, 0, 0, 20, 0, 0, 0xE0, 0x09, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x70, 0x06})
	f.Fuzz(func(t *testing.T, b []byte) {
		const max = 256
		fr, err := hostserver.UnmarshalFrame(b, max)
		if err != nil {
			return
		}
		if len(b) > max || int(fr.Header.Length) != len(b) {
			t.Fatalf("accepted %d bytes with length %d", len(b), fr.Header.Length)
		}
		again, err := fr.Marshal(max)
		if err != nil || !bytes.Equal(again, b) {
			t.Fatalf("re-marshal = % X, %v; want % X", again, err, b)
		}
	})
}

func FuzzParseFields(f *testing.F) {
	f.Add([]byte{0, 0, 0, 9, 0x11, 0x01, 'o', 'n', 'e'})
	f.Add([]byte{0, 0, 0, 6, 0x11, 0x03, 0, 0, 0, 6, 0x11, 0x03})
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0, 1})
	f.Fuzz(func(t *testing.T, b []byte) {
		fs, err := hostserver.ParseFields(b, hostserver.FieldPolicy{MaxFields: 8})
		if err != nil {
			return
		}
		var re []byte
		for _, fl := range fs {
			if re, err = hostserver.AppendField(re, fl.CodePoint, fl.Value); err != nil {
				t.Fatal(err)
			}
		}
		if !bytes.Equal(re, b) || len(fs) > 8 {
			t.Fatalf("fields did not consume input exactly")
		}
	})
}
