package modules

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/grafana/sobek"
)

type TypeKitComboResolver struct {
	mu           sync.Mutex
	cache        map[string]cacheElement
	reverseCache map[sobek.ModuleRecord]string
	custom       func(interface{}, string) (sobek.ModuleRecord, error)
}
type cacheElement struct {
	m   sobek.ModuleRecord
	err error
}
type unresolvedBinding struct {
	module  string
	binding string
}

func NewTypeKitComboResolver(vm *sobek.Runtime) *TypeKitComboResolver {
	r := &TypeKitComboResolver{cache: make(map[string]cacheElement), reverseCache: make(map[sobek.ModuleRecord]string)}
	r.custom = r.customResolver
	vm.SetGetImportMetaProperties(func(m sobek.ModuleRecord) []sobek.MetaProperty {
		specifier, ok := r.reverseCache[m]
		if !ok {
			panic("we got import.meta for module that wasn't imported: " + specifier)
		}
		return []sobek.MetaProperty{
			{
				Key:   "url",
				Value: vm.ToValue("file:///" + specifier),
			},
		}
	})
	return r
}

func (s *TypeKitComboResolver) Resolve(referencingScriptOrModule interface{}, specifier string) (sobek.ModuleRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k, ok := s.cache[specifier]
	if ok {
		return k.m, k.err
	}
	if strings.HasPrefix(specifier, "tk:") {
		p, err := s.custom(referencingScriptOrModule, specifier)
		s.cache[specifier] = cacheElement{m: p, err: err}
		return p, err
	}
	b, err := os.ReadFile(specifier)
	if err != nil {
		s.cache[specifier] = cacheElement{err: err}
		return nil, err
	}
	p, err := sobek.ParseModule(specifier, string(b), s.Resolve)
	if err != nil {
		s.cache[specifier] = cacheElement{err: err}
		return nil, err
	}
	s.cache[specifier] = cacheElement{m: p}
	s.reverseCache[p] = specifier
	return p, nil
}

func (r *TypeKitComboResolver) customResolver(_ interface{}, specifier string) (sobek.ModuleRecord, error) {
	switch specifier {
	case "custom:coolstuff":
		return &simpleModuleImpl{}, nil
	case "custom:coolstuff2":
		return &cyclicModuleImpl{
			resolve:          r.Resolve,
			requestedModules: []string{"custom:coolstuff3", "custom:coolstuff"},
			exports: map[string]unresolvedBinding{
				"coolStuff": {
					binding: "coolStuff",
					module:  "custom:coolstuff",
				},
				"otherCoolStuff": { // request it from third module which will request it back from us
					binding: "coolStuff",
					module:  "custom:coolstuff3",
				},
			},
		}, nil
	case "custom:coolstuff3":
		return &cyclicModuleImpl{
			resolve:          r.Resolve,
			requestedModules: []string{"custom:coolstuff2"},
			exports: map[string]unresolvedBinding{
				"coolStuff": { // request it back from the module
					binding: "coolStuff",
					module:  "custom:coolstuff2",
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown module %q", specifier)
	}
}

type simpleModuleImpl struct{}

var _ sobek.ModuleRecord = &simpleModuleImpl{}

func (s *simpleModuleImpl) Link() error {
	// this does nothing on this
	return nil
}

func (s *simpleModuleImpl) ResolveExport(exportName string, resolveset ...sobek.ResolveSetElement) (*sobek.ResolvedBinding, bool) {
	if exportName == "coolStuff" {
		return &sobek.ResolvedBinding{
			BindingName: exportName,
			Module:      s,
		}, false
	}
	return nil, false
}

func (s *simpleModuleImpl) Evaluate(rt *sobek.Runtime) *sobek.Promise {
	p, res, _ := rt.NewPromise()
	res(&simpleModuleInstanceImpl{rt: rt})
	return p
}

func (s *simpleModuleImpl) GetExportedNames(callback func([]string), records ...sobek.ModuleRecord) bool {
	callback([]string{"coolStuff"})
	return true
}

type simpleModuleInstanceImpl struct {
	rt *sobek.Runtime
}

func (si *simpleModuleInstanceImpl) GetBindingValue(exportName string) sobek.Value {
	if exportName == "coolStuff" {
		return si.rt.ToValue(5)
	}
	return nil
}

// START of cyclic module implementation
type cyclicModuleImpl struct {
	requestedModules []string
	exports          map[string]unresolvedBinding
	resolve          sobek.HostResolveImportedModuleFunc
}

var _ sobek.CyclicModuleRecord = &cyclicModuleImpl{}

func (s *cyclicModuleImpl) InitializeEnvironment() error {
	return nil
}

func (s *cyclicModuleImpl) Instantiate(_ *sobek.Runtime) (sobek.CyclicModuleInstance, error) {
	return &cyclicModuleInstanceImpl{module: s}, nil
}

func (s *cyclicModuleImpl) RequestedModules() []string {
	return s.requestedModules
}

func (s *cyclicModuleImpl) Link() error {
	// this does nothing on this
	return nil
}

func (s *cyclicModuleImpl) Evaluate(rt *sobek.Runtime) *sobek.Promise {
	return rt.CyclicModuleRecordEvaluate(s, s.resolve)
}

func (s *cyclicModuleImpl) ResolveExport(exportName string, resolveset ...sobek.ResolveSetElement) (*sobek.ResolvedBinding, bool) {
	b, ok := s.exports[exportName]
	if !ok {
		return nil, false
	}

	m, err := s.resolve(s, b.module)
	if err != nil {
		panic(err)
	}

	return &sobek.ResolvedBinding{
		Module:      m,
		BindingName: b.binding,
	}, false
}

func (s *cyclicModuleImpl) GetExportedNames(callback func([]string), records ...sobek.ModuleRecord) bool {
	result := make([]string, 0, len(s.exports))
	for k := range s.exports {
		result = append(result, k)
	}
	sort.Strings(result)
	callback(result)
	return true
}

type cyclicModuleInstanceImpl struct {
	rt     *sobek.Runtime
	module *cyclicModuleImpl
}

func (si *cyclicModuleInstanceImpl) HasTLA() bool {
	return false
}

func (si *cyclicModuleInstanceImpl) ExecuteModule(rt *sobek.Runtime, _, _ func(interface{}) error) (sobek.CyclicModuleInstance, error) {
	si.rt = rt
	return si, nil
}

func (si *cyclicModuleInstanceImpl) GetBindingValue(exportName string) sobek.Value {
	b, ambigious := si.module.ResolveExport(exportName)
	if ambigious || b == nil {
		panic("fix this")
	}
	return si.rt.GetModuleInstance(b.Module).GetBindingValue(exportName)
}
