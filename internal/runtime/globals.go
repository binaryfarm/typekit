package runtime

import (
	"github.com/binaryfarm/typekit/internal/common"
	"github.com/grafana/sobek"
)

func set_globals(vm *sobek.Runtime) {
	global := vm.GlobalObject()
	global.Set("console", common.Console{})
}
