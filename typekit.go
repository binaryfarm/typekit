package typekit

import (
	_ "embed"
)

//go:generate sh -c "git rev-parse --short HEAD > commit"
//go:embed commit
var Sha string

//go:embed version
var Version string
