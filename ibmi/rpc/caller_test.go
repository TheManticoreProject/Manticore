package rpc_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
)

// fakeCaller is a transport-free rpc.Caller: it records each call and
// answers by filling every output-capable parameter with a pattern.
type fakeCaller struct {
	calls []rpc.Call
}

func (f *fakeCaller) Call(ctx context.Context, c rpc.Call) (*rpc.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.calls = append(f.calls, c)
	res := &rpc.Result{
		RequestedMessageOption: c.MessageOption(),
		EffectiveMessageOption: c.MessageOption(),
		Messages:               []rpc.IBMMessage{{ID: "CPC2191", Severity: 0}},
		CorrelationID:          uint32(len(f.calls)),
	}
	for _, p := range c.Parameters() {
		slot := rpc.ParameterResult{}
		switch p.Direction() {
		case rpc.DirectionOutput:
			slot = rpc.ParameterResult{Present: true, Data: make([]byte, p.Maximum())}
		case rpc.DirectionInputOutput:
			slot = rpc.ParameterResult{Present: true, Data: append(p.InputBytes(), 0xEE)}
		}
		res.Parameters = append(res.Parameters, slot)
	}
	return res, res.Validate(c)
}

// fakeExclusive serializes callbacks and invalidates the scoped caller once
// the callback returns.
type fakeExclusive struct {
	mu    sync.Mutex
	inner rpc.Caller
}

var errCallerReleased = errors.New("fake: exclusive caller used after return")

type scopedCaller struct {
	inner    rpc.Caller
	released bool
}

func (s *scopedCaller) Call(ctx context.Context, c rpc.Call) (*rpc.Result, error) {
	if s.released {
		return nil, errCallerReleased
	}
	return s.inner.Call(ctx, c)
}

func (f *fakeExclusive) ExecuteExclusive(ctx context.Context, fn func(rpc.Caller) error) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	scoped := &scopedCaller{inner: f.inner}
	defer func() { scoped.released = true }()
	return fn(scoped)
}

var (
	_ rpc.Caller          = (*fakeCaller)(nil)
	_ rpc.ExclusiveCaller = (*fakeExclusive)(nil)
)

// TestFakeCallerRoundTrip is the Phase 1 exit gate: a complete program call
// is built, observed and answered through the neutral interfaces alone.
func TestFakeCallerRoundTrip(t *testing.T) {
	must := mustParam(t)
	call, err := rpc.NewCall(mustName(t, "testlib"), mustName(t, "testpgm"), []rpc.Parameter{
		must(rpc.Input([]byte{0xC1, 0xC2})),
		rpc.EmptyInput(),
		must(rpc.Output(5)),
		must(rpc.InputOutput([]byte{0x01}, 4)),
	}, rpc.MessagesAll)
	if err != nil {
		t.Fatal(err)
	}

	fake := &fakeCaller{}
	var caller rpc.Caller = fake
	res, err := caller.Call(context.Background(), call)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}

	if len(fake.calls) != 1 {
		t.Fatalf("fake observed %d calls", len(fake.calls))
	}
	seen := fake.calls[0]
	if seen.Library().String() != "TESTLIB" || seen.Program().String() != "TESTPGM" || seen.MessageOption() != rpc.MessagesAll {
		t.Errorf("fake observed %v option %v", seen, seen.MessageOption())
	}
	wantDirs := []rpc.Direction{rpc.DirectionInput, rpc.DirectionInput, rpc.DirectionOutput, rpc.DirectionInputOutput}
	for i, p := range seen.Parameters() {
		if p.Direction() != wantDirs[i] {
			t.Errorf("parameter %d direction %v, want %v", i, p.Direction(), wantDirs[i])
		}
	}

	if err := res.Validate(call); err != nil {
		t.Fatalf("result does not align with call: %v", err)
	}
	for i, want := range []struct {
		present bool
		length  int
	}{{false, 0}, {false, 0}, {true, 5}, {true, 2}} {
		d, present := res.Output(i)
		if present != want.present || len(d) != want.length {
			t.Errorf("Output(%d) = %d bytes, present %v", i, len(d), present)
		}
	}
	if len(res.Messages) != 1 || res.Messages[0].ID != "CPC2191" || res.CorrelationID != 1 {
		t.Errorf("result metadata = %+v", res)
	}
}

func TestFakeExclusiveScopesCaller(t *testing.T) {
	call := newTestCall(t, rpc.EmptyInput())
	ex := &fakeExclusive{inner: &fakeCaller{}}
	var leaked rpc.Caller
	err := ex.ExecuteExclusive(context.Background(), func(c rpc.Caller) error {
		leaked = c
		for i := 0; i < 3; i++ {
			if _, err := c.Call(context.Background(), call); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := leaked.Call(context.Background(), call); !errors.Is(err, errCallerReleased) {
		t.Errorf("use after return error = %v", err)
	}
}

func TestFakeCallerHonorsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fake := &fakeCaller{}
	if _, err := fake.Call(ctx, newTestCall(t)); !errors.Is(err, context.Canceled) || len(fake.calls) != 0 {
		t.Errorf("cancelled call = %v, recorded %d", err, len(fake.calls))
	}
}
