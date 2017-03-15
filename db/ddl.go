package db

const (
	imageMetaTable = `
CREATE TABLE IF NOT EXISTS imageMeta (
	ID SERIAL PRIMARY KEY,
	userID INT NOT NULL,
	type STRING,
	mimeType STRING,
	width FLOAT,
	height FLOAT,
	dateCreated TIMESTAMP NOT NULL,
	dateUpdated TIMESTAMP NOT NULL,
	deleted BOOL NOT NULL DEFAULT FALSE,
	INDEX(userID),
	INDEX(type),
	INDEX(mimeType)
);
`
)
