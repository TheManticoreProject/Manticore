package server

import (
	"strings"
	"testing"
)

// resolvePath is the containment boundary: every path a backend sees passes
// through it, and a backend is entitled to assume the result cannot escape the
// share. These tests are the case for that assumption, so they are deliberately
// adversarial rather than illustrative.

// TestResolvePathAccepts asserts the forms a client legitimately sends resolve to
// the share-relative form a backend is promised.
func TestResolvePathAccepts(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{"", ""},
		{"\\", ""},
		{"/", ""},
		{"\\file.txt", "file.txt"},
		{"file.txt", "file.txt"},
		{"\\dir\\file.txt", "dir/file.txt"},
		{"dir\\sub\\file.txt", "dir/sub/file.txt"},
		// Forward slashes mean the same thing.
		{"/dir/file.txt", "dir/file.txt"},
		// A trailing separator is redundant but clients send it.
		{"\\dir\\", "dir"},
		{"\\dir\\sub\\", "dir/sub"},
		// A dot inside a name is ordinary.
		{"\\my.file.txt", "my.file.txt"},
		{"\\.hidden", ".hidden"},
		// A name that merely starts with a reserved name is not one.
		{"\\CONSOLE.txt", "CONSOLE.txt"},
		{"\\NULLABLE", "NULLABLE"},
		// Non-ASCII names are ordinary.
		{"\\каталог\\файл.txt", "каталог/файл.txt"},
		{"\\日本語.txt", "日本語.txt"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.raw, func(t *testing.T) {
			got, err := resolvePath(tc.raw)
			if err != nil {
				t.Fatalf("resolvePath(%q) error = %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("resolvePath(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

// TestResolvePathRefusesTraversal asserts every shape of escape is refused rather
// than normalised. Normalising is the trap: rewriting "..\..\x" into "x" turns a
// traversal attempt into a successful access somewhere unintended.
func TestResolvePathRefusesTraversal(t *testing.T) {
	cases := []string{
		"..",
		"\\..",
		"..\\",
		"\\..\\",
		"\\..\\..\\windows\\win.ini",
		"..\\..\\..\\etc\\passwd",
		"dir\\..\\..\\file",
		"\\dir\\..\\..\\..\\secret",
		// A traversal reached through forward slashes.
		"../../etc/passwd",
		"/../secret",
		"dir/../../file",
		// Mixed separators.
		"dir\\../file",
		"dir/..\\file",
		// A "." element, which is harmless but is still not the promised form.
		".",
		"\\.\\file",
		"dir\\.\\file",
		// An empty element, which a doubled separator produces. Refused rather
		// than collapsed, because collapsing is the normalising this avoids.
		"\\\\",
		"dir\\\\file",
		"dir//file",
		// "...." is not a traversal on its own, but it ends in a dot, which
		// Windows strips — so it is refused for that reason.
		"....\\",
	}
	for _, raw := range cases {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			if got, err := resolvePath(raw); err == nil {
				t.Fatalf("resolvePath(%q) = %q, want a refusal", raw, got)
			}
		})
	}
}

// TestResolvePathRefusesAbsoluteAndDeviceForms asserts a path that names something
// outside the share by construction is refused.
func TestResolvePathRefusesAbsoluteAndDeviceForms(t *testing.T) {
	cases := []string{
		// Drive letters.
		"C:\\windows",
		"C:",
		"\\C:\\windows",
		"dir\\C:\\file",
		// The NT device and long-path prefixes.
		"\\\\?\\C:\\windows",
		"\\??\\C:\\windows",
		"\\\\.\\PhysicalDrive0",
		// A UNC path, which names another host entirely.
		"\\\\server\\share\\file",
		// An alternate data stream, which names something other than the file's
		// contents.
		"file.txt:hidden",
		"file.txt:$DATA",
		"dir\\file.txt:stream:$DATA",
	}
	for _, raw := range cases {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			if got, err := resolvePath(raw); err == nil {
				t.Fatalf("resolvePath(%q) = %q, want a refusal", raw, got)
			}
		})
	}
}

// TestResolvePathRefusesReservedDeviceNames asserts a DOS device name is refused
// at any level and with any extension, because Windows resolves it to the device
// rather than to a file however it is dressed up.
func TestResolvePathRefusesReservedDeviceNames(t *testing.T) {
	cases := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM9", "LPT1", "LPT9",
		// Case does not matter.
		"con", "NuL", "Com1",
		// Nor does an extension: "NUL.txt" still resolves to the device.
		"NUL.txt", "CON.log",
		// Nor does depth.
		"dir\\NUL", "dir\\sub\\CON.txt",
	}
	for _, raw := range cases {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			if got, err := resolvePath(raw); err == nil {
				t.Fatalf("resolvePath(%q) = %q, want a refusal", raw, got)
			}
		})
	}
}

