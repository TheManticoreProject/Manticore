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


def fetch_idl(url: str) -> str:
    """Download a Microsoft Open Specifications "Appendix A: Full IDL" page and
    return the IDL text. The IDL is rendered as one or more <pre> code blocks; we
    concatenate them in document order, strip any inner markup, and unescape HTML
    entities. Needs network access (honours the standard HTTP(S)_PROXY env vars).

    Example:
        https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-efsr/4a25b8e1-fd90-41b6-9301-62ed71334436
    """
    import html as _html
    import urllib.request

    req = urllib.request.Request(url, headers={"User-Agent": "Mozilla/5.0 (idlgen)"})
    doc = urllib.request.urlopen(req, timeout=30).read().decode("utf-8", "replace")
    blocks = re.findall(r"<pre[^>]*>(.*?)</pre>", doc, re.S)
    if not blocks:
        raise ParseError(f"no <pre> code blocks found at {url} (is it an "
                         f"'Appendix A: Full IDL' page?)")
    parts = [_html.unescape(re.sub(r"<[^>]+>", "", b)) for b in blocks]
    # MS Learn renders indentation with &nbsp; (U+00A0) and may use Unicode
    # spaces/zero-width chars; normalize to plain ASCII whitespace.
    text = "\n".join(parts)
    return text.replace("\xa0", " ").replace("​", "").replace("﻿", "")


def spec_tag_from_url(url: str) -> Optional[str]:
    """Best-effort [MS-XXX] spec tag from an openspecs URL (.../ms-efsr/... -> MS-EFSR)."""
    m = re.search(r"/(ms-[a-z0-9]+)/", url)
    return m.group(1).upper() if m else None


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
        if c in " \t\r\xa0":  # \xa0 = non-breaking space, common in pasted/HTML IDL
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
    is_pipe: bool = False          # `typedef pipe <base> NAME` — an NDR pipe of <base>

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
        # NDR `pipe` modifier (a streaming type, e.g. `typedef pipe unsigned char T`):
        # skip it and parse the element type. Pipe streaming is not modeled — the
        # methods that use the pipe type need manual review.
        if self.peek().value == "pipe":
            self.next()
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
        is_pipe = self.peek().value == "pipe"  # `typedef pipe <base> NAME` (parse_base_type skips the keyword)
        base = self.parse_base_type()
        names = self._parse_declarators()
        self.expect(";")
        return Typedef(kind="alias", names=names, attrs=attrs, base=base, is_pipe=is_pipe)

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


# --------------------------------------------------------------------------- #
# Phase 3: structures generator (structures/<TYPE>.go)
# --------------------------------------------------------------------------- #

# Base C / MIDL scalars (multi-word spellings) -> Go type.
_SCALAR = {
    "unsigned long": "ndr.DWORD", "long": "int32",
    "unsigned short": "uint16", "short": "int16",
    "unsigned int": "uint32", "int": "int32",
    "unsigned char": "uint8", "char": "byte", "byte": "uint8",
    "unsigned hyper": "uint64", "hyper": "int64",
    "wchar_t": "uint16", "boolean": "bool",
    "float": "float32", "double": "float64",
}
# [MS-DTYP] scalar typedef names (imported) -> Go type.
_DTYP_SCALAR = {
    "DWORD": "ndr.DWORD", "ULONG": "ndr.DWORD", "LONG": "int32",
    "UINT": "uint32", "INT": "int32", "USHORT": "uint16", "SHORT": "int16",
    "WORD": "uint16", "UCHAR": "uint8", "BYTE": "uint8", "CHAR": "int8",
    "BOOL": "ndr.BOOL", "BOOLEAN": "bool", "ULONG64": "uint64",
    "ULONGLONG": "uint64", "DWORD64": "uint64", "LONGLONG": "int64",
    "HYPER": "int64", "WCHAR": "uint16", "NTSTATUS": "ndr.DWORD",
    "NET_API_STATUS": "ndr.DWORD", "LCID": "ndr.DWORD",
    "ACCESS_MASK": "ndr.DWORD", "SECURITY_INFORMATION": "ndr.DWORD",
    "RPC_STATUS": "int32",
}
# [MS-DTYP] aggregate types provided by windows/ms-dtyp (reuse, don't emit).
_DTYP_STRUCT = {"RPC_SID", "RPC_UNICODE_STRING", "LARGE_INTEGER",
                "ULARGE_INTEGER", "LUID"}
# Built-in [MS-DTYP] pointer aliases -> canonical msdtyp type.
_DTYP_PALIAS = {
    "PRPC_SID": "RPC_SID", "PRPC_UNICODE_STRING": "RPC_UNICODE_STRING",
    "PLARGE_INTEGER": "LARGE_INTEGER", "PLUID": "LUID",
}
# Wide-string typedefs ([string] wchar_t*): modeled as a [unique] *ndr.WSTR.
_STR_TYPES = {"LMSTR", "LMCSTR", "LPWSTR", "PWSTR", "LPCWSTR", "PWCHAR"}


