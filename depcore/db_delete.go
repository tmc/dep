package depcore

import (
	"database/sql"
	"fmt"
)

func _deleteExports(pkgPath string, tx *sql.Tx) (err error) {
	_, err = tx.Exec(`delete from exports where package = ?`, pkgPath)
	return
}

func _deleteImports(pkgPath string, tx *sql.Tx) (err error) {
	_, err = tx.Exec(`delete from imports where package = ?`, pkgPath)
	return
}

func _deletePackage(pkgPath string, tx *sql.Tx) (err error) {
	_, err = tx.Exec(`delete from packages where package = ?`, pkgPath)
	return
}

func (ø *db) DeletePackage(pkgPath string) (err error) {
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
	err = _deleteExports(pkgPath, tx)
	if err != nil {
		return
	}
	err = _deleteImports(pkgPath, tx)
	if err != nil {
		return
	}

	err = _deletePackage(pkgPath, tx)
	if err != nil {
		return
	}
	tx.Commit()
	fmt.Printf("deleted: %s\n", pkgPath)
	return
}
