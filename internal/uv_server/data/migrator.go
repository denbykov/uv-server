package data

import (
	"github.com/sirupsen/logrus"

	"uv_server/internal/uv_server/common/loggers"
	"uv_server/internal/uv_server/config"
)

var db_version int = 1

type Migrator struct {
	log        *logrus.Entry
	repo       *MigrationRepository
	changesets map[int]func() error
}

func NewMigrator(
	config *config.Config,
	repo *MigrationRepository) *Migrator {
	object := &Migrator{}

	object.log = loggers.PresentationLogger.
		WithField("component", "Migrator")
	object.repo = repo
	object.changesets = make(map[int]func() error)
	object.registerChangesets()

	return object
}

func (m *Migrator) registerChangesets() {
	m.changesets[0] = m.repo.From0to1
}

func (m *Migrator) MigrateIfNeeded() {
	m.log.Trace("Checking for migration")

	ver, err := m.repo.GetVersion()

	m.log.Infof("Current db version: %v", ver)
	m.log.Infof("Required db version: %v", db_version)

	if err != nil {
		m.log.Fatal(err)
	}

	if db_version == ver {
		m.log.Info("Migration is not needed")
		return
	}

	if db_version < ver {
		m.log.Fatalf("Migration from %v to %v is impossible", ver, db_version)
	}

	m.log.Infof("Migrating from %v to %v", ver, db_version)

	for v := ver; v < db_version; v++ {
		m.log.Debugf("Doing migration %v", v)

		err := m.changesets[v]()

		if err != nil {
			m.log.Fatalf("Migration %v has failed: %v", v, err)
		}
	}
}