class TypeResolver:
    """Catalogs an interface's typedefs so field/param types can be resolved to
    Go types and pointer aliases expanded."""

    def __init__(self, iface: Interface):
        self.kinds: dict[str, str] = {}               # canonical name -> kind
        self.palias: dict[str, tuple[str, int]] = {}  # PFOO -> (canonical, depth)
        self.scalar_alias: dict[str, str] = {}        # alias -> Go scalar
        self.ctx_handles: set[str] = set()
        self.str_handles: set[str] = set()            # [handle] wchar_t* string handles
        self.pipes: dict[str, str] = {}               # pipe type name -> Go element type
        self.extra_names: dict[str, list[str]] = {}   # primary -> [extra non-ptr aliases]
        self.unresolved: set[str] = set()             # type names referenced but not defined
        self.enum_values: dict[str, int] = {}         # enumerator name -> numeric value
        for td in iface.typedefs:
            if td.kind == "enum":
                nv = 0
                for en, ev in td.enumerators:
                    v = _eval_const(_norm_attr(ev)) if (ev and str(ev).strip()) else nv
                    if v is None:
                        nv += 1
                        continue
                    self.enum_values[en] = v
                    nv = v + 1
        for td in iface.typedefs:
            canon = td.name
            if canon:
                self.kinds[canon] = td.kind
            # The first declarator is always the primary (canonical) type and is
            # never a pointer alias of itself. Later pointer declarators are
            # pointer aliases of the primary; later non-pointer declarators are
            # plain Go aliases of it.
            for idx, d in enumerate(td.names):
                if idx == 0:
                    primary = d.name
                    continue
                if d.ptr >= 1:
                    if d.name != primary:
                        self.palias[d.name] = (primary, d.ptr)
                else:
                    self.extra_names.setdefault(canon, []).append(d.name)
                    self.kinds[d.name] = td.kind
            if td.kind == "alias":
                has_value_decl = any(d.ptr == 0 for d in td.names)
                if td.is_pipe:
                    # NDR pipe of <base>; model as a Go slice with the `pipe` tag.
                    if canon:
                        self.pipes[canon] = _SCALAR.get(td.base) or _DTYP_SCALAR.get(td.base) or "byte"
                elif "context_handle" in td.attrs and td.base == "void":
                    if canon:
                        self.ctx_handles.add(canon)
                elif td.base == "wchar_t" and not has_value_decl:
                    if canon:
                        self.str_handles.add(canon)  # [handle] wchar_t* string handle
                elif canon and td.base and has_value_decl:
                    # Only a by-value scalar alias; a pointer-only alias
                    # (e.g. [handle] wchar_t *PSAMPR_SERVER_NAME) is a string
                    # handle resolved at use sites, not a scalar type.
                    go = _SCALAR.get(td.base) or _DTYP_SCALAR.get(td.base)
                    if go:
                        self.scalar_alias[canon] = go

    def is_local(self, name: str) -> bool:
        return name in self.kinds and name not in _DTYP_STRUCT

    def resolve_base(self, base: str, _depth: int = 0) -> tuple[str, int, str]:
        """Return (go_type, extra_ptr, import) for a base type spelling;
        import is "" or "msdtyp"."""
        if _depth > 16:
            return base, 0, ""  # guard against any pathological alias cycle
        if base == "GUID":
            return "guid.GUID", 0, "guid"
        if base in _STR_TYPES or base in self.str_handles:
            return "ndr.WSTR", 1, ""
        if base in _DTYP_STRUCT:
            return f"msdtyp.{base}", 0, "msdtyp"
        if base in _DTYP_PALIAS:
            return f"msdtyp.{_DTYP_PALIAS[base]}", 1, "msdtyp"
        if base in self.palias:
            canon, depth = self.palias[base]
            go, _, imp = self.resolve_base(canon, _depth + 1)
            return go, depth, imp
        if base in self.scalar_alias:
            return self.scalar_alias[base], 0, ""
        if base in _SCALAR:
            return _SCALAR[base], 0, ""
        if base in _DTYP_SCALAR:
            return _DTYP_SCALAR[base], 0, ""
        if base in self.ctx_handles or self.is_local(base):
            return base, 0, ""
        if base.startswith(("struct ", "union ", "enum ")):
            return base.split()[-1], 0, ""
        # MS pointer prefix convention: P<X> / LP<X> is a pointer to X when X
        # resolves to a known type.
        for pfx in ("LP", "P"):
            if base.startswith(pfx) and len(base) > len(pfx):
                inner = base[len(pfx):]
                if (inner in _DTYP_STRUCT or inner in _DTYP_SCALAR or inner in _SCALAR
                        or inner == "GUID" or inner in self.ctx_handles
                        or self.is_local(inner)):
                    go, _, imp = self.resolve_base(inner, _depth + 1)
                    return go, 1, imp
        # Unknown: a type-like name referenced but not defined in this IDL.
        if base[:1].isupper() and " " not in base:
            self.unresolved.add(base)
        return base, 0, ""

    def case_value(self, label: str) -> Optional[int]:
        """Resolve a union case label (a number or an enumerator name) to its
        numeric discriminant value."""
        label = label.strip()
        if re.fullmatch(r"\d+|0[xX][0-9A-Fa-f]+", label):
            return int(label, 0)
        return self.enum_values.get(label)


