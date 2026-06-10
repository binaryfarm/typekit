package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/binaryfarm/typekit/internal/engine"
)

type TypeKitResolver struct {
	mu           sync.Mutex
	cache        map[string]cacheElement
	reverseCache map[engine.ModuleRecord]string
	baseDirs     map[engine.ModuleRecord]string
	vm           *engine.Runtime
}
type cacheElement struct {
	m   engine.ModuleRecord
	err error
}

func NewTypeKitResolver(vm *engine.Runtime) *TypeKitResolver {
	r := &TypeKitResolver{
		cache:        make(map[string]cacheElement),
		reverseCache: make(map[engine.ModuleRecord]string),
		baseDirs:     make(map[engine.ModuleRecord]string),
		vm:           vm,
	}
	vm.SetGetImportMetaProperties(func(m engine.ModuleRecord) []engine.MetaProperty {
		specifier, ok := r.reverseCache[m]
		if !ok {
			panic("we got import.meta for module that wasn't imported: " + specifier)
		}
		fmt.Printf("Resolving module %s", specifier)
		return []engine.MetaProperty{
			{
				Key:   "url",
				Value: vm.ToValue("file:///" + specifier),
			},
		}
	})
	return r
}

func (r *TypeKitResolver) loadModuleFromFile(specifier, baseDir string) (engine.ModuleRecord, error) {
	path := specifier
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
		path, _ = filepath.Abs(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(path)
	mod, err := engine.ParseModule(specifier, string(data), r.Resolve)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.baseDirs[mod] = dir
	r.mu.Unlock()
	return mod, nil
}

func (s *TypeKitResolver) Resolve(referencingScriptOrModule interface{}, specifier string) (engine.ModuleRecord, error) {
	s.mu.Lock()
	k, ok := s.cache[specifier]
	if ok {
		m, err := k.m, k.err
		s.mu.Unlock()
		return m, err
	}

	var baseDir string
	if referencingScriptOrModule != nil {
		if mr, ok := referencingScriptOrModule.(engine.ModuleRecord); ok {
			baseDir = s.baseDirs[mr]
		}
	}
	if baseDir == "" {
		baseDir = "."
	}

	// Check for built-in prefixes that don't need file loading
	isNode := strings.HasPrefix(specifier, "node:")
	isTk := strings.HasPrefix(specifier, "typekit:")
	isRelative := strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../")

	s.mu.Unlock()

	var m engine.ModuleRecord
	var e error

	if isNode {
		m, e = resolveNodePackage(specifier)
	} else if isTk {
		m, e = resolveTkPackage(specifier)
	} else if isRelative {
		m, e = s.loadModuleFromFile(specifier, baseDir)
	} else {
		m, e = resolveFsPackage(specifier)
	}

	if e != nil {
		s.mu.Lock()
		s.cache[specifier] = cacheElement{err: e}
		s.mu.Unlock()
		return nil, e
	}

	s.mu.Lock()
	s.cache[specifier] = cacheElement{m: m}
	s.reverseCache[m] = specifier
	s.mu.Unlock()
	return m, nil
}

func (s *TypeKitResolver) AddMainModule(m engine.ModuleRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reverseCache[m] = "main"
	s.baseDirs[m] = "."
}
