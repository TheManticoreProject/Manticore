package ibmi_test

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// importGraph holds the module-relative import edges of every package found
// under a module root, split between production and _test.go files.
type importGraph struct {
	module string
	prod   map[string]map[string]bool
	test   map[string]map[string]bool
}

// loadImportGraph parses only the import clauses of every .go file below root.
// Directories named testdata, vendor or starting with "." or "_" are skipped,
// exactly as the go tool skips them. Imports outside the module are ignored.
func loadImportGraph(root string) (*importGraph, error) {
	module, err := readModulePath(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, err
	}
	g := &importGraph{
		module: module,
		prod:   map[string]map[string]bool{},
		test:   map[string]map[string]bool{},
	}
	fset := token.NewFileSet()
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		pkg := filepath.ToSlash(rel)
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		edges := g.prod
		if strings.HasSuffix(name, "_test.go") {
			edges = g.test
		}
		if edges[pkg] == nil {
			edges[pkg] = map[string]bool{}
		}
		for _, spec := range file.Imports {
			target, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			if target == module {
				edges[pkg]["."] = true
			} else if strings.HasPrefix(target, module+"/") {
				edges[pkg][strings.TrimPrefix(target, module+"/")] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

func readModulePath(goMod string) (string, error) {
	f, err := os.Open(goMod)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "module ")), `"`), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%s: no module directive", goMod)
}

// closure returns every module package reachable from pkg through
// production imports, excluding pkg itself.
func (g *importGraph) closure(pkg string) []string {
	seen := map[string]bool{pkg: true}
	queue := []string{pkg}
	var out []string
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for next := range g.prod[cur] {
			if !seen[next] {
				seen[next] = true
				out = append(out, next)
				queue = append(queue, next)
			}
		}
	}
	sort.Strings(out)
	return out
}

// within reports whether pkg is root or one of its descendants.
func within(pkg, root string) bool {
	return pkg == root || strings.HasPrefix(pkg, root+"/")
}

// testInfra is the scripted fake-server package; only _test.go files may
// import it.
const testInfra = "network/ibmi/internal/hosttest"

// networkMayUse lists the only root-ibmi packages network/ibmi may depend on.
var networkMayUse = map[string]bool{
	"ibmi":       true,
	"ibmi/ccsid": true,
	"ibmi/rpc":   true,
}

// checkIBMiBoundaries returns every violation of the IBM i dependency rules
// documented in package ibmi and package network/ibmi.
func checkIBMiBoundaries(g *importGraph) []string {
	var violations []string
	add := func(rule, from, to string) {
		violations = append(violations, fmt.Sprintf("%s: %s -> %s", rule, from, to))
	}
	pkgs := map[string]bool{}
	for p := range g.prod {
		pkgs[p] = true
	}
	for p := range g.test {
		pkgs[p] = true
	}

	for pkg := range pkgs {
		reach := g.closure(pkg)

		if within(pkg, "ibmi") {
			for _, dep := range reach {
				if within(dep, "network/ibmi") {
					add("platform-imports-network", pkg, dep)
				}
			}
		}

		if within(pkg, "ibmi/rpc") {
			for dep := range g.test[pkg] {
				if within(dep, "network/ibmi") {
					add("rpc-test-imports-network", pkg, dep)
				}
			}
		}

		if within(pkg, "network/ibmi") {
			for _, dep := range reach {
				if within(dep, "ibmi") && !networkMayUse[dep] {
					add("network-imports-concrete-api", pkg, dep)
				}
			}
			if !within(pkg, "network/ibmi/client") {
				for _, dep := range reach {
					if within(dep, "network/ibmi/client") {
						add("lower-layer-imports-client", pkg, dep)
					}
				}
			}
		}

		for _, root := range []string{"ibmi", "ibmi/rpc"} {
			if pkg != root {
				continue
			}
			for dep := range g.prod[pkg] {
				if dep != root && within(dep, root) {
					add("root-imports-child", pkg, dep)
				}
			}
		}

		if !within(pkg, testInfra) {
			for dep := range g.prod[pkg] {
				if within(dep, testInfra) {
					add("production-imports-testinfra", pkg, dep)
				}
			}
		}
	}
	sort.Strings(violations)
	return violations
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the test directory")
		}
		dir = parent
	}
}

