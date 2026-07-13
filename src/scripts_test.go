package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec" 
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

	pkg := PackageJSON{
		PackageManager: "",
		Scripts:        scripts,
	}
	pkgData, err := json.Marshal(pkg)
	if err != nil {
		t.Fatal(err)
	}

	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, pkgData, 0644); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listScripts(tmpDir, Npm, tmpDir)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
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

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	listScripts(tmpDir, Npm, tmpDir)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("No scripts found")) {
		t.Errorf("expected 'No scripts found', got %q", output)
	}
}

func TestListScriptsNoPackageJsonSubprocess(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS") == "1" {
		// Use a single temp dir for both start and stop
		tmpDir := os.Getenv("TEST_TMP_DIR")
		if tmpDir == "" {
			tmpDir = t.TempDir() // fallback
		}
		listScripts(tmpDir, Npm, tmpDir)
		return
	}
	tmpDir := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestListScriptsNoPackageJsonSubprocess")
	cmd.Env = append(os.Environ(), "TEST_SUBPROCESS=1", "TEST_TMP_DIR="+tmpDir)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected error (exit code 1)")
	}
	expected := "Error: No package.json found in tree."
	if !bytes.Contains(out, []byte(expected)) {
		t.Errorf("expected stderr to contain %q, got %q", expected, string(out))
	}
}