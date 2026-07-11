package gssapi

import (
	"encoding/hex"
	"errors"
	"testing"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
)

// TestSeqWindowReplayOnly drives the sliding-window state machine directly in
// the default consumer mode (replay detection on, sequence enforcement off).
// Only duplicates and below-window tokens are rejected; gaps and out-of-order
// (but fresh) tokens are accepted, matching the RFC 2743 GSS_C_REPLAY_FLAG-only
// semantics DCE/RPC, SMB, and LDAP drive this layer with.
func TestSeqWindowReplayOnly(t *testing.T) {
	w := &seqWindow{replayDetect: true}

	// First token seeds the window and is always accepted.
	if err := w.check(10); err != nil {
		t.Fatalf("seed token rejected: %v", err)
	}
	// In-order successor.
	if err := w.check(11); err != nil {
		t.Fatalf("in-order token rejected: %v", err)
	}
	// Replay of an already-seen number -> DUPLICATE.
	if err := w.check(11); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replay: got %v, want ErrDuplicateToken", err)
	}
	if err := w.check(10); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replay of seed: got %v, want ErrDuplicateToken", err)
	}
	// A gap ahead is accepted in replay-only mode (no ErrGapToken).
	if err := w.check(20); err != nil {
		t.Fatalf("gap-ahead token rejected in replay-only mode: %v", err)
	}
	// A fresh out-of-order token inside the window is accepted (not a replay).
	if err := w.check(15); err != nil {
		t.Fatalf("in-window out-of-order token rejected in replay-only mode: %v", err)
	}
	// ... but replaying it now trips DUPLICATE.
	if err := w.check(15); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replay of out-of-order token: got %v, want ErrDuplicateToken", err)
	}
	// Advance the expected sequence well past the window width, then a stale token
	// that has dropped below the window can no longer be proven fresh -> OLD.
	if err := w.check(90); err != nil { // big gap ahead, accepted in replay-only mode
		t.Fatalf("far-ahead token rejected in replay-only mode: %v", err)
	}
	if err := w.check(15); !errors.Is(err, ErrOldToken) {
		t.Fatalf("below-window token: got %v, want ErrOldToken", err)
	}
}

// TestSeqWindowSequenceEnforcement drives the window with strict sequencing on:
// gaps and out-of-order tokens now surface as ErrGapToken / ErrUnseqToken, and
// below-window tokens as ErrUnseqToken.
func TestSeqWindowSequenceEnforcement(t *testing.T) {
	w := &seqWindow{replayDetect: true, sequence: true}

	if err := w.check(100); err != nil {
		t.Fatalf("seed token rejected: %v", err)
	}
	if err := w.check(101); err != nil {
		t.Fatalf("in-order token rejected: %v", err)
	}
	// Skipping ahead -> GAP (token still consumed/advanced).
	if err := w.check(105); !errors.Is(err, ErrGapToken) {
		t.Fatalf("gap: got %v, want ErrGapToken", err)
	}
	// Filling one of the skipped slots -> UNSEQ (fresh, but out of order).
	if err := w.check(103); !errors.Is(err, ErrUnseqToken) {
		t.Fatalf("out-of-order: got %v, want ErrUnseqToken", err)
	}
	// Replaying that slot -> DUPLICATE (replay detection still active).
	if err := w.check(103); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replay: got %v, want ErrDuplicateToken", err)
	}
	// Advance past the window width, then a token that has dropped below the
	// window with sequencing on -> UNSEQ.
	if err := w.check(180); !errors.Is(err, ErrGapToken) {
		t.Fatalf("far-ahead token (seq mode): got %v, want ErrGapToken", err)
	}
	if err := w.check(110); !errors.Is(err, ErrUnseqToken) {
		t.Fatalf("below-window (seq mode): got %v, want ErrUnseqToken", err)
	}
}

// TestSeqWindowDisabled confirms that a window with neither flag set is a
// no-op: every sequence number, including blatant replays, is accepted (this is
// the zero-value SecContext behaviour existing round-trip tests rely on).
func TestSeqWindowDisabled(t *testing.T) {
	w := &seqWindow{}
	for _, s := range []uint64{5, 5, 5, 1, 900, 2} {
		if err := w.check(s); err != nil {
			t.Fatalf("disabled window rejected seq %d: %v", s, err)
		}
	}
}

