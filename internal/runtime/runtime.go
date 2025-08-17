package runtime

import (
	"github.com/binaryfarm/typekit/internal/engine"
	"github.com/binaryfarm/typekit/internal/modules"
)

type Runtime struct {
	vm      *engine.Runtime
	modules *modules.TypeKitResolver
}

func NewRuntime() *Runtime {

	vm := engine.New()
	resolver := modules.NewTypeKitResolver(vm)

	vm.SetFieldNameMapper(engine.TagFieldNameMapper("json", true))
	setGlobals(vm)
	return &Runtime{
		vm:      vm,
		modules: resolver,
	}
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
	e = record.Link()
	if e != nil {
		return nil, e
	}
	promise := record.Evaluate(r.vm)
	if promise.State() != engine.PromiseStateFulfilled {
		err := promise.Result().Export().(error)
		return nil, err
	}
	return promise.Result(), nil
}

func (r *Runtime) Close() error {
	return nil
}
