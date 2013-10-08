package depcore

import (
	"fmt"
	"github.com/go-dep/gdf"
	"strings"
)

/*
import (
	"github.com/go-dep/cli"
)

func CLILint(c *cli.Context, o *Options) ErrorCode {
	// parseGlobalFlags(c)
	return 0
}

*/

func lintInit(pkg *gdf.Package) error {
	if pkg.InitMd5 == "" {
		return nil
	}
	if len(pkg.RawInits) > 1 {
		fs := []string{}
		for k, _ := range pkg.RawInits {
			fs = append(fs, k)
		}
		return fmt.Errorf("package has more than one init function:\n%s", strings.Join(fs, "\n"))
	}

	for k, v := range pkg.RawInits {
		if strings.Index(v, ";") != -1 {
			return fmt.Errorf("init function in %s has more than one statement", k)
		}
	}
	return nil
}

func (env *Environment) Lint(pkg *gdf.Package) error {
	return lintInit(pkg)
}
