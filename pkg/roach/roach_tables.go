package roach

const (
	Version = 1

	TblConfigurations = "configurations"
	TblImageMeta      = "image_meta"

	ColID         = "ID"
	ColUserID     = "user_id"
	ColType       = "type"
	ColMimeType   = "mime_type"
	ColWidth      = "width"
	ColHeight     = "height"
	ColDeleted    = "deleted"
	ColKey        = "key"
	ColValue      = "value"
	ColCreateDate = "create_date"
	ColUpdateDate = "update_date"

	// CREATE TABLE DESCRIPTIONS
	TblDescConfigurations = `
	CREATE TABLE IF NOT EXISTS ` + TblConfigurations + ` (
		` + ColKey + ` VARCHAR(56) PRIMARY KEY NOT NULL CHECK (` + ColKey + ` != ''),
		` + ColValue + ` BYTEA NOT NULL CHECK (` + ColValue + ` != ''),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`

	TblDescImageMeta = `
	CREATE TABLE IF NOT EXISTS ` + TblImageMeta + ` (
		` + ColID + ` SERIAL PRIMARY KEY,
		` + ColUserID + ` INT NOT NULL,
		` + ColType + ` STRING,
		` + ColMimeType + ` STRING,
		` + ColWidth + ` FLOAT,
		` + ColHeight + ` FLOAT,
		` + ColCreateDate + ` TIMESTAMP NOT NULL,
		` + ColUpdateDate + ` TIMESTAMP NOT NULL,
		` + ColDeleted + ` BOOL NOT NULL DEFAULT FALSE,
		INDEX(` + ColUserID + `),
		INDEX(` + ColType + `),
		INDEX(` + ColMimeType + `)
	);
	`
)

var (
	// TblNames lists all table names in order of dependency
	// (tables with foreign key references listed after parent table descriptions).
	TblNames = []string{
		TblConfigurations,
		TblImageMeta,
	}

	// TblDescs lists all CREATE TABLE DESCRIPTIONS in order of dependency
	// (tables with foreign key references listed after parent table descriptions).
	TblDescs = []string{
		TblDescConfigurations,
		TblDescImageMeta,
	}
)
