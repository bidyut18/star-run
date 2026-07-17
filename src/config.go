package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

type ConfigType int

const (
	ConfigPackageJSON ConfigType = iota
	ConfigDenoJSON
	ConfigDenoJSONC
)

type Config struct {
	Path string
	Type ConfigType
}

type ConfigReader interface {
	ReadScripts(path string) ([]Script, error)
	HasScript(path, name string) (bool, error)
}

type PackageJSONReader struct{}

func (r PackageJSONReader) ReadScripts(path string) ([]Script, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return scriptsFromMap(pkg.Scripts), nil
}

func (r PackageJSONReader) HasScript(path, name string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("opening package.json: %w", err)
	}
	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false, fmt.Errorf("parsing package.json: %w", err)
	}
	_, ok := pkg.Scripts[name]
	return ok, nil
}

type DenoConfigReader struct{}

func (r DenoConfigReader) ReadScripts(path string) ([]Script, error) {
	data, err := readConfigFile(path)
	if err != nil {
		return nil, err
	}
	var cfg struct {
		Tasks map[string]string `json:"tasks"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing deno config: %w", err)
	}
	return scriptsFromMap(cfg.Tasks), nil
}

func (r DenoConfigReader) HasScript(path, name string) (bool, error) {
	data, err := readConfigFile(path)
	if err != nil {
		return false, err
	}
	var cfg struct {
		Tasks map[string]string `json:"tasks"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false, fmt.Errorf("parsing deno config: %w", err)
	}
	_, ok := cfg.Tasks[name]
	return ok, nil
}

var trailingCommaRE = regexp.MustCompile(`,(\s*[}\]])`)

func stripTrailingCommas(data []byte) []byte {
	return trailingCommaRE.ReplaceAll(data, []byte("$1"))
}

func stripJSONC(data []byte) ([]byte, error) {
	var out bytes.Buffer
	inString := false
	escape := false
	i := 0
	for i < len(data) {
		b := data[i]

		if inString {
			if escape {
				out.WriteByte(b)
				escape = false
				i++
				continue
			}
			if b == '\\' {
				out.WriteByte(b)
				escape = true
				i++
				continue
			}
			if b == '"' {
				inString = false
			}
			out.WriteByte(b)
			i++
			continue
		}

		if b == '"' {
			inString = true
			out.WriteByte(b)
			i++
			continue
		}

		// Line comment
		if b == '/' && i+1 < len(data) && data[i+1] == '/' {
			for i < len(data) && data[i] != '\n' {
				i++
			}
			continue
		}

		// Block comment
		if b == '/' && i+1 < len(data) && data[i+1] == '*' {
			i += 2
			for i < len(data) {
				if data[i-1] == '*' && data[i] == '/' {
					i++
					break
				}
				i++
			}
			continue
		}

		out.WriteByte(b)
		i++
	}

	if inString {
		return nil, errors.New("unterminated string in JSONC")
	}
	return out.Bytes(), nil
}

func readConfigFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening config: %w", err)
	}
	stripped, err := stripJSONC(data)
	if err != nil {
		return nil, fmt.Errorf("stripping comments from %s: %w", path, err)
	}
	cleaned := stripTrailingCommas(stripped)
	return cleaned, nil
}

func scriptsFromMap(m map[string]string) []Script {
	scripts := make([]Script, 0, len(m))
	for name, cmd := range m {
		scripts = append(scripts, Script{Name: name, Command: cmd})
	}
	sort.Sort(byName(scripts))
	return scripts
}

func LocateConfig(startDir, stopDir string, pm PackageManager) (*Config, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("resolving start dir: %w", err)
	}
	var absStop string
	if stopDir != "" {
		absStop, err = filepath.Abs(stopDir)
		if err != nil {
			return nil, fmt.Errorf("resolving stop dir: %w", err)
		}
	}

	var candidates []struct {
		name string
		typ  ConfigType
	}
	if pm == Deno {
		candidates = []struct {
			name string
			typ  ConfigType
		}{
			{"deno.jsonc", ConfigDenoJSONC},
			{"deno.json", ConfigDenoJSON},
		}
	} else {
		candidates = []struct {
			name string
			typ  ConfigType
		}{
			{"package.json", ConfigPackageJSON},
		}
	}

	dir := absStart
	for {
		for _, cand := range candidates {
			fullPath := filepath.Join(dir, cand.name)
			if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
				return &Config{Path: fullPath, Type: cand.typ}, nil
			}
		}
		if absStop != "" && dir == absStop {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil, ErrConfigNotFound
}

var (
	ErrConfigNotFound = errors.New("no configuration file found")
	ErrScriptNotFound = errors.New("script not found")
)