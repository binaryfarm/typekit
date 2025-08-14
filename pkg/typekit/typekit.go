package typekit

import (
	"path"

	"github.com/binaryfarm/typekit/internal/common"
	"github.com/binaryfarm/typekit/internal/compiler"
	"github.com/binaryfarm/typekit/internal/runtime"
)

type App struct {
	runtime *runtime.Runtime
	options Options
}
type Options struct {
	Watch      bool
	EntryPoint string
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
func setGlobals(runtime *runtime.Runtime) {
	vm := runtime.VM()
	global := vm.NewObject()
	global.Set("console", common.Console{})
}
func NewApp(options Options) *App {
	runtime := runtime.NewRuntime()
	setGlobals(runtime)
	return &App{
		options: options,
		runtime: runtime,
	}
}

func (a *App) Run() error {
	c := compiler.NewCompiler()
	e := c.Build(a.options.EntryPoint, "")
	if e != nil {
		return e
	}
	for _, f := range c.Result.OutputFiles {
		ext := path.Ext(f.Path)
		if f.Path == "<stdout>" || ext == ".js" || ext == ".ts" {
			_, e := a.runtime.Eval(string(f.Contents))
			if e != nil {
				return e
			}
		}
	}
	return nil
}
func (a *App) Destroy() error {
	return a.runtime.Close()
}
