package runtime

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/binaryfarm/typekit/internal/engine"
	"github.com/binaryfarm/typekit/internal/modules"
)

type Runtime struct {
	vm           *engine.Runtime
	modules      *modules.TypeKitResolver
	importSem    chan struct{}
	importCancel context.CancelFunc
	importWG     sync.WaitGroup
	closed       bool
	closeMu      sync.Mutex
}

func NewRuntime() *Runtime {
	vm := engine.New()
	resolver := modules.NewTypeKitResolver(vm)

	vm.SetFieldNameMapper(engine.TagFieldNameMapper("json", true))
	setGlobals(vm)

	_, cancel := context.WithCancel(context.Background())
	r := &Runtime{
		vm:           vm,
		modules:      resolver,
		importSem:    make(chan struct{}, 10),
		importCancel: cancel,
	}

	vm.SetImportModuleDynamically(r.handleDynamicImport)

	return r
}

func (r *Runtime) VM() *engine.Runtime {
	return r.vm
}

func (r *Runtime) Repl(src string) (engine.Value, error) {
	return r.vm.RunString(src)
}

func (r *Runtime) Eval(src string) (engine.Value, error) {
	record, e := engine.ParseModule("main", src, r.modules.Resolve)
	if e != nil {
		return nil, e
	}

	r.modules.AddMainModule(record)

	e = record.Link()
	if e != nil {
		return nil, e
	}
	promise := record.Evaluate(r.vm)

	for promise.State() == engine.PromiseStatePending {
		r.vm.RunJobs()
	}

	if promise.State() == engine.PromiseStateRejected {
		err := promise.Result().Export().(error)
		return nil, err
	}
	return promise.Result(), nil
}

func (r *Runtime) Close() error {
	r.closeMu.Lock()
	defer r.closeMu.Unlock()

	if r.closed {
		return nil
	}
	r.closed = true

	r.importCancel()

	done := make(chan struct{})
	go func() {
		r.importWG.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	r.vm.Interrupt(nil)
	r.vm.ClearInterrupt()

	return nil
}

func (r *Runtime) handleDynamicImport(referencingScriptOrModule interface{}, specifier engine.Value, promiseCapability interface{}) {
	specifierStr := specifier.String()

	r.importWG.Add(1)
	go func() {
		defer r.importWG.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		select {
		case r.importSem <- struct{}{}:
			defer func() { <-r.importSem }()
		case <-ctx.Done():
			r.vm.FinishLoadingImportModule(referencingScriptOrModule, specifier, promiseCapability, nil, ctx.Err())
			return
		}

		resolvedPath := specifierStr
		if !filepath.IsAbs(resolvedPath) {
			resolvedPath, _ = filepath.Abs(resolvedPath)
		}

		sourceText, err := os.ReadFile(resolvedPath)
		if err != nil {
			r.vm.FinishLoadingImportModule(referencingScriptOrModule, specifier, promiseCapability, nil, err)
			return
		}

		record, err := engine.ParseModule(specifierStr, string(sourceText), r.modules.Resolve)
		if err != nil {
			r.vm.FinishLoadingImportModule(referencingScriptOrModule, specifier, promiseCapability, nil, err)
			return
		}

		err = record.Link()
		if err != nil {
			r.vm.FinishLoadingImportModule(referencingScriptOrModule, specifier, promiseCapability, nil, err)
			return
		}

		r.vm.FinishLoadingImportModule(referencingScriptOrModule, specifier, promiseCapability, record, nil)
	}()
}
