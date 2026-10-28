package server

import "testing"

func TestCreditWindowAcceptsEachMessageIdOnce(t *testing.T) {
	w := newCreditWindow()
	if !w.consume(0, 1) || w.consume(0, 1) {
		t.Fatal("MessageId 0 must be accepted once")
	}
	w.grant(8) // MessageIds 1 to 8

	// Out of order is fine, reuse is not.
	for _, id := range []uint64{5, 2, 1} {
		if !w.consume(id, 1) {
			t.Fatalf("MessageId %d refused", id)
		}
	}
	for _, id := range []uint64{5, 2, 1, 0} {
		if w.consume(id, 1) {
			t.Fatalf("MessageId %d accepted twice", id)
		}
	}

	// A multi-credit request is refused whole when any MessageId it spans was used.
	if w.consume(4, 2) {
		t.Fatal("a charge spanning the used MessageId 5 was accepted")
	}
	if !w.consume(3, 2) || !w.consume(6, 3) {
		t.Fatal("unused MessageIds 3, 4 and 6 to 8 were refused")
	}
	if w.consume(9, 1) {
		t.Fatal("a MessageId beyond the granted ones was accepted")
	}
	if w.available() != 0 || len(w.used) != 0 || w.low != 9 {
		t.Fatalf("window = low %d, %d used, %d available; want low 9 and nothing outstanding", w.low, len(w.used), w.available())
	}
}

func TestCreditWindowExpiresSkippedMessageIds(t *testing.T) {
	w := newCreditWindow()
	w.grant(maxCredits)
	// MessageId 0 is never used; every other granted one is, as soon as granted.
	next := uint64(1)
	for i := 0; i < 4*maxWindow; i++ {
		for ; next < w.granted; next++ {
			if !w.consume(next, 1) {
				t.Fatalf("MessageId %d refused", next)
			}
		}
		w.grant(1)
	}
	if span := w.granted - w.low; span > maxWindow {
		t.Fatalf("window spans %d MessageIds, want at most %d", span, maxWindow)
	}
	if len(w.used) > maxWindow {
		t.Fatalf("%d used MessageIds recorded, want at most %d", len(w.used), maxWindow)
	}
	if w.consume(0, 1) {
		t.Fatal("the expired MessageId 0 was accepted")
	}
}
