package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/tomogoma/go-commons/database/cockroach"
	"github.com/tomogoma/imagems/model"
)

type DB struct {
	db *sql.DB
}

func New(dsnF cockroach.DSNFormatter) (*DB, error) {
	d, err := cockroach.DBConn(dsnF)
	if err != nil {
		return nil, fmt.Errorf("error connecting to cockroach: %v", err)
	}
	err = cockroach.InstantiateDB(d, dsnF.DBName(), imageMetaTable)
	err = cockroach.CloseDBOnError(d, err)
	if err != nil {
		return nil, fmt.Errorf("error instantiating db vals: %v", err)
	}
	return &DB{db: d}, nil
}

func (d *DB) SaveMeta(m *model.ImageMeta) error {
	if m == nil {
		return errors.New("ImageMeta was nil")
	}
	q := `
	INSERT INTO imageMeta (userID, type, mimeType, width, height, dateCreated, dateUpdated)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING ID
	`
	return d.db.QueryRow(q, m.UserID, m.Type, m.MimeType, m.Width, m.Height).Scan(&m.ID)
}
func (d *DB) DeleteMeta(id int64) error {
	q := `
	UPDATE imageMeta SET deleted=TRUE, dateUpdated=CURRENT_TIMESTAMP
		WHERE ID=$1
	`
	rslt, err := d.db.Exec(q, id)
	if err != nil {
		return err
	}
	return checkNumRowsAffected(rslt, 1)
}

func checkNumRowsAffected(rslt sql.Result, expCount int64) error {
	c, err := rslt.RowsAffected()
	if err != nil {
		return fmt.Errorf("error asserting row count affected: %v", err)
	}
	if c != expCount {
		return errors.New("updated row count not as expected")
	}
	return nil
}
