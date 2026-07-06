---
description: Implement a DCE/RPC interface from a spec URL, bug-check it, open a PR, heal CI until green, then squash-merge and close the issue.
argument-hint: <spec-idl-url>
---

You are implementing one **classic DCE/RPC** interface end to end and driving it to a
green PR — or, if it already exists in Manticore, bringing it into full conformance with
the dcerpc-interface-structure skill (same loop, same end state). This loop does NOT
implement DCOM / COM object interfaces (see Phase 0.7). Work autonomously through the
phases below. Do NOT ask for confirmation between phases; only stop to ask if a spec URL
is missing or the IDL fails to parse.

**An already-present protocol/interface is NEVER a reason to skip.** Run the whole
loop as if it did not exist: regenerate, cross-check every file against the skill,
enforce the split layout (functions under `network/dcerpc/interfaces/`, structures
under `windows/protocols/<ms-xxx>/`), and PR the reconciliation.

Input (from `$ARGUMENTS`): the spec "Appendix A: Full IDL" URL. That is the ONLY
argument — you derive everything else from the link and the page it points to.

## Phase 0 — Derive the spec name and pipe from the link
1. **MS-XXX spec name** comes from the URL path segment, e.g.
   `.../windows_protocols/ms-efsr/...` → `MS-EFSR`. Uppercase it.
