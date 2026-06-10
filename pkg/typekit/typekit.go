package typekit

import (
	_ "embed"
	"fmt"
	"os"
	"runtime/debug"

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

func (a *App) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v\n%s", r, debug.Stack())
		}
	}()

	c := compiler.NewCompiler()
	js, e := c.Compile(readEntryPoint(a.options.EntryPoint))
	if e != nil {
		return fmt.Errorf("build failed: %w", e)
	}
	_, e = a.runtime.Eval(js)
	if e != nil {
		return fmt.Errorf("runtime error: %w", e)
	}
	return nil
}

func readEntryPoint(entry string) string {
	content, err := os.ReadFile(entry)
	if err != nil {
		return ""
	}
	return string(content)
}
func (a *App) Destroy() error {
	return a.runtime.Close()
}
