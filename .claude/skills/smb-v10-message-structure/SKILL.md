---
name: smb-v10-message-structure
description: Implement and organize SMB 1.0 (CIFS) message code in the Manticore project — command Request/Response structs, the command-casting dispatcher, reserved/obsolete command stubs, AndX extended-response variants, and NT_TRANSACT subcommand payloads with their NtTransactRequest/Response builders and parsers. Use this skill whenever adding or fixing an SMB_COM_* command, wiring a command code into the dispatcher, implementing an MS-SMB extended response, or adding an NT_TRANSACT / TRANSACTION2 subcommand payload and its builders. Triggers include "add an SMB command", "implement SMB_COM_X", "wire a command code", "add an extended response", "implement an NT_TRANSACT subcommand", "wire subcommand structs into NtTransact builders", or "add an SMB transaction payload". For DCE/RPC interfaces (NDR/opnums/idlgen) use the dcerpc-interface-structure skill instead — this skill is the SMB message layer below it.
---

# SMB 1.0 (CIFS) Message Code Structure

How to lay out and implement SMB 1.0 message code in the Manticore repo under `network/smb/smb_v10`. This is the **SMB message layer** — distinct from the DCE/RPC interface layer (see the `dcerpc-interface-structure` skill). DCE/RPC rides on `SMB_COM_TRANSACTION`/named-pipe reads, *not* on the NT_TRANSACT subcommands covered here, so the two skills do not overlap.

Everything here was derived from implementing the full CIFS command set, the MS-SMB server-response extensions, and the NT_TRANSACT subcommand payloads against the official MS-CIFS / MS-SMB / MS-FSCC specifications.

## Core principle

**Verify every wire layout against the official spec before writing code, and never guess a format the spec does not pin down.**

- WebFetch the exact MS-CIFS / MS-SMB / MS-FSCC section for the command/structure and copy field names and byte sizes verbatim. The section is linked from the issue or found via the [MS-CIFS] SMB_COM command-code table.
- Cross-spec: a command in MS-CIFS often references a record defined in MS-FSCC (e.g. `FILE_NOTIFY_INFORMATION` [MS-FSCC] 2.7.1, `FILE_QUOTA_INFORMATION` [MS-FSCC] 2.4.41) or a self-relative `SECURITY_DESCRIPTOR` in MS-DTYP 2.4.6. Fetch the referenced page too.
- If the spec gives no marshallable format (obsolete/reserved commands; opaque blobs like a security descriptor), model it as a documented empty stub or opaque `[]byte` — do **not** reconstruct a legacy `[XOPEN-SMB]`/`[SMB-LM1X]` layout from memory.
- All multi-byte integers are **little-endian**.

## Directory layout

```
network/smb/smb_v10/
  message/
    commands/                      one file per SMB_COM_* command
      <Command>Request.go          e.g. NtTransactRequest.go, NtCreateAndxResponse.go
      <Command>Response.go
      0.command_casting.go         the command-code → struct dispatcher (two switches)
      command_interface/           the CommandInterface + embedded Command base
      andx/                        the AndX block (chained commands)
      codes/codes.go               SMB_COM_* CommandCode constants + name map
    parameters/                    the SMB parameter block (WordCount + Words)
    data/                          the SMB data block (ByteCount + Bytes)
    header/, header/flags{,2}/     SMB header + flag bits
  subcommands/                     TRANSACTION / TRANSACTION2 / NT_TRANSACT subcommand
                                   payload structs + enums (FSCTL codes, FILE_* records)
  types/                           SMB/MS-DTYP type aliases (UCHAR, USHORT, ULONG, …)
```

Key fact: **`types.UCHAR = uint8` is an alias**, so `[]types.UCHAR` *is* `[]byte` — pass command `Parameters`/`Data` slices straight to functions taking `[]byte` with no conversion.

## A command Request/Response struct

Each command embeds `command_interface.Command` and implements `Marshal`/`Unmarshal` plus a `New<Command>Request()` constructor that sets the command code. The base `Command` provides `GetParameters()`/`GetData()` (the `parameters.Parameters` and `data.Data` blocks). Mirror an existing well-formed command (e.g. `OpenResponse.go`) rather than starting blank.

Marshal pattern:
1. Lazily create Parameters/Data if nil.
2. If AndX, fold the AndX words into the parameter block first (`for _, p := range c.GetAndX().GetParameters() { c.GetParameters().AddWord(p) }`).
3. Build `rawParametersContent []byte` field-by-field (little-endian), then `c.GetParameters().AddWordsFromBytesStream(rawParametersContent)`.
4. Build `rawDataContent []byte`, then `c.GetData().Add(rawDataContent)`.
5. Marshal parameters then data.

**The `parameters` block enforces `WordCount == len(Words)`** (`parameters.Marshal` errors otherwise, and `AddWordsFromBytesStream` sets `WordCount = len(bytes)/2`). You therefore cannot emit a `WordCount` that disagrees with the actual word bytes — see the extended-response quirk below.

