package commands_test

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestSecondaryRequestWithOneEmptyBlockRoundTrips asserts each secondary
// request decodes when it carries only parameters or only data.
//
// A secondary continues whichever block the primary left unfinished, so one of
// the two is routinely empty, and Marshal writes the empty block's offset as
// zero. A decoder that measured the gap between the blocks from that zero offset
// refused the message, and the server answered the fragment with
// STATUS_INVALID_SMB.
func TestSecondaryRequestWithOneEmptyBlockRoundTrips(t *testing.T) {
	payload := []types.UCHAR{0x01, 0x02, 0x03, 0x04}

	type secondary struct {
		message interface {
			Marshal() ([]byte, error)
			Unmarshal([]byte) (int, error)
		}
		blocks func() (parameters, data []types.UCHAR)
	}
	build := map[string]func(parameters, data []types.UCHAR) (out, in secondary){
		"TRANSACTION_SECONDARY": func(parameters, data []types.UCHAR) (secondary, secondary) {
			out := commands.NewTransactionSecondaryRequest()
			out.Trans2_Parameters, out.ParameterCount = parameters, types.USHORT(len(parameters))
			out.Trans2_Data, out.DataCount = data, types.USHORT(len(data))
			in := commands.NewTransactionSecondaryRequest()
			return secondary{out, nil}, secondary{in, func() ([]types.UCHAR, []types.UCHAR) { return in.Trans2_Parameters, in.Trans2_Data }}
		},
		"TRANSACTION2_SECONDARY": func(parameters, data []types.UCHAR) (secondary, secondary) {
			out := commands.NewTransaction2SecondaryRequest()
			out.Trans2_Parameters, out.ParameterCount = parameters, types.USHORT(len(parameters))
			out.Trans2_Data, out.DataCount = data, types.USHORT(len(data))
			in := commands.NewTransaction2SecondaryRequest()
			return secondary{out, nil}, secondary{in, func() ([]types.UCHAR, []types.UCHAR) { return in.Trans2_Parameters, in.Trans2_Data }}
		},
		"NT_TRANSACT_SECONDARY": func(parameters, data []types.UCHAR) (secondary, secondary) {
			out := commands.NewNtTransactSecondaryRequest()
			out.NT_Trans_Parameters, out.ParameterCount = parameters, types.ULONG(len(parameters))
			out.NT_Trans_Data, out.DataCount = data, types.ULONG(len(data))
			in := commands.NewNtTransactSecondaryRequest()
			return secondary{out, nil}, secondary{in, func() ([]types.UCHAR, []types.UCHAR) { return in.NT_Trans_Parameters, in.NT_Trans_Data }}
		},
	}

	for name, construct := range build {
		for _, layout := range []struct {
			name             string
			parameters, data []types.UCHAR
		}{
			{"parameters only", payload, []types.UCHAR{}},
			{"data only", []types.UCHAR{}, payload},
		} {
			out, in := construct(layout.parameters, layout.data)
			raw, err := out.message.Marshal()
			if err != nil {
				t.Fatalf("%s with %s: Marshal() error = %v", name, layout.name, err)
			}
			if _, err := in.message.Unmarshal(raw); err != nil {
				t.Errorf("%s with %s: Unmarshal() error = %v", name, layout.name, err)
				continue
			}
			parameters, data := in.blocks()
			if !bytes.Equal(parameters, layout.parameters) || !bytes.Equal(data, layout.data) {
				t.Errorf("%s with %s decoded parameters % x and data % x, want % x and % x",
					name, layout.name, parameters, data, layout.parameters, layout.data)
			}
		}
	}
}
