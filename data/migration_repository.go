package data

import (
	"database/sql"
	"fmt"
	"log"
	"server/common/loggers"

	"github.com/sirupsen/logrus"
)

const (
	AppTable string = "app"
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
	select name from sqlite_master 
	where type='table' 
	and name=$1; 
	`

	r.log.Debugf("Executing statement: %v", statement)

	rows, err := r.db.Query(statement, AppTable)

	r.log.Debugf("Executing statement: %v", statement)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0,
			fmt.Errorf("failed to check for %v table existance: %v",
				AppTable,
				err)
	}

	defer rows.Close()

	var version int

	for rows.Next() {
		err = rows.Scan(&version)
		if err != nil {
			log.Fatal(err)
		}
	}

	return version, nil
}

// func From0to1() error {

// 	return nil
// }
