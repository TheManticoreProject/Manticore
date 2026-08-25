package server

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Path limits. SMB carries a path in a 16-bit-counted field, but a server has no
// reason to accept anything near that: these bounds match what a file system will
// take and keep a malicious path from turning into work.
const (
	// MaxPathLength is the longest share-relative path accepted, in bytes.
	MaxPathLength = 4096

	// MaxPathComponentLength is the longest single element of a path, in bytes.
	// 255 is what every file system this could sit on enforces.
	MaxPathComponentLength = 255
)

// invalidNameCharacters are the characters a file name cannot carry. The
// separators and the colon are checked separately, before elements are split.
const invalidNameCharacters = `*?"<>|`

// reservedDeviceNames are the DOS device names Windows resolves specially at
// every level of a path. A file called "CON" is not a file, so a path naming one
// is refused rather than passed to a backend that might open a device.
var reservedDeviceNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// resolvePath converts a client-supplied SMB path into the share-relative form a
// FileSystem is promised: forward-slash separated, with no leading slash, no
// empty element, and no "." or ".." element.
//
// This is the containment boundary. Every path that reaches a backend passes
// through here, and a backend is entitled to assume the result cannot escape the
// share, so everything that could make it escape is refused here rather than
// being normalised into something plausible. Normalising is the trap: a resolver
// that quietly rewrites "..\..\x" into "x" turns a traversal attempt into a
// successful access somewhere unintended, so a path containing ".." is rejected
// outright instead.
//
// Refused: an absolute path, a drive letter, a UNC or NT device prefix, an
// alternate data stream, a NUL or other control byte, a "." or ".." element, an
// empty element (which "\\" produces), a reserved DOS device name at any level,
// invalid UTF-8, and anything over the length bounds.
//
// Parameters:
//   - raw: the path exactly as the client sent it
//
// Returns:
//   - The share-relative path, "" for the share root
//   - An error naming why the path was refused
func resolvePath(raw string) (string, error) {
	if len(raw) > MaxPathLength {
		return "", fmt.Errorf("path is %d bytes, over the %d-byte limit", len(raw), MaxPathLength)
	}
	if !utf8.ValidString(raw) {
		return "", fmt.Errorf("path is not valid UTF-8")
	}

	// A control byte, NUL above all, can truncate a path in a consumer written
	// in a language where a string ends at the first zero.
	for i := 0; i < len(raw); i++ {
		if raw[i] < 0x20 {
			return "", fmt.Errorf("path contains the control byte 0x%02X at offset %d", raw[i], i)
		}
	}

	// A stream name selects something other than the file's contents, and nothing
	// here implements streams. It is checked before separators are normalised,
	// because a colon anywhere is enough to refuse.
	if strings.Contains(raw, ":") {
		return "", fmt.Errorf("path contains a colon, which names a drive or an alternate data stream")
	}

	// SMB uses backslashes; accept forward slashes too, since a client that sends
	// them means the same thing, and normalise before the element checks so that
	// neither separator can smuggle an element past them.
	path := strings.ReplaceAll(raw, "\\", "/")

	// A doubled separator is refused here rather than after the trims below,
	// because the trims would collapse it: "\\" reduces to the share root, which
	// is exactly the normalising this function must not do.
	if strings.Contains(path, "//") {
		return "", fmt.Errorf("path contains a doubled separator")
	}

	// A leading separator is how a client names a path from the share root, and
	// is the only one that is stripped rather than refused.
	path = strings.TrimPrefix(path, "/")

	// The share root itself.
	if path == "" {
		return "", nil
	}

	// A trailing separator is redundant but harmless, and clients send it.
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		return "", nil
	}

	elements := strings.Split(path, "/")
	for _, element := range elements {
		switch element {
		case "":
			// Produced by a doubled separator. Refused rather than collapsed,
			// because collapsing is the normalising this function does not do.
			return "", fmt.Errorf("path contains an empty element")
		case ".":
			return "", fmt.Errorf("path contains a %q element", ".")
		case "..":
			return "", fmt.Errorf("path contains a %q element", "..")
		}
		if len(element) > MaxPathComponentLength {
			return "", fmt.Errorf("path element %q is %d bytes, over the %d-byte limit",
				element, len(element), MaxPathComponentLength)
		}
		// A trailing dot or space is stripped by Windows when it opens a file, so
		// "secret.txt." and "secret.txt" would name the same thing while looking
		// different to anything checking names. Refuse rather than reproduce that.
		if strings.HasSuffix(element, ".") || strings.HasSuffix(element, " ") {
			return "", fmt.Errorf("path element %q ends with a dot or a space", element)
		}
		if reservedDeviceNames[deviceNameOf(element)] {
			return "", fmt.Errorf("path element %q is a reserved device name", element)
		}
		// A wildcard or one of the other characters a name cannot carry. A
		// wildcard matters beyond validity: reaching a backend as a literal name
		// it might expand there instead, so a path is never allowed to hold one.
		// The commands that legitimately take a pattern go through
		// resolvePathPattern, which lifts it out first.
		if i := strings.IndexAny(element, invalidNameCharacters); i >= 0 {
			return "", fmt.Errorf("path element %q contains %q, which a name cannot carry", element, element[i])
		}
	}

	return strings.Join(elements, "/"), nil
}

