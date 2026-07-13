package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestListScripts(t *testing.T) {
	tmpDir := t.TempDir()

	scripts := map[string]string{
		"start":  "node index.js",
		"build":  "tsc",
		"test":   "jest",
		"lint":   "eslint .",
		"format": "prettier --write .",
	}

	pkg := PackageJSON{Scripts: scripts}
	pkgData, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), pkgData, 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	svc := ScriptService{
		Locator:  PackageLocator{},
		Reader:   PackageReader{},
		Renderer: ScriptRenderer{Writer: &buf},
	}

	if err := svc.ListScripts(tmpDir, Npm, tmpDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	for name, cmd := range scripts {
		if !bytes.Contains([]byte(output), []byte(name)) {
			t.Errorf("output missing script name: %s", name)
		}
		if !bytes.Contains([]byte(output), []byte(cmd)) {
			t.Errorf("output missing script command: %s", cmd)
		}
	}
	if !bytes.Contains([]byte(output), []byte("Detected: npm")) {
		t.Error("output missing detection header")
	}
}

func TestListScriptsNoScripts(t *testing.T) {
	tmpDir := t.TempDir()

	pkg := PackageJSON{Scripts: map[string]string{}}
	pkgData, _ := json.Marshal(pkg)
	_ = os.WriteFile(filepath.Join(tmpDir, "package.json"), pkgData, 0644)

	var buf bytes.Buffer
	svc := ScriptService{
		Locator:  PackageLocator{},
		Reader:   PackageReader{},
		Renderer: ScriptRenderer{Writer: &buf},
	}

	if err := svc.ListScripts(tmpDir, Npm, tmpDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("No scripts found")) {
		t.Errorf("expected 'No scripts found', got %q", output)
	}
}

func TestListScriptsNoPackageJson(t *testing.T) {
	tmpDir := t.TempDir()

	var buf bytes.Buffer
	svc := ScriptService{
		Locator:  PackageLocator{},
		Reader:   PackageReader{},
		Renderer: ScriptRenderer{Writer: &buf},
	}

	err := svc.ListScripts(tmpDir, Npm, tmpDir)
	if !errors.Is(err, ErrPackageNotFound) {
		t.Fatalf("expected ErrPackageNotFound, got %v", err)
	}
}

// ─── NEW: ValidateScript tests ───
func TestValidateScriptExists(t *testing.T) {
	tmpDir := t.TempDir()

	pkg := PackageJSON{Scripts: map[string]string{"build": "tsc"}}
	data, _ := json.Marshal(pkg)
	_ = os.WriteFile(filepath.Join(tmpDir, "package.json"), data, 0644)

	svc := ScriptService{Locator: PackageLocator{}, Reader: PackageReader{}}
	if err := svc.ValidateScript(tmpDir, tmpDir, "build"); err != nil {
		t.Fatalf("expected 'build' to exist, got: %v", err)
	}
}

func TestValidateScriptMissing(t *testing.T) {
	tmpDir := t.TempDir()

	pkg := PackageJSON{Scripts: map[string]string{"build": "tsc"}}
	data, _ := json.Marshal(pkg)
	_ = os.WriteFile(filepath.Join(tmpDir, "package.json"), data, 0644)

	svc := ScriptService{Locator: PackageLocator{}, Reader: PackageReader{}}
	err := svc.ValidateScript(tmpDir, tmpDir, "deploy")
	if err == nil {
		t.Fatal("expected error for missing script, got nil")
	}
	want := "script 'deploy' not found in package.json"
	if err.Error() != want {
		t.Errorf("expected %q, got %q", want, err.Error())
	}
}

func TestValidateScriptNoPackageJson(t *testing.T) {
	tmpDir := t.TempDir()

	svc := ScriptService{Locator: PackageLocator{}, Reader: PackageReader{}}
	err := svc.ValidateScript(tmpDir, tmpDir, "build")
	if !errors.Is(err, ErrPackageNotFound) {
		t.Fatalf("expected ErrPackageNotFound, got %v", err)
	}
}