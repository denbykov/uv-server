package data

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
	"uv_server/internal/uv_server/business/data"
	"uv_server/internal/uv_server/common/loggers"

	"github.com/sirupsen/logrus"

	gfw "uv_server/internal/uv_server/business/workflows/get_files/job_messages"
	gsw "uv_server/internal/uv_server/business/workflows/get_settings/job_messages"
	ssw "uv_server/internal/uv_server/business/workflows/set_settings/job_messages"
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

func (d *Database) GetSorageDir() (string, error) {
	var dir string

	statement := `
	SELECT
		storage_dir
	FROM app
	`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	err := d.db.QueryRow(statement).Scan(&dir)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		d.log.Errorf("failed to get storage dir: %v", err)
		return "", fmt.Errorf("failed to get storage dir")
	}

	return dir, nil
}

func (d *Database) SetSorageDir(pathDir string) error {
	statement := `
	UPDATE app
		SET storage_dir =?
	`

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	_, err := d.db.Exec(statement, pathDir)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to set storage dir: %v", err)
		return fmt.Errorf("failed to set storage dir")
	}

	return nil
}

func getStructFieldsAndValues(structValue reflect.Value) ([]string, []interface{}) {
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	setFields := make([]string, 0, structValue.NumField())
	values := make([]interface{}, 0, structValue.NumField())

	for i, field := range reflect.VisibleFields(structValue.Type()) {
		jsonTag := field.Tag.Get("json")
		dbColumn := jsonTag
		if dbColumn == "" {
			dbColumn = strings.ToLower(field.Name)
		}
		setFields = append(setFields, fmt.Sprintf("%s = ?", dbColumn))
		values = append(values, structValue.Field(i).Interface())
	}

	return setFields, values
}

func getStructFieldsAndDestinations(structPtr interface{}) ([]string, []interface{}) {
	structValue := reflect.ValueOf(structPtr)
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	structType := structValue.Type()
	fields := make([]string, 0, structType.NumField())
	destinations := make([]interface{}, 0, structType.NumField())

	for _, field := range reflect.VisibleFields(structType) {
		jsonTag := field.Tag.Get("json")
		dbColumn := jsonTag
		if dbColumn == "" {
			dbColumn = strings.ToLower(field.Name)
		}
		fields = append(fields, dbColumn)
		destinations = append(destinations, structValue.FieldByName(field.Name).Addr().Interface())
	}

	return fields, destinations
}

func (d *Database) SetSettingsForSSW(request *ssw.Request) (*ssw.Result, error) {
	setFields, values := getStructFieldsAndValues(reflect.ValueOf(request.Settings))
	statement := "UPDATE app SET " + strings.Join(setFields, ", ")

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	_, err := d.db.Exec(statement, values...)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to set settings: %v", err)
		return nil, fmt.Errorf("failed to set settings")
	}
	result := &ssw.Result{}

	fields, destinations := getStructFieldsAndDestinations(&result.Settings)

	selectStatement := fmt.Sprintf(`
	SELECT
		%s
	FROM app
	`, strings.Join(fields, ",\n\t\t"))

	d.log.Debugf("executing statement: %v", selectStatement)

	err = d.db.QueryRow(selectStatement).Scan(destinations...)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil {
		d.log.Errorf("failed to get settings: %v", err)
		return nil, fmt.Errorf("failed to get settings")
	}

	return result, nil
}

func (d *Database) GetSettingsForGSW() (*gsw.Result, error) {
	result := &gsw.Result{}
	fields, destinations := getStructFieldsAndDestinations(result)
	statement := fmt.Sprintf(`
	SELECT
		%s
	FROM app
	`, strings.Join(fields, ",\n\t\t"))

	d.log.Debugf("executing statement: %v", statement)
	startedAt := time.Now()

	err := d.db.QueryRow(statement).Scan(destinations...)

	d.log.Debugf("execution took %v us", time.Since(startedAt).Microseconds())

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		d.log.Errorf("failed to get settings: %v", err)
		return nil, fmt.Errorf("failed to get settings")
	}
	return result, nil
}
