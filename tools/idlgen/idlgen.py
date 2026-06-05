#!/usr/bin/env python3
"""idlgen — MIDL (.idl) parser and Manticore DCE/RPC Go skeleton generator.

This is Phase 1: an encoding-tolerant tokenizer plus a recursive-descent parser
for the MIDL subset used by the Microsoft Open Specifications IDLs (MS-LSAD,
MS-SRVS, MS-SAMR, …), producing an AST. Later phases turn the AST into the
Go skeleton tree under network/dcerpc/interfaces/<uuid>/<ver>/.

Usage (Phase 1):
    python3 tools/idlgen/idlgen.py parse <file.idl> [--json] [--quiet]

`parse` prints a one-line-per-interface summary (counts of typedefs by kind,
methods, params); `--json` dumps the full AST as JSON instead. Exit status is
non-zero on a parse error (with file:line).

The grammar deliberately covers more than the corpus strictly needs (the
"broader MIDL subset" decision): multiple typedef declarators, pointer-prefix
aliases (`T, *PT`), struct/union/enum bodies, encapsulated and non-encapsulated
unions (`[switch_type]`/`[case]`/`[default]`), field/parameter attribute lists,
array suffixes and conformant `[*]`, `#pragma`, `const`, and `import`.
"""

from __future__ import annotations

import argparse
import dataclasses
import json
import re
import sys
from dataclasses import dataclass, field
from typing import Optional


# --------------------------------------------------------------------------- #
# Source loading (encoding-tolerant)
# --------------------------------------------------------------------------- #

def read_idl(path: str) -> str:
    """Read an .idl file, tolerating the Windows-1252 smart quotes that appear
    in the Microsoft Open Specifications license headers. Only comments contain
    non-ASCII bytes, and the tokenizer strips comments, so a lossy decode is
    safe."""
    with open(path, "rb") as fh:
        raw = fh.read()
    if raw.startswith(b"\xef\xbb\xbf"):  # UTF-8 BOM
        raw = raw[3:]
    for enc in ("utf-8", "cp1252", "latin-1"):
        try:
            return raw.decode(enc)
        except UnicodeDecodeError:
            continue
    return raw.decode("utf-8", errors="replace")


# --------------------------------------------------------------------------- #
# Tokenizer
# --------------------------------------------------------------------------- #

@dataclass
class Token:
    kind: str        # IDENT | NUMBER | STRING | PUNCT | PREPROC
    value: str
    line: int

    def __repr__(self) -> str:  # pragma: no cover - debug aid
        return f"{self.kind}({self.value!r})@{self.line}"


# Punctuation we care about as single-character tokens.
_PUNCT = set("[](){};,*=/+-.<>:")
_IDENT_RE = re.compile(r"[A-Za-z_]\w*")
_NUMBER_RE = re.compile(r"0[xX][0-9A-Fa-f]+|\d+")


class LexError(Exception):
    pass


def tokenize(src: str) -> list[Token]:
    """Turn IDL source into a flat token list. Comments are removed; lines that
    start with `#` become a single PREPROC token carrying the raw line (so the
    parser can record `#pragma pack` and skip the rest)."""
    tokens: list[Token] = []
    i, n, line = 0, len(src), 1
    at_line_start = True
    while i < n:
        c = src[i]
        if c == "\n":
            line += 1
            i += 1
            at_line_start = True
            continue
        if c in " \t\r":
            i += 1
            continue
        # Preprocessor line: from `#` to end of line, emitted whole.
        if c == "#" and at_line_start:
            j = src.find("\n", i)
            j = n if j == -1 else j
            tokens.append(Token("PREPROC", src[i:j].strip(), line))
            i = j
            continue
        at_line_start = False
        # Comments.
        if c == "/" and i + 1 < n and src[i + 1] == "/":
            j = src.find("\n", i)
            i = n if j == -1 else j
            continue
        if c == "/" and i + 1 < n and src[i + 1] == "*":
            j = src.find("*/", i + 2)
            if j == -1:
                raise LexError(f"unterminated /* comment at line {line}")
            line += src.count("\n", i, j)
            i = j + 2
            continue
        # String literal.
        if c == '"':
            j = i + 1
            while j < n and src[j] != '"':
                if src[j] == "\\":
                    j += 1
                j += 1
            if j >= n:
                raise LexError(f"unterminated string at line {line}")
            tokens.append(Token("STRING", src[i + 1:j], line))
            i = j + 1
            continue
        # Identifier / keyword.
        m = _IDENT_RE.match(src, i)
        if m:
            tokens.append(Token("IDENT", m.group(), line))
            i = m.end()
            continue
        # Number (decimal or hex).
        m = _NUMBER_RE.match(src, i)
        if m:
            tokens.append(Token("NUMBER", m.group(), line))
            i = m.end()
            continue
        # Punctuation.
        if c in _PUNCT:
            tokens.append(Token("PUNCT", c, line))
            i += 1
            continue
        raise LexError(f"unexpected character {c!r} at line {line}")
    tokens.append(Token("EOF", "", line))
    return tokens


