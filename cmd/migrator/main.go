package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationsPath, migrationsTable string

	flag.StringVar(&storagePath, "storage-path", "", "the path of storage")
	flag.StringVar(&migrationsPath, "migration-path", "", "the path of mirations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "table name with migarions")
	flag.Parse()

	if storagePath == "" {
		panic("flag storage-path is required")
	}
	if migrationsPath == "" {
		panic("flag migration-path is required")
	}

	m, err := migrate.New("file://"+migrationsPath, fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable))
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
