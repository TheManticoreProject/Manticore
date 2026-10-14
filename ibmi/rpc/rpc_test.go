package rpc_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
)

func mustName(t *testing.T, s string) rpc.ObjectName {
	t.Helper()
	n, err := rpc.NewObjectName(s)
	if err != nil {
		t.Fatalf("NewObjectName(%q): %v", s, err)
	}
	return n
}

func mustParam(t *testing.T) func(rpc.Parameter, error) rpc.Parameter {
	return func(p rpc.Parameter, err error) rpc.Parameter {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
}

func wantInvalid(t *testing.T, what string, err error) {
	t.Helper()
	var ve *rpc.ValidationError
	if !errors.Is(err, rpc.ErrInvalid) || !errors.As(err, &ve) {
		t.Errorf("%s: error = %v; want a ValidationError matching ErrInvalid", what, err)
	}
}

func TestObjectName(t *testing.T) {
	for in, want := range map[string]string{
		"QGY": "QGY", "qgyolaus": "QGYOLAUS", "QSYS": "QSYS", "A": "A",
		"ABCDEFGHIJ": "ABCDEFGHIJ", "$LIB": "$LIB", "#P_1@": "#P_1@", "@X9": "@X9", "MiXeD_9": "MIXED_9",
	} {
		n, err := rpc.NewObjectName(in)
		if err != nil || n.String() != want || n.IsZero() {
			t.Errorf("NewObjectName(%q) = %q, %v; want %q", in, n, err, want)
		}
	}
	for _, in := range []string{
		"", "ABCDEFGHIJK", "1ABC", "_ABC", "*LIBL", "QGY/QGYOLAUS", "\"QGY\"", "A B", "A.B", "A-B", "ÄBC", "QGY\x00", "qgy€",
	} {
		n, err := rpc.NewObjectName(in)
		if !n.IsZero() {
			t.Errorf("NewObjectName(%q) returned %q", in, n)
		}
		wantInvalid(t, fmt.Sprintf("NewObjectName(%q)", in), err)
	}
}

func TestParameterConstructors(t *testing.T) {
	must := mustParam(t)

	in := must(rpc.Input([]byte{1, 2, 3}))
	if in.Direction() != rpc.DirectionInput || in.Maximum() != 3 || in.InputLength() != 3 || !bytes.Equal(in.InputBytes(), []byte{1, 2, 3}) {
		t.Errorf("Input = %v", in)
	}
	empty := rpc.EmptyInput()
	if empty.Direction() != rpc.DirectionInput || empty.Maximum() != 0 || empty.InputBytes() == nil || len(empty.InputBytes()) != 0 {
		t.Errorf("EmptyInput = %v, bytes %v", empty, empty.InputBytes())
	}
	if nilIn := must(rpc.Input(nil)); nilIn.InputBytes() == nil || nilIn.IsZero() {
		t.Error("Input(nil) must be an explicit empty input")
	}
	out := must(rpc.Output(62))
	if out.Direction() != rpc.DirectionOutput || out.Maximum() != 62 || out.InputLength() != 0 || out.InputBytes() != nil {
		t.Errorf("Output = %v", out)
	}
	io := must(rpc.InputOutput([]byte{9}, 4))
	if io.Direction() != rpc.DirectionInputOutput || io.Maximum() != 4 || !bytes.Equal(io.InputBytes(), []byte{9}) {
		t.Errorf("InputOutput = %v", io)
	}
	if exact := must(rpc.InputOutput([]byte{1, 2}, 2)); exact.Maximum() != 2 {
		t.Errorf("exact InputOutput = %v", exact)
	}
	if emptyIO := must(rpc.InputOutput(nil, 8)); emptyIO.InputLength() != 0 {
		t.Errorf("empty InputOutput = %v", emptyIO)
	}
}

func TestParameterRejects(t *testing.T) {
	p, err := rpc.Output(0)
	wantInvalid(t, "Output(0)", err)
	if !p.IsZero() {
		t.Error("Output(0) returned a usable parameter")
	}
	_, err = rpc.InputOutput([]byte{1}, 0)
	wantInvalid(t, "InputOutput(max 0)", err)
	_, err = rpc.InputOutput([]byte{1, 2, 3}, 2)
	wantInvalid(t, "InputOutput(max < input)", err)

	if err := rpc.CheckInputLength(math.MaxUint32); err != nil {
		t.Errorf("CheckInputLength(MaxUint32) = %v", err)
	}
	wantInvalid(t, "CheckInputLength(MaxUint32+1)", rpc.CheckInputLength(math.MaxUint32+1))
}

func TestParameterOwnership(t *testing.T) {
	must := mustParam(t)
	src := []byte{1, 2, 3}
	in := must(rpc.Input(src))
	io := must(rpc.InputOutput(src, 8))
	src[0] = 0xFF
	if in.InputBytes()[0] != 1 || io.InputBytes()[0] != 1 {
		t.Fatal("mutating the caller's slice changed a Parameter")
	}
	got := in.InputBytes()
	got[1] = 0xFF
	if in.InputBytes()[1] != 2 {
		t.Fatal("mutating InputBytes changed the Parameter")
	}
}

func TestParameterFormattingHidesData(t *testing.T) {
	p := mustParam(t)(rpc.Input([]byte("SECRETPASSWORD")))
	c := newTestCall(t, p)
	for _, s := range []string{
		fmt.Sprint(p), fmt.Sprintf("%+v", p), fmt.Sprintf("%#v", p),
		fmt.Sprint(c), fmt.Sprintf("%+v", c), fmt.Sprintf("%#v", c),
		fmt.Sprintf("%v", []rpc.Parameter{p}), fmt.Sprintf("%#v", []rpc.Call{c}),
	} {
		if strings.Contains(s, "SECRET") || strings.Contains(s, "83") {
			t.Errorf("formatted parameter leaks data: %s", s)
		}
	}
}

func newTestCall(t *testing.T, params ...rpc.Parameter) rpc.Call {
	t.Helper()
	c, err := rpc.NewCall(mustName(t, "TESTLIB"), mustName(t, "TESTPGM"), params, rpc.MessagesUpToTen)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewCallRejects(t *testing.T) {
	lib, pgm := mustName(t, "TESTLIB"), mustName(t, "TESTPGM")
	_, err := rpc.NewCall(rpc.ObjectName{}, pgm, nil, rpc.MessagesAll)
	wantInvalid(t, "zero library", err)
	_, err = rpc.NewCall(lib, rpc.ObjectName{}, nil, rpc.MessagesAll)
	wantInvalid(t, "zero program", err)
	for _, o := range []rpc.MessageOption{0, 4, 255} {
		_, err = rpc.NewCall(lib, pgm, nil, o)
		wantInvalid(t, fmt.Sprintf("option %d", o), err)
	}
	_, err = rpc.NewCall(lib, pgm, []rpc.Parameter{rpc.EmptyInput(), {}}, rpc.MessagesAll)
	wantInvalid(t, "zero parameter", err)
	var ve *rpc.ValidationError
	if errors.As(err, &ve) && ve.Field != "parameter[1]" {
		t.Errorf("zero parameter Field = %q", ve.Field)
	}
	if c, err := rpc.NewCall(lib, pgm, nil, rpc.MessagesNone); err != nil || c.ParameterCount() != 0 || c.IsZero() {
		t.Errorf("parameterless call = %v, %v", c, err)
	}
}

func TestCallOwnership(t *testing.T) {
	must := mustParam(t)
	params := []rpc.Parameter{must(rpc.Input([]byte{1})), must(rpc.Output(4))}
	c := newTestCall(t, params...)
	params[0] = must(rpc.Output(99))
	if c.Parameters()[0].Direction() != rpc.DirectionInput {
		t.Fatal("mutating the caller's parameter slice changed the Call")
	}
	got := c.Parameters()
	got[1] = must(rpc.Output(99))
	if c.Parameters()[1].Maximum() != 4 {
		t.Fatal("mutating Parameters() changed the Call")
	}
	if s := c.String(); s != "TESTLIB/TESTPGM(2 parameters)" {
		t.Errorf("String() = %q", s)
	}
}

func TestResultValidate(t *testing.T) {
	must := mustParam(t)
	c := newTestCall(t, must(rpc.Input([]byte{1})), must(rpc.Output(4)), must(rpc.InputOutput([]byte{2}, 3)))
	ok := func() *rpc.Result {
		return &rpc.Result{
			Parameters:             []rpc.ParameterResult{{}, {Present: true, Data: []byte{}}, {Present: true, Data: []byte{7, 8, 9}}},
			RequestedMessageOption: rpc.MessagesUpToTen,
			EffectiveMessageOption: rpc.MessagesUpToTen,
		}
	}
	if err := ok().Validate(c); err != nil {
		t.Fatalf("valid result rejected: %v", err)
	}

	// An input-only slot is absent; a present empty output is distinct.
	r := ok()
	if d, present := r.Output(0); present || d != nil {
		t.Error("input-only slot reported as present")
	}
	if d, present := r.Output(1); !present || d == nil || len(d) != 0 {
		t.Error("present empty output not distinguished from absent")
	}
	if _, present := r.Output(3); present {
		t.Error("out-of-range Output reported present")
	}
	if _, present := r.Output(-1); present {
		t.Error("negative Output reported present")
	}

	// Outputs may be absent when the server returned only messages.
	absent := ok()
	absent.Parameters[1], absent.Parameters[2] = rpc.ParameterResult{}, rpc.ParameterResult{}
	if err := absent.Validate(c); err != nil {
		t.Errorf("absent outputs rejected: %v", err)
	}

	for name, mutate := range map[string]func(*rpc.Result){
		"too few slots":        func(r *rpc.Result) { r.Parameters = r.Parameters[:2] },
		"too many slots":       func(r *rpc.Result) { r.Parameters = append(r.Parameters, rpc.ParameterResult{}) },
		"input slot present":   func(r *rpc.Result) { r.Parameters[0] = rpc.ParameterResult{Present: true, Data: []byte{}} },
		"data without present": func(r *rpc.Result) { r.Parameters[1] = rpc.ParameterResult{Data: []byte{1}} },
		"output over maximum":  func(r *rpc.Result) { r.Parameters[1].Data = []byte{1, 2, 3, 4, 5} },
		"inout over maximum":   func(r *rpc.Result) { r.Parameters[2].Data = []byte{1, 2, 3, 4} },
		"requested mismatch":   func(r *rpc.Result) { r.RequestedMessageOption = rpc.MessagesAll },
		"effective undefined":  func(r *rpc.Result) { r.EffectiveMessageOption = 0 },
	} {
		r := ok()
		mutate(r)
		wantInvalid(t, name, r.Validate(c))
	}
	wantInvalid(t, "nil result", (*rpc.Result)(nil).Validate(c))
}

func TestIBMMessageStringOmitsText(t *testing.T) {
	m := rpc.IBMMessage{ID: "CPF9801", Severity: 40, Text: "Object SECRETOBJ not found", SubstitutionData: []byte("SECRETOBJ")}
	if s := m.String(); s != "CPF9801 (severity 40)" {
		t.Errorf("String() = %q", s)
	}
}
