package registry

import (
	"reflect"
	"testing"
)

func TestValueAccessors(t *testing.T) {
	if got := StringValue("hello").String(); got != "hello" {
		t.Errorf("StringValue/String round-trip = %q, want %q", got, "hello")
	}
	if got := ExpandStringValue("%PATH%").String(); got != "%PATH%" {
		t.Errorf("ExpandStringValue/String round-trip = %q, want %q", got, "%PATH%")
	}

	if n, ok := DwordValue(0x1f).Uint32(); !ok || n != 0x1f {
		t.Errorf("DwordValue/Uint32 = (%d, %v), want (31, true)", n, ok)
	}
	if n, ok := QwordValue(0x1122334455667788).Uint64(); !ok || n != 0x1122334455667788 {
		t.Errorf("QwordValue/Uint64 = (%#x, %v), want (0x1122334455667788, true)", n, ok)
	}

	items := []string{"a", "bb", "ccc"}
	if got := MultiStringValue(items).MultiString(); !reflect.DeepEqual(got, items) {
		t.Errorf("MultiStringValue/MultiString round-trip = %v, want %v", got, items)
	}
}

func TestUint32Short(t *testing.T) {
	if _, ok := (Value{Type: RegDword, Data: []byte{0x01, 0x02}}).Uint32(); ok {
		t.Error("Uint32 on short data: ok = true, want false")
	}
}

func TestDwordEncodingLittleEndian(t *testing.T) {
	if got := DwordValue(0x01020304).Data; !reflect.DeepEqual(got, []byte{0x04, 0x03, 0x02, 0x01}) {
		t.Errorf("DwordValue data = % x, want 04 03 02 01", got)
	}
}
