// Package functions holds the SASec method stubs (one per opnum) for the MS-TSCH
// SASec interface. Each stub marshals its request and unmarshals the response through
// an ndr.Invoker.
package functions

import "unicode/utf16"

// decodeWideBuffer converts a fixed-size UTF-16 output buffer (as returned by the
// SAGet* methods in their [in,out] size_is(ccBufferSize) wchar_t wszBuffer[] parameter)
// into a Go string, stopping at the first NUL terminator ([MS-TSCH] 3.2.5.3.6/3.2.5.3.7).
func decodeWideBuffer(buf []uint16) string {
	for i, c := range buf {
		if c == 0 {
			buf = buf[:i]
			break
		}
	}
	return string(utf16.Decode(buf))
}
