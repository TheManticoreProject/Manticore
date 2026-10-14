package qgy_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy"
)

func TestLibraries(t *testing.T) {
	if qgy.DefaultLibrary().String() != "QGY" {
		t.Errorf("DefaultLibrary = %s", qgy.DefaultLibrary())
	}
	if qgy.LibraryOr(rpc.ObjectName{}).String() != "QGY" {
		t.Error("the zero library must default to QGY")
	}
	qsys, _ := rpc.NewObjectName("QSYS")
	if qgy.LibraryOr(qsys) != qsys {
		t.Error("an explicit library must be kept")
	}
}

func TestBinary4(t *testing.T) {
	for v, want := range map[int32][]byte{
		0:       {0, 0, 0, 0},
		62:      {0, 0, 0, 0x3e},
		-1:      {0xff, 0xff, 0xff, 0xff},
		1 << 24: {1, 0, 0, 0},
	} {
		if got := qgy.Binary4(v); !bytes.Equal(got, want) {
			t.Errorf("Binary4(%d) = % x, want % x", v, got, want)
		}
	}
}
