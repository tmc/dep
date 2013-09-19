package depcore

/*
import (
	"fmt"
	"github.com/metakeule/cli"
)

func CLIShow(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	fmt.Printf("%s\n", asJson(packages(o)...))
	return 0
}
*/

import (
	"fmt"
	"github.com/metakeule/exports"
)

func (env *Environment) CLIShow(pkgs ...*exports.Package) string {
	return fmt.Sprintf("%s\n", asJson(pkgs...))
}