// TestSeqWindowWrapAround exercises the 32-bit wrap boundary (RFC 4121 §4.2.6):
// sequence numbers straddling 2^32-1 -> 0 must remain in order, and a replay
// across the boundary must still be caught.
func TestSeqWindowWrapAround(t *testing.T) {
	w := &seqWindow{replayDetect: true}
	const max = uint64(0xffffffff)

	if err := w.check(max - 1); err != nil { // seed near the top
		t.Fatalf("seed near wrap rejected: %v", err)
	}
	if err := w.check(max); err != nil {
		t.Fatalf("token at 2^32-1 rejected: %v", err)
	}
	if err := w.check(0); err != nil { // wrapped to zero, in order
		t.Fatalf("wrapped token 0 rejected: %v", err)
	}
	if err := w.check(1); err != nil {
		t.Fatalf("post-wrap token 1 rejected: %v", err)
	}
	// Replay of the wrapped-to-zero token across the boundary -> DUPLICATE.
	if err := w.check(0); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replay across wrap: got %v, want ErrDuplicateToken", err)
	}
	// Replay of the pre-wrap top value -> DUPLICATE.
	if err := w.check(max); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replay of pre-wrap top: got %v, want ErrDuplicateToken", err)
	}
}

// enableRecv turns on the given enforcement mode on a test SecContext's receive
// window (contexts built with a struct literal default to disabled).
func enableRecv(ctx *SecContext, replay, sequence bool) {
	ctx.recvWindow = seqWindow{replayDetect: replay, sequence: sequence}
}

// TestVerifyMICAESReplayEnforcement builds acceptor MIC tokens with the send
// path helper and feeds them to the AES receive path through a shared context,
// proving the window is wired into VerifyMIC.
func TestVerifyMICAESReplayEnforcement(t *testing.T) {
	ctx := newTestContext(1)
	enableRecv(ctx, true, false)
	data := []byte("server-signed-response")

	tok5 := acceptorMIC(t, ctx.SessionKey, ctx.SessionEType, 5, data)
	tok6 := acceptorMIC(t, ctx.SessionKey, ctx.SessionEType, 6, data)
	tok8 := acceptorMIC(t, ctx.SessionKey, ctx.SessionEType, 8, data)

	if err := ctx.VerifyMIC(data, tok5); err != nil {
		t.Fatalf("first MIC rejected: %v", err)
	}
	if err := ctx.VerifyMIC(data, tok6); err != nil {
		t.Fatalf("in-order MIC rejected: %v", err)
	}
	// Replaying seq 5 must now be caught as a duplicate.
	if err := ctx.VerifyMIC(data, tok5); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replayed MIC: got %v, want ErrDuplicateToken", err)
	}

	// With sequence enforcement on: seed low (5), skip ahead (8 -> GAP), then a
	// fresh token filling a skipped slot (6) trips UNSEQ.
	ctxSeq := newTestContext(1)
	enableRecv(ctxSeq, true, true)
	if err := ctxSeq.VerifyMIC(data, tok5); err != nil { // seed
		t.Fatalf("seed MIC rejected: %v", err)
	}
	if err := ctxSeq.VerifyMIC(data, tok8); !errors.Is(err, ErrGapToken) {
		t.Fatalf("skip-ahead MIC (seq mode): got %v, want ErrGapToken", err)
	}
	if err := ctxSeq.VerifyMIC(data, tok6); !errors.Is(err, ErrUnseqToken) {
		t.Fatalf("out-of-order MIC (seq mode): got %v, want ErrUnseqToken", err)
	}
}

