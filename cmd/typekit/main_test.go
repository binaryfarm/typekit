package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/binaryfarm/typekit/pkg/typekit"
)

func TestApp(t *testing.T) {
	d, _ := os.Getwd()
	d = filepath.Clean(fmt.Sprintf("%s/../../", d))
	entryPoint := filepath.Join(d, "test2.ts")
	fmt.Println(entryPoint)
	app := typekit.NewApp(typekit.Options{
		Watch:      false,
		EntryPoint: entryPoint,
	})
	e := app.Run()
	if e != nil {
		t.Error(e)
	}
}
