package keycredential_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ldap"
	"github.com/TheManticoreProject/Manticore/windows/guid"
	"github.com/TheManticoreProject/Manticore/windows/keycredential"
	"github.com/TheManticoreProject/Manticore/windows/keycredential/crypto"
	"github.com/TheManticoreProject/Manticore/windows/keycredential/utils"
	"github.com/TheManticoreProject/Manticore/windows/keycredential/version"
)

func TestKeyCredential_Unmarshal(t *testing.T) {
	tests := []struct {
		name                       string
		msDsKeyCredentialLinkValue string
		expectedKC                 *keycredential.KeyCredential
		wantErr                    bool
	}{
		{
			name:                       "Valid KeyCredential with specific identifier",
			msDsKeyCredentialLinkValue: "B:10:9012345678:CN=POC,CN=Computers,DC=MANTICORE,DC=local",
			expectedKC:                 nil,
			wantErr:                    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kc := keycredential.NewKeyCredential(
				version.KeyCredentialVersion{Value: version.KeyCredentialVersion_1},
				"",
				crypto.RSAKeyMaterial{},
				guid.GUID{},
				utils.DateTime{},
				utils.DateTime{},
			)

			dnb := ldap.DNWithBinary{}
			bytesRead, err := dnb.Unmarshal([]byte(tt.msDsKeyCredentialLinkValue))
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Unmarshal() error = %v", err)
				}
				return
			}
			if bytesRead != len([]byte(tt.msDsKeyCredentialLinkValue)) {
				t.Errorf("Unmarshal() bytesRead = %v, want %v", bytesRead, len([]byte(tt.msDsKeyCredentialLinkValue)))
				return
			}

			_, err = kc.Unmarshal(dnb.BinaryData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
