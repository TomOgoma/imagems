package roach

import (
	"github.com/tomogoma/imagems/pkg/model"
)

func (r *Roach) SaveMeta(m model.ImageMeta) (int64, error) {

	if err := r.InitDBIfNot(); err != nil {
		return -1, err
	}

	cols := ColDesc(ColUserID, ColType, ColMimeType, ColWidth, ColHeight,
		ColCreateDate, ColUpdateDate)
	q := `
	INSERT INTO ` + TblImageMeta + ` (` + cols + `)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING ` + ColID + `
	`
	var ID int64
	err := r.db.QueryRow(q, m.UserID, m.Type, m.MimeType, m.Width, m.Height).
		Scan(&ID)

	return ID, err
}

func (r *Roach) DeleteMeta(id int64) error {

	if err := r.InitDBIfNot(); err != nil {
		return err
	}

	q := `
	UPDATE ` + TblImageMeta + `
		SET ` + ColDeleted + `=TRUE, ` + ColUpdateDate + `=CURRENT_TIMESTAMP
		WHERE ` + ColID + `=$1
	`
	rslt, err := r.db.Exec(q, id)
	return checkRowsAffected(rslt, err, 1)
}