def _go_field_name(name: str) -> str:
    # Exported (capitalized) so the NDR codec, which skips unexported fields,
    # marshals it. MS-IDL fields are usually PascalCase already; this fixes the
    # occasional lowercase member such as `char data[16]`.
    if not name:
        return "Field"
    return name[0].upper() + name[1:]


def _cap_ref(token: str) -> str:
    """Capitalize a size_is/length_is reference the same way field names are
    exported, so the tag points at the real Go field (the IDL may spell the
    count field lowercase, e.g. size_is(cbData) -> CbData). Numbers and any
    non-identifier expression pass through unchanged."""
    return _go_field_name(token) if re.fullmatch(r"[A-Za-z_]\w*", token) else token


def _eval_const(expr: str) -> Optional[int]:
    """Evaluate a constant integer array bound such as `16` or `(256*2)+4`.
    Returns None if the expression references an identifier (e.g. a size_is
    field) or otherwise isn't a pure integer constant."""
    s = (expr or "").strip()
    if not s or not re.fullmatch(r"[0-9xXa-fA-F+\-*/() ]+", s):
        return None
    try:
        v = eval(s, {"__builtins__": {}}, {})  # guarded: digits/operators only
        return int(v)
    except Exception:
        return None


def _field_decl(res: TypeResolver, f: Field) -> tuple[str, str, str]:
    """Return (go_type, ndr_tag, import) for a struct field (embedded context:
    pointers default to [unique])."""
    go, extra_ptr, imp = res.resolve_base(f.type.base)
    ptr = f.type.ptr + extra_ptr
    size = f.attrs.get("size_is")
    length = f.attrs.get("length_is")
    arr = f.type.array
    # Fixed-size array: a numeric `[N]`/`[(256*2)+4]` suffix with no size_is.
    if arr is not None and arr not in ("*", "") and not isinstance(size, str):
        n = _eval_const(arr)
        if n is not None:
            return f"[{n}]{'*' * ptr}{go}", "", imp
    # An array is either a `[...]`/`[*]` suffix, or a sized pointer (one pointer
    # level is the array indirection and is consumed by the slice).
    suffix_array = arr is not None
    ptr_array = (not suffix_array) and isinstance(size, str) and ptr >= 1
    if suffix_array or ptr_array:
        elem_ptr = ptr if suffix_array else ptr - 1
        tags: list[str] = []
        if elem_ptr >= 1:
            tags.append("elem=unique")
        tags.append("unique")
        # size_is: a literal constant or a sibling field name is usable directly;
        # an arithmetic expression (e.g. MaximumLength/2) is not a field, so fall
        # back to a bare conformant array (the count derives from the slice).
        sized = False
        if isinstance(size, str):
            sn = _norm_attr(size)
            if re.fullmatch(r"\d+|0[xX][0-9A-Fa-f]+|[A-Za-z_]\w*", sn):
                tags.append(f"size_is={_cap_ref(sn)}")
                sized = True
        varying = False
        if isinstance(length, str):
            tags.append("varying")
            varying = True
            ln = _norm_attr(length)
            if re.fullmatch(r"[A-Za-z_]\w*", ln):
                tags.append(f"length_is={_cap_ref(ln)}")
        if not sized and not varying:
            tags.append("conformant")
        elem = ("*" * elem_ptr) + go
        return f"[]{elem}", ",".join(tags), imp
    if ptr == 0:
        return go, "", imp
    if "string" in f.attrs:
        return "*ndr.WSTR", "unique", ""
    return "*" * ptr + go, "unique", imp


def gen_enum(td: Typedef, spec: str) -> str:
    name = td.name
    out: list[str] = []
    w = out.append
    w(f"// {name} is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [{spec}]).")
    w(f"type {name} uint16")
    w("")
    w("const (")
    next_val = 0
    for ename, eval in td.enumerators:
        if eval is not None and str(eval).strip():
            v = _norm_attr(eval)
            w(f"\t{ename} {name} = {v}")
            try:
                next_val = int(v, 0) + 1
            except ValueError:
                next_val += 1
        else:
            w(f"\t{ename} {name} = {next_val}")
            next_val += 1
    w(")")
    w("")
    return "\n".join(out)


