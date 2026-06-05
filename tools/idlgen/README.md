# idlgen — MIDL (.idl) → Go DCE/RPC skeleton generator

`idlgen.py` parses a Microsoft Open Specifications IDL (MS-LSAD, MS-SRVS,
MS-SAMR, …) and generates the Manticore DCE/RPC interface skeleton in the
UUID-versioned layout described by the `dcerpc-interface-structure` skill:

```
network/dcerpc/interfaces/<uuid>/<maj>.<min>/
  interface.go        descriptor: SyntaxID, PipeName, opnums, status, name maps
  structures/<T>.go   one NDR type per file
  functions/<NN>_<M>.go  one method stub per on-the-wire opnum
```

Stdlib-only; no third-party dependencies. Requires `gofmt` on `PATH` for
formatted output (falls back to unformatted if absent).

## Scope

This is a **skeleton generator**, not a wire-perfect one. It gets the mechanical
~80–90% right and marks the judgment calls with `TODO(idlgen)`. Some information
is simply not in the IDL and must be supplied or reviewed by hand:

- the named pipe (`--pipe`), the status/NTSTATUS code table, and doc comments;
- a few NDR tags that were only nailed via live testing (fixed encrypted-password
  buffers, `SAMPR_LOGON_HOURS`' literal bound, per-method tolerance of
  `STATUS_MORE_ENTRIES` / `SOME_NOT_MAPPED`);
- ergonomic exported-function signatures (the generator mirrors the request
  fields; hand-written stubs often take `string`/`uint32` and convert).

Live integration testing against a real server remains the acceptance gate.

## Commands

```sh
# inspect the AST
python3 tools/idlgen/idlgen.py parse <file.idl> [--json]

# generate one layer
python3 tools/idlgen/idlgen.py gen-descriptor <file.idl> --pipe '\samr' --spec MS-SAMR --out interface.go
python3 tools/idlgen/idlgen.py gen-structures <file.idl> --spec MS-SAMR --out <dir>
python3 tools/idlgen/idlgen.py gen-functions  <file.idl> --spec MS-SAMR --import-base <module-path> --out <dir>

# generate the whole interface tree under an interfaces root (import path auto-derived from go.mod)
python3 tools/idlgen/idlgen.py generate <file.idl> --out-root network/dcerpc/interfaces --spec MS-SAMR --pipe '\samr'

# diff the generated tree against the committed one (drift / regression report; --strict to fail on any code diff)
python3 tools/idlgen/idlgen.py check <file.idl> --out-root network/dcerpc/interfaces --spec MS-SAMR --pipe '\samr'
```

## Validation

Developed against the three hand-implemented, live-validated interfaces as ground
truth. A fully generated interface (descriptor + structures + functions) builds
and `go vet`s cleanly for lsarpc (58 methods), srvsvc (47), and samr (64). The
generated descriptors are byte-identical (mechanical parts) to the hand-written
ones, the existing structure round-trip tests pass against the generated samr
types, and 92/102 generated samr structure files are code-identical to
hand-written — the rest being the documented review surface above.
