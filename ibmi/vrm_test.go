package ibmi_test

import (
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi"
)

func TestNewVRMBoundaries(t *testing.T) {
	valid := [][3]uint8{{1, 0, 0}, {5, 4, 0}, {7, 5, 0}, {9, 9, 9}}
	for _, c := range valid {
		v, err := ibmi.NewVRM(c[0], c[1], c[2])
		if err != nil {
			t.Fatalf("NewVRM%v: %v", c, err)
		}
		if v.Version() != c[0] || v.Release() != c[1] || v.Modification() != c[2] || v.IsZero() {
			t.Errorf("NewVRM%v = %v", c, v)
		}
	}
	invalid := [][3]uint8{{0, 0, 0}, {0, 9, 9}, {10, 0, 0}, {5, 10, 0}, {5, 4, 10}, {255, 255, 255}}
	for _, c := range invalid {
		if v, err := ibmi.NewVRM(c[0], c[1], c[2]); !errors.Is(err, ibmi.ErrInvalidVRM) || !v.IsZero() {
			t.Errorf("NewVRM%v = %v, %v; want ErrInvalidVRM", c, v, err)
		}
	}
}

func TestParseVRM(t *testing.T) {
	v, err := ibmi.ParseVRM("V5R4M0")
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "V5R4M0" {
		t.Errorf("String() = %q", v.String())
	}
	for _, s := range []string{"", "V5R4", "v5r4m0", "V0R0M0", "V5R4M0 ", "V10R1M0", "VXR4M0", "5R4M0V", "V5-4M0"} {
		if _, err := ibmi.ParseVRM(s); !errors.Is(err, ibmi.ErrInvalidVRM) {
			t.Errorf("ParseVRM(%q) error = %v; want ErrInvalidVRM", s, err)
		}
	}
}

func TestVRMOrdering(t *testing.T) {
	v5r4, _ := ibmi.ParseVRM("V5R4M0")
	v5r4m5, _ := ibmi.ParseVRM("V5R4M5")
	v7r1, _ := ibmi.ParseVRM("V7R1M0")
	if !v7r1.AtLeast(v5r4) || v5r4.AtLeast(v7r1) || !v5r4.AtLeast(v5r4) {
		t.Error("AtLeast ordering is wrong")
	}
	if v5r4.Compare(v5r4m5) != -1 || v5r4m5.Compare(v5r4) != 1 || v5r4.Compare(v5r4) != 0 {
		t.Error("Compare ordering is wrong")
	}
	if (ibmi.VRM{}).String() != "V?R?M?" || !(ibmi.VRM{}).IsZero() {
		t.Error("zero VRM must report IsZero and a placeholder string")
	}
}
