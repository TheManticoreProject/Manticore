package hostserver_test

import (
	"reflect"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

// TestEventCarriesNoPayload pins the tracer's safety by construction: an
// Event has no field that can hold frame bytes, field values, credentials
// or message text. Adding one must be a deliberate, reviewed change.
func TestEventCarriesNoPayload(t *testing.T) {
	allowed := map[string]reflect.Kind{
		"Kind": reflect.Uint8, "ServerID": reflect.Uint16, "RequestID": reflect.Uint16,
		"CorrelationID": reflect.Uint32, "Bytes": reflect.Int, "Err": reflect.Interface,
	}
	typ := reflect.TypeOf(hostserver.Event{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if kind, ok := allowed[f.Name]; !ok || f.Type.Kind() != kind {
			t.Errorf("Event field %s %s is not on the reviewed metadata list", f.Name, f.Type)
		}
	}
	if typ.NumField() != len(allowed) {
		t.Errorf("Event has %d fields, reviewed %d", typ.NumField(), len(allowed))
	}
}
