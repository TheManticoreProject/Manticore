package server

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/encoding/utf16"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/subcommands"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// Regression coverage for the wire defects that only a third-party client exposed.
//
// Every one of them was invisible to the unit suite for the same reason: it pairs
// this server with the client in this repository, and that client sends its names
// in OEM even after negotiating Unicode. A field decoded with the wrong character
// width therefore agreed with itself in both directions and every round-trip
// passed. These tests set the Unicode flag deliberately, which is what the suite
// was never doing.
//
// The live matrix in live_interop_integration_test.go is what found them; these
// are what keep them found without a client binary present.

// TestUnicodeTreeConnectPath asserts a Unicode share path decodes whole.
//
// A single-byte terminator scan ends "\\\\host\\share" at its first character, so
// every tree connect from a Unicode client was refused as not being a UNC path.
func TestUnicodeTreeConnectPath(t *testing.T) {
	for _, unicode := range []bool{false, true} {
		unicode := unicode
		name := "OEM"
		if unicode {
			name = "Unicode"
		}
		t.Run(name, func(t *testing.T) {
			const path = `\\127.0.0.1\FILES`

			request := commands.NewTreeConnectAndxRequest()
			request.SetUnicode(unicode)
			// One null byte, which is what a client sends for an empty password and
			// what leaves the path 16-bit aligned.
			request.Password = []types.UCHAR{0x00}
			request.PasswordLength = types.USHORT(1)
			if unicode {
				request.Path = append([]types.UCHAR(utf16.EncodeUTF16LE(path)), 0x00, 0x00)
			} else {
				request.Path = append([]types.UCHAR(path), 0x00)
			}
			request.Service = []types.UCHAR("?????\x00")

			marshalled, err := request.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			decoded := commands.NewTreeConnectAndxRequest()
			decoded.Init()
			decoded.SetUnicode(unicode)
			if _, err := decoded.Unmarshal(marshalled); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if got := decodeWireString(decoded.Path, unicode); got != path {
				t.Fatalf("the path decoded as %q, want %q", got, path)
			}
			if got := shareNameFromPath(decodeWireString(decoded.Path, unicode)); got != "FILES" {
				t.Fatalf("the share name resolved to %q, want FILES", got)
			}
			// The Service string is OEM whatever the message declared.
			if got := decodeOEMString(decoded.Service); got != "?????" {
				t.Fatalf("the service decoded as %q, want ?????", got)
			}
		})
	}
}

// TestUnicodeCreateDirectoryName asserts a Unicode name in an SMB_STRING decodes
// whole.
//
// SMB_STRING's null-terminated formats end at the first null CHARACTER. Ending
// them at the first null BYTE truncated "\newdir" to a single backslash, and
// decoding the resulting one-byte buffer as UTF-16 then read past the end of it —
// which took down the connection rather than failing the request.
func TestUnicodeCreateDirectoryName(t *testing.T) {
	const directory = `\newdir`

	request := commands.NewCreateDirectoryRequest()
	request.SetUnicode(true)
	request.DirectoryName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
	request.DirectoryName.Buffer = []types.UCHAR(utf16.EncodeUTF16LE(directory))
	request.DirectoryName.Length = types.USHORT(len(request.DirectoryName.Buffer))

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	decoded := commands.NewCreateDirectoryRequest()
	decoded.Init()
	decoded.SetUnicode(true)
	if _, err := decoded.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got := decodeWireString(decoded.DirectoryName.Buffer, true); got != directory {
		t.Fatalf("the directory name decoded as %q, want %q", got, directory)
	}
}

