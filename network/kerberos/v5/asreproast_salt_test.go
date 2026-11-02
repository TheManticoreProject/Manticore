package kerberos

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// TestASRepSalt checks the salt for the enc-part etype is taken from the AS-REP's
// PA-ETYPE-INFO2, or from a legacy PA-ETYPE-INFO, and that an AS-REP without one
// for that etype yields "" (the default salt).
func TestASRepSalt(t *testing.T) {
	info2, err := messages.ETypeInfo2{
		{EType: messages.ETypeRC4HMAC},
		{EType: messages.ETypeAES256CTSHMACSHA196, Salt: "CORP.LOCALolduser"},
	}.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := messages.ETypeInfo{{EType: messages.ETypeAES128CTSHMACSHA196, Salt: []byte("CORP.LOCALlegacy")}}.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	padata := []messages.PAData{
		{PADataType: messages.PAETypeInfo2, PADataValue: info2},
		{PADataType: messages.PAETypeInfo, PADataValue: legacy},
	}

	for _, tc := range []struct {
		etype int
		want  string
	}{
		{messages.ETypeAES256CTSHMACSHA196, "CORP.LOCALolduser"},
		{messages.ETypeAES128CTSHMACSHA196, "CORP.LOCALlegacy"},
		{messages.ETypeRC4HMAC, ""},
	} {
		if got := asRepSalt(padata, tc.etype); got != tc.want {
			t.Errorf("etype %d: salt %q, want %q", tc.etype, got, tc.want)
		}
	}
	if got := asRepSalt(nil, messages.ETypeAES256CTSHMACSHA196); got != "" {
		t.Errorf("no padata: salt %q, want empty", got)
	}
}
