package roach

const (
	Version = 1

	TblConfigurations = "configurations"
	TblImageMeta      = "image_meta"
	TblAPIKeys        = "api_keys"

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

	TblDescAPIKeys = `
	CREATE TABLE IF NOT EXISTS ` + TblAPIKeys + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` VARCHAR(256) NOT NULL CHECK (` + ColUserID + ` != ''),
		` + ColKey + ` VARCHAR(256) NOT NULL CHECK ( LENGTH(` + ColKey + `) >= 56 ),
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL
	);
	`

	TblDescImageMeta = `
	CREATE TABLE IF NOT EXISTS ` + TblImageMeta + ` (
		` + ColID + ` BIGSERIAL PRIMARY KEY NOT NULL CHECK (` + ColID + `>0),
		` + ColUserID + ` BIGINT NOT NULL CHECK (` + ColUserID + `>0),
		` + ColType + ` VARCHAR(256),
		` + ColMimeType + ` VARCHAR(256),
		` + ColWidth + ` FLOAT,
		` + ColHeight + ` FLOAT,
		` + ColCreateDate + ` TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		` + ColUpdateDate + ` TIMESTAMPTZ NOT NULL,
		` + ColDeleted + ` BOOL NOT NULL DEFAULT FALSE
	);
	`
)

var (
	// TblNames lists all table names in order of dependency
	// (tables with foreign key references listed after parent table descriptions).
	TblNames = []string{
		TblConfigurations,
		TblAPIKeys,
		TblImageMeta,
	}

	// TblDescs lists all CREATE TABLE DESCRIPTIONS in order of dependency
	// (tables with foreign key references listed after parent table descriptions).
	TblDescs = []string{
		TblDescConfigurations,
		TblDescAPIKeys,
		TblDescImageMeta,
	}
)
