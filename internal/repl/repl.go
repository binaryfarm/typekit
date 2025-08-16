package repl

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	"github.com/binaryfarm/typekit/internal/compiler"
	"github.com/binaryfarm/typekit/internal/engine"
	tkr "github.com/binaryfarm/typekit/internal/runtime"
)

type REPL struct {
	runtime  *tkr.Runtime
	compiler *compiler.Compiler
}

func NewREPL() *REPL {
	return &REPL{
		runtime:  tkr.NewRuntime(),
		compiler: compiler.NewCompiler(),
	}
}

func (r *REPL) Run(in *os.File) error {
	defer r.runtime.Close()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("")
		fmt.Println("Goodbye!")
		os.Exit(0)
	}()
	inReader := bufio.NewScanner(in)
	for {
		if inReader.Scan() {
			line := inReader.Text()
			if line == "/quit" {
				c <- os.Interrupt
				continue
			}
			js, e := r.compiler.Compile(line)
			if e != nil {
				fmt.Printf("ERROR!! %v\n", e)
				continue
			}
			v, e := r.runtime.Repl(js)
			if e != nil {
				fmt.Printf("ERROR!! %v\n", e)
			}
			if v != nil {
				if !engine.IsUndefined(v) {
					fmt.Printf("%v\n", v)
				}
			}
		}
	}
}
