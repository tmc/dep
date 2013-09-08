package db

import (
	"database/sql"
)

func InsertPackages(p []*Pkg, e []*Exp, im []*Imp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = db.Begin()
	if err != nil {
		return
	}
	stmt, err := tx.Prepare("insert into packages(package, importsmd5, exportsmd5, mainmd5, initmd5, json, jsonmd5) values(?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	for i := 0; i < len(p); i++ {
		_, err = stmt.Exec(p[i].Package, p[i].ImportsMd5, p[i].ExportsMd5, p[i].MainMd5, p[i].InitMd5, p[i].Json, p[i].JsonMd5)
		if err != nil {
			return
		}
	}

	err = _insertExports(tx, e...)
	if err != nil {
		return
	}
	err = _insertImports(tx, im...)
	if err != nil {
		return
	}

	tx.Commit()
	return
}

func _insertExports(tx *sql.Tx, e ...*Exp) (err error) {
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("insert into exports(package, name, value) values(?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	for i := 0; i < len(e); i++ {
		_, err = stmt.Exec(e[i].Package, e[i].Name, e[i].Value)
		if err != nil {
			return
		}
	}
	return
}

func InsertExports(e ...*Exp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = db.Begin()
	if err != nil {
		return
	}
	err = _insertExports(tx, e...)
	if err != nil {
		return
	}
	tx.Commit()
	return
}

func _insertImports(tx *sql.Tx, im ...*Imp) (err error) {
	var stmt *sql.Stmt
	stmt, err = tx.Prepare("insert into imports(package, import, name, value) values(?, ?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	for i := 0; i < len(im); i++ {
		_, err = stmt.Exec(im[i].Package, im[i].Import, im[i].Name, im[i].Value)
		if err != nil {
			return err
		}
	}
	return
}

func InsertImports(im ...*Imp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = db.Begin()
	if err != nil {
		return
	}
	err = _insertImports(tx, im...)
	if err != nil {
		return
	}
	tx.Commit()
	return
}
