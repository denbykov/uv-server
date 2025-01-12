package data

import (
	"database/sql"
	"fmt"
	"server/common/loggers"

	"github.com/sirupsen/logrus"
)

type MigrationRepository struct {
	log *logrus.Entry
	db  *sql.DB
}

func NewMigrationRepositry(db *sql.DB) *MigrationRepository {
	object := &MigrationRepository{}

	object.log = loggers.DataLogger.
		WithField("component", "MigrationRepository")
	object.db = db

	return object
}

// Gets
func (r *MigrationRepository) GetVersion() (int, error) {
	statement := `
	SELECT name FROM sqlite_master
	WHERE type='table'
	AND name=$1;
	`

	r.log.Debugf("Executing statement: %v", statement)

	var dummy interface{}
	err := r.db.QueryRow(statement, "app").Scan(&dummy)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0,
			fmt.Errorf("failed to check for %v table existance: %v",
				"app",
				err)
	}

	statement = `
	SELECT db_version FROM app; 
	`

	r.log.Debugf("Executing statement: %v", statement)

	var version int
	err = r.db.QueryRow(statement).Scan(&version)

	if err != nil {
		return 0, fmt.Errorf("failed to get app version: %v", err)
	}

	return version, nil
}

func (r *MigrationRepository) From0to1() error {
	statement := `
	CREATE TABLE app (
		db_version INTEGER NOT NULL
	)
	`

	r.log.Debugf("Executing statement: %v", statement)

	_, err := r.db.Exec(statement)

	if err != nil {
		return fmt.Errorf("failed to create app table: %v", err)
	}

	statement = `
	INSERT INTO app (db_version)
	VALUES($1)
	`

	_, err = r.db.Exec(statement, 1)

	if err != nil {
		return fmt.Errorf("failed to insert app record: %v", err)
	}

	return nil
}

// func (r *MigrationRepository) From1to2() error {
// 	statement := `
// 	create table files (
// 		person_id INTEGER PRIMARY KEY,
// 		path VARCHAR(255) NULL UNIQUE,
// 		source_url VARCHAR(255) NOT NULL UNIQUE,
// 		status VARCHAR(1) NOT NULL,
// 		added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
// 		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
// 	)
// 	`

// 	return nil
// }
