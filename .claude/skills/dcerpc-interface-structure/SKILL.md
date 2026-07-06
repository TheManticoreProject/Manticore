---
name: dcerpc-interface-structure
description: Scaffold and organize DCE/RPC interface code in the Manticore project using the UUID-versioned directory layout (interfaces/<uuid>/<major>.<minor>/ with structures/ and functions/ subpackages). Translating an IDL to a whole interface is automated by the idlgen generator bundled with this skill (idlgen.py) — run it first, then review with this skill. Use this skill whenever creating a new RPC interface, adding a method (opnum) or NDR structure to an existing one, splitting a monolithic interface file, translating an IDL to Go, or wiring an MS-protocol package to the interfaces it composes. Triggers include "add an RPC interface", "scaffold lsarpc/srvsvc/efsr", "add an opnum/function", "add an NDR structure", "implement the methods from this IDL", "organize the dcerpc interface", or "create an ms-protocols package".
---

# DCE/RPC Interface Code Structure

How to lay out and implement DCE/RPC interface code in the Manticore repo. An *interface* is a single RPC abstract syntax (a UUID + version); a *protocol* (MS-LSAD, MS-EFSR, …) composes one or more interfaces into higher-level workflows. Interfaces are the reusable building blocks; protocols sit above them.

**Translating an IDL to a whole interface is automated by `idlgen.py`, bundled in this skill directory — run it first (see [Implementing a whole interface from an IDL](#implementing-a-whole-interface-from-an-idl)), then use the rest of this document to review and refine the generated skeleton.** The generator is the single source of truth for the conventions below, so they stay consistent across every interface.

## Core principle

An RPC interface is identified by its **UUID and version, never by the named pipe** it is reached over. One pipe (e.g. `\lsarpc`) multiplexes several distinct interfaces, so the pipe name cannot name an interface unambiguously. Therefore directories are named by UUID, and the pipe name is just a transport detail recorded inside the descriptor.

## Directory layout

```
network/dcerpc/interfaces/<UUID-with-dashes>/<major>.<minor>/
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

- **Folder = full canonical UUID with dashes**, e.g. `12345778-1234-abcd-ef00-0123456789ab`. Use the *real* UUID; double-check the digit count (32 hex digits / 5 groups). A wrong UUID baked into a path is silent — verify it. (We once shipped a folder missing the trailing `b`.)
- **Version is a nested directory** `<major>.<minor>`, e.g. `0.0`. Different versions of the same interface are different import paths and can coexist.

## Package naming

Go identifiers cannot start with a digit or contain hyphens, so the package name is an **encoding** of the UUID, not the UUID verbatim:

- **Root descriptor package:** `rpcinterface_<UUID-no-dashes>_<major>_<minor>`
  e.g. folder `12345778-1234-abcd-ef00-0123456789ab/0.0/` → `package rpcinterface_123457781234abcdef000123456789ab_0_0`
  The `_<major>_<minor>` suffix keeps versions from colliding when imported side by side.
- **Subpackages keep plain names:** `package structures`, `package functions`.

Because the root package name is unwieldy, **importers always alias it** — by convention to the pipe/interface short name (`lsarpc`, `srvsvc`, …):

```go
import (
    lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/functions"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/12345778-1234-abcd-ef00-0123456789ab/0.0/structures"
)
```

## Dependency direction (must stay acyclic)

```
msdtyp (shared [MS-DTYP] base types)   structures  →  import ndr (+ msdtyp, guid)
                          ▲                ▲
                          └────────────────┤
functions   →  imports ndr, structures, msdtyp, and the root descriptor (aliased)
   ▲
callers / ms-protocols  →  import the descriptor (to bind) + functions (to call)

interface.go (descriptor)  →  imports only syntax, guid, fmt   (depends on nothing in this tree)
```

The descriptor never imports `functions` or `structures`. `functions` imports the descriptor for opnums + status. No cycles.

## Transport neutrality: stubs take `ndr.Invoker`, not a concrete client

The interface tree lives at `network/dcerpc/interfaces/` (NOT under `v5/`), because an interface's abstract syntax and NDR types are independent of the RPC wire-protocol version. A method stub therefore depends only on the small `ndr.Invoker` interface (`Invoke(in ndr.Call, out any) error`), which the connection-oriented `network/dcerpc/v5/client.Client` satisfies (and a future connectionless v4 client can too):

```go
func <Method>(rpc ndr.Invoker, /* args */) (..., error) { ... rpc.Invoke(req, &resp) ... }
```

No `functions` file imports a concrete client or transport. Only the integration/unit **tests** (package `functions_test`) import `v5/client` + transport to build a real client and pass it in (a `*client.Client` is an `ndr.Invoker`).

## What goes where

| Item | Location |
|---|---|
| `SyntaxID()` (UUID + version) | `interface.go` |
| `PipeName` (transport endpoint) | `interface.go` |
| Opnum constants (`OpnumXxx`) | `interface.go` |
| `OpnumToName` map + derived `NameToOpnum` | `interface.go` |
| Status / NTSTATUS codes + `StatusString` | `interface.go` |
| Access-mask / flag constants | `interface.go` |
| Interface-specific NDR data types (one per file) | `structures/` |
| Shared [MS-DTYP] types (`RPC_SID`, `RPC_UNICODE_STRING`, `LUID`, …) | reuse `windows/ms-dtyp` — do **not** redefine |
| Method stubs (one per opnum file) | `functions/` |
| Request/response shapes used by >1 method | `functions/functions.go` |
| Request/response shapes used by 1 method | that method's `functions/<NN>_<Name>.go` |

Naming follows the spec: type names as in the IDL (`LSAPR_HANDLE`, `LSAPR_OBJECT_ATTRIBUTES`), method names with their interface prefix (`LsarOpenPolicy2`, `LsarClose`), opnum constants `Opnum<MethodName>`.

## API surface: direct

Callers use the subpackages directly — `functions.LsarOpenPolicy2(rpc, …)`, `structures.LSAPR_HANDLE`. The root stays a thin descriptor and does **not** re-export. No facade boilerplate to keep in sync.

---

## NDR modeling reference

Translating an IDL to Go is mostly about getting the `ndr` struct tags right. The declarative codec (`network/dcerpc/ndr`) is driven entirely by tags — reach for the `ndr.Marshaler` escape hatch (`AlignmentNDR`/`MarshalNDR`/`UnmarshalNDR`) only when a layout genuinely can't be expressed (rare; we implemented all of lsarpc without it).

### Tag vocabulary (`ndr:"..."`)

| Tag | Meaning |
|---|---|
| `unique` / `ref` / `ptr` | pointer attribute: `[unique]`, `[ref]`, full `[ptr]` |
| `conformant` | conformant array (has `maximum_count`) |
| `varying` | conformant-varying array (`max`, `offset`, `actual_count`) |
| `size_is=Field` | array max count comes from sibling `Field` (implies conformant) |
| `size_is=N` | array max count is the **literal constant `N`** (e.g. `size_is(1000)`); transmitted as `maximum_count` even when fewer elements are sent |
| `length_is=Field` | array actual count from sibling `Field` (implies varying) |
| `elem=ref` / `elem=unique` / `elem=ptr` | pointer attribute of **array elements** (array of pointers) |
| `switch` / `case=N` / `default` | discriminated union (see below) |
| `align=N` | explicit alignment override |
| `wstr` / `str` | force wide/ASCII string mode |

Scalar type mapping: `unsigned long`/`ACCESS_MASK`/`SECURITY_INFORMATION` → `ndr.DWORD` (uint32); `long` → `int32`; `unsigned short` → `uint16`; `short` → `int16`; `unsigned char` → `uint8`; `LARGE_INTEGER` → `msdtyp.LARGE_INTEGER`.

**NDR enums are 16-bit** ([C706] §14.3.6; no `v1_enum` in these IDLs). Model every enum as a named `uint16` with its constants — *not* `uint32`. A `uint32` silently emits 4 bytes and corrupts the wire.

### Reuse the `msdtyp` base types

`windows/ms-dtyp` (package `msdtyp`) holds the [MS-DTYP] common types. Reuse them; never redefine:
- `msdtyp.RPC_SID` — conformant `SubAuthority`; helpers `ParseSID`, `String`.
- `msdtyp.RPC_UNICODE_STRING` — counted UTF-16. **Byte-vs-char gotcha:** `Length`/`MaximumLength` are *byte* counts; the buffer is a `[]uint16` char array. Build with `msdtyp.NewUnicodeString(goString)`; read with `.String()`. Cannot be modeled by a bare `ndr.WSTR`.
- `msdtyp.LUID` (`LowPart`/`HighPart`), `msdtyp.LARGE_INTEGER`, `msdtyp.ULARGE_INTEGER`.

If a needed base type is missing from `msdtyp`, add it there (with a round-trip test), not in an interface's `structures/`.

### The single-vs-double pointer rule (the one that bites)

Mapping IDL parameters/fields to Go fields hinges on pointer depth:

- A **single** top-level pointer `P<TYPE> X` is `[ref]` (NDR transmits no referent id; the referent is in place) → model as the **inline value** `structures.<TYPE>` (no pointer, no tag).
- A **double** pointer `P<TYPE> *X` (or a `[unique]`-marked single pointer) → model the inner as `*structures.<TYPE> \`ndr:"unique"\``.
- A top-level `[out] scalar *X` (e.g. `unsigned long *`) → inline scalar in the response.
- `PLUID Value` at top level → inline `msdtyp.LUID`.

### Arrays

Canonical patterns (see `network/dcerpc/ndr/array_referents_test.go`):
- Array of structs (possibly containing pointers): `[]T \`ndr:"conformant,size_is=N"\``.
- Array of `[unique]` pointers: `[]*T \`ndr:"conformant"\``; of `[ref]` pointers: `[]*T \`ndr:"conformant,elem=ref"\``.
- **`[unique]` pointer to a conformant array** (the enum/lookup-buffer shape `[size_is(n)] PFOO Field`): `Field []FOO \`ndr:"unique,size_is=n"\`` (the walker emits a referent id, then defers the array body). Use `[]*FOO` only if the IDL element is itself a pointer (`PFOO *`).
- Counted byte blob `[size_is(M),length_is(L)] uchar *Buf` → `[]byte \`ndr:"unique,varying,size_is=M,length_is=L"\``.
- **Top-level `[in, size_is(N_literal), length_is(Count)] TYPE Name[*]`** (the SAMR Lookup* shape, e.g. `Names[*]` with `size_is(1000)`) → `[]TYPE \`ndr:"ref,size_is=1000,varying"\``. Three things at once: (1) `ref` (a pointer-to-conformant-array) so the `maximum_count` is **not hoisted ahead of the preceding context handle** — a bare `conformant` field hoists it and the server faults `nca_s_fault_context_mismatch`; (2) the **literal** `size_is=1000` because the server requires that exact constant as `maximum_count` (deriving it from the element count faults `nca_s_fault_ndr`); (3) `varying` for the `offset`/`actual_count` words. `actual_count` is the live element count.

### Discriminated unions (`switch_is` / `switch_type`)

A union is a Go struct with a `switch`-tagged discriminant field plus one field per arm:

```go
type LSAPR_POLICY_INFORMATION struct {
    Class           POLICY_INFORMATION_CLASS `ndr:"switch"`        // discriminant (an enum → uint16)
    AuditLog        POLICY_AUDIT_LOG_INFO    `ndr:"case=1"`        // arms are VALUE fields, matching the IDL
    PrimaryDomain   LSAPR_POLICY_PRIMARY_DOM_INFO `ndr:"case=3"`
    // ... `ndr:"default"` for the [default] arm
}
```

- The discriminant is transmitted **inline** as the first part of the union, even for a non-encapsulated `switch_is` union (its value goes on the wire twice — once as the external param, once here). So add a synthetic `switch` field even when the IDL union has no tag member.
- `case=N` takes the **numeric** discriminant value (compute it from the enum order). If two case labels map to one arm, give each its own field with its own `case=`.
- Arms are **value** fields (the IDL arms are values like `POLICY_AUDIT_LOG_INFO Foo`), not pointers, unless the IDL declares a pointer arm.
- In a request, set the union's discriminant field to match the method's info-class argument before marshalling. See `network/dcerpc/ndr/union.go` for the full model.
- **A `switch_is` union passed as a method argument by pointer** (`[in][switch_is(V)] UNION *Arg`, e.g. SamrConnect5's `SAMPR_REVISION_INFO *InRevisionInfo`) is transmitted **inline** (its own discriminant + arm) — model it as the **inline value** `structures.UNION` (no `*`, no `unique`), and set its `switch` field to the discriminant argument `V`. Wrapping it in `*UNION \`ndr:"unique"\`` emits a stray referent id → `nca_s_fault_ndr`. (Same for the matching `[out] UNION *` parameter.)
- Discriminant **width**: an enum `switch_type` is 16-bit (named `uint16`); a `switch_type(unsigned long)` is 4-byte `ndr.DWORD` (e.g. SAMPR_REVISION_INFO). Within one interface both can coexist.

### Response-shape conventions

- **NTSTATUS only** (Set/Add/Remove/Delete) → use shared `statusResponse{ Status ndr.DWORD }`; Go func returns `error`.
- **`[out] LSAPR_HANDLE *X`** (Open/Create) → use shared `handleResponse{ Handle, Status }`; func returns `(structures.LSAPR_HANDLE, error)`.
- **Other `[out]`/`[in,out]`** → a per-method `<method>Response{ <out/inout fields in IDL order>; Status ndr.DWORD }` with `Status` last.
- An `[in,out]` parameter appears in **both** the request and response structs.
- A `[in] handle_t RpcHandle` (the explicit binding handle, e.g. `LsarLookupSids3`/`Names4`) is **not** marshalled — omit it from the request struct entirely; the Go func still takes only `rpc ndr.Invoker`.

---

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
// endpoint, opnums, opnum<->name maps, status and flag constants). NDR types live in
// the structures subpackage and method stubs in functions; both depend on this package,
// never the reverse.
package rpcinterface_<uuidnodashes>_<maj>_<min>

import (
    "fmt"

    "github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
    "github.com/TheManticoreProject/Manticore/windows/guid"
)

// PipeName is the IPC$-relative named pipe for this interface.
const PipeName = `\<pipe>`

// Opnums ([MS-XXX] section 3.x.4). Opnums "not used on the wire" are omitted.
const (
    OpnumFoo uint16 = 0
    OpnumBar uint16 = 44
)

const ( StatusSuccess uint32 = 0x00000000 /* ... */ )

// SyntaxID returns the abstract syntax identifier: <UUID>, version <maj>.<min>.
func SyntaxID() syntax.SyntaxID {
    return syntax.SyntaxID{
        UUID:         guid.GUID{A: 0x........, B: 0x...., C: 0x...., D: 0x...., E: 0x............},
        MajorVersion: <maj>,
        MinorVersion: <min>,
    }
}

func StatusString(status uint32) string { /* switch over known codes, else 0x%08x */ }

// OpnumToName maps each on-the-wire opnum to its method name; the single source of truth.
var OpnumToName = map[uint16]string{ OpnumFoo: "Foo", OpnumBar: "Bar" }

// NameToOpnum is the reverse, built from OpnumToName so the two never drift.
var NameToOpnum = func() map[string]uint16 {
    m := make(map[string]uint16, len(OpnumToName))
    for op, name := range OpnumToName { m[name] = op }
    return m
}()
```

Build the GUID literal by splitting the UUID `AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE` into `guid.GUID{A: 0xAAAAAAAA, B: 0xBBBB, C: 0xCCCC, D: 0xDDDD, E: 0xEEEEEEEEEEEE}` (A `uint32`, B/C/D `uint16`, E `uint64`). Do **not** use `guid.GUID.FromString` in `SyntaxID` — it returns `(*GUID, error)` and does not fit a struct field.

### `structures/<TYPE>.go`

```go
package structures

import "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
// + msdtyp / guid when the type embeds RPC_SID, RPC_UNICODE_STRING, GUID, etc.

// <TYPE> models <TYPE> ([MS-XXX] 2.2.x). <notes on NDR layout>.
type <TYPE> struct {
    Field ndr.DWORD
    // ...
}
```

Structures must not import `client`.

### `functions/functions.go` (shared shapes)

```go
package functions

import (
    "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/<UUID>/<maj>.<min>/structures"
)

type handleResponse struct { Handle structures.LSAPR_HANDLE; Status ndr.DWORD }
type statusResponse struct { Status ndr.DWORD }
```

### `functions/<NN>_<Method>.go` (one per opnum)

```go
package functions

import (
    "fmt"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
    lsarpc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/<UUID>/<maj>.<min>"
    "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/<UUID>/<maj>.<min>/structures"
)

type <method>Request struct { /* [in] + [in,out] params in IDL order */ }
func (*<method>Request) Opnum() uint16 { return lsarpc.Opnum<Method> }

// <Method> calls <Method> (opnum <NN>) ([MS-XXX] 3.x.y).
func <Method>(rpc ndr.Invoker, /* args */) (structures.LSAPR_HANDLE, error) {
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

The request type implements `ndr.Call` (a single `Opnum() uint16` method); `client.Invoke(in ndr.Call, out any)` marshals the request and unmarshals the response stub. **Gotcha:** a `Set*`-style file that uses only `statusResponse` and an inline union/struct often does *not* reference `ndr.*` — drop the `ndr` import or the build fails with "imported and not used".

## Tests

- **Unit tests** for marshalling/round-trips go in `functions/functions_test.go` as an **external** package `functions_test`, importing the descriptor (aliased), `functions`, `structures`, `client`, `pdu`, `syntax`. Drive the client with an in-memory `fakeTransport` and canned `bind_ack` / response PDUs.
- **Structure round-trip tests** in `structures/` (`ndr.Marshal` → `ndr.Unmarshal` → `reflect.DeepEqual`) for every non-trivial shape — especially unique-pointer-to-conformant-array buffers, arrays of pointer-bearing structs, and each union with ≥2 arms selected. This is where wire-shape bugs surface without a live server.
- **Descriptor tests** (`StatusString`, `SyntaxID`, the `OpnumToName`/`NameToOpnum` round trip) in `interface_test.go`, **internal** package.
- **Integration tests** in `functions/integration_test.go` behind `//go:build integration`, package `functions_test`. SMB → IPC$ → bind → call against a live host via `DCERPC_TEST_HOST` / `DCERPC_TEST_USER` / `DCERPC_TEST_PASS` (optional `DCERPC_TEST_DOMAIN`, `DCERPC_TEST_PORT`). Skip when `DCERPC_TEST_HOST` is unset.

Live-test note: set `smb.NativeOS`/`smb.NativeLanMan` before `SessionSetup`, and open pipes by their IPC$-relative name (`\lsarpc`, not `\pipe\lsarpc`).

**Context-handle scope — per-bind isolation vs. handle chains.** For interfaces where each call is independent (lsarpc/srvsvc), bind a **fresh pipe per method** so a fault can't desync the next call. But context handles are bound to the RPC **association**: an interface whose handles *chain* (SAMR `server→domain→account`, any Open*→use→Close pattern) must run the whole chain on **one** pipe/bind — a handle from a different pipe faults `nca_s_fault_context_mismatch`. Isolate only the fault-prone probes onto their own pipes. The SMB transport also intermittently returns `STATUS_PIPE_EMPTY` (0xc00000d9) on a read; it is transient (retry), not a wire bug, and affects all interfaces.

**Go-level round-trip ≠ wire-correct.** Tags can round-trip in Go yet be wrong against Windows (e.g. discriminant width, `[in,out]` buffers sent empty on request). Treat live integration testing as the real acceptance gate and call out the unverified spots.

---

## Implementing a whole interface from an IDL

**Generate the skeleton with the bundled `idlgen.py` first — do not hand-write the tree.** The generator (`idlgen.py`, in this skill directory) encodes every convention in this document — package naming, the single-vs-double pointer rule, `msdtyp` reuse, the union/array tag rules, the opnum maps — and emits a skeleton that **builds and `go vet`s with zero errors** out of the box (validated against lsarpc, srvsvc, and samr). This skill is then the reference for reviewing the ~10–20% the IDL can't express. Everything below (the NDR modeling reference, pointer/response rules, templates) is exactly what the generator emits and what you verify by hand.

It is a stdlib-only Python script; run it from the repo root (it resolves the Go import path from the nearest `go.mod` above `--out-root`) with `gofmt` on `PATH`. Subcommands: `fetch` (download the IDL from the spec), `parse` (AST/summary), `gen-descriptor`, `gen-structures`, `gen-functions`, `generate` (whole tree), `check` (drift report).

### 1. Get the IDL

Each interface's IDL is published on its Microsoft Open Specifications **"Appendix A: Full IDL"** page. Fetch it straight from there:

```sh
python3 .claude/skills/dcerpc-interface-structure/idlgen.py fetch \
    https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-efsr/4a25b8e1-fd90-41b6-9301-62ed71334436 \
    --out ms-efsr.idl
```

`fetch` downloads the page, concatenates its `<pre>` IDL blocks in order, normalizes the HTML entities and `&nbsp;` indentation, writes the `.idl`, sanity-parses it (printing the interface name + method/typedef counts), and prints a suggested `generate` command. It needs network access and honours the standard `HTTP(S)_PROXY` env vars. (If you already have the `.idl` on disk, skip this step.)

### 2. Generate the tree

```sh
python3 .claude/skills/dcerpc-interface-structure/idlgen.py generate <iface>.idl \
    --out-root network/dcerpc/interfaces --spec MS-XXX --pipe '\<pipe>'
```

This writes `interface.go`, `structures/*.go`, and `functions/<NN>_<Method>.go` under `network/dcerpc/interfaces/<uuid>/<maj>.<min>/` (opnums skip `OpnumNNotUsedOnWire`; the import path is derived from `go.mod`), runs `gofmt`, and prints the count of `TODO(idlgen)` markers. Drive the layers individually when needed: `parse` (AST/summary), `gen-descriptor`, `gen-structures`, `gen-functions`.

### 3. Review what the IDL cannot carry (`TODO(idlgen)` markers + known hard cases)

Grep for `TODO(idlgen)` and reconcile each against the rules above:

- **`interface.go`** — `PipeName` (set via `--pipe`; verify it), the **status / NTSTATUS code table** (not in the IDL — add the interface's codes from [MS-ERREF]/[MS-XXX] and extend `StatusString`), and doc comments.
- **NDR tags the generator can't infer** — verify against this skill: fixed encrypted-password buffers, `SAMPR_LOGON_HOURS`' literal `size_is(1260)` bound, and **per-method tolerance of `STATUS_MORE_ENTRIES`/`SOME_NOT_MAPPED`** (a generated stub accepts only `StatusSuccess`; relax it for `Enumerate*`/`Lookup*`/`QueryDisplay*`).
- **Exported-function ergonomics** — stubs mirror the request fields and use named returns; refine to friendly `string`/`uint32` parameters with conversions where it reads better.
- **Types referenced but absent from the IDL** get a `type X struct{}` placeholder with a `TODO(idlgen)` (e.g. MS-SRVS `SERVER_INFO_100/101`); fill in their fields by hand.
- **NDR `pipe` types** (`typedef pipe …`, e.g. MS-EFSR's `EFS_EXIM_PIPE`) are generated as a slice type plus `ndr:"pipe"` on the parameter (the codec marshals the [C706] 14.7 chunked stream). The whole pipe is buffered in one chunk rather than streamed incrementally, and live validation of pipe methods needs the real service — review those.

If the codec genuinely lacks a feature a type needs (rather than the generator mis-modeling it), file an enhancement issue against `network/dcerpc/ndr` (or `msdtyp`) and defer those methods rather than shipping code that can't be correct.

### 4. Verify

`gofmt -w`, then `go build`, `go vet`, `go test ./<path>/...`, and `go build -tags integration ./<path>/...`. Then **live-validate** — the real acceptance gate (see Tests). Cross-check completeness (count parity is not enough — verify each opnum number):

```sh
grep -aoE '\b[A-Z][A-Za-z0-9]+\(' <iface>.idl   # real methods (filter to the interface prefix)
```

Confirm: (a) every IDL method has exactly one `functions/<NN>_<Name>.go` and exported func; (b) no extras; (c) the file prefix `NN` == the IDL `// Opnum NN` == the `Opnum<Name>` constant; (d) `NotUsedOnWire` opnums are absent.

### 5. Keep regenerations and edits reconcilable: `check`

```sh
python3 .claude/skills/dcerpc-interface-structure/idlgen.py check <iface>.idl \
    --out-root network/dcerpc/interfaces --spec MS-XXX --pipe '\<pipe>'
```

`check` does a code-only diff (ignoring comments) of a freshly generated tree against the committed one and reports **code-identical / differing / generated-only / committed-only** counts plus the differing files. Use it to see what your hand-edits changed, to catch accidental drift, and as a regression guard when `idlgen` itself changes (`--strict` exits non-zero on any code difference). Hand-added files (`functions/functions.go`, `*_test.go`) show up as *committed-only* and are expected.

### Editing by hand (no IDL, or surgical changes)

When there is no IDL or you are adding a single opnum/type, the manual order still applies: **descriptor → structures (one coherent unit, with round-trip tests) → one file per opnum**. Function files are disjoint and parallelize well; the contract when fanning out is: don't touch `interface.go`/`functions.go`/`structures/`/other `NN_*.go`, follow the pointer/response rules above, and run a single authoritative `gofmt` + build/vet/test at the end.

---

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

1. **New opnum:** add `functions/<NN>_<Method>.go` (zero-padded), add `Opnum<Method>` to `interface.go` **and** an entry to `OpnumToName`, put request/response shapes in the method file (or `functions.go` if shared). Apply the pointer/response rules above.
2. **New NDR type:** add `structures/<TYPE>.go` (`package structures`); reuse `msdtyp` for base types; add a round-trip test.
3. **New status/flag constant:** add to `interface.go` (extend `StatusString` if it is a status code).
4. Run `gofmt -w`, `go build`, `go vet`, `go test ./<path>/...`, and `go build -tags integration ./<path>/...` before finishing.
