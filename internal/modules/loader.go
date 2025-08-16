package modules

import (
	"fmt"
	"os"
	"path/filepath"
)

type Loader interface {
	Data() ([]byte, error)
}

func get_loader(specifier string) Loader {
	ext := filepath.Ext(specifier)
	switch ext {
	case "js":
		return new_ts_loader(specifier)
	case "ts":
		return new_ts_loader(specifier)
	default:
		return new_ts_loader(fmt.Sprintf("%s.ts", specifier))
	}
}

type tsLoader struct {
	path string
}

func (t *tsLoader) Data() ([]byte, error) {
	return os.ReadFile(t.path)
}

func new_ts_loader(specifier string) Loader {
	path, _ := filepath.Abs(specifier)
	return &tsLoader{
		path: path,
	}
}
