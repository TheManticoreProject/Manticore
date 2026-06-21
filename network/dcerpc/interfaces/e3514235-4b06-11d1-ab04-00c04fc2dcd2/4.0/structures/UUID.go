package structures

import "github.com/TheManticoreProject/Manticore/windows/guid"

// UUID is the [MS-DTYP] GUID/UUID (2.3.4.1) carried by drsuapi messages such as
// puuidClientDsa, UuidDsa, and the various invocation-id fields. The IDL imports it
// from ms-dtyp.idl rather than defining it, so it is aliased to the project's canonical
// guid.GUID (which already has the NDR layout used inline by DSNAME and others).
type UUID = guid.GUID
