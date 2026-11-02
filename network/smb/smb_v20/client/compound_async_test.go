package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/types"
)

// TestCompoundAwaitsPendingAndSeparateFrames covers a compound answered the way
// Windows answers QUERY_INFO followed by CHANGE_NOTIFY: the QUERY_INFO result
// first, then an interim STATUS_PENDING for the notify, then its final response
// — each in a frame of its own, or the interim chained with the first. The
// compound must return both final responses in request order and leave no frame
// unread, so the next request gets its own answer.
func TestCompoundAwaitsPendingAndSeparateFrames(t *testing.T) {
	queryResp := commands.NewQueryInfoResponse()
	queryResp.OutputBuffer = []byte{0xAA}
	notifyFinal := commands.NewChangeNotifyResponse()
	nextResp := commands.NewQueryInfoResponse()
	nextResp.OutputBuffer = []byte{0xBB}

	for _, tc := range []struct {
		name   string
		frames func(t *testing.T) [][]byte
	}{
		{"every response in its own frame", func(t *testing.T) [][]byte {
			return [][]byte{
				cannedCompound(t, []compoundSegment{{queryResp, 0, 0}}),
				cannedCompound(t, []compoundSegment{{commands.NewChangeNotifyResponse(), 1, ntStatusPending}}),
				cannedCompound(t, []compoundSegment{{notifyFinal, 1, 0}}),
			}
		}},
		{"interim chained with the first response", func(t *testing.T) [][]byte {
			return [][]byte{
				cannedCompound(t, []compoundSegment{{queryResp, 0, 0}, {commands.NewChangeNotifyResponse(), 1, ntStatusPending}}),
				cannedCompound(t, []compoundSegment{{notifyFinal, 1, 0}}),
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			frames := tc.frames(t)
			frames = append(frames, cannedCompound(t, []compoundSegment{{nextResp, 2, 0}}))
			ft := &fakeTransport{responses: frames}
			c := withConnectedTree(ft)

			query := commands.NewQueryInfoRequest()
			query.InfoType = types.UCHAR(commands.SMB2_0_INFO_FILE)
			query.FileInfoClass = 4
			query.OutputBufferLength = 0x1000
			notify := commands.NewChangeNotifyRequest()
			notify.OutputBufferLength = 0x1000

			responses, err := c.SendUnrelatedCompound([]*message.Message{c.NewRequest(query), c.NewRequest(notify)})
			if err != nil {
				t.Fatalf("SendUnrelatedCompound: %v", err)
			}
			if len(responses) != 2 {
				t.Fatalf("got %d responses, want 2", len(responses))
			}
			if _, ok := responses[0].Command.(*commands.QueryInfoResponse); !ok || responses[0].Header.Status != 0 {
				t.Errorf("response 0 = %T status 0x%08x, want the QUERY_INFO result", responses[0].Command, responses[0].Header.Status)
			}
			if _, ok := responses[1].Command.(*commands.ChangeNotifyResponse); !ok || responses[1].Header.Status != 0 {
				t.Errorf("response 1 = %T status 0x%08x, want the notify's final response", responses[1].Command, responses[1].Header.Status)
			}

			// The next request reads its own answer, not a leftover of the compound.
			out, err := c.QueryInfo(types.SMB2_FILEID{}, 1, 4, 0)
			if err != nil || len(out) != 1 || out[0] != 0xBB {
				t.Fatalf("the request after the compound got %x, %v", out, err)
			}
		})
	}
}

// TestCompoundRejectsADuplicateFinalResponse checks a server answering one
// request twice is refused rather than silently overwriting the first answer.
func TestCompoundRejectsADuplicateFinalResponse(t *testing.T) {
	a, b := commands.NewQueryInfoResponse(), commands.NewQueryInfoResponse()
	ft := &fakeTransport{responses: [][]byte{
		cannedCompound(t, []compoundSegment{{a, 0, 0}, {b, 0, 0}}),
	}}
	c := withConnectedTree(ft)
	q1, q2 := commands.NewQueryInfoRequest(), commands.NewQueryInfoRequest()
	if _, err := c.SendUnrelatedCompound([]*message.Message{c.NewRequest(q1), c.NewRequest(q2)}); err == nil {
		t.Fatal("a compound answering one request twice was accepted")
	}
}

// TestCompoundRejectsAMismatchedCommand checks a segment answering a request
// with another command is refused.
func TestCompoundRejectsAMismatchedCommand(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{
		cannedCompound(t, []compoundSegment{{commands.NewQueryInfoResponse(), 0, 0}, {commands.NewCloseResponse(), 1, 0}}),
	}}
	c := withConnectedTree(ft)
	q1, q2 := commands.NewQueryInfoRequest(), commands.NewQueryInfoRequest()
	if _, err := c.SendUnrelatedCompound([]*message.Message{c.NewRequest(q1), c.NewRequest(q2)}); err == nil {
		t.Fatal("a CLOSE response to a QUERY_INFO request was accepted")
	}
}

// TestCompoundClearsPendingStateOnError checks an error after an interim
// response leaves no recorded async operation for a later Cancel to target.
func TestCompoundClearsPendingStateOnError(t *testing.T) {
	ft := &fakeTransport{responses: [][]byte{
		cannedCompound(t, []compoundSegment{{commands.NewQueryInfoResponse(), 0, 0}, {commands.NewChangeNotifyResponse(), 1, ntStatusPending}}),
		// The final response answers request 0 a second time: an error.
		cannedCompound(t, []compoundSegment{{commands.NewQueryInfoResponse(), 0, 0}}),
	}}
	c := withConnectedTree(ft)
	query, notify := commands.NewQueryInfoRequest(), commands.NewChangeNotifyRequest()
	if _, err := c.SendUnrelatedCompound([]*message.Message{c.NewRequest(query), c.NewRequest(notify)}); err == nil {
		t.Fatal("the compound succeeded")
	}
	if messageId, asyncId, ok := c.pendingAsync(); ok {
		t.Errorf("a pending async operation is still recorded after the error: message %d async %d", messageId, asyncId)
	}
}
