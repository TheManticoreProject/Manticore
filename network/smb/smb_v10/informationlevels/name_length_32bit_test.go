package informationlevels

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
)

// TestNameLengthOverflowRejected sets the 32-bit name length of every information
// level that carries one to values that overflow offset+length in a 32-bit int,
// and requires an error. Before the length was bounded as an unsigned value, a
// 32-bit build wrapped the sum negative, passed the bounds check, and panicked on
// the slice; the test is therefore only meaningful under GOARCH=386.
func TestNameLengthOverflowRejected(t *testing.T) {
	type codec interface {
		Marshal() ([]byte, error)
		Unmarshal([]byte) (int, error)
	}
	cases := []codec{
		&SMB_QUERY_FS_ATTRIBUTE_INFO{},
		&SMB_QUERY_FS_VOLUME_INFO{},
		&SMB_FIND_FILE_NAMES_INFO{},
		&SMB_FIND_FILE_BOTH_DIRECTORY_INFO{},
		&SMB_FIND_FILE_FULL_DIRECTORY_INFO{},
		&SMB_FIND_FILE_DIRECTORY_INFO{},
		&SMB_QUERY_FILE_NAME_INFO{},
		&SMB_QUERY_FILE_ALT_NAME_INFO{},
		&SMB_QUERY_FILE_ALL_INFO{},
		&SMB_QUERY_FILE_STREAM_INFO{},
	}
	const nameLength = 0x37

	for _, c := range cases {
		typ := reflect.TypeOf(c).Elem()
		value := reflect.ValueOf(c).Elem()
		var name reflect.Value
		for _, field := range []string{"Filename", "Filesystemname", "Volumelabel", "Streamname"} {
			if name = value.FieldByName(field); name.IsValid() {
				break
			}
		}
		if !name.IsValid() {
			t.Fatalf("%s: no name field", typ.Name())
		}
		name.Set(reflect.ValueOf(bytes.Repeat([]types.UCHAR{'A'}, nameLength)))
		raw, err := c.Marshal()
		if err != nil {
			t.Fatalf("%s: Marshal: %v", typ.Name(), err)
		}
		// Every other field is zero, so the first occurrence is the length.
		at := bytes.Index(raw, []byte{nameLength, 0, 0, 0})
		if at < 0 {
			t.Fatalf("%s: name length not found in % x", typ.Name(), raw)
		}

		for _, length := range []uint32{0x7FFFFFF8, 0x80000000, 0xFFFFFFF0} {
			binary.LittleEndian.PutUint32(raw[at:], length)
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s: length 0x%08X: Unmarshal panicked: %v", typ.Name(), length, r)
					}
				}()
				_, err = reflect.New(typ).Interface().(codec).Unmarshal(raw)
			}()
			if err == nil {
				t.Errorf("%s: length 0x%08X: Unmarshal accepted a name past the end of the data", typ.Name(), length)
			}
		}
	}
}
