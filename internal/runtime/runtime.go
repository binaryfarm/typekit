package runtime

import (
	"github.com/binaryfarm/typekit/internal/modules"
	api "github.com/grafana/sobek"
)

type Runtime struct {
	vm      *api.Runtime
	modules *modules.TypeKitComboResolver
}

func NewRuntime() *Runtime {

	vm := api.New()
	resolver := modules.NewTypeKitComboResolver(vm)

	vm.SetFieldNameMapper(api.TagFieldNameMapper("json", true))
	set_globals(vm)
	return &Runtime{
		vm:      vm,
		modules: resolver,
	}
}

func (r *Runtime) VM() *api.Runtime {
	return r.vm
}

func (r *Runtime) Eval(src string) (api.Value, error) {
	record, e := api.ParseModule("app", src, r.modules.Resolve)
	if e != nil {
		return nil, e
	}
	e = record.Link()
	if e != nil {
		return nil, e
	}
	e = record.InitializeEnvironment()
	if e != nil {
		return nil, e
	}
	promise := record.Evaluate(r.vm)
	if promise.State() != api.PromiseStateFulfilled {
		err := promise.Result().Export().(error)
		return nil, err
	}
	return promise.Result(), nil
}

func (r *Runtime) Close() error {
	return nil
}
