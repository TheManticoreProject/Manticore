package uuid_v1_test

import (
	"testing"

	uuid_v1 "github.com/TheManticoreProject/Manticore/crypto/uuid/v1"
)

func TestUUIDv1Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"ValidData", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, false},
		{"InvalidLength", []byte{0x01, 0x02}, true},
		{"InvalidVersion", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x20, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
			_, err := u.Unmarshal(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDv1FromString(t *testing.T) {
	tests := []struct {
		name    string
		uuidStr string
		wantErr bool
	}{
		{"ValidString", "01020304-0506-0708-1020-304050607080", false},
		{"InvalidLength", "01020304-0506-0708-1020-30405060708", true},
		{"InvalidFormat", "01020304-0506-0708-1020-30405060708G", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
			err := u.FromString(tt.uuidStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUUIDv1FromBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{"ValidData", []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80}, false},
		{"InvalidLength", []byte{0x01, 0x02}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u uuid_v1.UUIDv1
			err := u.FromBytes(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