# --------------------------------------------------------------------------- #
# AST
# --------------------------------------------------------------------------- #

@dataclass
class TypeRef:
    """A type use: a base type name plus pointer depth and any array suffix.

    `base` is the spelled base type (e.g. "unsigned long", "RPC_SID",
    "void"). `ptr` is the number of `*`. `array` is the raw text inside a
    trailing `[...]` (None if not an array; "*" for a conformant `[*]`;
    "" for an empty `[]`)."""
    base: str
    ptr: int = 0
    array: Optional[str] = None


@dataclass
class Field:
    name: str
    type: TypeRef
    attrs: dict = field(default_factory=dict)
    nested: Optional["Typedef"] = None  # inline struct/union/enum defined as this member's type


@dataclass
class Param:
    name: str
    type: TypeRef
    attrs: dict = field(default_factory=dict)  # in/out/size_is/switch_is/...


@dataclass
class Declarator:
    """One name in a typedef declarator list, e.g. `*PFOO` -> name=PFOO, ptr=1."""
    name: str
    ptr: int = 0
    array: Optional[str] = None


@dataclass
class Typedef:
    kind: str                      # "struct" | "union" | "enum" | "alias"
    names: list[Declarator]        # declarator list (canonical + pointer aliases)
    attrs: dict = field(default_factory=dict)
    tag: Optional[str] = None      # the `_TAG` after struct/union/enum
    fields: list[Field] = field(default_factory=list)   # struct/union members
    enumerators: list[tuple[str, Optional[str]]] = field(default_factory=list)
    base: Optional[str] = None     # for "alias": the aliased base type spelling
    switch_type: Optional[str] = None  # for unions: [switch_type(...)]

    @property
    def name(self) -> Optional[str]:
        """The canonical (non-pointer) type name, if any."""
        for d in self.names:
            if d.ptr == 0:
                return d.name
        return self.names[0].name if self.names else None


@dataclass
class Method:
    name: str
    ret: TypeRef
    params: list[Param] = field(default_factory=list)
    attrs: dict = field(default_factory=dict)
    opnum: int = -1  # assigned by declaration order within the interface


@dataclass
class Interface:
    name: str
    attrs: dict = field(default_factory=dict)
    imports: list[str] = field(default_factory=list)
    typedefs: list[Typedef] = field(default_factory=list)
    methods: list[Method] = field(default_factory=list)
    consts: list[dict] = field(default_factory=list)
    pragmas: list[str] = field(default_factory=list)


# --------------------------------------------------------------------------- #
# Parser
# --------------------------------------------------------------------------- #

class ParseError(Exception):
    pass


# Storage-class / sign / length keywords that may precede a base type and that,
# together, still name a single type (e.g. "unsigned long", "wchar_t").
_TYPE_LEAD = {"unsigned", "signed", "long", "short", "char", "int",
              "small", "hyper", "float", "double", "void", "wchar_t",
              "byte", "boolean"}


