package data

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
)

var db_version int = 2

type Migrator struct {
	log        *logrus.Entry
	config     *config.Config
	db         *sql.DB
	changesets map[int]func() error
}

func NewMigrator(config *config.Config, db *sql.DB) *Migrator {
	object := &Migrator{}

	object.log = loggers.PresentationLogger.
		WithField("component", "db_migrator")
	object.config = config

	object.db = db
	object.changesets = make(map[int]func() error)
	object.registerChangesets(config.ChangesetsLocation)

	return object
}

func (r *Migrator) GetVersion() (int, error) {
	statement := `
	SELECT name FROM sqlite_master
	WHERE type='table'
	AND name=$1;
	`

	r.log.Debugf("executing statement: %v", statement)

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

func (m *Migrator) setDbVersion(version int) error {
	statement := `
	INSERT OR REPLACE INTO app (id, db_version)
	VALUES (1, $1);
	`

	_, err := m.db.Exec(statement, version)

	if err != nil {
		return fmt.Errorf("failed to set db version: %v", err)
	}

	return nil
}

func (m *Migrator) executeFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	queries := strings.Split(string(content), ";")

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		_, err := m.db.Exec(query)
		if err != nil {
			return fmt.Errorf("error executing query %q: %w", query, err)
		}
	}

	return tx.Commit()
}

func (m *Migrator) parseFileVersion(filename string) (int, error) {
	base := filepath.Base(filename)

	re := regexp.MustCompile(`^(\d+)`)
	match := re.FindStringSubmatch(base)

	if len(match) < 2 {
		return 0, fmt.Errorf("invalid filename format: %s", filename)
	}

	version, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse version number: %w", err)
	}

	return version, nil
}

func (m *Migrator) registerChangesets(dir string) {
	entries, err := os.ReadDir(dir)

	if err != nil {
		m.log.Fatal(err)
	}

	for _, e := range entries {
		m.log.Tracef("registering changeset: %v", e.Name())

		ver, err := m.parseFileVersion(e.Name())

		if err != nil {
			m.log.Fatal(err)
		}

		m.changesets[ver] = func() error {
			return m.executeFile(path.Join(dir, e.Name()))
		}
	}
}

func (m *Migrator) MigrateIfNeeded() {
	m.log.Trace("checking for migration")

	ver, err := m.GetVersion()

	m.log.Infof("current db version: %v", ver)
	m.log.Infof("required db version: %v", db_version)

	if err != nil {
		m.log.Fatal(err)
	}

	if db_version == ver {
		m.log.Info("migration is not needed")
		return
	}

	if db_version < ver {
		m.log.Fatalf("migration from %v to %v is impossible", ver, db_version)
	}

	m.log.Infof("Migrating from %v to %v", ver, db_version)

	for v := ver; v < db_version; v++ {
		nextVer := v + 1

		m.log.Infof("migrating to %v", nextVer)

		err := m.changesets[nextVer]()
		if err != nil {
			m.log.Fatalf("migration to %v has failed: %v", nextVer, err)
		}

		err = m.setDbVersion(nextVer)
		if err != nil {
			m.log.Fatalf("migration to %v has failed: %v", nextVer, err)
		}
	}
}