// TestResolvePathRefusesControlBytesAndBadEncoding asserts a path that could be
// truncated or misread downstream is refused. A NUL matters most: a consumer
// written in a language where a string ends at the first zero would see a
// different path from the one that was checked.
func TestResolvePathRefusesControlBytesAndBadEncoding(t *testing.T) {
	cases := map[string]string{
		"embedded NUL":             "file\x00.txt",
		"NUL after a traversal":    "..\x00\\file",
		"trailing NUL":             "file.txt\x00",
		"newline":                  "file\n.txt",
		"carriage return":          "file\r.txt",
		"tab":                      "file\t.txt",
		"escape":                   "file\x1b.txt",
		"invalid UTF-8":            "file\xff\xfe.txt",
		"truncated UTF-8 sequence": "file\xc3",
	}
	for name, raw := range cases {
		raw := raw
		t.Run(name, func(t *testing.T) {
			if got, err := resolvePath(raw); err == nil {
				t.Fatalf("resolvePath(%q) = %q, want a refusal", raw, got)
			}
		})
	}
}

// TestResolvePathRefusesTrailingDotsAndSpaces asserts a name Windows would strip
// down to a different name is refused. Without this, "secret.txt." and
// "secret.txt" would name one file while looking like two to anything comparing
// names.
func TestResolvePathRefusesTrailingDotsAndSpaces(t *testing.T) {
	cases := []string{
		"file.txt.",
		"file.txt ",
		"file.txt...",
		"dir.\\file.txt",
		"dir \\file.txt",
		"dir\\file.txt.",
	}
	for _, raw := range cases {
		raw := raw
		t.Run(raw, func(t *testing.T) {
			if got, err := resolvePath(raw); err == nil {
				t.Fatalf("resolvePath(%q) = %q, want a refusal", raw, got)
			}
		})
	}
}

// TestResolvePathEnforcesLengthBounds asserts the bounds hold, so a path cannot
// become work by being enormous.
func TestResolvePathEnforcesLengthBounds(t *testing.T) {
	// A single element at the limit is fine; one byte over is not.
	atLimit := strings.Repeat("a", MaxPathComponentLength)
	if _, err := resolvePath("\\" + atLimit); err != nil {
		t.Fatalf("a %d-byte element was refused: %v", MaxPathComponentLength, err)
	}
	if _, err := resolvePath("\\" + atLimit + "a"); err == nil {
		t.Fatalf("a %d-byte element was accepted", MaxPathComponentLength+1)
	}

	// A whole path over the limit is refused whatever its elements look like.
	long := strings.Repeat("a\\", MaxPathLength)
	if _, err := resolvePath(long); err == nil {
		t.Fatal("a path over the length limit was accepted")
	}
}

// TestResolvePathIsIdempotent asserts a resolved path resolves to itself. A
// resolver that changed its answer on a second pass would mean a backend and the
// server disagreed about what a path named.
func TestResolvePathIsIdempotent(t *testing.T) {
	inputs := []string{"", "\\", "\\file.txt", "\\dir\\sub\\file.txt", "dir/file.txt"}
	for _, raw := range inputs {
		once, err := resolvePath(raw)
		if err != nil {
			t.Fatalf("resolvePath(%q) error = %v", raw, err)
		}
		twice, err := resolvePath(once)
		if err != nil {
			t.Fatalf("resolvePath(%q) on its own output error = %v", once, err)
		}
		if once != twice {
			t.Fatalf("resolvePath is not idempotent: %q then %q", once, twice)
		}
	}
}

// TestResolvePathNeverEscapes is the property the whole function exists for: no
// accepted path, from a corpus built to escape, ever produces a result that
// climbs out of the share.
func TestResolvePathNeverEscapes(t *testing.T) {
	// A corpus assembled from the pieces an escape is built out of.
	pieces := []string{"", "a", ".", "..", "...", "\\", "/", "C:", "?", "*", " ", "NUL", "\x00"}

	checked := 0
	var walk func(prefix string, depth int)
	walk = func(prefix string, depth int) {
		if depth == 0 {
			resolved, err := resolvePath(prefix)
			if err != nil {
				return
			}
			checked++
			// An accepted path must be exactly the promised form.
			if strings.HasPrefix(resolved, "/") {
				t.Fatalf("resolvePath(%q) = %q, which is absolute", prefix, resolved)
			}
			if strings.Contains(resolved, "\\") {
				t.Fatalf("resolvePath(%q) = %q, which still contains a backslash", prefix, resolved)
			}
			if resolved == "" {
				return
			}
			for _, element := range strings.Split(resolved, "/") {
				switch element {
				case "", ".", "..":
					t.Fatalf("resolvePath(%q) = %q, which contains the element %q", prefix, resolved, element)
				}
			}
			return
		}
		for _, piece := range pieces {
			walk(prefix+piece, depth-1)
		}
	}
	walk("", 3)

	if checked == 0 {
		t.Fatal("the corpus produced no accepted paths, so the property was never exercised")
	}
	t.Logf("checked %d accepted paths out of a %d-combination corpus", checked, len(pieces)*len(pieces)*len(pieces))
}