def gen_ctx_handle(name: str, spec: str) -> str:
    return (f"// {name} is an RPC context handle: 20 bytes ([MS-RPCE] 2.3.2.2, "
            f"[{spec}]).\ntype {name} [20]byte\n")


def gen_scalar_alias(name: str, go: str, spec: str) -> str:
    return f"// {name} is a scalar typedef ([{spec}]).\ntype {name} {go}\n"


def gen_struct(res: TypeResolver, td: Typedef, spec: str) -> tuple[str, set]:
    name = td.name
    imports: set = set()
    # Hoist inline (nested/anonymous) struct/union/enum members into their own
    # generated sibling types and rewrite the member to reference them.
    nested_defs: list[Typedef] = []
    for f in td.fields:
        if f.nested is not None:
            nname = (f.nested.tag or "").lstrip("_") or f"{name}_{_go_field_name(f.name)}"
            res.kinds[nname] = f.nested.kind
            f.nested.names = [Declarator(name=nname)]
            f.type = TypeRef(base=nname, ptr=f.type.ptr, array=f.type.array)
            nested_defs.append(f.nested)
    out: list[str] = []
    w = out.append
    w(f"// {name} ([{spec}]). Generated by tools/idlgen; verify pointer/array tags.")
    w(f"type {name} struct {{")
    for f in td.fields:
        go, tag, imp = _field_decl(res, f)
        if imp:
            imports.add(imp)
        if tag:
            w(f'\t{_go_field_name(f.name)} {go} `ndr:"{tag}"`')
        else:
            w(f"\t{_go_field_name(f.name)} {go}")
    w("}")
    w("")
    body = "\n".join(out)
    for nd in nested_defs:
        if nd.kind == "union":
            sub, simp = gen_union(res, nd, spec)
        elif nd.kind == "enum":
            sub, simp = gen_enum(nd, spec), set()
        else:
            sub, simp = gen_struct(res, nd, spec)
        body += "\n" + sub
        imports |= simp
    return body, imports


def gen_union(res: TypeResolver, td: Typedef, spec: str) -> tuple[str, set]:
    name = td.name
    imports: set = set()
    out: list[str] = []
    w = out.append
    sw = (td.switch_type or "").strip()
    if sw and sw not in ("unsigned long", "DWORD", "ULONG"):
        disc_go, _, dimp = res.resolve_base(sw)
        if dimp:
            imports.add(dimp)
    else:
        disc_go = "ndr.DWORD"
    w(f"// {name} is a discriminated union ([{spec}]); the discriminant precedes the")
    w("// selected arm ([C706] 14.3.8). Generated by tools/idlgen — verify case values.")
    w(f"type {name} struct {{")
    w(f'\tTag {disc_go} `ndr:"switch"`')
    for f in td.fields:
        go, _, imp = res.resolve_base(f.type.base)
        if imp:
            imports.add(imp)
        typ = ("*" + go) if f.type.ptr >= 1 else go
        if "default" in f.attrs:
            ctag = "default"
        elif isinstance(f.attrs.get("case"), str):
            # The IDL case label may be a number or an enumerator name, and there
            # may be several labels for one arm; resolve the first to its numeric
            # discriminant value (the codec's case= expects a number).
            first = _norm_attr(f.attrs["case"]).split(",")[0]
            num = res.case_value(first)
            ctag = f"case={num}" if num is not None else f"case={first}"
        else:
            ctag = "case=TODO"
        tag = ctag + (",unique" if f.type.ptr >= 1 else "")
        w(f'\t{_go_field_name(f.name)} {typ} `ndr:"{tag}"`')
    w("}")
    w("")
    return "\n".join(out), imports


