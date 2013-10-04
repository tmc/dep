package depcore

import (
	"fmt"
	"github.com/metakeule/gdf"
	"os"
	"path/filepath"
)

func (o *Environment) Get(pkgPath string, overrides []*gdf.Package, confirmation func(candidates ...*gdf.Package) bool) (conflicts map[string]map[string][3]string, changed map[string][2]string, err error) {
	o.Open()
	defer o.Close()

	t := o.newTentative()
	var er error
	conflicts, changed, er = t.updatePackage(pkgPath, overrides, confirmation)

	if len(conflicts) > 0 {
		er = fmt.Errorf("Error: there are %v conflicts", len(conflicts))
	}

	if er != nil {
		dir := filepath.Dir(t.GOPATH)
		new_path := filepath.Join(dir, fmt.Sprintf(TempGOPATHPreFix+"%v", now()))
		os.Rename(t.GOPATH, new_path)
		err = fmt.Errorf(er.Error()+"\ncheck or remove the temporary gopath at %s\n", new_path)
	}
	return
}
