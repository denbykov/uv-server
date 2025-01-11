package data

import (
	"server/common/loggers"
	"server/config"

	"github.com/sirupsen/logrus"
)

var db_version int

type Migrator struct {
	log  *logrus.Entry
	repo *MigrationRepository
}

func NewMigrator(
	config *config.Config,
	repo *MigrationRepository) *Migrator {
	object := &Migrator{}

	object.log = loggers.PresentationLogger.
		WithField("component", "Migrator")
	object.repo = repo

	return object
}

func (m *Migrator) MigrateIfNeeded() {
	m.log.Trace("Checking for migration")

	ver, err := m.repo.GetVersion()

	if err != nil {
		m.log.Fatal(err)
	}

	if db_version == ver {
		m.log.Info("Migration is not needed")
	}

	if db_version < ver {
		m.log.Fatalf("Migration from %v to %v is impossible", ver, db_version)
	}

	m.log.Infof("Migrating from %v to %v", ver, db_version)
}
