package main

import (
	"fmt"
	"os"

	"gesprekken/dbsetup"
	"gesprekken/handlers"
	"gesprekken/routes"
	"gesprekken/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun/extra/bundebug"
)

func main() {
	// Laad .env configuratie
	if err := godotenv.Load(); err != nil {
		fmt.Println("Geen .env bestand gevonden (bestaande environment variabelen worden gebruikt)")
	}

	// Database verbinding
	db, err := dbsetup.ConnectToDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database verbinding mislukt: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// SQL debug logging
	if os.Getenv("BUNDEBUG") == "1" {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	// Tabellen aanmaken
	if err := dbsetup.CreateTables(db); err != nil {
		fmt.Fprintf(os.Stderr, "Tabellen aanmaken mislukt: %v\n", err)
		os.Exit(1)
	}

	// Gin instellen
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}
	r := gin.Default()

	// Handlers en routes registreren
	h := handlers.New(db)
	routes.Register(r, h)

	// Object storage (MinIO) voor documentbijlagen
	store, err := storage.NewMinIOStorage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "MinIO verbinding mislukt: %v\n", err)
		os.Exit(1)
	}
	dh := handlers.NewDocumentHandler(db, store)
	routes.RegisterDocumentRoutes(r, dh)

	// Server starten
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Gesprekken API gestart op http://localhost:%s\n", port)
	fmt.Printf("OpenAPI spec beschikbaar op http://localhost:%s/v1/openapi.json\n", port)

	if err := r.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Server starten mislukt: %v\n", err)
		os.Exit(1)
	}
}
