package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/binaryfarm/typekit/internal/engine"
)

const tkModuleHome = ".typekit/modules"

type npmDist struct {
	Shasum  string `json:"shasum"`
	Tarball string `json:"tarball"`
}
type nodePackage struct {
	Main    string   `json:"main"`
	Version string   `json:"version"`
	Dist    *npmDist `json:"dist,omitempty"`
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func findPackageJSON(startDir, specifier string) (string, *nodePackage, error) {
	dir := startDir
	for {
		p := filepath.Join(dir, "node_modules", specifier, "package.json")
		if fileExists(p) {
			data, err := os.ReadFile(p)
			if err != nil {
				return "", nil, err
			}
			var pkg nodePackage
			if err := json.Unmarshal(data, &pkg); err != nil {
				return "", nil, err
			}
			return p, &pkg, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	home, _ := os.UserHomeDir()
	globalPath := filepath.Join(home, tkModuleHome, specifier, "package.json")
	if fileExists(globalPath) {
		data, err := os.ReadFile(globalPath)
		if err != nil {
			return "", nil, err
		}
		var pkg nodePackage
		if err := json.Unmarshal(data, &pkg); err != nil {
			return "", nil, err
		}
		return globalPath, &pkg, nil
	}

	return "", nil, fmt.Errorf("package %s not found", specifier)
}

func resolvePackage(specifier, referrerDir string, resolveFn engine.HostResolveImportedModuleFunc) (engine.ModuleRecord, error) {
	pkgPath, pkg, err := findPackageJSON(referrerDir, specifier)
	if err != nil {
		return nil, err
	}

	pkgDir := filepath.Dir(pkgPath)
	mainFile := pkg.Main
	if mainFile == "" {
		mainFile = "index.js"
	}

	indexPath := filepath.Join(pkgDir, mainFile)
	ext := filepath.Ext(indexPath)
	if ext == "" {
		indexPath += ".js"
	}

	if !fileExists(indexPath) {
		if ext == "" {
			indexPath = strings.TrimSuffix(indexPath, ".js") + ".ts"
		}
		if !fileExists(indexPath) {
			return nil, fmt.Errorf("main file not found for package %s", specifier)
		}
	}

	sourceText, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	return engine.ParseModule(specifier, string(sourceText), resolveFn)
}

func resolveFsPackage(specifier string) (engine.ModuleRecord, error) {
	var resolver engine.HostResolveImportedModuleFunc
	resolver = func(referencingScriptOrModule interface{}, spec string) (engine.ModuleRecord, error) {
		if strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") {
			baseDir := "."
			if referencingScriptOrModule != nil {
				if _, ok := referencingScriptOrModule.(engine.ModuleRecord); ok {
					// baseDir should be tracked by TypeKitResolver.baseDirs
					// but we can't access it here, so fallback to "."
				}
			}
			resolvedPath := filepath.Join(baseDir, spec)
			resolvedPath = filepath.Clean(resolvedPath)
			ext := filepath.Ext(resolvedPath)
			if ext == "" {
				if fileExists(resolvedPath + ".js") {
					resolvedPath += ".js"
				} else if fileExists(resolvedPath + ".ts") {
					resolvedPath += ".ts"
				} else if fileExists(filepath.Join(resolvedPath, "index.js")) {
					resolvedPath = filepath.Join(resolvedPath, "index.js")
				} else if fileExists(filepath.Join(resolvedPath, "index.ts")) {
					resolvedPath = filepath.Join(resolvedPath, "index.ts")
				}
			}

			sourceText, err := os.ReadFile(resolvedPath)
			if err != nil {
				return nil, err
			}
			return engine.ParseModule(spec, string(sourceText), resolver)
		}
		return resolvePackage(spec, ".", resolver)
	}
	return resolver(nil, specifier)
}

func resolveTkPackage(specifier string) (engine.ModuleRecord, error) {
	name := strings.TrimPrefix(specifier, "typekit:")

	builtins := map[string]string{
		"console": `export const log = (...args: any[]) => { console.log(...args); };`,
		"fs":      `export const readFile = (path: string) => { return typekit.fs.readFile(path); };`,
		"fetch":   `export const fetch = (url: string, opts?: any) => { return typekit.fetch.fetch(url, opts); };`,
	}

	if src, ok := builtins[name]; ok {
		return engine.ParseModule(specifier, src, func(referencingScriptOrModule interface{}, s string) (engine.ModuleRecord, error) {
			return nil, fmt.Errorf("builtin module %s cannot import %s", name, s)
		})
	}
	return nil, fmt.Errorf("unknown typekit module: %s", name)
}

func resolveNodePackage(specifier string) (engine.ModuleRecord, error) {
	name := strings.TrimPrefix(specifier, "node:")
	return resolvePackage(name, ".", func(referencingScriptOrModule interface{}, spec string) (engine.ModuleRecord, error) {
		return resolveNodePackage("node:" + spec)
	})
}
