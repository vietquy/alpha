package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

// Config defines the options that are used when connecting to a PostgreSQL instance
type Config struct {
	Host        string
	Port        string
	User        string
	Pass        string
	Name        string
}

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate
// failure.
func Connect(cfg Config) (*sqlx.DB, error) {
	url := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s  sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass)

	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := migrateDB(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDB(db *sqlx.DB) error {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "things_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS things (
						id       UUID,
						owner    VARCHAR(254),
						key      VARCHAR(4096) UNIQUE NOT NULL,
						name     VARCHAR(1024),
						metadata JSON,
						PRIMARY KEY (id, owner)
					)`,
					`CREATE TABLE IF NOT EXISTS projects (
						id       UUID,
						owner    VARCHAR(254),
						name     VARCHAR(1024),
						metadata JSON,
						PRIMARY KEY (id, owner)
					)`,
					`CREATE TABLE IF NOT EXISTS connections (
						project_id    UUID,
						project_owner VARCHAR(254),
						thing_id      UUID,
						thing_owner   VARCHAR(254),
						FOREIGN KEY (project_id, project_owner) REFERENCES projects (id, owner) ON DELETE CASCADE ON UPDATE CASCADE,
						FOREIGN KEY (thing_id, thing_owner) REFERENCES things (id, owner) ON DELETE CASCADE ON UPDATE CASCADE,
						PRIMARY KEY (project_id, project_owner, thing_id, thing_owner)
					)`,
				},
				Down: []string{
					"DROP TABLE connections",
					"DROP TABLE things",
					"DROP TABLE projects",
				},
			},
			{
				Id: "things_2",
				Up: []string{
					`ALTER TABLE IF EXISTS things ALTER COLUMN
					 metadata TYPE JSONB using metadata::text::jsonb
					`,
				},
			},
			{
				Id: "things_3",
				Up: []string{
					`ALTER TABLE IF EXISTS projects ALTER COLUMN
					 metadata TYPE JSONB using metadata::text::jsonb
					`,
				},
			},
		},
	}

	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}
