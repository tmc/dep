package depcore

import (
	"database/sql"
)

func (ø *db) UpdatePackage(p *dbPkg, i []*imp, e []*exp) (err error) {
	var tx *sql.Tx
	defer func() {
		if err != nil && tx != nil {
			tx.Rollback()
		}
	}()
	tx, err = ø.Begin()
	if err != nil {
		return
	}

	_, err = tx.Exec(`
        update 
            packages
        set 
            initmd5 = ?, 
            json = ?, 
            jsonmd5 = ?
        where
            package = ?
        `,
		p.InitMd5, p.Json, p.JsonMd5, p.Package)
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
