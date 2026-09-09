package server

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/fileflags"
)

// TestAttributesForSetsArchiveOnFiles guards the change: every regular file now
// carries FILE_ATTRIBUTE_ARCHIVE, as Windows reports, and no entry is described as
// FILE_ATTRIBUTE_NORMAL — which is valid only when it stands alone ([MS-FSCC] 2.6)
// and now has no entry left to describe.
func TestAttributesForSetsArchiveOnFiles(t *testing.T) {
	tests := []struct {
		name string
		attr FileAttr
		want uint32
	}{
		{
			name: "regular file",
			attr: FileAttr{Name: "plain.txt"},
			want: fileflags.FILE_ATTRIBUTE_ARCHIVE,
		},
		{
			name: "read-only file",
			attr: FileAttr{Name: "ro.txt", ReadOnly: true},
			want: fileflags.FILE_ATTRIBUTE_ARCHIVE | fileflags.FILE_ATTRIBUTE_READONLY,
		},
		{
			name: "directory carries no archive bit",
			attr: FileAttr{Name: "subdir", IsDir: true},
			want: fileflags.FILE_ATTRIBUTE_DIRECTORY,
		},
		{
			name: "read-only directory",
			attr: FileAttr{Name: "rodir", IsDir: true, ReadOnly: true},
			want: fileflags.FILE_ATTRIBUTE_DIRECTORY | fileflags.FILE_ATTRIBUTE_READONLY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := attributesFor(tt.attr)
			if got != tt.want {
				t.Errorf("attributesFor = %#08x, want %#08x", got, tt.want)
			}
			if got&fileflags.FILE_ATTRIBUTE_NORMAL != 0 {
				t.Errorf("attributesFor = %#08x, which sets FILE_ATTRIBUTE_NORMAL alongside other bits", got)
			}
		})
	}
}

// TestLegacyAttributesAgreeWithExtended checks the 16-bit and 32-bit views of the
// same entry describe it the same way. They are computed by separate functions, so
// they can drift apart; the archive bit is the case where they just did.
func TestLegacyAttributesAgreeWithExtended(t *testing.T) {
	for _, attr := range []FileAttr{
		{Name: "plain.txt"},
		{Name: "ro.txt", ReadOnly: true},
		{Name: "subdir", IsDir: true},
		{Name: "rodir", IsDir: true, ReadOnly: true},
	} {
		extended := attributesFor(attr)
		legacy := legacyAttributesFor(attr)

		if (extended&fileflags.FILE_ATTRIBUTE_ARCHIVE != 0) != (legacy&smbFileAttributeArchive != 0) {
			t.Errorf("%q: archive bit differs (extended %#08x, legacy %#04x)", attr.Name, extended, legacy)
		}
		if (extended&fileflags.FILE_ATTRIBUTE_DIRECTORY != 0) != (legacy&smbFileAttributeDirectory != 0) {
			t.Errorf("%q: directory bit differs (extended %#08x, legacy %#04x)", attr.Name, extended, legacy)
		}
		if (extended&fileflags.FILE_ATTRIBUTE_READONLY != 0) != (legacy&smbFileAttributeReadOnly != 0) {
			t.Errorf("%q: read-only bit differs (extended %#08x, legacy %#04x)", attr.Name, extended, legacy)
		}

		// The legacy field reserves 0x0080; the extended NORMAL value must never
		// leak into it.
		if legacy&0x0080 != 0 {
			t.Errorf("%q: legacy attributes %#04x set the reserved 0x0080 bit", attr.Name, legacy)
		}
	}
}
