package db

import (
	"database/sql"
)

func UpdatePackage(db *sql.DB, p *Pkg, i []*Imp, e []*Exp) (err error) {
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

	_, err = tx.Exec(`
        update 
            packages
        set 
            importsmd5 = ?, 
            exportsmd5 = ?, 
            initmd5 = ?, 
            json = ?, 
            jsonmd5 = ?
        where
            package = ?
        `,
		p.ImportsMd5, p.ExportsMd5, p.InitMd5, p.Json, p.JsonMd5, p.Package)
	if err != nil {
		return
	}

	err = _deleteExports(p.Package, tx)
	if err != nil {
		return
	}

	err = _deleteImports(p.Package, tx)
	if err != nil {
		return
	}

	err = _insertImports(tx, i...)
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
