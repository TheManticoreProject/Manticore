package dsrepl_test

// strptr returns a pointer to s, for populating the optional (*string) offset
// string fields of the DS_REPL_*_BLOB structures in tests.
func strptr(s string) *string {
	return &s
}

// sameStr reports whether two optional (*string) values are equal: both nil, or
// both non-nil and pointing to equal strings.
func sameStr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
