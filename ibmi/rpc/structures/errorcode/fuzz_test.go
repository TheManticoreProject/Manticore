package errorcode_test

import (
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc/structures/errorcode"
)

func FuzzCheck(f *testing.F) {
	seed := make([]byte, 48)
	binary.BigEndian.PutUint32(seed[4:], 48)
	binary.BigEndian.PutUint32(seed[8:], 40)
	binary.BigEndian.PutUint32(seed[24:], 32)
	binary.BigEndian.PutUint32(seed[28:], 0xFFFFFFFF)
	f.Add(seed)
	f.Fuzz(func(t *testing.T, b []byte) {
		for _, p := range []errorcode.Parameter{must(errorcode.ERRC0100(48)), must(errorcode.ERRC0200(48))} {
			err := p.Check(lib, pgm, rpc.ParameterResult{Present: true, Data: b}, codec)
			if ae, ok := err.(*errorcode.APIError); ok && len(ae.ExceptionData) > len(b) {
				t.Fatalf("exception data larger than the buffer")
			}
		}
	})
}

func must(p errorcode.Parameter, err error) errorcode.Parameter {
	if err != nil {
		panic(err)
	}
	return p
}