def gen_structures(iface: Interface, spec: str) -> dict[str, str]:
    """Return {filename: Go source} for each generated structure type."""
    res = TypeResolver(iface)
    files: dict[str, str] = {}
    for td in iface.typedefs:
        name = td.name
        if not name or name in _DTYP_STRUCT:
            continue
        imports: set = set()
        if td.kind == "enum":
            body = gen_enum(td, spec)
        elif td.kind == "alias":
            if name in res.pipes:
                # NDR pipe ([C706] 14.7): a chunked stream of the element type,
                # marshalled with the `ndr:"pipe"` tag at the parameter site.
                body = (f"// {name} is an NDR pipe of {res.pipes[name]} ([C706] 14.7, "
                        f"[{spec}]); transmit with the `ndr:\"pipe\"` tag.\n"
                        f"type {name} []{res.pipes[name]}\n")
            elif name in res.ctx_handles:
                body = gen_ctx_handle(name, spec)
            elif name in res.scalar_alias:
                body = gen_scalar_alias(name, res.scalar_alias[name], spec)
            elif td.base and (res.is_local(td.base) or td.base in _DTYP_STRUCT):
                # typedef <aggregate-or-msdtyp> NewName; -> a Go type alias.
                tgt, _, timp = res.resolve_base(td.base)
                if timp:
                    imports.add(timp)
                body = (f"// {name} is an alias for {td.base} ([{spec}]).\n"
                        f"type {name} = {tgt}\n")
            else:
                continue  # pointer-only alias: resolved at use sites
        elif td.kind == "struct":
            body, imports = gen_struct(res, td, spec)
        elif td.kind == "union":
            body, imports = gen_union(res, td, spec)
        else:
            continue
        # Additional non-pointer declarators on the same typedef are Go aliases.
        for extra in res.extra_names.get(name, []):
            body += f"\n// {extra} is an alias for {name} (same IDL typedef).\n" \
                    f"type {extra} = {name}\n"
        header = ["package structures", ""]
        imps = []
        if "ndr." in body:  # struct tags are strings; only ndr.<Type> needs the import
            imps.append('\t"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"')
        if "msdtyp" in imports or "msdtyp." in body:
            imps.append('\tmsdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"')
        if "guid." in body:
            imps.append('\t"github.com/TheManticoreProject/Manticore/windows/guid"')
        if imps:
            header += ["import ("] + sorted(imps) + [")", ""]
        files[f"{name}.go"] = "\n".join(header) + "\n" + body
    # Emit TODO placeholders for type-like names referenced but never defined in
    # this IDL (e.g. MS-SRVS's SERVER_INFO_100/101, which the spec defines in a
    # different document). This keeps the package compiling with a clear marker.
    defined = set(res.kinds)
    for u in sorted(res.unresolved):
        fname = f"{u}.go"
        if u in defined or fname in files:
            continue
        files[fname] = (
            "package structures\n\n"
            f"// TODO(idlgen): {u} is referenced by this interface but not defined in "
            f"the IDL\n// ([{spec}] may define it elsewhere); fill in its fields by hand.\n"
            f"type {u} struct{{}}\n")
    return files


# --------------------------------------------------------------------------- #
# Phase 4: functions generator (functions/<NN>_<Method>.go)
# --------------------------------------------------------------------------- #

def _param_field(res: TypeResolver, p: Param) -> tuple[str, str, set]:
    """Map a method parameter to a (go_type, ndr_tag, imports) request/response
    field. Applies the top-level pointer rule: a single pointer ([ref]) is the
    inline value; [unique] or a double pointer is a `*T` referent. Top-level
    array parameters use `ref` (not `unique`) so the count is not hoisted ahead
    of a preceding context handle."""
    # An NDR pipe parameter (the IDL declares it as `PIPE_TYPE *`) is the slice
    # type tagged `ndr:"pipe"`; the pointer is just MIDL's pipe-param convention.
    if p.type.base in res.pipes:
        return "structures." + p.type.base, "pipe", set()
    go, extra_ptr, imp = res.resolve_base(p.type.base)
    imports = {imp} if imp else set()
    # Local interface types live in the structures package and must be qualified
    # when referenced from the functions package (unlike inside structures/).
    if go and "." not in go and not go[0].islower() and (
            res.is_local(go) or go in res.ctx_handles):
        go = "structures." + go
    ptr = p.type.ptr + extra_ptr
    size = p.attrs.get("size_is")
    length = p.attrs.get("length_is")
    arr = p.type.array
    is_unique = "unique" in p.attrs

    if "string" in p.attrs and arr is None:
        # A [string] wchar_t* parameter. Its pointer attribute decides the framing:
        # [unique]/[ptr] -> a referent id then the wide string; [ref] (the default for a
        # top-level parameter with no explicit attribute — pointer_default governs only
        # embedded pointers) -> the wide string inline, no referent id.
        if is_unique:
            return "*ndr.WSTR", "unique", set()
        if "ptr" in p.attrs or "full" in p.attrs:
            return "*ndr.WSTR", "ptr", set()
        return "ndr.WSTR", "", set()

    ptr_array = (arr is None) and isinstance(size, str) and ptr >= 1
    if arr is not None or ptr_array:
        if arr is not None and arr not in ("*", "") and not isinstance(size, str):
            n = _eval_const(arr)
            if n is not None:
                return f"[{n}]{'*' * ptr}{go}", "", imports
        elem_ptr = ptr if arr is not None else ptr - 1
        tags: list[str] = []
        if elem_ptr >= 1:
            tags.append("elem=unique")
        tags.append("ref")  # pointer-to-array, no count hoisting before the handle
        sized = False
        if isinstance(size, str):
            sn = _norm_attr(size)
            if re.fullmatch(r"\d+|0[xX][0-9A-Fa-f]+|[A-Za-z_]\w*", sn):
                tags.append(f"size_is={_cap_ref(sn)}")
                sized = True
        varying = False
        if isinstance(length, str):
            tags.append("varying")
            varying = True
            ln = _norm_attr(length)
            if re.fullmatch(r"[A-Za-z_]\w*", ln):
                tags.append(f"length_is={_cap_ref(ln)}")
        if not sized and not varying:
            tags.append("conformant")
        return f"[]{'*' * elem_ptr}{go}", ",".join(tags), imports

    if ptr == 0:
        return go, "", imports
    if ptr == 1 and not is_unique:
        return go, "", imports             # single top-level [ref] -> inline value
    return "*" + go, "unique", imports     # [unique] or double pointer -> referent