// TestIBMiImportBoundaries enforces the IBM i dependency direction on the
// real module.
func TestIBMiImportBoundaries(t *testing.T) {
	g, err := loadImportGraph(findModuleRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if g.module != "github.com/TheManticoreProject/Manticore" {
		t.Fatalf("unexpected module path %q", g.module)
	}
	if _, ok := g.prod["ibmi"]; !ok {
		t.Fatal("package ibmi was not discovered; the checker is not walking the module")
	}
	for _, v := range checkIBMiBoundaries(g) {
		t.Error(v)
	}
}

// writeForbiddenEdgeModule writes a throwaway module whose packages contain
// deliberately forbidden IBM i import edges, plus allowed test-only edges.
// Each entry maps a module-relative file to the packages it imports.
func writeForbiddenEdgeModule(t *testing.T) string {
	t.Helper()
	const module = "example.com/fixture"
	files := map[string][]string{
		"ibmi/root.go":                           {"ibmi/rpc"},
		"ibmi/rpc/rpc.go":                        {"ibmi"},
		"ibmi/rpc/rpc_test.go":                   {"network/ibmi/hostserver"},
		"ibmi/helper/helper.go":                  {"network/ibmi/hostserver"},
		"ibmi/rpc/qgy/qgyolaus/qgyolaus.go":      {"ibmi/rpc", "ibmi/helper"},
		"ibmi/api/users/users_test.go":           {"ibmi/rpc/qgy/qgyolaus"},
		"network/ibmi/hostserver/hostserver.go":  {"network/ibmi/internal/hosttest"},
		"network/ibmi/rmtcmd/rmtcmd.go":          {"ibmi/rpc", "network/ibmi/client"},
		"network/ibmi/rmtcmd/rmtcmd_test.go":     {"network/ibmi/client", "network/ibmi/internal/hosttest"},
		"network/ibmi/internal/hosttest/fake.go": nil,
		"network/ibmi/client/client.go":          {"ibmi/rpc/qgy/qgyolaus"},
	}
	root := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module "+module+"\n\ngo 1.24.0\n")
	for rel, imports := range files {
		var b strings.Builder
		fmt.Fprintf(&b, "package %s\n\n", filepath.Base(filepath.Dir(filepath.FromSlash(rel))))
		for _, imp := range imports {
			fmt.Fprintf(&b, "import _ %q\n", module+"/"+imp)
		}
		write(rel, b.String())
	}
	return root
}

// TestIBMiImportBoundariesRejectForbiddenEdges proves the checker fails on a
// deliberately broken module and accepts its allowed test-only edges.
func TestIBMiImportBoundariesRejectForbiddenEdges(t *testing.T) {
	g, err := loadImportGraph(writeForbiddenEdgeModule(t))
	if err != nil {
		t.Fatal(err)
	}
	got := checkIBMiBoundaries(g)
	want := []string{
		"lower-layer-imports-client: network/ibmi/rmtcmd -> network/ibmi/client",
		"network-imports-concrete-api: network/ibmi/client -> ibmi/helper",
		"network-imports-concrete-api: network/ibmi/client -> ibmi/rpc/qgy/qgyolaus",
		"network-imports-concrete-api: network/ibmi/rmtcmd -> ibmi/helper",
		"network-imports-concrete-api: network/ibmi/rmtcmd -> ibmi/rpc/qgy/qgyolaus",
		"platform-imports-network: ibmi/helper -> network/ibmi/hostserver",
		"platform-imports-network: ibmi/helper -> network/ibmi/internal/hosttest",
		"platform-imports-network: ibmi/rpc/qgy/qgyolaus -> network/ibmi/hostserver",
		"platform-imports-network: ibmi/rpc/qgy/qgyolaus -> network/ibmi/internal/hosttest",
		"production-imports-testinfra: network/ibmi/hostserver -> network/ibmi/internal/hosttest",
		"root-imports-child: ibmi -> ibmi/rpc",
		"rpc-test-imports-network: ibmi/rpc -> network/ibmi/hostserver",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations mismatch\n got: %q\nwant: %q", got, want)
	}
}