// TestResolvePathPattern asserts a wildcard is lifted out of the final element and
// the rest of the path is still resolved, and that a wildcard cannot smuggle a
// separator past the element checks.
func TestResolvePathPattern(t *testing.T) {
	cases := []struct {
		raw           string
		wantDirectory string
		wantPattern   string
	}{
		{"\\*", "", "*"},
		{"\\*.txt", "", "*.txt"},
		{"\\dir\\*.txt", "dir", "*.txt"},
		{"\\dir\\sub\\a?c.*", "dir/sub", "a?c.*"},
		// No wildcard: the whole thing is a path.
		{"\\dir\\file.txt", "dir/file.txt", ""},
		{"\\dir\\", "dir", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.raw, func(t *testing.T) {
			directory, pattern, err := resolvePathPattern(tc.raw)
			if err != nil {
				t.Fatalf("resolvePathPattern(%q) error = %v", tc.raw, err)
			}
			if directory != tc.wantDirectory || pattern != tc.wantPattern {
				t.Fatalf("resolvePathPattern(%q) = (%q, %q), want (%q, %q)",
					tc.raw, directory, pattern, tc.wantDirectory, tc.wantPattern)
			}
		})
	}

	// The directory part is still subject to every rule above.
	refused := []string{
		"\\..\\*",
		"..\\..\\*.txt",
		"C:\\*",
		"\\dir\\..\\*",
		"\\NUL\\*",
		"\\dir\\*\\file",  // a wildcard that is not the final element
		"\\dir\\*:stream", // a stream on the wildcard
		"\\dir\\*\x00",    // a control byte in the wildcard
	}
	for _, raw := range refused {
		raw := raw
		t.Run("refused "+raw, func(t *testing.T) {
			if directory, pattern, err := resolvePathPattern(raw); err == nil {
				t.Fatalf("resolvePathPattern(%q) = (%q, %q), want a refusal", raw, directory, pattern)
			}
		})
	}
}

// TestSplitAndJoinPath asserts the two helpers are inverses, including at the
// share root where the parent is the empty path.
func TestSplitAndJoinPath(t *testing.T) {
	cases := []struct {
		path   string
		parent string
		name   string
	}{
		{"file.txt", "", "file.txt"},
		{"dir/file.txt", "dir", "file.txt"},
		{"dir/sub/file.txt", "dir/sub", "file.txt"},
	}
	for _, tc := range cases {
		parent, name := splitPath(tc.path)
		if parent != tc.parent || name != tc.name {
			t.Fatalf("splitPath(%q) = (%q, %q), want (%q, %q)", tc.path, parent, name, tc.parent, tc.name)
		}
		if rejoined := joinPath(parent, name); rejoined != tc.path {
			t.Fatalf("joinPath(%q, %q) = %q, want %q", parent, name, rejoined, tc.path)
		}
	}
}

// TestMatchPattern asserts the wildcard matcher, including the forms a client
// sends to mean "everything" and a pattern of many stars, which must not become
// expensive.
func TestMatchPattern(t *testing.T) {
	cases := []struct {
		pattern string
		name    string
		matches bool
	}{
		{"", "anything.txt", true},
		{"*", "anything.txt", true},
		{"*.*", "anything.txt", true},
		{"*.txt", "file.txt", true},
		{"*.txt", "file.log", false},
		{"file.*", "file.txt", true},
		{"a?c.txt", "abc.txt", true},
		{"a?c.txt", "ac.txt", false},
		{"a?c.txt", "abbc.txt", false},
		{"*file*", "my file here", true},
		{"*file*", "nothing", false},
		// Case-insensitive, as the file system commands are.
		{"*.TXT", "file.txt", true},
		{"FILE.*", "file.txt", true},
		// Exact names.
		{"file.txt", "file.txt", true},
		{"file.txt", "file.txt.bak", false},
		// A pattern of many stars must terminate quickly, not blow up.
		{strings.Repeat("*", 40) + "z", strings.Repeat("a", 200), false},
		{strings.Repeat("*a", 20), strings.Repeat("a", 200), true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.pattern+" vs "+tc.name, func(t *testing.T) {
			if got := matchPattern(tc.pattern, tc.name); got != tc.matches {
				t.Fatalf("matchPattern(%q, %q) = %t, want %t", tc.pattern, tc.name, got, tc.matches)
			}
		})
	}
}
