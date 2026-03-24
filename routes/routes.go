package routes

import (
	"gesprekken/handlers"

	"github.com/gin-gonic/gin"
)

const apiVersion = "1.0.0"

// versionHeader middleware voegt API-Version header toe aan elk antwoord
// conform NL API Design Rules /core/version-header.
func versionHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("API-Version", apiVersion)
		c.Next()
	}
}

// securityHeaders voegt verplichte HTTP security headers toe
// conform NL API Design Rules /core/transport/security-headers.
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Security-Policy", "frame-ancestors 'none'")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	}
}

// corsMiddleware staat cross-origin requests toe
// conform NL API Design Rules /core/transport/cors.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Accept")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// Register koppelt alle routes aan de Gin engine.
// URI-structuur conform NL API Design Rules:
//   - /core/uri-version: major versie in URI (/v1)
//   - /core/naming-collections: meervoud voor collecties
//   - /core/path-segments-kebab-case: lowercase, kebab-case
//   - /core/nested-child: geneste URIs voor child resources
//   - /core/no-trailing-slash: geen trailing slash
func Register(r *gin.Engine, h *handlers.Handler) {
	r.Use(corsMiddleware(), versionHeader(), securityHeaders())

	v1 := r.Group("/v1")

	// OAS document op standaardlocatie (/core/publish-openapi)
	v1.StaticFile("/openapi.json", "./openapi.json")

	// Gesprekken
	v1.GET("/gesprekken", h.ListGesprekken)
	v1.POST("/gesprekken", h.CreateGesprek)
	v1.GET("/gesprekken/:id", h.GetGesprek)
	v1.DELETE("/gesprekken/:id", h.DeleteGesprek)

	// Gespreksdeelnemers (top-level: bestaan onafhankelijk)
	v1.GET("/gespreksdeelnemers", h.ListGespreksdeelnemers)
	v1.POST("/gespreksdeelnemers", h.CreateGespreksdeelnemer)
	v1.GET("/gespreksdeelnemers/:id", h.GetGespreksdeelnemer)
	v1.PUT("/gespreksdeelnemers/:id", h.UpdateGespreksdeelnemer)
	v1.DELETE("/gespreksdeelnemers/:id", h.DeleteGespreksdeelnemer)

	// Deelnemertypen (opzoektabel, read-only via API)
	v1.GET("/deelnemertypen", h.ListDeelnemertypen)
	v1.GET("/deelnemertypen/:id", h.GetDeelnemertype)

	// Deelnames (genest onder gesprekken — child resource)
	v1.GET("/gesprekken/:id/deelnames", h.ListDeelnames)
	v1.POST("/gesprekken/:id/deelnames", h.CreateDeelname)
	v1.PATCH("/gesprekken/:id/deelnames/:deelnameId", h.UpdateDeelname)

	// Bijdragen (genest onder gesprekken)
	v1.GET("/gesprekken/:id/bijdragen", h.ListBijdragen)
	v1.POST("/gesprekken/:id/bijdragen", h.CreateBijdrage)
	v1.GET("/gesprekken/:id/bijdragen/:bijdrageId", h.GetBijdrage)

	// Lezingen (genest onder bijdragen)
	v1.GET("/gesprekken/:id/bijdragen/:bijdrageId/lezingen", h.ListLezingen)
	v1.POST("/gesprekken/:id/bijdragen/:bijdrageId/lezingen", h.CreateLezing)
}

// RegisterDocumentRoutes koppelt document-upload/download/metadata routes.
func RegisterDocumentRoutes(r *gin.Engine, dh *handlers.DocumentHandler) {
	v1 := r.Group("/v1")
	v1.Use(corsMiddleware(), versionHeader(), securityHeaders())

	// Documenten
	v1.POST("/documenten", dh.UploadDocument)
	v1.GET("/documenten/:id", dh.GetDocumentMetadata)
	v1.GET("/documenten/:id/download", dh.DownloadDocument)
}
