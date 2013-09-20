package depcore

import (
	"github.com/metakeule/exports"
)

/*
import (
	"github.com/metakeule/cli"
)

func CLILint(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	return 0
}

*/

func (env *Environment) Lint(pkg *exports.Package) error {
	// parseGlobalFlags(c)
	return nil
}
