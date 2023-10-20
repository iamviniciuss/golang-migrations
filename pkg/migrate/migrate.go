package migrate

import (
	"time"

	"github.com/Vinicius-Santos-da-Silva/mongo-migrate/pkg/entity"
	"go.mongodb.org/mongo-driver/mongo"
)

const defaultMigrationsCollection = "migrations"

const AllAvailable = -1

type MigrationRepository interface {
	Insert(rec *entity.VersionRecord) (*entity.VersionRecord, error)
	FindAll() ([]*entity.VersionRecord, error)
	FindOne() (*entity.VersionRecord, error)
	CreateCollectionIfNotExists(name string) error
}

type Migrate struct {
	migrationRepository  MigrationRepository
	db                   *mongo.Database
	migrations           []Migration
	migrationsCollection string
}

func NewMigrate(db *mongo.Database, migrations ...Migration) *Migrate {
	internalMigrations := make([]Migration, len(migrations))
	copy(internalMigrations, migrations)
	return &Migrate{
		db:                   db,
		migrations:           internalMigrations,
		migrationsCollection: defaultMigrationsCollection,
	}
}

func (m *Migrate) SetMigrationsCollection(name string) {
	m.migrationsCollection = name
}

func (m *Migrate) Version() (uint64, string, error) {
	if err := m.migrationRepository.CreateCollectionIfNotExists(m.migrationsCollection); err != nil {
		return 0, "", err
	}

	rec, err := m.migrationRepository.FindOne()

	if err != nil {
		return 0, "", nil
	}

	return rec.Version, rec.Description, nil
}

func (m *Migrate) SetVersion(version uint64, description string, typing string) error {
	rec := &entity.VersionRecord{
		Version:     version,
		Timestamp:   time.Now().UTC(),
		Description: description,
		Type:        typing,
	}

	m.migrationRepository.Insert(rec)

	return nil
}

func (m *Migrate) Up(n int) error {
	currentVersion, _, err := m.Version()
	if err != nil {
		return err
	}
	if n <= 0 || n > len(m.migrations) {
		n = len(m.migrations)
	}
	migrationSort(m.migrations)

	for i, p := 0, 0; i < len(m.migrations) && p < n; i++ {
		migration := m.migrations[i]
		if migration.Version <= currentVersion || migration.Handler == nil {
			continue
		}
		p++
		if err := migration.Handler.Up(); err != nil {
			return err
		}
		if err := m.SetVersion(migration.Version, migration.Description, migration.Handler.GetType()); err != nil {
			return err
		}
	}
	return nil
}

func (m *Migrate) Down(n int) error {
	currentVersion, _, err := m.Version()
	if err != nil {
		return err
	}
	if n <= 0 || n > len(m.migrations) {
		n = len(m.migrations)
	}
	migrationSort(m.migrations)

	for i, p := len(m.migrations)-1, 0; i >= 0 && p < n; i-- {
		migration := m.migrations[i]
		if migration.Version > currentVersion || migration.Handler == nil {
			continue
		}
		p++
		if err := migration.Handler.Down(); err != nil {
			return err
		}

		var prevMigration Migration
		if i == 0 {
			prevMigration = Migration{Version: 0}
		} else {
			prevMigration = m.migrations[i-1]
		}
		if err := m.SetVersion(prevMigration.Version, prevMigration.Description, migration.Handler.GetType()); err != nil {
			return err
		}
	}
	return nil
}