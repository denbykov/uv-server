package data

import "database/sql"

type Resources struct {
	Db       *sql.DB
	To_clean chan<- string
}
