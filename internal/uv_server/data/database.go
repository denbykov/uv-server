package data

import (
	"database/sql"
	"errors"
	"time"
	"uv_server/internal/uv_server/business/data"
	"uv_server/internal/uv_server/common/loggers"

	"github.com/sirupsen/logrus"
)

type Database struct {
	log *logrus.Entry
	db  *sql.DB
}

func NewDatabase(db *sql.DB) *Database {
	object := &Database{}

	object.log = loggers.DataLogger.
		WithField("component", "datbaase")
	object.db = db

	return object
}

func (d *Database) GetFileByUrl(url string) (*data.File, error) {
	var file data.File

	statement := `
	SELECT 
		id,
		"path",
		source_url,
		"source",
		status,
		added_at,
		updated_at
	FROM files
		WHERE source_url=?
	`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	err := d.db.QueryRow(statement, url).Scan(
		&file.Id,
		&file.Path,
		&file.SourceUrl,
		&file.Source,
		&file.Status,
		&file.AddedAt,
		&file.UpdatedAt,
	)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		d.log.Errorf("failed to get file by url: %v", err)
		return nil, err
	}

	return &file, nil
}

func (d *Database) InsertFile(file *data.File) (int64, error) {
	statement := `
	INSERT INTO files (
		path,
		source_url,
		source,
		status
	)
	VALUES (
		?,
		?,
		?,
		?
	)
	RETURNING id
	`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	result, err := d.db.Exec(statement,
		file.Path,
		file.SourceUrl,
		file.Source,
		file.Status,
	)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to insert a file: %v", err)
		return 0, err
	}

	return result.LastInsertId()
}

func (d *Database) UpdateFileStatus(file *data.File) error {
	statement := `
	UPDATE files
		SET status = ?
	WHERE
		id = ?
	`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	_, err := d.db.Exec(statement,
		file.Status,
		file.Id,
	)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to update file status: %v", err)
		return nil
	}

	return nil
}

func (d *Database) UpdateFilePath(file *data.File) error {
	statement := `
	UPDATE files
		SET path = ?
	WHERE
		id = ?
	`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	_, err := d.db.Exec(statement,
		file.Path,
		file.Id,
	)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to update file path: %v", err)
		return err
	}

	return nil
}

func (d *Database) DeleteFile(file *data.File) error {
	statement := `
	DELETE FROM files
	WHERE
		id = ?
	`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	_, err := d.db.Exec(statement,
		file.Id,
	)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to delete file: %v", err)
		return err
	}

	return nil
}
