package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	core "github.com/binaryfarm/typekit"
	"github.com/binaryfarm/typekit/pkg/typekit"
)

var (
	ShowVersion = flag.Bool("version", false, "Version")
	Verbose     = flag.Bool("verbose", false, "Verbose")
	Watch       = flag.Bool("watch", false, "Watch for changes")
)

func usage() {
	fmt.Print("TypeKit CLI Usage:\n")
	fmt.Print("typekit [OPTIONS] <src>\n")
	fmt.Print("--------------------------------\n")
	fmt.Print("OPTIONS:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 && flag.NFlag() < 1 {
		usage()
		return
	}
	if ShowVersion != nil {
		if *ShowVersion {
			fmt.Println("TypeKit")
			fmt.Println(core.Version)
			fmt.Println(core.Sha)
			return
		}
	}
	args := flag.Args()
	fmt.Printf("%v", args)
	entryPoint := args[len(args)-1]

	// repl power
	if strings.ToLower(entryPoint) == "repl" {
		fmt.Println("TypeKit")
		fmt.Println(core.Version)
		fmt.Println(core.Sha)
		fmt.Println(strings.Repeat("-", 24))
		fmt.Println("/quit to exit...")
		repl := typekit.NewREPL()
		e := repl.Run(os.Stdin)
		if e != nil {
			fmt.Println(e.Error())
		}
		os.Exit(0)
	}

	app := typekit.NewApp(typekit.Options{
		Watch:      *Watch,
		EntryPoint: entryPoint,
	})
	e := app.Run()
	if e != nil {
		fmt.Println(e.Error())
	}
	defer app.Destroy()
}