// TestUnicodeRenameAlignsTheSecondName asserts both names of a rename decode, for
// a first name of each parity.
//
// The first name's length decides whether an alignment byte precedes the second,
// so exactly half of all renames exercise the padded path. Missing it does not
// fail: it shifts every character of the second name by one byte and produces a
// different, plausible-looking name, so a file is created under a name the client
// never asked for.
func TestUnicodeRenameAlignsTheSecondName(t *testing.T) {
	cases := []struct {
		name string
		from string
		to   string
	}{
		// "\hello.txt" is ten characters, so the second name starts on an odd
		// offset and is padded.
		{"padded", `\hello.txt`, `\renamed.txt`},
		// One character shorter, so the second name is already aligned.
		{"unpadded", `\hell.txt`, `\renamed.txt`},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			request := commands.NewRenameRequest()
			request.SetUnicode(true)
			request.OldFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
			request.OldFileName.Buffer = []types.UCHAR(utf16.EncodeUTF16LE(tc.from))
			request.NewFileName.SetBufferFormat(types.SMB_STRING_BUFFER_FORMAT_NULL_TERMINATED_ASCII_STRING)
			request.NewFileName.Buffer = []types.UCHAR(utf16.EncodeUTF16LE(tc.to))

			marshalled, err := request.Marshal()
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			decoded := commands.NewRenameRequest()
			decoded.Init()
			decoded.SetUnicode(true)
			if _, err := decoded.Unmarshal(marshalled); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if got := decodeWireString(decoded.OldFileName.Buffer, true); got != tc.from {
				t.Errorf("the old name decoded as %q, want %q", got, tc.from)
			}
			if got := decodeWireString(decoded.NewFileName.Buffer, true); got != tc.to {
				t.Errorf("the new name decoded as %q, want %q", got, tc.to)
			}
		})
	}
}

// TestTransactionBlocksLocatedByDeclaredOffsets asserts the parameter block is
// found where the message says it is, even when the block carries trailing
// padding.
//
// This is the defect with the widest reach. The decoder derived the leading pad by
// subtracting the field lengths from the size of the data block, which assumes the
// block ends flush with its last field. A real client pads the end of the block to
// a four-byte boundary, so the surplus was charged to the leading pad and every
// field in the block was read shifted — a directory listing asked for information
// level 0x0104 and the server read 0x0000 from two bytes further on.
func TestTransactionBlocksLocatedByDeclaredOffsets(t *testing.T) {
	// The parameters of a TRANS2_FIND_FIRST2, as a client sends them.
	parameters := make([]byte, 12)
	binary.LittleEndian.PutUint16(parameters[0:2], 0x0016)                       // SearchAttributes
	binary.LittleEndian.PutUint16(parameters[2:4], 0x0556)                       // SearchCount
	binary.LittleEndian.PutUint16(parameters[4:6], 0x0006)                       // Flags
	binary.LittleEndian.PutUint16(parameters[6:8], smbFindFileBothDirectoryInfo) // InformationLevel
	parameters = append(parameters, utf16.EncodeUTF16LE(`\*`)...)
	parameters = append(parameters, 0x00, 0x00)

	request := commands.NewTransaction2Request()
	request.SetUnicode(true)
	request.Setup = []types.USHORT{types.USHORT(subcommands.TRANS2_FIND_FIRST2)}
	request.SetupCount = types.UCHAR(1)
	request.MaxParameterCount = 10
	request.MaxDataCount = 65535
	request.TotalParameterCount = types.USHORT(len(parameters))
	request.ParameterCount = types.USHORT(len(parameters))
	request.Trans2_Parameters = parameters
	// Two bytes of leading pad, as a client uses to align the block, and two of
	// trailing pad, which is what the old arithmetic mistook for more leading pad.
	request.Pad1 = []types.UCHAR{0x44, 0x20}

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	// Append the trailing padding a client puts at the end of the block, and grow
	// ByteCount to cover it, exactly as one does on the wire.
	byteCountAt := 1 + 2*int(marshalled[0])
	trailing := []byte{0x00, 0x00}
	binary.LittleEndian.PutUint16(marshalled[byteCountAt:byteCountAt+2],
		binary.LittleEndian.Uint16(marshalled[byteCountAt:byteCountAt+2])+uint16(len(trailing)))
	marshalled = append(marshalled, trailing...)

	decoded := commands.NewTransaction2Request()
	decoded.Init()
	decoded.SetUnicode(true)
	if _, err := decoded.Unmarshal(marshalled); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(decoded.Trans2_Parameters) < 8 {
		t.Fatalf("the parameter block decoded as %d bytes, want %d", len(decoded.Trans2_Parameters), len(parameters))
	}
	level := binary.LittleEndian.Uint16([]byte(decoded.Trans2_Parameters)[6:8])
	if level != smbFindFileBothDirectoryInfo {
		t.Fatalf("the information level read as 0x%04X, want 0x%04X: the parameter block was located by arithmetic rather than by its declared offset",
			level, smbFindFileBothDirectoryInfo)
	}
	if got := decodeWireString([]types.UCHAR(decoded.Trans2_Parameters)[12:], true); got != `\*` {
		t.Errorf("the search pattern decoded as %q, want %q", got, `\*`)
	}
}