2. **Named pipe** is NOT in the URL — find it. It is stated in the same spec, in the
   binding/transport section (search the spec for `\PIPE\`, `\pipe\`, "named pipe", or
   the `endpoint(...)` attribute in the IDL). Concretely:
   - After fetching the IDL (Phase 1), grep it for an `endpoint("ncacn_np:...")` line —
     the pipe is the `\pipe\<name>` inside it; the IPC$-relative form is `\<name>`.
   - If the IDL has no endpoint attribute, WebFetch the spec's "Protocol Details /
     Transport" section (or the "Appendix A" page's binding note) and read the pipe
     from there.
   - As a cross-check, the pipe usually matches the interface's common short name
     (lsarpc→`\lsarpc`, efsr→`\efsrpc`, srvsvc→`\srvsvc`).
   Record the derived `MS-XXX` and `\pipe`, and print both before generating so the
   choice is auditable. If the pipe genuinely cannot be determined, stop and ask.

## Phase 0.5 — Detect an existing implementation and choose a mode (never skip)
After fetching the IDL (start of Phase 1) you know the UUID, version, and `<ms-xxx>`.
Probe the committed tree and pick a mode — existing code changes HOW you produce the
result, never WHETHER you do:
- Interface tree present? `network/dcerpc/interfaces/<uuid>/<maj>.<min>/` (interface.go, functions/).
- Structures present? `windows/protocols/<ms-xxx>/`.
- **Legacy inline structures?** `network/dcerpc/interfaces/<uuid>/<maj>.<min>/structures/` — its
  presence means the code predates the split layout and MUST be migrated.
Modes:
- **Fresh** — nothing exists → generate normally (Phase 1, fresh path).
- **Reconcile** — any of the above exists → Phase 1, reconcile path. The deliverable is a
  tree that (a) matches the skill's split layout exactly, (b) covers every opnum/structure the
  IDL defines, and (c) preserves legitimate hand-refinements the generator cannot reproduce
  (status/NTSTATUS tables, per-method `STATUS_*` tolerance, existing tests).

## Phase 0.7 — RPC only: skip DCOM / COM object interfaces
This loop implements classic DCE/RPC interfaces (explicit binding handle, `ncacn_np` /
`ncacn_ip_tcp`), NOT DCOM — which needs object activation (IRemoteSCMActivator, ORPCTHIS/
ORPCTHAT, IRemUnknown) this loop does not model. Immediately after the IDL is fetched
(Phase 1 step 1), classify the interface(s) before generating:
- **DCOM/COM markers:** the interface carries the `[object]` attribute; inherits from
  `IUnknown` / `IDispatch` / another object interface (`interface IFoo : IUnknown`); or the
  IDL imports `ms-dcom` / `objidl` / `oaidl` or references `ORPCTHIS`/`ORPCTHAT`. The
  `[object]` attribute is the definitive marker — classic RPC interfaces never have it.
- Check with `idlgen.py parse <iface>.idl` (inspect the interface attributes) or grep the
  interface header for `object` / `: IUnknown`.
Decision:
- If the interface is an object/DCOM interface → **SKIP**: do NOT generate, branch, or PR.
  End the session immediately (no wakeup) with a one-line report:
  `skipped: DCOM/object interface (<name>)`.
- If the IDL mixes object and non-object interfaces → implement ONLY the non-object (classic
  RPC) interfaces and note which were skipped.
- Otherwise → proceed.

## Phase 1 — Scaffold or reconcile (skill: dcerpc-interface-structure)
1. Invoke the `dcerpc-interface-structure` skill and follow it exactly. Fetch the IDL:
   `idlgen.py fetch <url> --out <iface>.idl`; derive the pipe per Phase 0 (grep for `endpoint(`).
   Then apply the **Phase 0.7 DCOM gate** and the Phase 0.5 existing-code check before generating.
2. **Fresh path:** `idlgen.py generate <iface>.idl --out-root network/dcerpc/interfaces
   --spec <derived MS-XXX> --pipe '<derived \pipe>'`.
3. **Reconcile path (existing protocol) — do NOT blindly overwrite hand-tuned code:**
   a. If a LEGACY `structures/` subdir exists under the interface, migrate it to the split
      layout FIRST (the skill documents this): `git mv` the type files to
      `windows/protocols/<ms-xxx>/`, rename `package structures` → `package <msxxx>`, rewire
      every functions file (structures import → the protocol package, aliased; `structures.` →
      `<msxxx>.`), and delete the now-empty `structures/`.
   b. Generate a fresh tree into a scratch dir and run
      `idlgen.py check <iface>.idl --out-root network/dcerpc/interfaces --spec <MS-XXX>
      --pipe '<\pipe>'` to get the `[iface]`/`[struct]` drift report. For each
      differing/generated-only file, reconcile TOWARD the skill: add missing opnums and
      structures, fix wrong NDR tags / package names / file placement — but keep correct
      hand-written refinements; never replace a fuller hand-tuned stub with a barer generated one.
   c. Assert the split invariants: NO `structures/` under the interface tree; every NDR type
      lives only in `windows/protocols/<ms-xxx>/`, one file per type, `package <msxxx>`, with a
      single `structures_test.go`; functions import the structures package aliased and qualify
      types with `<msxxx>.`.
4. **Cross-check the WHOLE result against the skill** (both paths): directory layout, package
   naming, dependency direction (functions → structures, never the reverse; structures import
   nothing under `interfaces/`), the single-vs-double pointer rule, union/array tags,
   response-shape conventions, and the two-"protocol"-locations distinction
   (`windows/protocols/<ms-xxx>` = wire structures vs `network/dcerpc/ms-protocols/<ms-xxx>` =
   composition). Fix every deviation.
5. Report: UUID, version, opnum count, derived MS-XXX/pipe, **mode (fresh/reconcile)**, TODO
   count, and — in reconcile mode — a short list of what was migrated / added / fixed.

## Phase 2 — Refine (same skill)
Reconcile every `TODO(idlgen)` against the skill's NDR / pointer / response rules:
status-code table + StatusString, enum widths (16-bit), single-vs-double pointer,
union discriminants, literal `size_is`, and `STATUS_MORE_ENTRIES`/`SOME_NOT_MAPPED`
tolerance on Enumerate*/Lookup*. Add structure round-trip tests.

## Phase 3 — Local verification (catch CI failures before pushing)
The generator emits the split layout: the descriptor + stubs under
`network/dcerpc/interfaces/<uuid>/...` and the NDR structures under
`windows/protocols/<ms-xxx>/...`. Verify BOTH trees, and fix until all pass:
- `gofmt -l` (must be empty)
- `go build ./network/dcerpc/interfaces/<uuid>/... ./windows/protocols/<ms-xxx>/...`
- `go vet ./network/dcerpc/interfaces/<uuid>/... ./windows/protocols/<ms-xxx>/...`
- `go test -count=1 ./network/dcerpc/interfaces/<uuid>/... ./windows/protocols/<ms-xxx>/...`
- Cross-compile the 32-bit leg CI runs (this is where int-overflow bugs hide):
  `GOARCH=386 go build ./...` and `GOARCH=arm64 go build ./...`
- `go build -tags integration ./network/dcerpc/interfaces/<uuid>/...`

## Phase 4 — Bug check
Run the `/code-review` skill on the diff at effort `high`. Confirm each finding by
code analysis before acting (no speculation). Fix confirmed bugs in place.

## Phase 5 — Commit & PR (skill: bug-review-and-fix / Bug-Fixes-Commit-and-PR)
Follow that skill's templates and constraints. CRITICAL OVERRIDE: include NO mention
of Claude anywhere — not in the commit message, branch name, issue, PR title, or body.
Do NOT add a Co-Authored-By trailer.
- Branch off `main`: fresh mode → `feature-<iface>-interface`; reconcile mode →
  `refactor-<ms-xxx>-split-structures` (or `bugfix-<slug>` if the change is purely a fix).
- Commit, push, open a PR against `main` using the PR template. In reconcile mode, state in
  the PR body that this brings an existing interface into the skill's split layout, and list
  what moved (interface ↔ `windows/protocols/<ms-xxx>`), what was added, and what was fixed.
- If reconcile mode produced NO diff (the committed tree already conforms exactly), skip the
  PR and report "already conformant" instead — that is the one case where no PR is opened.
- Capture the PR number and branch name for the next phase.

## Phase 6 — Heal CI until green (the loop's inner cycle)
1. `gh pr checks <pr> --watch=false` (or `gh run list --branch <branch> --limit 1`).
2. If a run is still in progress: end this turn WITHOUT finishing; schedule a wakeup
   ~270s out to re-check (stays in the prompt-cache window).
3. If a run FAILED: `gh run view <id> --log-failed`, reproduce locally (use the exact
   GOOS/GOARCH from the failing matrix leg), fix minimally, commit (clean message, no
   Claude mention), push, then re-check.
4. When ALL matrix legs are green, proceed to Phase 7.

## Phase 7 — Merge, then terminate
1. Confirm the PR is mergeable and every required check passed:
   `gh pr checks <pr>` shows all green, and `gh pr view <pr> --json mergeable,mergeStateStatus`
   reports `MERGEABLE` / `CLEAN`. If it is `BLOCKED` (branch protection requires reviews or
   an admin merge), do NOT force it — report that a human merge is needed and STOP.
2. Merge with a squash so the PR body's `Closes #<issue>` auto-closes the issue, and clean up:
   `gh pr merge <pr> --squash --delete-branch`
   (If the run is unattended and branch protection would otherwise block a same-author merge,
   `gh pr merge <pr> --squash --delete-branch --admin` — only when you own the repo and intend
   to bypass protection.)
3. Verify: `gh pr view <pr> --json state` returns `MERGED`, and the linked issue is closed.
4. Post a one-line summary (interface, opnums, PR link, "merged"), then **STOP — do not
   schedule another wakeup.** The loop ends here.

Terminate (stop scheduling wakeups) on any of: PR merged (success); CI red and unfixable
after a reasonable number of attempts; or merge BLOCKED by branch protection. In the last
two cases, report clearly why you stopped.

Report at the end: interface UUID/version, opnums implemented, PR link, and final CI status.
