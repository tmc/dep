package depcore

import (
	"database/sql"
)

// the the exports of the given package that were imported
// by other packages
func (ø *db) GetImported(importedPkg string) (imps []*imp, err error) {
	var rows *sql.Rows
	imps = []*imp{}
	rows, err = ø.Query("select package, import, name, value from imports where import = ?", importedPkg)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		i := &imp{}
		err = rows.Scan(&i.Package, &i.Import, &i.Name, &i.Value)
		if err != nil {
			return
		}
		imps = append(imps, i)
	}
	return
}

func (ø *db) GetPackage(packagePath string, withExports bool, withImports bool) (p *dbPkg, exps []*exp, imps []*imp, err error) {
	var row *sql.Row
	p = &dbPkg{}
	row = ø.QueryRow("select package,  initmd5, jsonmd5, json from packages where package = ? limit 1", packagePath)
	err = row.Scan(&p.Package, &p.InitMd5, &p.JsonMd5, &p.Json)
	if err != nil {
		return
	}
	if withExports {
		var rows *sql.Rows
		exps = []*exp{}
		rows, err = ø.Query("select package, name, value from exports where package = ?", packagePath)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			e := &exp{}
			err = rows.Scan(&e.Package, &e.Name, &e.Value)
			if err != nil {
				return
			}
			exps = append(exps, e)
		}
	}

	if withImports {
		var rows *sql.Rows
		imps = []*imp{}
		rows, err = ø.Query("select package, import, name, value from imports where package = ?", packagePath)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			i := &imp{}
			err = rows.Scan(&i.Package, &i.Import, &i.Name, &i.Value)
			if err != nil {
				return
			}
			imps = append(imps, i)
		}
	}
	return
}

func (ø *db) GetAllPackages() (ps []*dbPkg, err error) {
	var rows *sql.Rows
	ps = []*dbPkg{}
	rows, err = ø.Query("select package, initmd5, jsonmd5, json from packages order by package asc")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		p := &dbPkg{}
		err = rows.Scan(&p.Package, &p.InitMd5, &p.JsonMd5, &p.Json)
		if err != nil {
			return
		}
		ps = append(ps, p)
	}
	return
}

func (ø *db) GetAllImports() (is []*imp, err error) {
	var rows *sql.Rows
	is = []*imp{}
	rows, err = ø.Query("select package, import, name, value from imports")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		i := &imp{}
		err = rows.Scan(&i.Package, &i.Import, &i.Name, &i.Value)
		if err != nil {
			return
		}
		is = append(is, i)
	}
	return
}

func (ø *db) GetAllExports() (es []*exp, err error) {
	var rows *sql.Rows
	es = []*exp{}
	rows, err = ø.Query("select package, name, value from exports")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		e := &exp{}
		err = rows.Scan(&e.Package, &e.Name, &e.Value)
		if err != nil {
			return
		}
		es = append(es, e)
	}
	return
}
