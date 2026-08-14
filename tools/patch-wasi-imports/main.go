package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	wasiModule = "wasi_snapshot_preview1"
	enosys     = 52
	ebadf      = 8
)

var allowedWASIImports = map[string]bool{
	"sched_yield":       true,
	"proc_exit":         true,
	"args_get":          true,
	"args_sizes_get":    true,
	"clock_time_get":    true,
	"environ_get":       true,
	"environ_sizes_get": true,
	"fd_write":          true,
	"random_get":        true,
	"poll_oneoff":       true,
	"fd_close":          true,
}

var specialReturnByImport = map[string]int{
	"fd_fdstat_get":       0,
	"fd_fdstat_set_flags": 0,
	"fd_prestat_get":      ebadf,
}

type importFunc struct {
	Index     int
	Module    string
	Name      string
	TypeIndex int
	Line      string
}

func main() {
	in := flag.String("in", "main.raw.wasm", "input WASM path")
	out := flag.String("out", "main.wasm", "output WASM path")
	keepWat := flag.Bool("keep-wat", false, "keep intermediate WAT files next to the output")
	flag.Parse()

	if err := run(*in, *out, *keepWat); err != nil {
		fmt.Fprintf(os.Stderr, "patch-wasi-imports: %v\n", err)
		os.Exit(1)
	}
}

func run(in, out string, keepWat bool) error {
	watBytes, err := wasmToolsOutput("print", in)
	if err != nil {
		return err
	}

	patched, removed, err := patchWAT(string(watBytes))
	if err != nil {
		return err
	}
	if len(removed) == 0 {
		return fmt.Errorf("no unsupported WASI imports found to patch; refusing to produce ambiguous output")
	}

	watPath := out + ".wat"
	if err := os.WriteFile(watPath, []byte(patched), 0o600); err != nil {
		return fmt.Errorf("write patched WAT: %w", err)
	}
	if !keepWat {
		defer os.Remove(watPath)
	}

	if _, err := wasmToolsOutput("parse", watPath, "-o", out); err != nil {
		return err
	}
	if _, err := wasmToolsOutput("validate", out); err != nil {
		return err
	}

	imports, err := wasmToolsOutput("print", out)
	if err != nil {
		return err
	}
	if remaining := unsupportedImportsInWAT(string(imports)); len(remaining) > 0 {
		return fmt.Errorf("unsupported WASI imports remain after patching: %s", strings.Join(remaining, ", "))
	}

	fmt.Printf("Patched %d unsupported WASI import(s): %s\n", len(removed), strings.Join(removed, ", "))
	return nil
}

func patchWAT(wat string) (string, []string, error) {
	typeSignatures, err := parseTypeSignatures(wat)
	if err != nil {
		return "", nil, err
	}
	imports, err := parseImports(wat)
	if err != nil {
		return "", nil, err
	}

	removeByIndex := make(map[int]importFunc)
	var removedNames []string
	for _, imp := range imports {
		if imp.Module == wasiModule && !allowedWASIImports[imp.Name] {
			removeByIndex[imp.Index] = imp
			removedNames = append(removedNames, fmt.Sprintf("%s#%d", imp.Name, imp.Index))
		}
	}
	if len(removeByIndex) == 0 {
		return wat, nil, nil
	}

	var removedIndexes []int
	for idx := range removeByIndex {
		removedIndexes = append(removedIndexes, idx)
	}
	sort.Ints(removedIndexes)

	oldImportCount := len(imports)
	newImportCount := oldImportCount - len(removedIndexes)

	stubIndexByOldIndex := make(map[int]int)
	for i, oldIndex := range removedIndexes {
		stubIndexByOldIndex[oldIndex] = newImportCount + i
	}

	callIndexMap := make(map[int]int)
	for _, imp := range imports {
		if newIndex, ok := stubIndexByOldIndex[imp.Index]; ok {
			callIndexMap[imp.Index] = newIndex
			continue
		}
		removedBefore := countLessThan(removedIndexes, imp.Index)
		callIndexMap[imp.Index] = imp.Index - removedBefore
	}

	var stubs strings.Builder
	for _, oldIndex := range removedIndexes {
		imp := removeByIndex[oldIndex]
		sig, ok := typeSignatures[imp.TypeIndex]
		if !ok {
			return "", nil, fmt.Errorf("missing signature for type %d used by %s", imp.TypeIndex, imp.Name)
		}
		if !strings.Contains(sig, "(result i32)") {
			return "", nil, fmt.Errorf("unsupported non-i32-result WASI import %s with signature %q", imp.Name, sig)
		}
		ret := enosys
		if special, ok := specialReturnByImport[imp.Name]; ok {
			ret = special
		}
		fmt.Fprintf(&stubs, "  (func $__wasi_stub_%s_%d (type %d) %s\n    i32.const %d)\n", sanitizeName(imp.Name), oldIndex, imp.TypeIndex, sig, ret)
	}

	lines := strings.SplitAfter(wat, "\n")
	var filtered strings.Builder
	for _, line := range lines {
		if imp, ok := parseImportLine(line); ok {
			if _, remove := removeByIndex[imp.Index]; remove {
				continue
			}
		}
		filtered.WriteString(line)
	}

	patched := filtered.String()
	marker := "  (table (;0;)"
	if !strings.Contains(patched, marker) {
		return "", nil, fmt.Errorf("cannot find insertion point %q", marker)
	}
	patched = strings.Replace(patched, marker, stubs.String()+marker, 1)
	patched = rewriteNumericCalls(patched, callIndexMap)

	return patched, removedNames, nil
}

