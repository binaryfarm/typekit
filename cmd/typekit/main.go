package main

import (
	"flag"
	"fmt"

	"github.com/binaryfarm/typekit/pkg/typekit"
)

var (
	Watch = flag.Bool("watch", false, "Watch for changes")
)

func usage() {
	fmt.Print("TypeKit CLI Usage:\n")
	fmt.Print("typekit [OPTIONS] <src>\n")
	fmt.Print("--------------------------------\n")
	fmt.Print("OPTIONS:\n")
	flag.PrintDefaults()
}
func init() {
	flag.Parse()
}
func main() {
	if len(flag.Args()) < 1 {
		usage()
		return
	}
	args := flag.Args()
	entryPoint := args[len(args)-1]

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
