package uuid_v3_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v1"
	"github.com/TheManticoreProject/Manticore/crypto/uuid/uuid_v3"
)

func TestUUIDv3(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		domain         string
		wantUUIDstring string
		wantErr        bool
	}{
		{
			name:           "Basic UUID v3",
			namespace:      "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			domain:         "unicorn-utterances.com",
			wantUUIDstring: "8d9aeee5-d9ad-3934-84f4-ac533183424d",
			wantErr:        false,
		},
		{
			name:           "Basic UUID v3",
			namespace:      "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			domain:         "podaliri.us",
			wantUUIDstring: "c0819443-a39c-3e47-a949-303520cf9661",
			wantErr:        false,
		},
		{
			name:           "Basic UUID v3",
			namespace:      "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			domain:         "manticore.local",
			wantUUIDstring: "17aae0f3-3230-34cf-ad4c-ca7b64fecff6",
			wantErr:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var u uuid_v3.UUIDv3

			ui := uuid_v1.UUIDv1{}
			ui.FromString(test.namespace)
			u.Namespace = &ui
			u.Name = test.domain

			_, err := u.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
			}

			if u.String() != test.wantUUIDstring {
				t.Errorf("UUIDv3.String() \n\tgot  %v\n\twant %v", u.String(), test.wantUUIDstring)
			}
		})
	}
}