func parseTypeSignatures(wat string) (map[int]string, error) {
	re := regexp.MustCompile(`(?m)^  \(type \(;([0-9]+);\) \(func(.*?)\)\)$`)
	matches := re.FindAllStringSubmatch(wat, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no function type signatures found")
	}
	result := make(map[int]string, len(matches))
	for _, m := range matches {
		idx, err := strconv.Atoi(m[1])
		if err != nil {
			return nil, err
		}
		result[idx] = strings.TrimSpace(m[2])
	}
	return result, nil
}

func parseImports(wat string) ([]importFunc, error) {
	var imports []importFunc
	for _, line := range strings.SplitAfter(wat, "\n") {
		imp, ok := parseImportLine(line)
		if !ok {
			continue
		}
		imports = append(imports, imp)
	}
	if len(imports) == 0 {
		return nil, fmt.Errorf("no function imports found")
	}
	return imports, nil
}

func parseImportLine(line string) (importFunc, bool) {
	re := regexp.MustCompile(`^  \(import "([^"]+)" "([^"]+)" \(func \(;([0-9]+);\) \(type ([0-9]+)\)\)\)\s*$`)
	m := re.FindStringSubmatch(line)
	if m == nil {
		return importFunc{}, false
	}
	idx, err := strconv.Atoi(m[3])
	if err != nil {
		return importFunc{}, false
	}
	typeIdx, err := strconv.Atoi(m[4])
	if err != nil {
		return importFunc{}, false
	}
	return importFunc{Index: idx, Module: m[1], Name: m[2], TypeIndex: typeIdx, Line: line}, true
}

func rewriteNumericCalls(wat string, callIndexMap map[int]int) string {
	re := regexp.MustCompile(`\bcall ([0-9]+)\b`)
	return re.ReplaceAllStringFunc(wat, func(s string) string {
		idxText := strings.TrimPrefix(s, "call ")
		idx, err := strconv.Atoi(idxText)
		if err != nil {
			return s
		}
		if newIdx, ok := callIndexMap[idx]; ok {
			return fmt.Sprintf("call %d", newIdx)
		}
		return s
	})
}

func unsupportedImportsInWAT(wat string) []string {
	seen := make(map[string]bool)
	for _, imp := range mustParseImports(wat) {
		if imp.Module == wasiModule && !allowedWASIImports[imp.Name] {
			seen[imp.Name] = true
		}
	}
	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func mustParseImports(wat string) []importFunc {
	imports, _ := parseImports(wat)
	return imports
}

func countLessThan(sorted []int, n int) int {
	return sort.SearchInts(sorted, n)
}

func sanitizeName(name string) string {
	return strings.NewReplacer("-", "_", ".", "_", "/", "_").Replace(name)
}

func wasmToolsOutput(args ...string) ([]byte, error) {
	const name = "wasm-tools"

	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, stderr.String())
	}
	return out, nil
}