class Parser:
    def __init__(self, tokens: list[Token], filename: str = "<idl>"):
        self.toks = tokens
        self.pos = 0
        self.filename = filename

    # -- token helpers ----------------------------------------------------- #
    def peek(self, k: int = 0) -> Token:
        return self.toks[min(self.pos + k, len(self.toks) - 1)]

    def next(self) -> Token:
        t = self.toks[self.pos]
        if self.pos < len(self.toks) - 1:
            self.pos += 1
        return t

    def at(self, value: str) -> bool:
        t = self.peek()
        return t.value == value and t.kind in ("PUNCT", "IDENT")

    def at_kind(self, kind: str) -> bool:
        return self.peek().kind == kind

    def expect(self, value: str) -> Token:
        t = self.peek()
        if t.value != value:
            raise ParseError(
                f"{self.filename}:{t.line}: expected {value!r}, got {t.kind} {t.value!r}")
        return self.next()

    def expect_ident(self) -> str:
        t = self.peek()
        if t.kind != "IDENT":
            raise ParseError(
                f"{self.filename}:{t.line}: expected identifier, got {t.kind} {t.value!r}")
        return self.next().value

    # -- attribute blocks -------------------------------------------------- #
    def parse_attrs(self) -> dict:
        """Parse zero or more consecutive `[ name(args), ... ]` blocks into a
        dict mapping attribute name -> raw argument string (or True for flags).
        MIDL allows a declaration to carry several adjacent blocks, e.g.
        `[in] [switch_is(InVersion)]`."""
        attrs: dict = {}
        while self.at("["):
            self.expect("[")
            while not self.at("]"):
                name = self.expect_ident()
                arg = None
                if self.at("("):
                    arg = self._collect_parens()
                self._add_attr(attrs, name, True if arg is None else arg)
                if self.at(","):
                    self.next()
            self.expect("]")
        return attrs

    @staticmethod
    def _add_attr(attrs: dict, name: str, value):
        if name in attrs:
            key = name + "*"
            attrs.setdefault(key, [attrs[name]] if not isinstance(attrs.get(key), list) else attrs[key])
            if key in attrs and isinstance(attrs[key], list):
                attrs[key].append(value)
        attrs[name] = value

    def _collect_parens(self) -> str:
        """Collect the raw text inside a balanced (...) group as a string."""
        self.expect("(")
        depth, parts = 1, []
        while depth > 0:
            t = self.peek()
            if t.kind == "EOF":
                raise ParseError(f"{self.filename}: unterminated '(' ")
            if t.value == "(":
                depth += 1
            elif t.value == ")":
                depth -= 1
                if depth == 0:
                    self.next()
                    break
            parts.append(self.next().value)
        return " ".join(parts).strip()

    def _collect_brackets(self) -> str:
        """Collect raw text inside a balanced [...] group (array suffix)."""
        self.expect("[")
        depth, parts = 1, []
        while depth > 0:
            t = self.peek()
            if t.kind == "EOF":
                raise ParseError(f"{self.filename}: unterminated '['")
            if t.value == "[":
                depth += 1
            elif t.value == "]":
                depth -= 1
                if depth == 0:
                    self.next()
                    break
            parts.append(self.next().value)
        return " ".join(parts).strip()

    # -- type references --------------------------------------------------- #
    def parse_base_type(self) -> str:
        """Parse a (possibly multi-word) base type name and return its spelling.
        Handles `unsigned long`, `wchar_t`, `void`, struct/union/enum tag refs,
        and plain typedef names."""
        words: list[str] = []
        # struct/union/enum used inline as a type reference: `struct _FOO`
        if self.peek().value in ("struct", "union", "enum"):
            words.append(self.next().value)
            if self.at_kind("IDENT"):
                words.append(self.next().value)
            return " ".join(words)
        # leading sign/length keywords may stack: "unsigned long"
        while self.at_kind("IDENT") and self.peek().value in _TYPE_LEAD:
            words.append(self.next().value)
            # "unsigned" / "signed" / "long" can be followed by another lead word
            if self.peek().value in _TYPE_LEAD and words[-1] in (
                    "unsigned", "signed", "long", "short"):
                continue
            break
        if not words:
            words.append(self.expect_ident())
        return " ".join(words)

    def _count_stars(self) -> int:
        ptr = 0
        while self.at("*"):
            self.next()
            ptr += 1
        return ptr

    # -- top level --------------------------------------------------------- #
    def parse(self) -> list[Interface]:
        interfaces: list[Interface] = []
        pending_imports: list[str] = []
        pending_pragmas: list[str] = []
        while not self.at_kind("EOF"):
            t = self.peek()
            if t.kind == "PREPROC":
                self.next()
                if t.value.startswith("#pragma"):
                    pending_pragmas.append(t.value)
                continue
            if t.kind == "IDENT" and t.value == "import":
                self.next()
                while not self.at(";"):
                    if self.at_kind("STRING"):
                        pending_imports.append(self.next().value)
                    else:
                        self.next()
                self.expect(";")
                continue
            if t.value == "[" or (t.kind == "IDENT" and t.value == "interface"):
                iface = self.parse_interface()
                iface.imports = pending_imports
                iface.pragmas += pending_pragmas
                pending_imports, pending_pragmas = [], []
                interfaces.append(iface)
                continue
            # Unknown top-level token — skip defensively.
            self.next()
        return interfaces

    def parse_interface(self) -> Interface:
        attrs = self.parse_attrs()
        self.expect("interface")
        name = self.expect_ident()
        iface = Interface(name=name, attrs=attrs)
        # optional ': base' inheritance (not in our corpus, but tolerate)
        if self.at(":"):
            self.next()
            self.expect_ident()
        if not self.at("{"):
            # forward declaration `interface NAME;`
            if self.at(";"):
                self.next()
            return iface
        self.expect("{")
        opnum = 0
        while not self.at("}"):
            t = self.peek()
            if t.kind == "PREPROC":
                self.next()
                if t.value.startswith("#pragma"):
                    iface.pragmas.append(t.value)
                continue
            if t.kind == "IDENT" and t.value == "import":
                self.next()
                while not self.at(";"):
                    if self.at_kind("STRING"):
                        iface.imports.append(self.next().value)
                    else:
                        self.next()
                self.expect(";")
                continue
            if t.kind == "IDENT" and t.value == "typedef":
                iface.typedefs.append(self.parse_typedef())
                continue
            if t.kind == "IDENT" and t.value == "const":
                iface.consts.append(self.parse_const())
                continue
            if t.kind == "IDENT" and t.value in ("cpp_quote", "midl_pragma"):
                self.next()
                if self.at("("):
                    self._collect_parens()
                if self.at(";"):
                    self.next()
                continue
            # Otherwise: a method declaration (possibly attribute-prefixed).
            method = self.parse_method()
            method.opnum = opnum
            opnum += 1
            iface.methods.append(method)
        self.expect("}")
        if self.at(";"):
            self.next()
        return iface

    def parse_const(self) -> dict:
        self.expect("const")
        parts = []
        while not self.at(";"):
            parts.append(self.next().value)
        self.expect(";")
        return {"raw": " ".join(parts)}

    # -- typedef ----------------------------------------------------------- #
    def parse_typedef(self) -> Typedef:
        self.expect("typedef")
        attrs = self.parse_attrs()
        kw = self.peek().value
        if kw in ("struct", "union", "enum"):
            return self._parse_aggregate_typedef(kw, attrs)
        return self._parse_alias_typedef(attrs)

    def _parse_declarators(self) -> list[Declarator]:
        decls: list[Declarator] = []
        while True:
            ptr = self._count_stars()
            name = self.expect_ident()
            array = None
            if self.at("["):
                array = self._collect_brackets()
            decls.append(Declarator(name=name, ptr=ptr, array=array))
            if self.at(","):
                self.next()
                continue
            break
        return decls

    def _parse_aggregate_typedef(self, kw: str, attrs: dict) -> Typedef:
        self.expect(kw)
        tag = None
        if self.at_kind("IDENT") and not self.at("{"):
            tag = self.next().value
        td = Typedef(kind=kw, names=[], attrs=attrs, tag=tag)
        sw = attrs.get("switch_type")
        if isinstance(sw, str):
            td.switch_type = sw
        self.expect("{")
        if kw == "enum":
            td.enumerators = self._parse_enum_body()
        else:
            td.fields = self._parse_struct_body()
        self.expect("}")
        td.names = self._parse_declarators()
        self.expect(";")
        return td

    def _parse_enum_body(self) -> list[tuple[str, Optional[str]]]:
        out: list[tuple[str, Optional[str]]] = []
        while not self.at("}"):
            name = self.expect_ident()
            val = None
            if self.at("="):
                self.next()
                parts = []
                while not self.at(",") and not self.at("}"):
                    parts.append(self.next().value)
                val = " ".join(parts).strip()
            out.append((name, val))
            if self.at(","):
                self.next()
        return out

    def _parse_struct_body(self) -> list[Field]:
        fields: list[Field] = []
        while not self.at("}"):
            if self.peek().kind == "PREPROC":
                self.next()
                continue
            fields.extend(self._parse_member())
        return fields

    def _parse_member(self) -> list[Field]:
        """Parse one struct/union member line: `[attrs] type decl[, decl];`.
        A `[case(N)]`/`[default]` union arm is just a member with those attrs.
        The member type may itself be an inline (possibly anonymous) struct,
        union, or enum definition, e.g. `[switch_is(L)] union {...} Field;`."""
        attrs = self.parse_attrs()
        # A union arm can be `[case(N)] ;` with no member (empty arm) — tolerate.
        if self.at(";"):
            self.next()
            return []
        nested = None
        if self.peek().value in ("struct", "union", "enum") and self._is_inline_aggregate():
            nested = self._parse_inline_aggregate(attrs)
            base = nested.kind + (f" {nested.tag}" if nested.tag else "")
        else:
            base = self.parse_base_type()
        members: list[Field] = []
        while True:
            ptr = self._count_stars()
            # bitfields / anonymous — tolerate a missing name before ';'
            name = self.expect_ident() if self.at_kind("IDENT") else ""
            array = None
            if self.at("["):
                array = self._collect_brackets()
            members.append(Field(name=name,
                                 type=TypeRef(base=base, ptr=ptr, array=array),
                                 attrs=dict(attrs), nested=nested))
            if self.at(","):
                self.next()
                continue
            break
        self.expect(";")
        return members

    def _is_inline_aggregate(self) -> bool:
        """True if the current struct/union/enum keyword introduces an inline
        body (`kw {`/`kw _TAG {`) rather than a bare type reference."""
        k = 2 if self.peek(1).kind == "IDENT" else 1
        return self.peek(k).value == "{"

    def _parse_inline_aggregate(self, member_attrs: dict) -> Typedef:
        """Parse an inline `struct|union|enum [_TAG] { ... }` used as a member
        type (no declarator list, no trailing ';' consumed here)."""
        kw = self.next().value
        tag = self.next().value if self.at_kind("IDENT") else None
        td = Typedef(kind=kw, names=[], attrs={}, tag=tag)
        sw = member_attrs.get("switch_type")
        if isinstance(sw, str):
            td.switch_type = sw
        self.expect("{")
        if kw == "enum":
            td.enumerators = self._parse_enum_body()
        else:
            td.fields = self._parse_struct_body()
        self.expect("}")
        return td

    def _parse_alias_typedef(self, attrs: dict) -> Typedef:
        base = self.parse_base_type()
        names = self._parse_declarators()
        self.expect(";")
        return Typedef(kind="alias", names=names, attrs=attrs, base=base)

    # -- methods ----------------------------------------------------------- #
    def parse_method(self) -> Method:
        attrs = self.parse_attrs()
        # Return type: base type + stars. The method name is the ident right
        # before '('.
        ret_base = self.parse_base_type()
        ret_ptr = self._count_stars()
        name = self.expect_ident()
        ret = TypeRef(base=ret_base, ptr=ret_ptr)
        m = Method(name=name, ret=ret, attrs=attrs)
        self.expect("(")
        m.params = self.parse_params()
        self.expect(")")
        if self.at(";"):
            self.next()
        return m

    def parse_params(self) -> list[Param]:
        params: list[Param] = []
        if self.at(")"):
            return params
        while True:
            attrs = self.parse_attrs()
            base = self.parse_base_type()
            ptr = self._count_stars()
            # `void` alone denotes no parameters.
            if base == "void" and ptr == 0 and (self.at(")") or self.at(",")):
                if self.at(","):
                    self.next()
                    continue
                break
            name = self.expect_ident() if self.at_kind("IDENT") else ""
            array = None
            if self.at("["):
                array = self._collect_brackets()
            params.append(Param(name=name,
                                type=TypeRef(base=base, ptr=ptr, array=array),
                                attrs=attrs))
            if self.at(","):
                self.next()
                continue
            break
        return params