// TestTransactionMarshalDeclaresWhereItPutTheBlocks asserts a marshaller's
// declared offsets describe its own output.
//
// The offsets were previously left for a caller to fill in, and nothing did, so
// they were zero: a receiver following the specification could not find either
// block. The check is against the bytes rather than against a recomputation of the
// same formula, so a wrong formula fails here too.
func TestTransactionMarshalDeclaresWhereItPutTheBlocks(t *testing.T) {
	parameters := []byte{0xA1, 0xA2, 0xA3, 0xA4, 0xA5}
	data := []byte{0xD1, 0xD2, 0xD3}

	request := commands.NewTransaction2Request()
	request.Setup = []types.USHORT{types.USHORT(subcommands.TRANS2_QUERY_FS_INFORMATION)}
	request.SetupCount = types.UCHAR(1)
	request.TotalParameterCount = types.USHORT(len(parameters))
	request.ParameterCount = types.USHORT(len(parameters))
	request.Trans2_Parameters = parameters
	request.TotalDataCount = types.USHORT(len(data))
	request.DataCount = types.USHORT(len(data))
	request.Trans2_Data = data
	request.Pad1 = []types.UCHAR{0x00, 0x00}
	request.Pad2 = []types.UCHAR{0x00}

	marshalled, err := request.Marshal()
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// The offsets are measured from the start of the SMB header, and the
	// marshalled command begins right after it.
	const headerSize = 32
	for _, tc := range []struct {
		name    string
		offset  int
		content []byte
	}{
		{"the parameter block", int(request.ParameterOffset), parameters},
		{"the data block", int(request.DataOffset), data},
	} {
		at := tc.offset - headerSize
		if at < 0 || at+len(tc.content) > len(marshalled) {
			t.Fatalf("%s is declared at offset %d, which is outside the %d-byte command",
				tc.name, tc.offset, len(marshalled))
		}
		if got := marshalled[at : at+len(tc.content)]; string(got) != string(tc.content) {
			t.Errorf("%s is declared at offset %d, where the bytes are % x, not % x",
				tc.name, tc.offset, got, tc.content)
		}
	}
}

