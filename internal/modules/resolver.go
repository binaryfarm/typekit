package modules

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/binaryfarm/typekit/internal/engine"
	"golang.org/x/mod/semver"
)

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
	if s, ok := os.Stat(path); ok != nil {
		return false
	} else {
		return s != nil
	}
}

func searchForPackage(specifier, version string) string {
	home, _ := os.UserHomeDir()
	v := semver.Build(version)
	retpath := ""
	base := filepath.Join(home, tk_module_home, specifier)
	e := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			v2 := semver.Build(d.Name())
			if semver.Compare(v2, v) > 0 {
				v = v2
			}
			return nil
		}
		if d.Name() == "package.json" || d.Name() == "package" {
			fmt.Printf("searching package.json at @%s\n", path)
			js, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			var pkg nodePackage
			e = json.Unmarshal(js, &pkg)
			if e != nil {
				return e
			}
			v2 := semver.Build(pkg.Version)
			if semver.Compare(v2, v) > 0 {
				v = v2
				retpath = path
				return nil
			}
		}
		return nil
	})
	if e != nil {
		return ""
	}
	return retpath
}
func resolveFsPackage(specifier string) (engine.ModuleRecord, error) {
	var resolver func(referencingScriptOrModule interface{}, specifier string) (engine.ModuleRecord, error)
	resolver = func(referencingScriptOrModule interface{}, specifier string) (engine.ModuleRecord, error) {
		var p string
		p = filepath.Join(".", specifier, "package.json")
		if !fileExists(p) {
			p = filepath.Join("node_modules", specifier, "package.json")
			if !fileExists(p) {
				p = searchForPackage(specifier, "0.0.0")
			}
		}
		fmt.Printf("searching path %s\n", p)
		pjson, e := os.ReadFile(p)
		if e != nil {
			return nil, e
		}
		var np nodePackage
		e = json.Unmarshal(pjson, &np)
		if e != nil {
			return nil, e
		}
		index := filepath.Join("node_modules", specifier, np.Main)
		ext := filepath.Ext(index)
		if ext == "" {
			index = fmt.Sprintf("%s.js", index)
		}
		indexjs, e := os.ReadFile(index)
		if e != nil {
			return nil, e
		}
		return engine.ParseModule(specifier, string(indexjs), resolver)

	}
	return resolver(nil, specifier)
}

func resolveTkPackage(specifier string) (engine.ModuleRecord, error) {
	return nil, fmt.Errorf("not implemented")
}

func resolveNodePackage(specifier string) (engine.ModuleRecord, error) {
	var resolver func(referencingScriptOrModule interface{}, specifier string) (engine.ModuleRecord, error)
	resolver = func(referencingScriptOrModule interface{}, specifier string) (engine.ModuleRecord, error) {
		p := filepath.Join(".", specifier, "package.json")
		fmt.Printf("searching for package.json ./%s\n", p)
		pjson, e := os.ReadFile(p)
		if e != nil {
			return nil, e
		}
		var np nodePackage
		e = json.Unmarshal(pjson, &np)
		if e != nil {
			return nil, e
		}
		index := filepath.Join("node_modules", specifier, np.Main)
		ext := filepath.Ext(index)
		if ext == "" {
			index = fmt.Sprintf("%s.js", index)
		}
		indexjs, e := os.ReadFile(index)
		if e != nil {
			return nil, e
		}
		return engine.ParseModule(specifier, string(indexjs), resolver)

	}
	return resolver(nil, specifier)
}
