package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

const separator = "────────────────────────────────────────"

type ScriptRenderer struct {
	Writer io.Writer
}

func (r ScriptRenderer) Render(pm PackageManager, scripts []Script) error {
	if len(scripts) == 0 {
		_, err := fmt.Fprintln(r.Writer, "No scripts found in configuration.")
		return err
	}
	maxLen := 0
	for _, s := range scripts {
		if l := utf8.RuneCountInString(s.Name); l > maxLen {
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
	spaces := strings.Repeat(" ", maxLen)
	for _, s := range scripts {
		pad := maxLen - utf8.RuneCountInString(s.Name)
		line := "  " + s.Name + spaces[:pad] + "  " + s.Command + "\n"
		if _, err := bw.WriteString(line); err != nil {
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
	Renderer ScriptRenderer
}

func (s *ScriptService) ListScripts(startDir string, pm PackageManager, stopDir string) error {
	cfg, err := LocateConfig(startDir, stopDir, pm)
	if err != nil {
		return err
	}
	reader, err := getReader(cfg.Type)
	if err != nil {
		return err
	}
	scripts, err := reader.ReadScripts(cfg.Path)
	if err != nil {
		return err
	}
	return s.Renderer.Render(pm, scripts)
}

// scriptNotFoundError is returned when a script does not exist in the config.
// It preserves the exact error message for compatibility and matches ErrScriptNotFound.
type scriptNotFoundError struct {
	name string
}

func (e *scriptNotFoundError) Error() string {
	return fmt.Sprintf("script '%s' not found in configuration", e.name)
}

func (e *scriptNotFoundError) Is(target error) bool {
	return target == ErrScriptNotFound
}

func (s *ScriptService) ValidateScript(startDir, stopDir string, pm PackageManager, scriptName string) error {
	cfg, err := LocateConfig(startDir, stopDir, pm)
	if err != nil {
		return err
	}
	reader, err := getReader(cfg.Type)
	if err != nil {
		return err
	}
	has, err := reader.HasScript(cfg.Path, scriptName)
	if err != nil {
		return err
	}
	if !has {
		return &scriptNotFoundError{name: scriptName}
	}
	return nil
}

func getReader(typ ConfigType) (ConfigReader, error) {
	switch typ {
	case ConfigPackageJSON:
		return PackageJSONReader{}, nil
	case ConfigDenoJSON, ConfigDenoJSONC:
		return DenoConfigReader{}, nil
	default:
		return nil, fmt.Errorf("unsupported config type: %v", typ)
	}
}

func listScripts(startDir string, pm PackageManager, stopDir string) error {
	svc := ScriptService{Renderer: ScriptRenderer{Writer: os.Stdout}}
	return svc.ListScripts(startDir, pm, stopDir)
}