package dep

import (
	"fmt"
	"github.com/metakeule/cli"
)

func Show(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	fmt.Printf("%s\n", asJson(packages(o)...))
	return 0
}
