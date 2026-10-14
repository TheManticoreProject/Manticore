package qgyclst_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/qgy/qgyclst"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/openlist"
)

var codec, _ = ccsid.Lookup(ccsid.CCSID37)

type fakeCaller struct {
	calls []rpc.Call
	err   error
}

func (f *fakeCaller) Call(ctx context.Context, c rpc.Call) (*rpc.Result, error) {
	f.calls = append(f.calls, c)
	if f.err != nil {
		return nil, f.err
	}
	return &rpc.Result{Parameters: make([]rpc.ParameterResult, c.ParameterCount()), RequestedMessageOption: c.MessageOption(), EffectiveMessageOption: c.MessageOption()}, nil
}

func TestClose(t *testing.T) {
	fc := &fakeCaller{}
	h := openlist.Handle{1, 2, 3, 4}
	if err := qgyclst.Close(context.Background(), fc, qgyclst.Request{Handle: h}, codec); err != nil {
		t.Fatal(err)
	}
	c := fc.calls[0]
	p := c.Parameters()
	if c.Library().String() != "QGY" || c.Program().String() != "QGYCLST" || len(p) != 2 ||
		p[0].Direction() != rpc.DirectionInput || !bytes.Equal(p[0].InputBytes(), h[:]) ||
		p[1].Direction() != rpc.DirectionInput || p[1].Maximum() != 8 {
		t.Errorf("call = %v %v", c, p)
	}
	fc.err = errors.New("native failure")
	if err := qgyclst.Close(context.Background(), fc, qgyclst.Request{Handle: h}, codec); !errors.Is(err, fc.err) {
		t.Errorf("caller error = %v", err)
	}
}
