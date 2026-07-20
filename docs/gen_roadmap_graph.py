#!/usr/bin/env python3
import os, re

ROOT = "/home/claude/git_projects/TheManticoreProject/Manticore"
TOPS = ["crypto", "encoding", "logger", "network", "utils", "windows"]
EXCLUDE_NAMES = {"docs", "testdata"}  # documentation / test fixtures are not modules

RED, ORANGE, GREEN = 0, 1, 2
NAME = {RED: "red", ORANGE: "orange", GREEN: "green"}

# Explicit states from the human-curated roadmap (functional completeness).
# Keyed by repo-relative path; applied to that dir's OWN base color.
OVERRIDE = {
    "network/smb/smb_v10": GREEN,
    "network/smb/smb_v20": ORANGE,
    "network/smb/smb_v21": RED,
    "network/smb/smb_v30": RED,
    "network/smb/smb_v302": RED,
    "network/smb/smb_v311": RED,
    "network/ldap": ORANGE,
    "network/kerberos": GREEN,
    "network/kerberos/v5": GREEN,
    "network/gssapi": ORANGE,      # skeleton; real GSSAPI is crypto/spnego + kerberos/v5/gssapi
    "network/llmnr": GREEN,
    "network/netbios/nbns": GREEN,
    "network/netbios/nbdgm": GREEN,
    "network/netbios/nbt": ORANGE,
    "network/netbios/nbf": RED,
    "network/dns": RED,        # 1-file stub (MS-DNSP wire types live under windows/protocols)
    "network/tcp": RED,        # roadmap: Raw TCP/IP not implemented
    "network/ip": RED,         # roadmap: Raw TCP/IP not implemented
    "network/dcerpc/ndr": GREEN,
    "network/dcerpc/v4": GREEN,
    "network/dcerpc/v5": GREEN,
    "network/dcerpc/syntax": GREEN,
    "windows/ms-dtyp": ORANGE,
    "windows/filesystem": ORANGE,
    "windows/filesystem/infoclass": ORANGE,
}

def own_metrics(d):
    """non-recursive: (go_loc, test_loc, has_go) for *.go directly in d"""
    go_loc = 0
    test_loc = 0
    has_go = False
    try:
        for f in os.listdir(d):
            p = os.path.join(d, f)
            if not os.path.isfile(p) or not f.endswith(".go"):
                continue
            has_go = True
            with open(p, "r", errors="ignore") as fh:
                n = sum(1 for _ in fh)
            if f.endswith("_test.go"):
                test_loc += n
            else:
                go_loc += n
    except OSError:
        pass
    return go_loc, test_loc, has_go

NEUTRAL = GREEN  # structural dirs defer to their children via worst-color propagation

def metric_color(go_loc, test_loc, has_go, is_leaf):
    # RED is objective: no production Go code in this dir at all (empty / planned).
    if go_loc == 0:
        return RED if is_leaf else NEUTRAL
    # GREEN: real code (>=30 LOC) exercised by a real test (>=20 LOC) => working & verified.
    if go_loc >= 30 and test_loc >= 20:
        return GREEN
    # ORANGE: minimal/stub code, or code with no real test => partial / unverified.
    return ORANGE

def base_color(rel, go_loc, test_loc, has_go, is_leaf):
    if rel in OVERRIDE:
        return OVERRIDE[rel]
    return metric_color(go_loc, test_loc, has_go, is_leaf)

# Build directory tree
dirs = []  # (rel_path, parent_rel_or_None)
for top in TOPS:
    abspath = os.path.join(ROOT, top)
    for cur, subdirs, files in os.walk(abspath):
        subdirs[:] = [s for s in subdirs if s not in EXCLUDE_NAMES and not s.startswith(".")]
        rel = os.path.relpath(cur, ROOT)
        dirs.append(rel)

dirset = set(dirs)

# children map
def parent_of(rel):
    p = os.path.dirname(rel)
    return p if p in dirset else None

children = {rel: [] for rel in dirs}
for rel in dirs:
    p = parent_of(rel)
    if p is not None:
        children[p].append(rel)

# base colors (needs children to know if a dir is a leaf)
base = {}
for rel in dirs:
    go_loc, test_loc, has_go = own_metrics(os.path.join(ROOT, rel))
    is_leaf = len(children[rel]) == 0
    base[rel] = base_color(rel, go_loc, test_loc, has_go, is_leaf)

# propagate worst (min) color bottom-up via memoized recursion
color = {}
def resolve(rel):
    if rel in color:
        return color[rel]
    c = base[rel]
    for ch in children[rel]:
        c = min(c, resolve(ch))
    color[rel] = c
    return c
for rel in dirs:
    resolve(rel)

# root node = worst of tops
root_color = min(color[t] for t in TOPS)

# ---- emit mermaid ----
def nid(rel):
    return "n_" + re.sub(r'[^0-9a-zA-Z]', '_', rel)

# Display pruning (colors above already account for the full tree):
#  - collapse the UUID-keyed DCE/RPC interface registry to a single node
#  - drop the mechanical structures/ and functions/ NDR partitions
COLLAPSE = ("network/dcerpc/interfaces",)
PRUNE_NAMES = {"structures", "functions"}

def emitted(rel):
    b = os.path.basename(rel)
    if b in PRUNE_NAMES:
        return False
    for c in COLLAPSE:
        if rel != c and rel.startswith(c + os.sep):
            return False
    return True

lines = []
lines.append("```mermaid")
lines.append("graph LR")
DOT = {RED: "\U0001F534", ORANGE: "\U0001F7E0", GREEN: "\U0001F7E2"}
# root
lines.append(f'  root["Manticore {DOT[root_color]}"]:::c{root_color}')
for t in TOPS:
    lines.append(f'  root --> {nid(t)}')

vis = [rel for rel in sorted(dirs) if emitted(rel)]
emitted_edges = set()
for rel in vis:
    label = os.path.basename(rel)
    c = color[rel]
    lines.append(f'  {nid(rel)}["{label} {DOT[c]}"]:::c{c}')
for rel in vis:
    p = parent_of(rel)
    if p is not None and emitted(p):
        e = (p, rel)
        if e not in emitted_edges:
            lines.append(f'  {nid(p)} --> {nid(rel)}')
            emitted_edges.add(e)

lines.append("  classDef c2 fill:#1f8b4c,stroke:#155d33,color:#fff;")
lines.append("  classDef c1 fill:#d9822b,stroke:#a35d17,color:#fff;")
lines.append("  classDef c0 fill:#c0392b,stroke:#7d2419,color:#fff;")
lines.append("```")

out = "\n".join(lines)
print(out)
print("\n---- STATS ----", file=os.sys.stderr)
from collections import Counter
cnt = Counter(color[r] for r in vis)
print("visible nodes:", len(vis)+1, "green:", cnt[GREEN], "orange:", cnt[ORANGE], "red:", cnt[RED], file=os.sys.stderr)
