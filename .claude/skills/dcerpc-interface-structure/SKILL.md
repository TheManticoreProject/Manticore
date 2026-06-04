---
name: dcerpc-interface-structure
description: Scaffold and organize DCE/RPC interface code in the Manticore project using the UUID-versioned directory layout (interfaces/<uuid>/<major>.<minor>/ with structures/ and functions/ subpackages). Use this skill whenever creating a new RPC interface, adding a method (opnum) or NDR structure to an existing one, splitting a monolithic interface file, or wiring an MS-protocol package to the interfaces it composes. Triggers include "add an RPC interface", "scaffold lsarpc/srvsvc/efsr", "add an opnum/function", "add an NDR structure", "organize the dcerpc interface", or "create an ms-protocols package".
---

# DCE/RPC Interface Code Structure

How to lay out DCE/RPC interface code in the Manticore repo. An *interface* is a single RPC abstract syntax (a UUID + version); a *protocol* (MS-LSAD, MS-EFSR, …) composes one or more interfaces into higher-level workflows. Interfaces are the reusable building blocks; protocols sit above them.

## Core principle

An RPC interface is identified by its **UUID and version, never by the named pipe** it is reached over. One pipe (e.g. `\lsarpc`) multiplexes several distinct interfaces, so the pipe name cannot name an interface unambiguously. Therefore directories are named by UUID, and the pipe name is just a transport detail recorded inside the descriptor.

## Directory layout

```
network/dcerpc/v5/interfaces/<UUID-with-dashes>/<major>.<minor>/
  interface.go        package rpcinterface_<UUID-no-dashes>_<major>_<minor>
  interface_test.go
  structures/         package structures
    <NDR_TYPE_NAME>.go        (one file per NDR type)
  functions/          package functions
    functions.go              (shared request/response shapes only)
    <NN>_<MethodName>.go      (one file per opnum; NN = zero-padded opnum)
    functions_test.go
    integration_test.go       (//go:build integration)
```

- **Folder = full canonical UUID with dashes**, e.g. `12345778-1234-abcd-ef00-0123456789ab`. Use the *real* UUID; double-check the digit count (32 hex digits / 5 groups). A wrong UUID baked into a path is silent — verify it.
- **Version is a nested directory** `<major>.<minor>`, e.g. `0.0`. Different versions of the same interface are different import paths and can coexist.

## Package naming

Go identifiers cannot start with a digit or contain hyphens, so the package name is an **encoding** of the UUID, not the UUID verbatim:

- **Root descriptor package:** `rpcinterface_<UUID-no-dashes>_<major>_<minor>`
  e.g. folder `12345778-1234-abcd-ef00-0123456789ab/0.0/` → `package rpcinterface_123457781234abcdef000123456789ab_0_0`
  The `_<major>_<minor>` suffix keeps versions from colliding when imported side by side.
- **Subpackages keep plain names:** `package structures`, `package functions`. The import path disambiguates them across interfaces; identical-prefix names would only force aliases everywhere.

Because the root package name is unwieldy, **importers always alias it** — by convention to the pipe/interface short name:

```go
import (
    lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/functions"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)
```

## Dependency direction (must stay acyclic)

```
structures  →  imports only ndr
   ▲
functions   →  imports client, ndr, structures, and the root descriptor (aliased)
   ▲
callers / ms-protocols  →  import the descriptor (to bind) + functions (to call)

interface.go (descriptor)  →  imports only syntax, guid, fmt   (depends on nothing in this tree)
```

The descriptor never imports `functions` or `structures`. `functions` imports the descriptor for opnums + status. No cycles.

## What goes where

| Item | Location |
|---|---|
| `SyntaxID()` (UUID + version) | `interface.go` |
| `PipeName` (transport endpoint) | `interface.go` |
| Opnum constants (`OpnumXxx`) | `interface.go` |
| Status / NTSTATUS codes + `StatusString` | `interface.go` |
| Access-mask / flag constants | `interface.go` |
| NDR data types (one per file) | `structures/` |
| Method stubs (one per opnum file) | `functions/` |
| Request/response shapes used by >1 method | `functions/functions.go` |
| Request/response shapes used by 1 method | that method's `functions/<NN>_<Name>.go` |

Naming follows the spec: type names as in the IDL (`LSAPR_HANDLE`, `LSAPR_OBJECT_ATTRIBUTES`), method names with their interface prefix (`LsarOpenPolicy2`, `LsarClose`), opnum constants `Opnum<MethodName>`.

## API surface: direct

Callers use the subpackages directly — `functions.LsarOpenPolicy2(rpc, …)`, `structures.LSAPR_HANDLE`. The root stays a thin descriptor and does **not** re-export. No facade boilerplate to keep in sync.

## Templates

### `interface.go` (root descriptor)

```go
// Package rpcinterface_<uuidnodashes>_<maj>_<min> is the descriptor for the
// <Name> (<pipe>) RPC interface, abstract syntax <UUID> version <maj>.<min> ([MS-XXX]).
//
// An RPC interface is identified by its UUID and version, never by the named pipe it
// is reached over: that pipe multiplexes several interfaces, so the directory is named
// after the UUID with the version in the nested <maj>.<min>/ directory.
//
// This package holds only the interface-level descriptor (abstract syntax, transport
// endpoint, opnums, status and flag constants). NDR types live in the structures
// subpackage and method stubs in functions; both depend on this package, never the reverse.
//
// References:
//   - [MS-XXX] ... : <url>
package rpcinterface_<uuidnodashes>_<maj>_<min>

import (
    "fmt"

    "github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
    "github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for this interface.
const PipeName = `\<pipe>`

// Opnums for the implemented methods ([MS-XXX] section 3.x.4).
const (
    OpnumFoo uint16 = 0
    OpnumBar uint16 = 44
)

// Status codes ([MS-XXX] return values).
const (
    StatusSuccess uint32 = 0x00000000
    // ...
)

// SyntaxID returns the abstract syntax identifier: <UUID>, version <maj>.<min>.
func SyntaxID() syntax.SyntaxID {
    return syntax.SyntaxID{
        UUID:         guid.GUID{A: 0x........, B: 0x...., C: 0x...., D: 0x...., E: 0x............},
        MajorVersion: <maj>,
        MinorVersion: <min>,
    }
}

func StatusString(status uint32) string {
    switch status {
    case StatusSuccess:
        return "STATUS_SUCCESS"
    default:
        return fmt.Sprintf("0x%08x", status)
    }
}
```