Unmarshal pattern: unmarshal the Parameters and Data blocks, early-return on the all-empty case (it is an error response carrying only a header `Status`), then decode fields by offset. For AndX responses, skip the 4-octet AndX block (`offset += 4`) before the first real parameter.

Always add the command code to `allCommandCodes` in `nil_unmarshal_sweep_test.go` — a sweep test that calls `Unmarshal` on a fresh struct with 3 zero bytes and must not panic. (This is why `Unmarshal` must lazily init Parameters/Data.)

## The command-casting dispatcher

`0.command_casting.go` has **two switches**, `CreateRequestCommand` and `CreateResponseCommand`, mapping a `codes.CommandCode` to `New<Command>Request()` / `New<Command>Response()`. Both end in a `default` that returns `command code not supported: %d`. A command is only reachable once it has a `case` in **both** switches. The two `default` blocks are textually identical — when scripting an insertion, split the file on the default block and insert before the first (request) and second (response) occurrence.

## Reserved / obsolete command stubs

Many `SMB_COM_*` codes are declared in `codes.go` but were never implemented. Two MS-CIFS statuses:

- **N — reserved, never defined** (e.g. `SMB_COM_NEW_FILE_SIZE`, `SMB_COM_QUERY_SERVER`).
- **X — obsolete, real in a legacy dialect but the format lives only in `[SMB-LM1X]`/`[XOPEN-SMB]`/`[SMB-CORE]`, which MS-CIFS does not reproduce** (e.g. `SMB_COM_COPY`, `SMB_COM_MOVE`, `SMB_COM_READ_MPX_SECONDARY`).

For **both** N and X codes, the spec only says "servers SHOULD/MUST return `STATUS_NOT_IMPLEMENTED`" and gives no marshallable format, and the completeness issues ask only to "route the code to a defined handler instead of failing." So both get the **same empty-framing stub**, differing only in doc wording:

- `<Name>Request`/`<Name>Response` embedding `command_interface.Command`, constructor `New<Name>Request()` setting the code.
- `Marshal`/`Unmarshal` emit/consume only the empty `WordCount`(1) + `ByteCount`(2) framing — i.e. marshalled output is exactly `00 / 00 00`.
- Doc comment cites the MS-CIFS section and the SHOULD/MUST `STATUS_NOT_IMPLEMENTED` rule; for X codes note the legacy reference (`[SMB-LM1X]`/`[XOPEN-SMB]`/`[SMB-CORE]`) that MS-CIFS does not republish.
- Wire into both dispatcher switches; add to the `allCommandCodes` sweep.
- Test: dispatch resolves (not the `default` error) + the empty `00 / 00 00` framing round-trips.

`SMB_COM_NO_ANDX_COMMAND` (0xFF) terminates an AndX chain; `SMB_COM_INVALID` (0xFE) is not a real code.

## AndX extended-response variants (MS-SMB)

Some AndX responses have a base CIFS form and a larger MS-SMB form selected by a request flag (`NT_CREATE_ANDX` extended response [MS-SMB] 2.2.4.9.2; `TREE_CONNECT_ANDX` extended response [MS-SMB] 2.2.4.7.2). Model the extension on the existing response struct:

- Add an `Extended bool` selector field plus the extra fields (e.g. `VolumeGUID [16]byte`, `FileId uint64`, `MaximalAccessRights types.ULONG`).
- `Marshal` appends the extra fields after the base fields **only when `Extended`** — the base layout stays byte-identical otherwise.
- `Unmarshal` detects the extended form by the **presence of the extra parameter octets** after the base fields (e.g. "if ≥32 more parameter octets follow `Directory`, set `Extended` and read them"), then decodes them.

**WordCount quirk (NT_CREATE_ANDX extended):** [MS-SMB] says the extended `WordCount` SHOULD be `0x2A` (42), but the extended Words block is actually 50 words (the 4 extra fields add 32 octets to the 68-octet base). Because the `parameters` layer enforces `WordCount == len(Words)`, the implementation emits the **self-consistent** `0x32` (50) and documents the spec's fixed `0x2A` in a comment. Do not try to force `0x2A` — it would desync framing. `TREE_CONNECT_ANDX` has no such quirk (3 → 7 words, consistent).

## NT_TRANSACT subcommands

`SMB_COM_NT_TRANSACT` (0xA0) is a generic container ([MS-CIFS] 2.2.4.62.1). Its Words end with `…SetupCount(1), Function(2), Setup[SetupCount]` — **`Function` (the subcommand code) is separate from the `Setup` words**, and `Setup` is `SetupCount` two-byte words *after* `Function`. The `SMB_Data` is `Pad1, NT_Trans_Parameters[ParameterCount], Pad2, NT_Trans_Data[DataCount]`.

Where each subcommand puts its arguments varies — always check the per-subcommand section:

| Subcommand | Function | Args carried in | Notes |
|---|---|---|---|
| NOTIFY_CHANGE | 0x0004 | **Setup** (CompletionFilter4+FID2+WatchTree1+Reserved1) | response: `FILE_NOTIFY_INFORMATION` list in NT_Trans_Parameters |
| QUERY/SET_SECURITY_DESC | 0x0006 / 0x0003 | NT_Trans_Parameters (FID2+Reserved2+SecurityInformation4) | descriptor = opaque NT_Trans_Data; query response param = LengthNeeded4 |
| QUERY_QUOTA | 0x0007 | NT_Trans_Parameters (16 octets) | SidList = NT_Trans_Data; response = DataLength4 + `FILE_QUOTA_INFORMATION` list |
| IOCTL (copychunk, snapshots, resume-key) | 0x0002 | **Setup** = FunctionCode4+FID2+IsFsctl1+IsFlags1 | FSCTL payload = NT_Trans_Data |

Subcommand **payload structs** live in `subcommands/` (e.g. `srv_copychunk.go`, `srv_snapshot_array.go`, `nt_transact_quota.go`, `nt_transact_notify_change.go`, `nt_transact_security_desc.go`), each with `Marshal`/`Unmarshal` and `SetupCount`/`ChunkCount`/length fields **derived from slice lengths**, not stored redundantly. The FSCTL function codes (`FSCTL_SRV_REQUEST_RESUME_KEY` 0x00140078, `FSCTL_SRV_COPYCHUNK` 0x001440F2, `FSCTL_SRV_ENUMERATE_SNAPSHOTS` 0x00144064) live with their structs.

### Wiring subcommand structs into NtTransactRequest/Response

`NtTransactRequest` carries `Setup []types.USHORT` (added so notify-change and IOCTL subcommands are representable; `SetupCount` is derived from `len(Setup)` on marshal). Builders + parsers live in `message/commands/NtTransactSubcommandBuilders.go` (the `commands` package may import `subcommands` — `subcommands` imports nothing from `commands`, so there is no cycle).

- **Request builders** return a populated `*NtTransactRequest`: set `Function`, the `Setup` words (notify-change setup, or `NewNtTransactIoctlSetup(fsctl, fid, true, 0)` for the FSCTLs), `NT_Trans_Parameters`/`NT_Trans_Data` from the subcommand struct's `Marshal()`, and the counts (`TotalParameterCount`/`ParameterCount`, `TotalDataCount`/`DataCount`, `MaxParameterCount`/`MaxDataCount`).
- **Response parsers** are methods on `*NtTransactResponse` that read `c.Parameters`/`c.Data` (both `[]types.UCHAR` = `[]byte`) into the typed structs.
- To pack a subcommand's little-endian setup byte block into `Setup []USHORT`, read it back as LE `uint16` words.

**Offset-sensitivity gotcha:** a full `NtTransactRequest` `Marshal`→`Unmarshal` round-trip reconstructs `Pad1`/`Pad2` from the **header-relative** `ParameterOffset`/`DataOffset`. Builders deliberately do **not** set those offsets — they are computed at message-assembly time once the header position is known. Consequence for tests: round-trip the **setup-only** case (notify-change) end-to-end, and validate the param/data builders by **re-parsing the populated `NT_Trans_Parameters`/`NT_Trans_Data` fields** rather than a full request round-trip. Large transfers spanning `NT_TRANSACT_SECONDARY` are also a higher-level concern, not the builder's.

## Tests

- **Golden bytes** for fixed layouts: marshal a struct with known field values and assert the exact octet sequence (use real example bytes from the spec trace where available, e.g. the 24-byte copychunk resume key).
- **Round-trip** for variable/derived layouts: marshal → unmarshal → compare.
- **Dispatch** for new command codes: `CreateRequestCommand`/`CreateResponseCommand` resolve without the generic error.
- Add the code to the `nil_unmarshal_sweep_test.go` `allCommandCodes` list.
- None of these obscure paths are live-validated against a server — say so in the PR; golden + round-trip tests are the acceptance gate.

## Checklist when adding an SMB command or subcommand

1. WebFetch the MS-CIFS/MS-SMB/MS-FSCC section; copy field names + byte sizes verbatim. Note the SHOULD/MUST status and whether a format is even defined.
2. Decide the shape: full command struct, reserved/obsolete empty stub, AndX extended variant, or NT_TRANSACT subcommand payload + builder/parser.
3. Implement `Marshal`/`Unmarshal` (little-endian; derive counts from lengths; lazily init Parameters/Data).
4. For commands: wire **both** dispatcher switches + add to the `allCommandCodes` sweep. For subcommands: add builder(s) + response parser(s); ensure `Setup` is set for setup-carrying subcommands.
5. Tests: golden + round-trip (+ dispatch for commands). `gofmt`, `go build ./...`, `go vet`, `go test ./network/smb/smb_v10/...`.
6. Per the Bug-Fixes-Commit-and-PR / Issue-Commit-and-PR skills: branch from `main`, focused commit with a `Fixes #<n>` closing keyword, PR via the template, no Claude attribution. Note in the PR that the path is not live-validated.
