// Command errgen generates the Go form of a Windows error-code table from the
// tab-separated extract of its specification.
//
// The extract is committed next to this program, so generation needs no network
// access and a change in the specification appears as a diff in the TSV rather
// than in generated Go. See extract_ms_erref.py for how the TSV is produced.
//
// Usage:
//
//	go run ./windows/errors/errgen -in <table.tsv> -out <dir> -package win32 -type WIN32_ERROR
//
// The Go constant name for a row is derived from its symbolic name by stripping
// -strip-name-prefix and prepending -add-name-prefix, which is how the [MS-ERREF]
// NTSTATUS names become the NT_STATUS_* constants this tree already uses. Where
// the two differ, both resolve through FromName.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// record is one row of the TSV: a code, one of its symbolic names, and the
// description the specification gives for it.
type record struct {
	Code        uint32
	Name        string // the symbolic name as the specification writes it
	GoName      string // the Go constant name, which may differ from Name
	Description string
	Order       int // position in the file, so specification order can be preserved
}

func main() {
	in := flag.String("in", "", "path to the tab-separated table to read")
	out := flag.String("out", "", "directory to write the generated Go files into")
	pkg := flag.String("package", "", "package name for the generated files")
	typeName := flag.String("type", "", "name of the code type to generate")
	source := flag.String("source", "", "specification URL to record in the generated files")
	stripPrefix := flag.String("strip-name-prefix", "", "prefix to strip from a symbolic name before deriving its Go constant name")
	addPrefix := flag.String("add-name-prefix", "", "prefix to prepend when deriving a Go constant name")
	flag.Parse()

	for name, value := range map[string]string{"in": *in, "out": *out, "package": *pkg, "type": *typeName} {
		if value == "" {
			fatalf("-%s is required", name)
		}
	}

	records, err := readTable(*in, *stripPrefix, *addPrefix)
	if err != nil {
		fatalf("reading %s: %v", *in, err)
	}

	base := filepath.Base(*in)
	codes := renderCodes(records, *pkg, *typeName, base, *source)
	table := renderTable(records, *pkg, *typeName, base, *source)

	if err := writeFormatted(filepath.Join(*out, "codes.go"), codes); err != nil {
		fatalf("%v", err)
	}
	if err := writeFormatted(filepath.Join(*out, "table.go"), table); err != nil {
		fatalf("%v", err)
	}

	fmt.Fprintf(os.Stderr, "errgen: %d names over %d codes\n", len(records), countCodes(records))
}

// readTable parses the TSV, rejecting anything malformed rather than skipping it
// so that a change in the extract cannot silently drop codes.
func readTable(path, stripPrefix, addPrefix string) ([]record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []record
	seenNames := make(map[string]uint32)
	seenGoNames := make(map[string]uint32)

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}

		fields := strings.Split(text, "\t")
		if len(fields) != 3 {
			return nil, fmt.Errorf("line %d: got %d fields, want 3", line, len(fields))
		}

		code, err := strconv.ParseUint(fields[0], 16, 32)
		if err != nil {
			return nil, fmt.Errorf("line %d: %q is not a 32-bit hex code: %v", line, fields[0], err)
		}
		name, description := fields[1], strings.TrimSpace(fields[2])
		if !isIdentifier(name) {
			return nil, fmt.Errorf("line %d: %q is not a Go identifier", line, name)
		}
		if description == "" {
			return nil, fmt.Errorf("line %d: %s has no description", line, name)
		}
		if previous, duplicate := seenNames[name]; duplicate {
			return nil, fmt.Errorf("line %d: %s is already defined as 0x%08X", line, name, previous)
		}
		seenNames[name] = uint32(code)

		goName := addPrefix + strings.TrimPrefix(name, stripPrefix)
		if !isIdentifier(goName) {
			return nil, fmt.Errorf("line %d: %q is not a Go identifier", line, goName)
		}
		if previous, duplicate := seenGoNames[goName]; duplicate {
			return nil, fmt.Errorf("line %d: the Go name %s is already defined as 0x%08X", line, goName, previous)
		}
		seenGoNames[goName] = uint32(code)

		records = append(records, record{
			Code:        uint32(code),
			Name:        name,
			GoName:      goName,
			Description: description,
			Order:       len(records),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no rows found")
	}

	// Emit ascending by code, and by specification order within a code, so the
	// first name a code carries stays its canonical one.
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Code != records[j].Code {
			return records[i].Code < records[j].Code
		}
		return records[i].Order < records[j].Order
	})
	return records, nil
}

func renderCodes(records []record, pkg, typeName, base, source string) string {
	var b strings.Builder
	writeHeader(&b, pkg, base, source)

	fmt.Fprintf(&b, "// Every %s defined by the specification. A code carries more than one name\n", typeName)
	b.WriteString("// where the specification gives it more than one; the names are aliases of the\n")
	b.WriteString("// same value.\nconst (\n")
	for _, r := range records {
		fmt.Fprintf(&b, "\t// %s\n", r.Description)
		if r.GoName != r.Name {
			fmt.Fprintf(&b, "\t//\n\t// The specification names this %s.\n", r.Name)
		}
		fmt.Fprintf(&b, "\t%s %s = 0x%08X\n", r.GoName, typeName, r.Code)
	}
	b.WriteString(")\n")
	return b.String()
}

func renderTable(records []record, pkg, typeName, base, source string) string {
	var b strings.Builder
	writeHeader(&b, pkg, base, source)

	fmt.Fprintf(&b, "// table is the specification's row for each %s, keyed by code. Where a code\n", typeName)
	b.WriteString("// carries several names, the first the specification lists is the canonical one\n")
	b.WriteString("// held here; nameToCode resolves the rest.\n")
	fmt.Fprintf(&b, "var table = map[%s]Entry{\n", typeName)
	seen := make(map[uint32]bool, len(records))
	for _, r := range records {
		if seen[r.Code] {
			continue
		}
		seen[r.Code] = true
		fmt.Fprintf(&b, "\t%s: {Name: %q, Description: %q},\n", r.GoName, r.GoName, r.Description)
	}
	b.WriteString("}\n\n")

	b.WriteString("// nameToCode resolves every symbolic name, canonical or alias, to its code.\n")
	b.WriteString("// Where the Go constant name differs from the name the specification uses,\n")
	b.WriteString("// both are present, so a name copied out of the specification resolves too.\n")
	fmt.Fprintf(&b, "var nameToCode = map[string]%s{\n", typeName)
	for _, r := range records {
		fmt.Fprintf(&b, "\t%q: %s,\n", r.GoName, r.GoName)
		if r.GoName != r.Name {
			fmt.Fprintf(&b, "\t%q: %s,\n", r.Name, r.GoName)
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func writeHeader(b *strings.Builder, pkg, base, source string) {
	fmt.Fprintf(b, "// Code generated by windows/errors/errgen from %s. DO NOT EDIT.\n", base)
	if source != "" {
		fmt.Fprintf(b, "//\n// Source: %s\n", source)
	}
	fmt.Fprintf(b, "\npackage %s\n\n", pkg)
}

func writeFormatted(path, content string) error {
	formatted, err := format.Source([]byte(content))
	if err != nil {
		return fmt.Errorf("formatting %s: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0o644)
}

func countCodes(records []record) int {
	seen := make(map[uint32]bool, len(records))
	for _, r := range records {
		seen[r.Code] = true
	}
	return len(seen)
}

func isIdentifier(s string) bool {
	for i, r := range s {
		switch {
		case r == '_', r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return s != ""
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "errgen: "+format+"\n", args...)
	os.Exit(1)
}
