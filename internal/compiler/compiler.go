package compiler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Compiler struct{}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(src string) (string, error) {
	return c.compileInDir(src, ".")
}

func (c *Compiler) compileInDir(src string, workDir string) (string, error) {
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return "", err
	}

	projectRoot := findProjectRoot(absWorkDir)

	tmpDir, err := os.MkdirTemp("", "typekit-tsgo-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	copyDir(projectRoot, tmpDir)

	inputFile := filepath.Join(tmpDir, "input.ts")
	if err := os.WriteFile(inputFile, []byte(src), 0644); err != nil {
		return "", fmt.Errorf("failed to write input: %w", err)
	}

	packageJSON := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(packageJSON, []byte(`{"type": "module"}`), 0644); err != nil {
		return "", fmt.Errorf("failed to write package.json: %w", err)
	}

	tsconfigFile := filepath.Join(tmpDir, "tsconfig.json")
	tsconfig := `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "react-jsx",
    "strict": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "allowImportingTsExtensions": true,
    "rewriteRelativeImportExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": false,
    "noEmitOnError": true,
    "declaration": false,
    "sourceMap": false,
    "removeComments": false,
    "newLine": "lf",
    "outDir": "out",
    "rootDir": "."
  },
  "include": ["**/*.ts", "**/*.tsx"]
}`
	if err := os.WriteFile(tsconfigFile, []byte(tsconfig), 0644); err != nil {
		return "", fmt.Errorf("failed to write tsconfig: %w", err)
	}

	cmd := exec.Command("tsgo", "-p", tsconfigFile)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tsgo failed: %s\n%s", err, string(output))
	}

	// Copy compiled output back to project root
	outDir := filepath.Join(tmpDir, "out")
	if err := copyDir(outDir, projectRoot); err != nil {
		return "", fmt.Errorf("failed to copy output: %w", err)
	}

	outFile := filepath.Join(tmpDir, "out", "input.js")
	jsContent, err := os.ReadFile(outFile)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}

	return string(jsContent), nil
}

func (c *Compiler) Build(entry string, buildPath string) error {
	absEntry, err := filepath.Abs(entry)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(absEntry)
	if err != nil {
		return err
	}

	workDir := filepath.Dir(absEntry)
	_, err = c.compileInDir(string(content), workDir)
	return err
}

func findProjectRoot(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "tsconfig.json")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return startDir
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		if strings.HasPrefix(relPath, "node_modules") || strings.HasPrefix(relPath, ".git") || strings.HasPrefix(relPath, "bin") || strings.HasPrefix(relPath, "vendor") || strings.HasPrefix(relPath, "vendor-ts") || relPath == "test.ts" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, content, 0644)
	})
}