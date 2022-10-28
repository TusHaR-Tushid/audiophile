package database

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/sirupsen/logrus"

	// for importing migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
)

var (
	AudiophileDB *sqlx.DB
)

type SSLMode string

const (
	SSLModeEnable  SSLMode = "enable"
	SSLModeDisable SSLMode = "disable"
)

// ConnectAndMigrate function connects with a given database and returns error if any
func ConnectAndMigrate(host, port, databaseName, user, password string, sslMode SSLMode) error {
	connStr := fmt.Sprintf("host=%s  port=%s  user=%s  password=%s  dbname=%s  sslmode=%s", host, port, user, password, databaseName, sslMode)
	DB, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return err
	}
	err = DB.Ping()
	if err != nil {
		return err
	}
	AudiophileDB = DB
	return migrateUp(DB)
}

func ShutdownDatabase() error {
	return AudiophileDB.Close()
}

// migrate func migrate the database and handles the migration logic
func migrateUp(db *sqlx.DB) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}
	migrate1, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"postgres", driver)
	if err != nil {
		return err
	}
	if err := migrate1.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

// Tx provides the transaction wrapper
func Tx(fn func(tx *sqlx.Tx) error) error {
	tx, err := AudiophileDB.Beginx()
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %+v", err)
	}
	defer func() {
		if err != nil {
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				logrus.Errorf("failed to rollback tx: %s", rollBackErr)
			}
			return
		}
		if commitErr := tx.Commit(); commitErr != nil {
			logrus.Errorf("failed to commit tx: %s", commitErr)
		}
	}()
	err = fn(tx)
	return err
}
