package typekit

import (
	_ "embed"
	"path"

	"github.com/binaryfarm/typekit/internal/compiler"
	"github.com/binaryfarm/typekit/internal/repl"
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

func NewApp(options Options) *App {
	runtime := runtime.NewRuntime()
	return &App{
		options: options,
		runtime: runtime,
	}
}

func NewREPL() *repl.REPL {
	return repl.NewREPL()
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