// deviceNameOf returns the part of an element Windows compares against the
// reserved device names: the name before its first dot, upper-cased. "NUL.txt"
// resolves to the NUL device, so the extension does not make it safe.
func deviceNameOf(element string) string {
	if dot := strings.IndexByte(element, '.'); dot >= 0 {
		element = element[:dot]
	}
	return strings.ToUpper(element)
}

// splitPath separates a resolved path into its parent and its final element. The
// parent of a top-level entry is "", the share root.
func splitPath(path string) (parent, name string) {
	if slash := strings.LastIndexByte(path, '/'); slash >= 0 {
		return path[:slash], path[slash+1:]
	}
	return "", path
}

// joinPath appends an element to a resolved path.
func joinPath(parent, name string) string {
	if parent == "" {
		return name
	}
	return parent + "/" + name
}

// resolvePathPattern resolves a path whose final element may be a wildcard,
// returning the resolved directory and the pattern to match within it.
//
// The commands that take a pattern — delete above all — carry it in the same
// field as a path, so the wildcard has to be lifted out before the rest of the
// path is checked. Only the final element may contain one: a wildcard in the
// middle would name several directories, which no command here means.
//
// Parameters:
//   - raw: the path exactly as the client sent it
//
// Returns:
//   - The resolved directory the pattern applies within
//   - The pattern, which is "" when the path named no wildcard
//   - An error naming why the path was refused
func resolvePathPattern(raw string) (directory, pattern string, err error) {
	normalised := strings.ReplaceAll(raw, "\\", "/")
	lastSlash := strings.LastIndexByte(normalised, '/')

	final := normalised[lastSlash+1:]
	if !strings.ContainsAny(final, "*?") {
		resolved, err := resolvePath(raw)
		if err != nil {
			return "", "", err
		}
		return resolved, "", nil
	}

	// The wildcard element is validated separately: it cannot be a path element,
	// so the element rules do not apply to it, but it must still be free of the
	// bytes and separators that would make it more than one element.
	if strings.ContainsAny(final, "/\\:") {
		return "", "", fmt.Errorf("wildcard element %q contains a separator", final)
	}
	if len(final) > MaxPathComponentLength {
		return "", "", fmt.Errorf("wildcard element %q is over the %d-byte limit", final, MaxPathComponentLength)
	}
	for i := 0; i < len(final); i++ {
		if final[i] < 0x20 {
			return "", "", fmt.Errorf("wildcard element contains the control byte 0x%02X", final[i])
		}
	}

	// Whatever precedes the wildcard is an ordinary path and is resolved as one.
	parent := ""
	if lastSlash >= 0 {
		parent = normalised[:lastSlash]
	}
	directory, err = resolvePath(parent)
	if err != nil {
		return "", "", err
	}
	return directory, final, nil
}
