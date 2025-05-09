package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"uv_server/internal/uv_server/business/data"
	"uv_server/internal/uv_server/common/loggers"

	"github.com/sirupsen/logrus"

	gfw "uv_server/internal/uv_server/business/workflows/get_files/job_messages"
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

func (d *Database) GetFilesForGFW(request *gfw.Request) (*gfw.Result, error) {
	result := &gfw.Result{}

	statement :=
		`
		SELECT 
			COUNT (*) 
		FROM files
		`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	err := d.db.QueryRow(statement).Scan(
		&result.Total,
	)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		d.log.Errorf("failed to get file count: %v", err)
		return result, fmt.Errorf("failed to get files")
	}

	statement = fmt.Sprintf(
		`
		SELECT 
			f.id,
			f.source,
			f.status,
			f.added_at
		FROM files as f
		ORDER BY f.added_at DESC
		LIMIT %v
		OFFSET %v
		`,
		*request.Limit,
		*request.Offset,
	)

	d.log.Debugf("executing statement: %v", statement)
	startedAt = time.Now()

	rows, err := d.db.Query(statement)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		d.log.Errorf("failed to get files: %v", err)
		return result, fmt.Errorf("failed to get files")
	}
	defer rows.Close()

	for rows.Next() {
		var file gfw.File

		err = rows.Scan(&file.Id, &file.Source, &file.Status, &file.AddedAt)
		if err != nil {
			d.log.Errorf("failed to scan files: %v", err)
			return result, fmt.Errorf("failed to get files")
		}
		result.Files = append(result.Files, file)
	}

	return result, nil
}

func (d *Database) GetSettings() (*data.Settings, error) {
	var settings data.Settings

	statement := `
	SELECT
		storage_dir
	FROM settings
	`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	err := d.db.QueryRow(statement).Scan(
		&settings.StorageDir,
	)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		d.log.Errorf("failed to get settings: %v", err)
		return nil, fmt.Errorf("failed to get settings")
	}

	return &settings, nil
}

func (d *Database) UpdateSettings(request *data.Settings) (*data.Settings, error) {
	tx, err := d.db.Begin()
	if err != nil {
		d.log.Errorf("failed to begin transaction: %v", err)
		return nil, fmt.Errorf("failed to update settings")
	}
	defer tx.Rollback()

	deleteStmt := `DELETE FROM settings`

	d.log.Debugf("executing statement: %v", deleteStmt)
	startedAt := time.Now()

	_, err = tx.Exec(deleteStmt)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to delete existing settings: %v", err)
		return nil, fmt.Errorf("failed to update settings")
	}

	insertStmt := `
    INSERT INTO settings (
        storage_dir
    ) VALUES (
        ?
    )`

	d.log.Debugf("executing statement: %v", insertStmt)
	startedAt = time.Now()

	_, err = tx.Exec(insertStmt, request.StorageDir)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to insert new settings: %v", err)
		return nil, fmt.Errorf("failed to update settings")
	}

	if err = tx.Commit(); err != nil {
		d.log.Errorf("failed to commit transaction: %v", err)
		return nil, fmt.Errorf("failed to update settings")
	}

	return request, nil
}