def parse_file(path: str) -> list[Interface]:
    src = read_idl(path)
    tokens = tokenize(src)
    return Parser(tokens, filename=path).parse()


# --------------------------------------------------------------------------- #
# JSON / summary
# --------------------------------------------------------------------------- #

def ast_to_dict(ifaces: list[Interface]) -> list[dict]:
    def conv(o):
        if dataclasses.is_dataclass(o):
            return {k: conv(v) for k, v in dataclasses.asdict(o).items()}
        if isinstance(o, (list, tuple)):
            return [conv(x) for x in o]
        return o
    return [conv(i) for i in ifaces]


def summarize(ifaces: list[Interface]) -> str:
    lines = []
    for iface in ifaces:
        kinds: dict[str, int] = {}
        for td in iface.typedefs:
            kinds[td.kind] = kinds.get(td.kind, 0) + 1
        nparams = sum(len(m.params) for m in iface.methods)
        uuid = iface.attrs.get("uuid", "?")
        ver = iface.attrs.get("version", "?")
        kindstr = ", ".join(f"{k}:{v}" for k, v in sorted(kinds.items())) or "none"
        not_wire = sum(1 for m in iface.methods if "NotUsedOnWire" in m.name)
        lines.append(
            f"interface {iface.name}  uuid({uuid}) version({ver})\n"
            f"    imports={iface.imports or '[]'}  pragmas={len(iface.pragmas)}\n"
            f"    typedefs={len(iface.typedefs)} [{kindstr}]\n"
            f"    methods={len(iface.methods)} "
            f"(on-the-wire={len(iface.methods) - not_wire}, NotUsedOnWire={not_wire}) "
            f"params={nparams}")
    return "\n".join(lines)


