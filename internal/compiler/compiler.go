package compiler

import (
	"fmt"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

type Compiler struct {
	Result esbuild.BuildResult
}

func NewCompiler() *Compiler {
	return &Compiler{}
}
func (c *Compiler) Compile(src string) (string, error) {
	result := esbuild.Transform(src, esbuild.TransformOptions{
		Loader: esbuild.LoaderTS,
	})
	if len(result.Errors) != 0 {
		msg := ""
		for _, err := range result.Errors {
			msg += err.Text + "\n"
		}
		return "", fmt.Errorf("build error: %s", msg)
	}
	return string(result.Code), nil
}

func (c *Compiler) Build(entry string, buildPath string) error {
	options := esbuild.BuildOptions{
		EntryPoints: []string{entry},
		Outdir:      buildPath,
		Packages:    esbuild.PackagesExternal,
	}

	c.Result = esbuild.Build(options)

	if len(c.Result.Errors) != 0 {
		msg := ""
		for _, err := range c.Result.Errors {
			msg += err.Text + "\n"
		}
		return fmt.Errorf("build error: %s", msg)
	}
	return nil
}