_GO_KEYWORDS = {
    "break", "case", "chan", "const", "continue", "default", "defer", "else",
    "fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
    "map", "package", "range", "return", "select", "struct", "switch", "type",
    "var",
}


def _lower_first(s: str) -> str:
    name = (s[0].lower() + s[1:]) if s else "arg"
    return name + "_" if name in _GO_KEYWORDS else name


# Status names allowed (besides success) before a method is treated as failed.
_ENUM_OK = {"StatusMoreEntries", "StatusSomeNotMapped", "StatusNoneMapped"}


def gen_functions(iface: Interface, spec: str, import_base: str) -> dict[str, str]:
    res = TypeResolver(iface)
    short = iface.name
    files: dict[str, str] = {}
    for m in iface.methods:
        if not is_on_the_wire(m):
            continue
        req_fields: list[tuple[str, str, str, Param]] = []   # (FieldName, go, tag, param)
        resp_fields: list[tuple[str, str, str]] = []
        imports: set = set()
        for p in m.params:
            if p.type.base == "handle_t":
                continue  # implicit binding handle, not marshalled
            go, tag, imp = _param_field(res, p)
            imports |= imp
            fname = _go_field_name(p.name)
            if "in" in p.attrs:
                req_fields.append((fname, go, tag, p))
            if "out" in p.attrs:
                resp_fields.append((fname, go, tag))
        lname = m.name[0].lower() + m.name[1:]
        out: list[str] = []
        w = out.append
        # request
        w(f"// {lname}Request carries the [in] parameters of {m.name}.")
        w(f"type {lname}Request struct {{")
        for fname, go, tag, _ in req_fields:
            w(f'\t{fname} {go}' + (f' `ndr:"{tag}"`' if tag else ""))
        w("}")
        w("")
        w(f"func (*{lname}Request) Opnum() uint16 {{ return {short}.Opnum{m.name} }}")
        w("")
        # response
        w(f"// {lname}Response carries the [out] parameters and return value of {m.name}.")
        w(f"type {lname}Response struct {{")
        for fname, go, tag in resp_fields:
            w(f'\t{fname} {go}' + (f' `ndr:"{tag}"`' if tag else ""))
        w('\tStatus ndr.DWORD `ndr:"retval"`')
        w("}")
        w("")
        # exported wrapper with named returns (zero values handled automatically)
        params = ", ".join(f"{_lower_first(fn)} {go}" for fn, go, _, _ in req_fields)
        rets = ", ".join(f"{fn} {go}" for fn, go, _ in resp_fields)
        sig_rets = (rets + ", " if rets else "") + "err error"
        w(f"// {m.name} calls {m.name} (opnum {m.opnum}) ([{spec}] — verify the parameter")
        w("// modeling and status handling).")
        w(f"func {m.name}(rpc ndr.Invoker{', ' + params if params else ''}) ({sig_rets}) {{")
        w(f"\treq := &{lname}Request{{")
        for fn, _, _, _ in req_fields:
            w(f"\t\t{fn}: {_lower_first(fn)},")
        w("\t}")
        w(f"\tvar resp {lname}Response")
        w("\tif err = rpc.Invoke(req, &resp); err != nil {")
        w(f'\t\terr = fmt.Errorf("{m.name}: %w", err)')
        w("\t\treturn")
        w("\t}")
        for fn, _, _ in resp_fields:
            w(f"\t{fn} = resp.{fn}")
        w(f"\tif uint32(resp.Status) != {short}.StatusSuccess {{")
        w(f'\t\terr = fmt.Errorf("{m.name} failed: %s", {short}.StatusString(uint32(resp.Status)))')
        w("\t}")
        w("\treturn")
        w("}")
        w("")
        body = "\n".join(out)

        header = ["package functions", "", "import ("]
        header.append('\t"fmt"')
        header.append("")
        header.append('\t"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"')
        if "msdtyp." in body:
            header.append('\tmsdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"')
        if "guid." in body:
            header.append('\t"github.com/TheManticoreProject/Manticore/windows/guid"')
        header.append(f'\t{short} "{import_base}"')
        if "structures." in body:
            header.append(f'\t"{import_base}/structures"')
        header.append(")")
        header.append("")
        files[f"{m.opnum:02d}_{m.name}.go"] = "\n".join(header) + "\n" + body
    return files


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


# --------------------------------------------------------------------------- #
# Phase 5: whole-interface generation + regression check
# --------------------------------------------------------------------------- #