# --------------------------------------------------------------------------- #
# Phase 2: descriptor generator (interface.go)
# --------------------------------------------------------------------------- #

def _norm_attr(raw) -> str:
    """Collapse the whitespace the tokenizer introduces around punctuation in a
    raw attribute argument, e.g. '12345778 - 1234' -> '12345778-1234'."""
    if not isinstance(raw, str):
        return ""
    return re.sub(r"\s+", "", raw)


def parse_uuid(raw: str) -> Optional[tuple[str, str, str, str, str]]:
    s = _norm_attr(raw)
    parts = s.split("-")
    if len(parts) != 5:
        return None
    return tuple(p.lower() for p in parts)  # type: ignore[return-value]


def parse_version(raw: str) -> tuple[int, int]:
    s = _norm_attr(raw)
    if "." in s:
        a, _, b = s.partition(".")
    else:
        a, b = s, "0"
    return int(a or 0), int(b or 0)


def is_on_the_wire(method: Method) -> bool:
    return "NotUsedOnWire" not in method.name


def package_name(uuid_parts, maj: int, min: int) -> str:
    nodash = "".join(uuid_parts)
    return f"rpcinterface_{nodash}_{maj}_{min}"


def gen_descriptor(iface: Interface, pipe: Optional[str], spec_tag: str) -> str:
    uuid_parts = parse_uuid(iface.attrs.get("uuid", ""))
    if uuid_parts is None:
        raise ParseError(f"interface {iface.name}: missing/invalid uuid attribute "
                         f"(supply --uuid)")
    maj, min = parse_version(iface.attrs.get("version", "0.0"))
    pkg = package_name(uuid_parts, maj, min)
    a, b, c, d, e = uuid_parts
    short = iface.name
    pipe = pipe if pipe is not None else f"\\{short}"
    wire = [m for m in iface.methods if is_on_the_wire(m)]
    not_wire = [m.opnum for m in iface.methods if not is_on_the_wire(m)]

    def opnum_const(m: Method) -> str:
        return f"Opnum{m.name}"

    out: list[str] = []
    w = out.append
    w(f"// Package {pkg} is the descriptor for the {short} RPC interface, abstract")
    w(f"// syntax {'-'.join(uuid_parts)} version {maj}.{min} ([{spec_tag}]).")
    w("//")
    w("// Generated by tools/idlgen (descriptor phase) from the IDL, then reviewed:")
    w("// the PipeName, the status-code table, and doc comments are not derivable from")
    w("// the IDL and must be confirmed by hand (see the TODO markers).")
    w(f"package {pkg}")
    w("")
    w("import (")
    w('\t"fmt"')
    w("")
    w('\t"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"')
    w('\t"github.com/TheManticoreProject/Manticore/windows/guid"')
    w(")")
    w("")
    w(f"// PipeName is the IPC$-relative named pipe for the {short} interface.")
    w("// TODO(idlgen): the pipe name is a transport detail not present in the IDL — verify.")
    w(f"const PipeName = `{pipe}`")
    w("")
    if not_wire:
        rng = ", ".join(str(n) for n in not_wire)
        w(f"// Opnums for the on-the-wire methods. Opnums {rng} are \"not used on the wire\"")
        w("// and are omitted.")
    else:
        w("// Opnums for the on-the-wire methods.")
    w("const (")
    for m in wire:
        w(f"\t{opnum_const(m)} uint16 = {m.opnum}")
    w(")")
    w("")
    w("// Status codes returned by this interface.")
    w("// TODO(idlgen): the IDL does not carry the status/NTSTATUS code table; add the")
    w(f"// interface's codes from [MS-ERREF] / [{spec_tag}] and extend StatusString below.")
    w("const (")
    w("\tStatusSuccess uint32 = 0x00000000")
    w(")")
    w("")
    w(f"// SyntaxID returns the {short} abstract syntax identifier:")
    w(f"// {'-'.join(uuid_parts)}, version {maj}.{min}.")
    w("func SyntaxID() syntax.SyntaxID {")
    w("\treturn syntax.SyntaxID{")
    w(f"\t\tUUID:         guid.GUID{{A: 0x{a}, B: 0x{b}, C: 0x{c}, D: 0x{d}, E: 0x{e}}},")
    w(f"\t\tMajorVersion: {maj},")
    w(f"\t\tMinorVersion: {min},")
    w("\t}")
    w("}")
    w("")
    w("// StatusString returns a mnemonic for the documented status codes, otherwise the")
    w("// hex value.")
    w("func StatusString(status uint32) string {")
    w("\tswitch status {")
    w("\tcase StatusSuccess:")
    w('\t\treturn "STATUS_SUCCESS"')
    w("\tdefault:")
    w('\t\treturn fmt.Sprintf("0x%08x", status)')
    w("\t}")
    w("}")
    w("")
    w("// OpnumToName maps each on-the-wire opnum to its method name; the single source of")
    w("// truth.")
    w("var OpnumToName = map[uint16]string{")
    for m in wire:
        w(f'\t{opnum_const(m)}: "{m.name}",')
    w("}")
    w("")
    w("// NameToOpnum is the reverse of OpnumToName, built at init so the two never drift.")
    w("var NameToOpnum = func() map[string]uint16 {")
    w("\tm := make(map[string]uint16, len(OpnumToName))")
    w("\tfor op, name := range OpnumToName {")
    w("\t\tm[name] = op")
    w("\t}")
    w("\treturn m")
    w("}()")
    w("")
    return "\n".join(out)


