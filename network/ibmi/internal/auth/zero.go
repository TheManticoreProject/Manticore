package auth

// Best-effort clearing of secret intermediates. Go can still leave copies in
// strings, stack slots or reallocated buffers, so this narrows exposure but
// cannot guarantee erasure.

func wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func wipeRunes(r []rune) {
	for i := range r {
		r[i] = 0
	}
}

func wipeUint16(u []uint16) {
	for i := range u {
		u[i] = 0
	}
}
