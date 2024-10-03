package models

import (
	"database/sql"
)

type Song struct {
	ID          int64        `db:"id"`
	Song        string       `db:"song"`
	Group       string       `db:"nameGroup"`
	Text        string       `db:"text"`
	ReleaseDate sql.NullTime `db:"release_date"`
	Link        string       `db:"link"`
}
