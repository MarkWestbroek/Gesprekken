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
		(*model.Deelnemertype)(nil),
		(*model.Gespreksdeelnemer)(nil),
		(*model.GesprekDeelname)(nil),
		(*model.Gespreksbijdrage)(nil),
		(*model.BijdrageLezing)(nil),
		(*model.Document)(nil),
		(*model.GespreksbijdrageVersie)(nil),
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

	// Indexes voor documenten lookup
	_, _ = db.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS idx_documenten_bron ON documenten (bron_type, bron_id)`)
	_, _ = db.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS idx_documenten_bijdrage ON documenten (bijdrage_id) WHERE bijdrage_id IS NOT NULL`)

	// Vul de opzoektabel deelnemertypen met standaardwaarden
	if err := seedDeelnemertypen(db); err != nil {
		return fmt.Errorf("seeden deelnemertypen mislukt: %w", err)
	}

	// Voeg type_id kolom toe aan gespreksdeelnemers als deze nog niet bestaat
	if err := migrateDeelnemerTypeID(db); err != nil {
		return fmt.Errorf("migratie type_id mislukt: %w", err)
	}

	// Voeg reactie_op_id kolom toe aan gespreksbijdragen als deze nog niet bestaat
	if err := migrateReactieOpID(db); err != nil {
		return fmt.Errorf("migratie reactie_op_id mislukt: %w", err)
	}

	// Voeg laatst_bewerkt_op kolom toe aan gespreksbijdragen voor bewerkhistorie
	if err := migrateLaatstBewerktOp(db); err != nil {
		return fmt.Errorf("migratie laatst_bewerkt_op mislukt: %w", err)
	}

	// Voeg teruggetrokken kolom toe aan gespreksbijdragen
	if err := migrateTeruggetrokken(db); err != nil {
		return fmt.Errorf("migratie teruggetrokken mislukt: %w", err)
	}

	// Index voor snel ophalen van versiehistorie per bijdrage
	_, _ = db.ExecContext(ctx,
		`CREATE INDEX IF NOT EXISTS idx_bijdrage_versies_bijdrage ON gespreksbijdrage_versies (bijdrage_id, versie)`)

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

// seedDeelnemertypen vult de opzoektabel met de standaard deelnemertypen
// als deze nog niet bestaan. Gebruikt ON CONFLICT DO NOTHING.
func seedDeelnemertypen(db *bun.DB) error {
	ctx := context.Background()
	typen := []model.Deelnemertype{
		{Code: "interne_actor", Naam: "Interne actor (medewerker)"},
		{Code: "partij", Naam: "Partij (externe deelnemer)"},
	}
	for _, t := range typen {
		_, err := db.NewInsert().Model(&t).
			On("CONFLICT (code) DO NOTHING").
			Exec(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// migrateDeelnemerTypeID voegt de type_id kolom toe aan bestaande
// gespreksdeelnemers-tabellen en vult lege waarden met het standaard
// type "interne_actor". Idempotent: slaat over als de kolom al bestaat.
func migrateDeelnemerTypeID(db *bun.DB) error {
	ctx := context.Background()
	// Controleer of de kolom al bestaat
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'gespreksdeelnemers' AND column_name = 'type_id'
		)`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		// Kolom bestaat al; vul eventuele NULLs met het standaard type
		_, err = db.ExecContext(ctx,
			`UPDATE gespreksdeelnemers SET type_id = (
				SELECT id FROM deelnemertypen WHERE code = 'interne_actor' LIMIT 1
			) WHERE type_id IS NULL`)
		return err
	}
	// Kolom nog niet aanwezig: voeg nullable toe, vul in, maak NOT NULL
	_, err = db.ExecContext(ctx,
		`ALTER TABLE gespreksdeelnemers ADD COLUMN type_id uuid REFERENCES deelnemertypen(id)`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx,
		`UPDATE gespreksdeelnemers SET type_id = (
			SELECT id FROM deelnemertypen WHERE code = 'interne_actor' LIMIT 1
		)`)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx,
		`ALTER TABLE gespreksdeelnemers ALTER COLUMN type_id SET NOT NULL`)
	return err
}

// migrateLaatstBewerktOp voegt de laatst_bewerkt_op kolom toe aan gespreksbijdragen
// zodat direct zichtbaar is of een bericht bewerkt is. Idempotent.
func migrateLaatstBewerktOp(db *bun.DB) error {
	ctx := context.Background()
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'gespreksbijdragen' AND column_name = 'laatst_bewerkt_op'
		)`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.ExecContext(ctx,
		`ALTER TABLE gespreksbijdragen ADD COLUMN laatst_bewerkt_op timestamptz`)
	return err
}

// migrateTeruggetrokken voegt de teruggetrokken kolom toe aan gespreksbijdragen
// zodat een bericht als teruggetrokken gemarkeerd kan worden. Idempotent.
func migrateTeruggetrokken(db *bun.DB) error {
	ctx := context.Background()
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'gespreksbijdragen' AND column_name = 'teruggetrokken'
		)`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.ExecContext(ctx,
		`ALTER TABLE gespreksbijdragen ADD COLUMN teruggetrokken boolean NOT NULL DEFAULT false`)
	return err
}

// migrateReactieOpID voegt de reactie_op_id kolom toe aan gespreksbijdragen
// zodat een bijdrage kan verwijzen naar een eerdere bijdrage (reply).
// Idempotent: slaat over als de kolom al bestaat.
func migrateReactieOpID(db *bun.DB) error {
	ctx := context.Background()
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'gespreksbijdragen' AND column_name = 'reactie_op_id'
		)`).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.ExecContext(ctx,
		`ALTER TABLE gespreksbijdragen ADD COLUMN reactie_op_id uuid REFERENCES gespreksbijdragen(id)`)
	return err
}
