package rpcserver

import (
	"bytes"
	"fmt"
	"sync"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/pdu"
	"github.com/TheManticoreProject/Manticore/windows/credentials"
)

// echoService answers opnum 0 by returning the request stub unchanged, so a
// concurrent caller can confirm it received the answer to its own call (routing)
// and that the per-message signature over that stub verified (authentication).
type echoService struct{}

func (echoService) AbstractSyntax() syntax.SyntaxID { return testSyntax() }
func (echoService) Call(opnum uint16, stub []byte) ([]byte, error) {
	if opnum != 0 {
		return nil, ErrUnknownOpnum
	}
	return append([]byte(nil), stub...), nil
}

// concurrentLoopTransport is a loopTransport that is safe for the client's reader
// goroutine (Recv) to run concurrently with caller goroutines (Send). The plain
// loopTransport above touches its reply slice without synchronization and assumes
// serial use; this one guards the queue and blocks Recv until a reply is available,
// modelling a real transport so concurrent authenticated Call goroutines exercise
// the per-message protect (senders) and unprotect (reader) paths under -race.
type concurrentLoopTransport struct {
	association *Association
	mu          sync.Mutex
	cond        *sync.Cond
	replies     [][]byte
	closed      bool
}

func newConcurrentLoopTransport(a *Association) *concurrentLoopTransport {
	t := &concurrentLoopTransport{association: a}
	t.cond = sync.NewCond(&t.mu)
	return t
}

func (t *concurrentLoopTransport) Connect() error { return nil }

func (t *concurrentLoopTransport) Send(p []byte) error {
	// The client holds its send lock across Send, so association.Handle is reached
	// serially and in request order; the server association is therefore used by one
	// goroutine at a time, exactly as over a real ordered transport.
	reply, err := t.association.Handle(append([]byte(nil), p...))
	if err != nil {
		return err
	}
	if len(reply) > 0 {
		t.mu.Lock()
		t.replies = append(t.replies, reply)
		t.cond.Signal()
		t.mu.Unlock()
	}
	return nil
}

func (t *concurrentLoopTransport) Recv() ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for len(t.replies) == 0 && !t.closed {
		t.cond.Wait()
	}
	if len(t.replies) == 0 {
		return nil, fmt.Errorf("concurrentLoopTransport closed")
	}
	reply := t.replies[0]
	t.replies = t.replies[1:]
	return reply, nil
}

func (t *concurrentLoopTransport) Close() error {
	t.mu.Lock()
	t.closed = true
	t.cond.Broadcast()
	t.mu.Unlock()
	return nil
}

func (t *concurrentLoopTransport) MaxXmitFrag() uint16 { return 4280 }
func (t *concurrentLoopTransport) MaxRecvFrag() uint16 { return 4280 }

// TestNTLMConcurrentCallsPerMessageAuth drives many authenticated Calls at once on a
// single association and checks each one's per-message signature verifies and each
// goroutine gets its own stub back. The client's reader goroutine unprotects response
// fragments as they arrive while caller goroutines protect request fragments under the
// send lock; because the NTLM context keeps disjoint send (outSeq/client keys) and
// receive (inSeq/server keys) state, this is correct and, run under -race, data-race
// free. Covers #1958 for PKT_INTEGRITY and PKT_PRIVACY.
func TestNTLMConcurrentCallsPerMessageAuth(t *testing.T) {
	for _, lv := range []struct {
		name  string
		level uint8
	}{
		{"PKT_INTEGRITY", pdu.AuthLevelPktIntegrity},
		{"PKT_PRIVACY", pdu.AuthLevelPktPrivacy},
	} {
		t.Run(lv.name, func(t *testing.T) {
			d := ntlmEndpoint(t, echoService{})
			tr := newConcurrentLoopTransport(d.Open())
			c := client.NewClient(tr)
			creds, err := credentials.NewCredentials("LAB", "alice", "Passw0rd!", "")
			if err != nil {
				t.Fatal(err)
			}
			if err := c.SetAuth(pdu.AuthTypeNTLMSSP, lv.level, creds); err != nil {
				t.Fatal(err)
			}
			if err := c.Bind(testSyntax()); err != nil {
				t.Fatalf("bind: %v", err)
			}
			defer c.Close()

			const workers = 24
			var wg sync.WaitGroup
			errs := make([]error, workers)
			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					// A distinct stub per call, of a distinct length, so the response
					// must be matched to this call and the signature must verify over
					// exactly these bytes.
					want := []byte(fmt.Sprintf("call-%02d-%s", idx, bytes.Repeat([]byte{'x'}, idx%7)))
					got, err := c.Call(0, want)
					if err != nil {
						errs[idx] = err
						return
					}
					if !bytes.Equal(got, want) {
						errs[idx] = fmt.Errorf("got %q, want %q", got, want)
					}
				}(i)
			}
			wg.Wait()
			for i, e := range errs {
				if e != nil {
					t.Fatalf("worker %d: %v", i, e)
				}
			}
		})
	}
}