Build the GUID literal by splitting the UUID `AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE` into `guid.GUID{A: 0xAAAAAAAA, B: 0xBBBB, C: 0xCCCC, D: 0xDDDD, E: 0xEEEEEEEEEEEE}` (A `uint32`, B/C/D `uint16`, E `uint64`). Do **not** use `guid.GUID.FromString` in `SyntaxID` — it returns `(*GUID, error)` and does not fit a struct field.

### `structures/<TYPE>.go`

```go
package structures

import (
    "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
)

// <TYPE> models <TYPE> ([MS-XXX] 2.2.x). <notes on NDR layout>.
type <TYPE> struct {
    Field ndr.DWORD
    // ...
}
```

Use `ndr` types (`ndr.DWORD` = uint32, `ndr.WSTR`, etc.) and NDR struct tags (e.g. `` `ndr:"unique"` `` for `[unique]` pointers). Structures must not import `client`.

### `functions/functions.go` (shared shapes)

```go
// Package functions implements the method stubs of the <Name> interface.
// Each opnum lives in its own <NN>_<Method>.go file; this file holds request/response
// shapes shared across more than one method.
package functions

import (
    "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/<UUID>/<maj>.<min>/structures"
)

type handleResponse struct {
    Handle structures.LSAPR_HANDLE
    Status ndr.DWORD
}
```

### `functions/<NN>_<Method>.go` (one per opnum)

```go
package functions

import (
    "fmt"

    "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
    lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/<UUID>/<maj>.<min>"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/<UUID>/<maj>.<min>/structures"
)

// <method>Request is the [in] parameter set of <Method>.
type <method>Request struct {
    // ... ndr-tagged fields, referencing structures.* for shared types
}

func (*<method>Request) Opnum() uint16 { return lsarpc.Opnum<Method> }

// <Method> calls <Method> (opnum <NN>) ...
func <Method>(rpc *client.Client, /* args */) (structures.LSAPR_HANDLE, error) {
    req := &<method>Request{ /* ... */ }
    var resp handleResponse
    if err := rpc.Invoke(req, &resp); err != nil {
        return structures.LSAPR_HANDLE{}, fmt.Errorf("<Method>: %w", err)
    }
    if uint32(resp.Status) != lsarpc.StatusSuccess {
        return resp.Handle, fmt.Errorf("<Method> failed: %s", lsarpc.StatusString(uint32(resp.Status)))
    }
    return resp.Handle, nil
}
```

The request type implements `ndr.Call` (a single `Opnum() uint16` method); `client.Invoke(in ndr.Call, out any)` marshals the request and unmarshals the response stub.

## Tests

- **Unit tests** for marshalling/round-trips go in `functions/functions_test.go` as an **external** package `functions_test`, importing the descriptor (aliased), `functions`, `structures`, `client`, `pdu`, `syntax`. Drive the client with an in-memory `fakeTransport` and canned `bind_ack` / response PDUs.
- **Descriptor tests** (`StatusString`, `SyntaxID`) go in `interface_test.go` as an **internal** test (same package as the descriptor).
- **Integration tests** go in `functions/integration_test.go` behind `//go:build integration`, package `functions_test`. They establish SMB → IPC$ → bind → call against a live host using `DCERPC_TEST_HOST` / `DCERPC_TEST_USER` / `DCERPC_TEST_PASS` (optional `DCERPC_TEST_DOMAIN`, `DCERPC_TEST_PORT`). Skip when `DCERPC_TEST_HOST` is unset.

Live-test note: set `smb.NativeOS`/`smb.NativeLanMan` before `SessionSetup`, and open pipes by their IPC$-relative name (`\lsarpc`, not `\pipe\lsarpc`).

## Protocol layer (`network/dcerpc/ms-protocols/`)

A protocol composes interfaces. It binds via the descriptor's `SyntaxID()` / `PipeName` and calls into `functions`:

```go
// network/dcerpc/ms-protocols/ms-lsad/  package mslsad
import (
    lsarpc "github.com/.../interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
    "github.com/.../interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/functions"
)

rpc.Bind(lsarpc.SyntaxID())
h, _ := functions.LsarOpenPolicy2(rpc, lsarpc.MaximumAllowed)
functions.LsarClose(rpc, h)
```

One protocol may reference multiple interfaces; one interface may be reused by multiple protocols. Keep interface packages free of protocol-level logic.

## Checklist when adding to an interface

1. **New opnum:** add `functions/<NN>_<Method>.go` (zero-padded opnum), add `Opnum<Method>` to `interface.go`, put request/response shapes in the method file (or `functions.go` if shared).
2. **New NDR type:** add `structures/<TYPE>.go`, `package structures`, import `ndr` only.
3. **New status/flag constant:** add to `interface.go` (extend `StatusString` if it is a status code).
4. Run `go build`, `go vet`, `go test ./<path>/...`, and `go build -tags integration ./<path>/...` before finishing.