def gen_interface_tree(iface: Interface, spec: str, pipe: Optional[str],
                       import_base: str) -> dict[str, str]:
    """Build the whole UUID-versioned interface tree as {relpath: gofmt'd source}:
    interface.go, structures/<T>.go, functions/<NN>_<M>.go."""
    files: dict[str, str] = {"interface.go": _gofmt(gen_descriptor(iface, pipe, spec))}
    for n, s in gen_structures(iface, spec).items():
        files[f"structures/{n}"] = _gofmt(s)
    for n, s in gen_functions(iface, spec, import_base).items():
        files[f"functions/{n}"] = _gofmt(s)
    return files


def find_module(start: str) -> tuple[Optional[str], Optional[str]]:
    """Walk up from `start` to the nearest go.mod; return (module path, root dir)."""
    import os
    d = os.path.abspath(start)
    while True:
        gomod = os.path.join(d, "go.mod")
        if os.path.isfile(gomod):
            with open(gomod) as fh:
                for line in fh:
                    if line.startswith("module "):
                        return line.split()[1], d
            return None, None
        parent = os.path.dirname(d)
        if parent == d:
            return None, None
        d = parent


def interface_paths(iface: Interface, out_root: str) -> tuple[str, str]:
    """Return (target_dir, import_base) for an interface under out_root."""
    import os
    parts = parse_uuid(iface.attrs.get("uuid", ""))
    if parts is None:
        raise ParseError(f"interface {iface.name}: missing/invalid uuid")
    maj, min = parse_version(iface.attrs.get("version", "0.0"))
    target = os.path.join(out_root, "-".join(parts), f"{maj}.{min}")
    module, root = find_module(out_root)
    if module and root:
        rel = os.path.relpath(os.path.abspath(target), root).replace(os.sep, "/")
        import_base = f"{module}/{rel}"
    else:
        import_base = "/".join(["MODULE", "-".join(parts), f"{maj}.{min}"])  # TODO: set module
    return target, import_base


def _code_only(text: str) -> list[str]:
    """Strip comments and blank lines and normalize whitespace for comparison."""
    out = []
    for line in text.splitlines():
        s = line.strip()
        if not s or s.startswith("//"):
            continue
        out.append(re.sub(r"\s+", " ", s))
    return out


