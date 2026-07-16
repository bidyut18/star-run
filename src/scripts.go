package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const separator = "────────────────────────────────────────"

type PackageLocator struct{}

func (l PackageLocator) Find(startDir, stopDir string) (string, error) {
	if startDir == "" {
		return "", fmt.Errorf("startDir cannot be empty: %w", os.ErrInvalid)
	}

	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolving startDir: %w", err)
	}

	var absStop string
	if stopDir != "" {
		absStop, err = filepath.Abs(stopDir)
		if err != nil {
			return "", fmt.Errorf("resolving stopDir: %w", err)
		}
	}

	for {
		pkgPath := filepath.Join(dir, "package.json")

		info, err := os.Stat(pkgPath)
		if err == nil && !info.IsDir() {
			return pkgPath, nil
		}

		if absStop != "" && dir == absStop {
			return "", ErrPackageNotFound
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrPackageNotFound
		}
		dir = parent
	}
}

type PackageReader struct{}

func (r PackageReader) Read(path string) (PackageJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PackageJSON{}, fmt.Errorf("opening package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return PackageJSON{}, fmt.Errorf("parsing package.json: %w", err)
	}

	return pkg, nil
}

// ReadScripts unmarshals only the "scripts" field, avoiding the cost of
// parsing the rest of package.json. Returns the scripts pre-sorted.
func (r PackageReader) ReadScripts(path string) ([]Script, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening package.json: %w", err)
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parsing package.json: %w", err)
	}

	scripts := make([]Script, 0, len(pkg.Scripts))
	for name, cmd := range pkg.Scripts {
		scripts = append(scripts, Script{Name: name, Command: cmd})
	}

	sort.Sort(byName(scripts))
	return scripts, nil
}

// HasScript checks script existence by unmarshaling only the "scripts" field.
func (r PackageReader) HasScript(path, name string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("opening package.json: %w", err)
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false, fmt.Errorf("parsing package.json: %w", err)
	}

	_, ok := pkg.Scripts[name]
	return ok, nil
}

type ScriptRenderer struct {
	Writer io.Writer
}

func (r ScriptRenderer) Render(pm PackageManager, scripts []Script) error {
	if len(scripts) == 0 {
		_, err := fmt.Fprintln(r.Writer, "No scripts found in package.json.")
		return err
	}

	maxLen := 0
	for _, s := range scripts {
		if l := len(s.Name); l > maxLen {
			maxLen = l
		}
	}

	bw := bufio.NewWriter(r.Writer)

	if _, err := fmt.Fprintf(bw, "\n📦 Detected: %s\n", pm); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(bw, separator); err != nil {
		return err
	}

	// Pre-build a padding string and slice it instead of calling fmt per line.
	spaces := strings.Repeat(" ", maxLen)
	for _, s := range scripts {
		if _, err := bw.WriteString("  " + s.Name + spaces[len(s.Name):] + "  " + s.Command + "\n"); err != nil {
			return err
		}
	}

	if _, err := bw.WriteString("\n"); err != nil {
		return err
	}
	return bw.Flush()
}

type byName []Script

func (s byName) Len() int           { return len(s) }
func (s byName) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s byName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type ScriptService struct {
	Locator  PackageLocator
	Reader   PackageReader
	Renderer ScriptRenderer
}

func (s *ScriptService) ListScripts(startDir string, pm PackageManager, stopDir string) error {
	pkgPath, err := s.Locator.Find(startDir, stopDir)
	if err != nil {
		return err
	}

	scripts, err := s.Reader.ReadScripts(pkgPath)
	if err != nil {
		return err
	}

	return s.Renderer.Render(pm, scripts)
}

func (s *ScriptService) ValidateScript(startDir, stopDir, scriptName string) error {
	pkgPath, err := s.Locator.Find(startDir, stopDir)
	if err != nil {
		return err
	}

	has, err := s.Reader.HasScript(pkgPath, scriptName)
	if err != nil {
		return err
	}

	if !has {
		return fmt.Errorf("script '%s' not found in package.json", scriptName)
	}
	return nil
}

func listScripts(startDir string, pm PackageManager, stopDir string) {
	svc := ScriptService{
		Locator:  PackageLocator{},
		Reader:   PackageReader{},
		Renderer: ScriptRenderer{Writer: os.Stdout},
	}

	if err := svc.ListScripts(startDir, pm, stopDir); err != nil {
		fatalf("Error: %v\n", err)
	}
}