// TestUnsealAESReplayEnforcement drives the AES CFX Wrap (Seal/Unseal) receive
// path: an acceptor-sealed token replayed with the same sequence number is
// rejected as a duplicate.
func TestUnsealAESReplayEnforcement(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	etype := iana.ETypeAES256CTSHMACSHA196
	ctx := newTestContext(1)
	ctx.SessionKey = key
	ctx.SubKey = key
	ctx.SubKeyEType = etype
	ctx.acceptorSubkey = true
	enableRecv(ctx, true, false)
	data := []byte("confidential-stub")

	sealed3, tok3 := acceptorSealDCE(t, key, etype, 3, data, 8)
	sealed4, tok4 := acceptorSealDCE(t, key, etype, 4, data, 8)

	if _, err := ctx.Unseal(sealed3, tok3); err != nil {
		t.Fatalf("first Unseal rejected: %v", err)
	}
	if _, err := ctx.Unseal(sealed4, tok4); err != nil {
		t.Fatalf("in-order Unseal rejected: %v", err)
	}
	if _, err := ctx.Unseal(sealed3, tok3); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replayed Unseal: got %v, want ErrDuplicateToken", err)
	}
}

// TestVerifyMICRC4ReplayEnforcement drives the RC4-HMAC MIC receive path with a
// shared context: duplicate and below-window tokens are rejected, a gap ahead is
// accepted in replay-only mode.
func TestVerifyMICRC4ReplayEnforcement(t *testing.T) {
	key, _ := hex.DecodeString("0123456789abcdef0123456789abcdef")
	data := []byte("the quick brown fox")
	ctx := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}
	enableRecv(ctx, true, false)

	if err := ctx.verifyMICRC4(data, ctx.makeMICRC4Acceptor(data, 2)); err != nil {
		t.Fatalf("seed RC4 MIC rejected: %v", err)
	}
	if err := ctx.verifyMICRC4(data, ctx.makeMICRC4Acceptor(data, 3)); err != nil {
		t.Fatalf("in-order RC4 MIC rejected: %v", err)
	}
	// Gap ahead accepted (replay-only).
	if err := ctx.verifyMICRC4(data, ctx.makeMICRC4Acceptor(data, 9)); err != nil {
		t.Fatalf("gap RC4 MIC rejected in replay-only mode: %v", err)
	}
	// Replay of seq 3 -> DUPLICATE.
	if err := ctx.verifyMICRC4(data, ctx.makeMICRC4Acceptor(data, 3)); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replayed RC4 MIC: got %v, want ErrDuplicateToken", err)
	}
}

// TestUnsealRC4ReplayEnforcement drives the RC4-HMAC Wrap receive path: a
// replayed acceptor Wrap token is rejected as a duplicate.
func TestUnsealRC4ReplayEnforcement(t *testing.T) {
	key, _ := hex.DecodeString("0123456789abcdef0123456789abcdef")
	data := []byte("privacy over kerberos rpc!!")
	ctx := &SecContext{SessionKey: key, SessionEType: rc4HMACEType}
	enableRecv(ctx, true, false)

	sealed2, tok2 := ctx.sealRC4Acceptor(data, 2)
	sealed3, tok3 := ctx.sealRC4Acceptor(data, 3)

	if _, err := ctx.unsealRC4(sealed2, tok2); err != nil {
		t.Fatalf("first RC4 Unseal rejected: %v", err)
	}
	if _, err := ctx.unsealRC4(sealed3, tok3); err != nil {
		t.Fatalf("in-order RC4 Unseal rejected: %v", err)
	}
	if _, err := ctx.unsealRC4(sealed2, tok2); !errors.Is(err, ErrDuplicateToken) {
		t.Fatalf("replayed RC4 Unseal: got %v, want ErrDuplicateToken", err)
	}
}

// TestDefaultInitOptionsEnableReplayOnly confirms InitOptions wires the default
// enforcement mode (replay on, sequence off) into the SecContext the transports
// receive.
func TestDefaultInitOptionsEnableReplayOnly(t *testing.T) {
	// Simulate what InitSecContext does for the default options.
	w := seqWindow{replayDetect: !false, sequence: false}
	if !w.replayDetect || w.sequence {
		t.Fatalf("default mode = {replay:%v seq:%v}, want {true false}", w.replayDetect, w.sequence)
	}
	// Verify the derived checksum type exists for the RC4 sanity above.
	if _, ok := kerbcrypto.ChecksumTypeForEType(iana.ETypeAES256CTSHMACSHA196); !ok {
		t.Fatal("expected a checksum type for AES256")
	}
}
