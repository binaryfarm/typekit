package modules

import (
	"fmt"
	"strings"
	"sync"

	"github.com/binaryfarm/typekit/internal/engine"
)

type TypeKitResolver struct {
	mu           sync.Mutex
	cache        map[string]cacheElement
	reverseCache map[engine.ModuleRecord]string
}
type cacheElement struct {
	m   engine.ModuleRecord
	err error
}

func NewTypeKitResolver(vm *engine.Runtime) *TypeKitResolver {
	r := &TypeKitResolver{cache: make(map[string]cacheElement), reverseCache: make(map[engine.ModuleRecord]string)}
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

func (s *TypeKitResolver) Resolve(referencingScriptOrModule interface{}, specifier string) (engine.ModuleRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k, ok := s.cache[specifier]
	if ok {
		return k.m, k.err
	}
	// NodeJS support
	if strings.HasPrefix(specifier, "node:") {
		m, e := resolveNodePackage(specifier)
		if e != nil {
			s.cache[specifier] = cacheElement{err: e}
			return nil, e
		}
		s.cache[specifier] = cacheElement{m: m}
		s.reverseCache[m] = specifier
		return m, nil
	} else if strings.HasPrefix(specifier, "typekit:") { // Yeah baby!
		m, e := resolveTkPackage(specifier)
		if e != nil {
			s.cache[specifier] = cacheElement{err: e}
			return nil, e
		}
		s.cache[specifier] = cacheElement{m: m}
		s.reverseCache[m] = specifier
		return m, nil
	} else {
		m, e := resolveFsPackage(specifier)
		if e != nil {
			s.cache[specifier] = cacheElement{err: e}
			return nil, e
		}
		s.cache[specifier] = cacheElement{m: m}
		s.reverseCache[m] = specifier
		return m, nil
	}
}