def _gofmt(src: str) -> str:
    """Run gofmt on a Go source string; return the original on any failure."""
    import shutil
    import subprocess
    if shutil.which("gofmt") is None:
        return src
    try:
        r = subprocess.run(["gofmt"], input=src, capture_output=True, text=True)
        return r.stdout if r.returncode == 0 else src
    except OSError:
        return src


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser(prog="idlgen", description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)
    p = sub.add_parser("parse", help="parse an .idl and print a summary or JSON AST")
    p.add_argument("idl", help="path to the .idl file")
    p.add_argument("--json", action="store_true", help="dump the full AST as JSON")
    p.add_argument("--quiet", action="store_true", help="only report parse errors")

    g = sub.add_parser("gen-descriptor", help="generate interface.go from an .idl")
    g.add_argument("idl", help="path to the .idl file")
    g.add_argument("--pipe", default=None, help="named pipe (default: \\<interface>)")
    g.add_argument("--spec", default="MS-XXX", help="spec tag for doc comments, e.g. MS-SAMR")
    g.add_argument("--out", default=None, help="output file (default: stdout)")
    g.add_argument("--interface", default=None, help="interface name if the IDL has several")

    args = ap.parse_args(argv)

    if args.cmd == "parse":
        try:
            ifaces = parse_file(args.idl)
        except (LexError, ParseError) as e:
            print(f"PARSE ERROR: {e}", file=sys.stderr)
            return 1
        if args.json:
            print(json.dumps(ast_to_dict(ifaces), indent=2))
        elif not args.quiet:
            if not ifaces:
                print("no interfaces found", file=sys.stderr)
                return 1
            print(summarize(ifaces))
        return 0

    if args.cmd == "gen-descriptor":
        try:
            ifaces = parse_file(args.idl)
        except (LexError, ParseError) as e:
            print(f"PARSE ERROR: {e}", file=sys.stderr)
            return 1
        if args.interface:
            ifaces = [i for i in ifaces if i.name == args.interface]
        if not ifaces:
            print("no matching interface found", file=sys.stderr)
            return 1
        try:
            src = _gofmt(gen_descriptor(ifaces[0], args.pipe, args.spec))
        except ParseError as e:
            print(f"ERROR: {e}", file=sys.stderr)
            return 1
        if args.out:
            with open(args.out, "w") as fh:
                fh.write(src)
            print(f"wrote {args.out}", file=sys.stderr)
        else:
            sys.stdout.write(src)
        return 0
    return 2


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