def main(argv: list[str]) -> int:
    ap = argparse.ArgumentParser(prog="idlgen", description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)
    p = sub.add_parser("parse", help="parse an .idl and print a summary or JSON AST")
    p.add_argument("idl", help="path to the .idl file")
    p.add_argument("--json", action="store_true", help="dump the full AST as JSON")
    p.add_argument("--quiet", action="store_true", help="only report parse errors")

    ft = sub.add_parser("fetch", help="download the IDL from an MS Open Specs 'Appendix A: Full IDL' page")
    ft.add_argument("url", help="learn.microsoft.com 'Appendix A: Full IDL' page URL")
    ft.add_argument("--out", default=None, help="output .idl file (default: stdout)")

    g = sub.add_parser("gen-descriptor", help="generate interface.go from an .idl")
    g.add_argument("idl", help="path to the .idl file")
    g.add_argument("--pipe", default=None, help="named pipe (default: \\<interface>)")
    g.add_argument("--spec", default="MS-XXX", help="spec tag for doc comments, e.g. MS-SAMR")
    g.add_argument("--out", default=None, help="output file (default: stdout)")
    g.add_argument("--interface", default=None, help="interface name if the IDL has several")

    st = sub.add_parser("gen-structures", help="generate structures/*.go from an .idl")
    st.add_argument("idl", help="path to the .idl file")
    st.add_argument("--spec", default="MS-XXX", help="spec tag for doc comments")
    st.add_argument("--out", required=True, help="output directory for the structures package")
    st.add_argument("--interface", default=None, help="interface name if the IDL has several")

    fn = sub.add_parser("gen-functions", help="generate functions/*.go from an .idl")
    fn.add_argument("idl", help="path to the .idl file")
    fn.add_argument("--spec", default="MS-XXX", help="spec tag for doc comments")
    fn.add_argument("--out", required=True, help="output directory for the functions package")
    fn.add_argument("--import-base", required=True, dest="import_base",
                    help="module import path of the interface (the descriptor package)")
    fn.add_argument("--interface", default=None, help="interface name if the IDL has several")

    ge = sub.add_parser("generate", help="generate the whole interface tree from an .idl")
    ge.add_argument("idl", help="path to the .idl file")
    ge.add_argument("--out-root", required=True, dest="out_root",
                    help="interfaces root dir, e.g. network/dcerpc/interfaces")
    ge.add_argument("--spec", default="MS-XXX", help="spec tag for doc comments")
    ge.add_argument("--pipe", default=None, help="named pipe (default: \\<interface>)")
    ge.add_argument("--interface", default=None, help="interface name if the IDL has several")

    ck = sub.add_parser("check", help="diff the generated tree against the committed one")
    ck.add_argument("idl", help="path to the .idl file")
    ck.add_argument("--out-root", required=True, dest="out_root",
                    help="interfaces root dir the committed tree lives under")
    ck.add_argument("--spec", default="MS-XXX", help="spec tag for doc comments")
    ck.add_argument("--pipe", default=None, help="named pipe (default: \\<interface>)")
    ck.add_argument("--interface", default=None, help="interface name if the IDL has several")
    ck.add_argument("--strict", action="store_true",
                    help="exit non-zero if any code differs (not just on missing files)")

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

    if args.cmd == "fetch":
        try:
            idl = fetch_idl(args.url)
        except (ParseError, OSError) as e:
            print(f"FETCH ERROR: {e}", file=sys.stderr)
            return 1
        if args.out:
            with open(args.out, "w") as fh:
                fh.write(idl)
            print(f"wrote {args.out}", file=sys.stderr)
        else:
            sys.stdout.write(idl)
        # Best-effort sanity parse + a suggested generate command.
        try:
            ifaces = Parser(tokenize(idl), filename=args.url).parse()
            spec = spec_tag_from_url(args.url) or "MS-XXX"
            for iface in ifaces:
                wire = sum(1 for m in iface.methods if is_on_the_wire(m))
                print(f"  parsed interface {iface.name}: {wire} on-the-wire methods, "
                      f"{len(iface.typedefs)} typedefs", file=sys.stderr)
            if args.out and ifaces:
                print(f"  next: python3 {sys.argv[0]} generate {args.out} "
                      f"--out-root network/dcerpc/interfaces --spec {spec} "
                      f"--pipe '\\<pipe>'", file=sys.stderr)
        except (LexError, ParseError) as e:
            print(f"  warning: fetched IDL did not parse cleanly: {e}", file=sys.stderr)
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

    if args.cmd == "gen-structures":
        import os
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
        files = gen_structures(ifaces[0], args.spec)
        os.makedirs(args.out, exist_ok=True)
        for fname, src in files.items():
            with open(os.path.join(args.out, fname), "w") as fh:
                fh.write(_gofmt(src))
        print(f"wrote {len(files)} files to {args.out}", file=sys.stderr)
        return 0

    if args.cmd == "gen-functions":
        import os
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
        files = gen_functions(ifaces[0], args.spec, args.import_base)
        os.makedirs(args.out, exist_ok=True)
        for fname, src in files.items():
            with open(os.path.join(args.out, fname), "w") as fh:
                fh.write(_gofmt(src))
        print(f"wrote {len(files)} files to {args.out}", file=sys.stderr)
        return 0

    if args.cmd in ("generate", "check"):
        import os
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
        iface = ifaces[0]
        try:
            target, import_base = interface_paths(iface, args.out_root)
            files = gen_interface_tree(iface, args.spec, args.pipe, import_base)
        except ParseError as e:
            print(f"ERROR: {e}", file=sys.stderr)
            return 1

        if args.cmd == "generate":
            for rel, src in files.items():
                path = os.path.join(target, rel)
                os.makedirs(os.path.dirname(path), exist_ok=True)
                with open(path, "w") as fh:
                    fh.write(src)
            todos = sum(src.count("TODO(idlgen)") for src in files.values())
            print(f"generated {len(files)} files under {target}", file=sys.stderr)
            print(f"import base: {import_base}", file=sys.stderr)
            print(f"{todos} TODO(idlgen) markers to review", file=sys.stderr)
            return 0

        # check: code-only diff of generated vs committed tree.
        same = differ = gen_only = 0
        differing: list[str] = []
        for rel, src in sorted(files.items()):
            path = os.path.join(target, rel)
            if not os.path.isfile(path):
                gen_only += 1
                differing.append(f"{rel} (generated-only)")
                continue
            with open(path) as fh:
                committed = fh.read()
            if _code_only(committed) == _code_only(src):
                same += 1
            else:
                differ += 1
                differing.append(rel)
        committed_only = 0
        for dirpath, _, names in os.walk(target):
            for nm in names:
                if not nm.endswith(".go") or nm.endswith("_test.go"):
                    continue
                rel = os.path.relpath(os.path.join(dirpath, nm), target).replace(os.sep, "/")
                if rel not in files:
                    committed_only += 1
        print(f"check {iface.name} against {target}:")
        print(f"  code-identical: {same}")
        print(f"  differing:      {differ}")
        print(f"  generated-only: {gen_only}")
        print(f"  committed-only: {committed_only} (hand-added; not generated)")
        if differing:
            print("  files differing/new:")
            for f in differing:
                print(f"    - {f}")
        if gen_only or (args.strict and differ):
            return 1
        return 0
    return 2


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
