package ndr

import (
	"bytes"
	"testing"
)

// levelUnion is a sample encapsulated NDR union switched on Level: arm 1 is a uint32,
// arm 2 is a uint16.
type levelUnion struct {
	Level uint32
	AsU32 uint32
	AsU16 uint16
}

func (u *levelUnion) SwitchValue() uint32 { return u.Level }
func (u *levelUnion) MarshalArm(e *Encoder, sw uint32) error {
	switch sw {
	case 1:
		e.WriteUint32(u.AsU32)
	case 2:
		e.WriteUint16(u.AsU16)
	}
	return nil
}
func (u *levelUnion) UnmarshalArm(d *Decoder, sw uint32) error {
	u.Level = sw // the union records the discriminant the walker decoded
	switch sw {
	case 1:
		v, err := d.ReadUint32()
		u.AsU32 = v
		return err
	case 2:
		v, err := d.ReadUint16()
		u.AsU16 = v
		return err
	}
	return nil
}

func TestUnion_Arm1(t *testing.T) {
	type wrap struct {
		U levelUnion
	}
	raw, err := Marshal(&wrap{U: levelUnion{Level: 1, AsU32: 0xAABBCCDD}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x01, 0, 0, 0, // discriminant (switch value)
		0xDD, 0xCC, 0xBB, 0xAA, // arm 1 (uint32)
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("union arm 1:\n got %x\nwant %x", raw, want)
	}
	var out wrap
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.U.Level != 1 || out.U.AsU32 != 0xAABBCCDD {
		t.Errorf("round trip: got %+v", out.U)
	}
}

func TestUnion_Arm2(t *testing.T) {
	type wrap struct {
		U levelUnion
	}
	raw, err := Marshal(&wrap{U: levelUnion{Level: 2, AsU16: 0xBEEF}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := []byte{
		0x02, 0, 0, 0, // discriminant
		0xEF, 0xBE, // arm 2 (uint16)
	}
	if !bytes.Equal(raw, want) {
		t.Errorf("union arm 2:\n got %x\nwant %x", raw, want)
	}
	var out wrap
	if err := Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.U.Level != 2 || out.U.AsU16 != 0xBEEF {
		t.Errorf("round trip: got %+v", out.U)
	}
}
