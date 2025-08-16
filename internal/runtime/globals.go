package runtime

import (
	"github.com/binaryfarm/typekit/internal/engine"
	"github.com/binaryfarm/typekit/internal/stdlib"
)

func setGlobals(vm *engine.Runtime) {
	global := vm.GlobalObject()
	global.Set("console", stdlib.Console{})
}
