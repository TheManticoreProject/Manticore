#!/usr/bin/env python3
"""Extract an [MS-ERREF] error code table into the TSV that errgen reads.

The tables are published only as HTML on Microsoft Learn, so each one is
extracted once and the result committed next to this script. That keeps
generation offline and makes a change to a specification show up as a diff in the
TSV rather than as a change in this parser.

Every [MS-ERREF] code table has the same two-column shape — a cell holding the
hexadecimal value and the symbolic name, and a cell holding the description — so
one parser serves all of them.

Usage:
    python3 extract_ms_erref.py --table win32    > ms-erref-2.2-win32.tsv
    python3 extract_ms_erref.py --table ntstatus > ms-erref-2.3.1-ntstatus.tsv
    python3 extract_ms_erref.py --table win32 page.html   # extract a local copy

--min-rows guards against a page whose layout has changed enough to yield a
short table; each known table carries its own floor, set just below the count
observed when it was last extracted.
"""

import argparse
import html.parser
import re
import urllib.request

# The [MS-ERREF] tables this script knows how to extract, with the row count
# below which extraction is treated as a layout change rather than a result.
TABLES = {
    "win32": {
        "url": "https://learn.microsoft.com/en-us/openspecs/windows_protocols/"
        "ms-erref/18d8fbe8-a967-4f1c-ae50-99ca8e491d2d",
        "section": "2.2 Win32 Error Codes",
        "min_rows": 2700,
    },
    "ntstatus": {
        "url": "https://learn.microsoft.com/en-us/openspecs/windows_protocols/"
        "ms-erref/596a1078-e883-4972-9bbc-49e60bebca55",
        "section": "2.3.1 NTSTATUS Values",
        "min_rows": 1790,
    },
}

# The value cell holds "0xXXXXXXXX" and the symbolic name, normally in separate
# paragraphs but occasionally in one.
CODE_ONLY = re.compile(r"^0x([0-9A-Fa-f]{8})$")
CODE_AND_NAME = re.compile(r"^0x([0-9A-Fa-f]{8})\s+([A-Za-z_][A-Za-z0-9_]*)$")
NAME = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*$")


class TableParser(html.parser.HTMLParser):
    """Collect the paragraph texts of every two-column row of the code table."""

    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.rows = []
        self._in_table = 0
        self._cells = None
        self._paragraphs = None
        self._text = None

    def handle_starttag(self, tag, attrs):
        if tag == "table":
            self._in_table += 1
        elif tag == "tr" and self._in_table:
            self._cells = []
        elif tag in ("td", "th") and self._cells is not None:
            self._paragraphs = []
        elif tag == "p" and self._paragraphs is not None:
            self._text = []

    def handle_endtag(self, tag):
        if tag == "p" and self._text is not None:
            text = re.sub(r"\s+", " ", "".join(self._text)).strip()
            if text:
                self._paragraphs.append(text)
            self._text = None
        elif tag in ("td", "th") and self._paragraphs is not None:
            self._cells.append(self._paragraphs)
            self._paragraphs = None
        elif tag == "tr" and self._cells is not None:
            if len(self._cells) == 2:
                self.rows.append(self._cells)
            self._cells = None
        elif tag == "table" and self._in_table:
            self._in_table -= 1

    def handle_data(self, data):
        if self._text is not None:
            self._text.append(data)


def extract(page, min_rows):
    parser = TableParser()
    parser.feed(page)

    out = []
    for value_cell, description_cell in parser.rows:
        if not value_cell:
            continue

        match = CODE_ONLY.match(value_cell[0])
        if match:
            code, names = match.group(1), [p for p in value_cell[1:] if NAME.match(p)]
        else:
            match = CODE_AND_NAME.match(value_cell[0])
            if not match:
                continue  # header row, or a cell that is not a code
            code, names = match.group(1), [match.group(2)]

        description = " ".join(description_cell)
        if not names or not description:
            raise SystemExit(f"row 0x{code} is missing a name or a description")

        for name in names:
            out.append((int(code, 16), name, description))

    if len(out) < min_rows:
        raise SystemExit(
            f"only {len(out)} rows extracted, want at least {min_rows}; "
            "the page layout has changed"
        )
    return out


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--table", required=True, choices=sorted(TABLES), help="which table to extract")
    parser.add_argument("--min-rows", type=int, help="override the table's row floor")
    parser.add_argument("page", nargs="?", help="a local copy of the page, instead of fetching it")
    args = parser.parse_args()

    table = TABLES[args.table]
    min_rows = args.min_rows if args.min_rows is not None else table["min_rows"]

    if args.page:
        page = open(args.page, encoding="utf-8").read()
    else:
        with urllib.request.urlopen(table["url"]) as response:
            page = response.read().decode("utf-8")

    rows = extract(page, min_rows)

    print(f"# [MS-ERREF] {table['section']}, extracted by extract_ms_erref.py --table {args.table}.")
    print(f"# Source: {table['url']}")
    print("# Columns: code (hex, no prefix), symbolic name, description.")
    print("# Rows are in specification order; a code may carry more than one name.")
    for code, name, description in rows:
        print(f"{code:08X}\t{name}\t{description}")


if __name__ == "__main__":
    main()
