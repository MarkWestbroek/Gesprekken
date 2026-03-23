package dbsetup

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"gesprekken/model"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// EnsureDatabaseExists maakt de database aan als die nog niet bestaat.
// Gebruikt DATABASE_ADMIN_URL om te verbinden met de postgres admin-database.
func EnsureDatabaseExists(dsn string) error {
	adminDSN := os.Getenv("DATABASE_ADMIN_URL")
	if adminDSN == "" {
		return fmt.Errorf("DATABASE_ADMIN_URL is niet geconfigureerd")
	}

	dbName := extractDBName(dsn)
	if dbName == "" {
		return fmt.Errorf("kan database-naam niet bepalen uit DATABASE_URL")
	}

	adminDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(adminDSN)))
	defer adminDB.Close()

	var exists bool
	err := adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("controle database bestaan mislukt: %w", err)
	}

	if !exists {
		// Database-naam is geëxtraheerd uit onze eigen DSN, geen user input
		_, err = adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, dbName))
		if err != nil {
			return fmt.Errorf("database aanmaken mislukt: %w", err)
		}
		fmt.Printf("Database '%s' aangemaakt\n", dbName)
	}

	return nil
}

// CreateTables maakt alle tabellen aan in de juiste volgorde (FK-afhankelijkheden).
func CreateTables(db *bun.DB) error {
	ctx := context.Background()

	// Zorg dat de uuid-ossp extensie beschikbaar is
	_, err := db.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)
	if err != nil {
		return fmt.Errorf("pgcrypto extensie aanmaken mislukt: %w", err)
	}

	models := []interface{}{
		(*model.Gesprek)(nil),
		(*model.Gespreksdeelnemer)(nil),
		(*model.GesprekDeelname)(nil),
		(*model.Gespreksbijdrage)(nil),
		(*model.BijdrageLezing)(nil),
	}

	for _, m := range models {
		_, err := db.NewCreateTable().
			Model(m).
			IfNotExists().
			WithForeignKeys().
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("tabel aanmaken mislukt voor %T: %w", m, err)
		}
	}

	fmt.Println("Alle tabellen aangemaakt (of bestonden al)")
	return nil
}

// ConnectToDatabase verbindt met de PostgreSQL database via Bun.
func ConnectToDatabase() (*bun.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is niet geconfigureerd")
	}

	if os.Getenv("AUTO_CREATE_DATABASE") == "true" {
		if err := EnsureDatabaseExists(dsn); err != nil {
			return nil, err
		}
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	// Verbinding testen
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database verbinding mislukt: %w", err)
	}

	return db, nil
}

func extractDBName(dsn string) string {
	// Format: postgres://user:pass@host:port/dbname?params
	parts := strings.Split(dsn, "/")
	if len(parts) < 4 {
		return ""
	}
	dbPart := parts[len(parts)-1]
	if idx := strings.Index(dbPart, "?"); idx != -1 {
		dbPart = dbPart[:idx]
	}
	return dbPart
}