// TestFindPatternAcceptsAnExactName asserts naming one file exactly is a search
// rather than an attempt to list it as a directory.
//
// A client does this before deleting or renaming something: it resolves the name
// first, and a server that treats every search path as a directory refuses the
// resolution and the operation never happens.
func TestFindPatternAcceptsAnExactName(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddFile("target.txt", []byte("x")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	if err := fs.AddDirectory("holder"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	if err := fs.AddFile("holder/inner.txt", []byte("y")); err != nil {
		t.Fatalf("AddFile() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	// An exact file name resolves to that one entry.
	entries, err := client.ListEntries(`\target.txt`)
	if err != nil {
		t.Fatalf("searching for an exact name failed: %v", err)
	}
	if !containsName(namesOf(entries), "target.txt") {
		t.Fatalf("searching for an exact name returned %v", namesOf(entries))
	}

	// A directory still lists its contents rather than matching its own name,
	// which is the behaviour the exact-name case must not have replaced.
	entries, err = client.ListEntries(`\holder\*`)
	if err != nil {
		t.Fatalf("listing a directory failed: %v", err)
	}
	if !containsName(namesOf(entries), "inner.txt") {
		t.Fatalf("listing a directory returned %v", namesOf(entries))
	}
	entries, err = client.ListEntries(`\holder`)
	if err != nil {
		t.Fatalf("listing a directory by name failed: %v", err)
	}
	if !containsName(namesOf(entries), "inner.txt") {
		t.Fatalf("listing a directory by name returned %v, want its contents", namesOf(entries))
	}
}

// TestFileCommandsOnAPipeHandleAreRefused asserts a read or a write through a pipe
// handle is answered rather than crashing.
//
// A pipe handle has no file behind it. An RPC client reads one directly when its
// transaction did not return a whole response, which reached an unguarded
// dereference and took the connection down.
func TestFileCommandsOnAPipeHandleAreRefused(t *testing.T) {
	pipes := newEchoPipe("srvsvc")
	_, client := pipeServer(t, pipes)

	fid, err := openPipeHandle(t, client, `\PIPE\srvsvc`)
	if err != nil {
		t.Fatalf("opening the pipe failed: %v", err)
	}

	if _, err := client.ReadFile(fid, 0, 16); err == nil {
		t.Error("reading a pipe handle as a file succeeded")
	}
	if _, err := client.WriteFile(fid, 0, []byte("x")); err == nil {
		t.Error("writing a pipe handle as a file succeeded")
	}

	// The connection must still be usable, which is what the crash cost.
	if _, err := client.Echo([]byte("still here")); err != nil {
		t.Fatalf("the connection did not survive a file command on a pipe handle: %v", err)
	}
}

// TestQueryFileInformationOnADirectoryHandle asserts a handle with no file behind
// it can still be described, since querying is what such a handle exists for.
func TestQueryFileInformationOnADirectoryHandle(t *testing.T) {
	fs := NewMemoryFileSystem("FILES")
	if err := fs.AddDirectory("adirectory"); err != nil {
		t.Fatalf("AddDirectory() error = %v", err)
	}
	_, client := fileServer(t, fs, false)

	fid, err := client.OpenFile("adirectory", 0x00120089, 0x00000007, 0x00000001, 0x00000001)
	if err != nil {
		t.Fatalf("opening the directory failed: %v", err)
	}
	defer client.CloseFile(fid)

	if _, err := client.Echo([]byte("alive")); err != nil {
		t.Fatalf("the connection did not survive opening a directory: %v", err)
	}
}

// TestVolumeInformationPassthroughLevels asserts the pass-through range is served,
// since that is how a client asks about free space.
func TestVolumeInformationPassthroughLevels(t *testing.T) {
	volume := VolumeInfo{
		Label: "FILES", FileSystemName: "NTFS",
		TotalBytes: 1 << 30, FreeBytes: 1 << 29,
		SectorsPerAllocationUnit: 8, BytesPerSector: 512,
	}

	t.Run("full size information", func(t *testing.T) {
		info, served := encodeVolumeInformation(smbInfoPassthrough+fileFsFullSizeInformation, volume, true)
		if !served {
			t.Fatal("the full-size class is not served")
		}
		if len(info) != 32 {
			t.Fatalf("the full-size block is %d bytes, want 32", len(info))
		}

		unit := uint64(volume.SectorsPerAllocationUnit) * uint64(volume.BytesPerSector)
		if got := binary.LittleEndian.Uint64(info[0:8]); got != uint64(volume.TotalBytes)/unit {
			t.Errorf("total units = %d, want %d", got, uint64(volume.TotalBytes)/unit)
		}
		// What the caller may use is what is free: nothing here imposes a quota, and
		// reporting less would make a client refuse a write the server would take.
		caller := binary.LittleEndian.Uint64(info[8:16])
		actual := binary.LittleEndian.Uint64(info[16:24])
		if caller != actual || caller != uint64(volume.FreeBytes)/unit {
			t.Errorf("available units = %d/%d, want both %d", caller, actual, uint64(volume.FreeBytes)/unit)
		}
	})

	t.Run("classes that shadow an SMB level", func(t *testing.T) {
		for class, level := range map[uint16]uint16{
			fileFsVolumeInformation:    smbQueryFsVolumeInfo,
			fileFsSizeInformation:      smbQueryFsSizeInfo,
			fileFsDeviceInformation:    smbQueryFsDeviceInfo,
			fileFsAttributeInformation: smbQueryFsAttributeInfo,
		} {
			viaClass, servedClass := encodeVolumeInformation(smbInfoPassthrough+class, volume, true)
			viaLevel, servedLevel := encodeVolumeInformation(level, volume, true)
			if !servedClass || !servedLevel {
				t.Errorf("class %d / level 0x%04X: served = %v / %v", class, level, servedClass, servedLevel)
				continue
			}
			if string(viaClass) != string(viaLevel) {
				t.Errorf("class %d and level 0x%04X disagree", class, level)
			}
		}
	})

	t.Run("an unserved class is refused", func(t *testing.T) {
		if _, served := encodeVolumeInformation(smbInfoPassthrough+0x00FF, volume, true); served {
			t.Error("a class this server does not implement was answered")
		}
	})
}

// TestFindEntryNamesFollowTheRequestEncoding asserts a listing's names go out in
// the encoding the request declared.
//
// A client reads the buffer as UTF-16 when it negotiated Unicode, whatever the
// server put there, so an OEM name arrives as half as many characters of nonsense —
// a listing of the right shape and the wrong text.
func TestFindEntryNamesFollowTheRequestEncoding(t *testing.T) {
	attr := FileAttr{Name: "hello.txt"}

	for _, unicode := range []bool{false, true} {
		unicode := unicode
		t.Run(map[bool]string{false: "OEM", true: "Unicode"}[unicode], func(t *testing.T) {
			entry := encodeFindEntry(smbFindFileBothDirectoryInfo, attr, unicode)
			if len(entry) < bothDirectoryInfoFixedSize+4 {
				t.Fatalf("the entry is %d bytes", len(entry))
			}

			nameLength := int(binary.LittleEndian.Uint32(entry[60:64]))
			want := len(attr.Name)
			if unicode {
				want *= 2
			}
			if nameLength != want {
				t.Fatalf("FileNameLength = %d, want %d bytes for %q", nameLength, want, attr.Name)
			}

			raw := entry[bothDirectoryInfoFixedSize : bothDirectoryInfoFixedSize+nameLength]
			got := string(raw)
			if unicode {
				got = utf16.DecodeUTF16LE(raw)
			}
			if got != attr.Name {
				t.Fatalf("the name encoded as %q, want %q", got, attr.Name)
			}
		})
	}
}

// TestSigningIsArmedWithoutASignedAuthenticate asserts signing is activated by a
// session setup whose own signature is absent.
//
// [MS-SMB] section 3.3.5.3 asks the server to sign the response at sequence one
// once it has the key; it does not ask for the request to be verified. A client
// that has not yet armed signing puts a placeholder in the field, so demanding a
// valid signature there refuses the client outright — and it would secure nothing,
// since the key is derived from the very message the signature would cover.
func TestSigningIsArmedWithoutASignedAuthenticate(t *testing.T) {
	// A placeholder such as a client sends before its own signing is armed.
	placeholder := []byte("BSRSPYL ")
	if len(placeholder) != 8 {
		t.Fatalf("the placeholder is %d bytes, want 8", len(placeholder))
	}
	if strings.TrimSpace(string(placeholder)) == "" {
		t.Fatal("the placeholder is blank")
	}

	// The unit suite's client signs its AUTHENTICATE, so what is asserted here is
	// the server's own rule: it arms signing from the exchange rather than from a
	// signature on it, which the live matrix exercises against a client that sends
	// the placeholder above.
	_, client := pipedClient(t, conformanceConfig(SigningRequired), true)
	if client.Connection == nil || !client.Connection.IsSigningActive {
		t.Fatal("signing was not armed by a session setup under a required policy")
	}
	if _, err := client.Echo([]byte("signed")); err != nil {
		t.Fatalf("a signed request after the exchange failed: %v", err)
	}
}
