package db

import (
	"database/sql"
)

func GetImported(db *sql.DB, importedPkg string) (imps []*Imp, err error) {
	var rows *sql.Rows
	imps = []*Imp{}
	rows, err = db.Query("select package, import, name, value from imports where import = ?", importedPkg)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		i := &Imp{}
		err = rows.Scan(&i.Package, &i.Import, &i.Name, &i.Value)
		if err != nil {
			return
		}
		imps = append(imps, i)
	}
	return
}

func GetPackage(db *sql.DB, packagePath string, withExports bool, withImports bool) (p *Pkg, exps []*Exp, imps []*Imp, err error) {
	var row *sql.Row
	p = &Pkg{}
	row = db.QueryRow("select package, importsmd5, exportsmd5,  initmd5, jsonmd5, json from packages where package = ? limit 1", packagePath)
	err = row.Scan(&p.Package, &p.ImportsMd5, &p.ExportsMd5, &p.InitMd5, &p.JsonMd5, &p.Json)
	if err != nil {
		return
	}
	if withExports {
		var rows *sql.Rows
		exps = []*Exp{}
		rows, err = db.Query("select package, name, value from exports where package = ?", packagePath)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			e := &Exp{}
			err = rows.Scan(&e.Package, &e.Name, &e.Value)
			if err != nil {
				return
			}
			exps = append(exps, e)
		}
	}

	if withImports {
		var rows *sql.Rows
		imps = []*Imp{}
		rows, err = db.Query("select package, import, name, value from imports where package = ?", packagePath)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			i := &Imp{}
			err = rows.Scan(&i.Package, &i.Import, &i.Name, &i.Value)
			if err != nil {
				return
			}
			imps = append(imps, i)
		}
	}
	return
}

func GetAllPackages(db *sql.DB) (ps []*Pkg, err error) {
	var rows *sql.Rows
	ps = []*Pkg{}
	rows, err = db.Query("select package, importsmd5, exportsmd5,  initmd5, jsonmd5, json from packages")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		p := &Pkg{}
		err = rows.Scan(&p.Package, &p.ImportsMd5, &p.ExportsMd5, &p.InitMd5, &p.JsonMd5, &p.Json)
		if err != nil {
			return
		}
		ps = append(ps, p)
	}
	return
}

func GetAllImports(db *sql.DB) (is []*Imp, err error) {
	var rows *sql.Rows
	is = []*Imp{}
	rows, err = db.Query("select package, import, name, value from imports")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		i := &Imp{}
		err = rows.Scan(&i.Package, &i.Import, &i.Name, &i.Value)
		if err != nil {
			return
		}
		is = append(is, i)
	}
	return
}

func GetAllExports(db *sql.DB) (es []*Exp, err error) {
	var rows *sql.Rows
	es = []*Exp{}
	rows, err = db.Query("select package, name, value from exports")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		e := &Exp{}
		err = rows.Scan(&e.Package, &e.Name, &e.Value)
		if err != nil {
			return
		}
		es = append(es, e)
	}
	return
}
