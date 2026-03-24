package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gesprekken/model"
	"gesprekken/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Maximale bestandsgrootte: 25 MB
const maxUploadSize = 25 << 20

// Toegestane MIME-types voor uploads
var allowedMIMETypes = map[string]bool{
	"image/jpeg":         true,
	"image/png":          true,
	"image/gif":          true,
	"image/webp":         true,
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain": true,
}

// DocumentHandler bevat de storage-referentie naast de database.
type DocumentHandler struct {
	DB      *bun.DB
	Storage storage.ObjectStorage
}

func NewDocumentHandler(db *bun.DB, store storage.ObjectStorage) *DocumentHandler {
	return &DocumentHandler{DB: db, Storage: store}
}

// UploadDocument verwerkt een multipart/form-data upload en slaat het bestand
// op in object storage met metadata in de database.
//
//	POST /v1/documenten
//	Form fields: bestand (file), naam, brontype, bronId, bronUrn, bronUrl
func (dh *DocumentHandler) UploadDocument(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize+1024)

	file, header, err := c.Request.FormFile("bestand")
	if err != nil {
		if err.Error() == "http: request body too large" {
			problemJSON(c, http.StatusRequestEntityTooLarge, "Bestand te groot",
				fmt.Sprintf("Maximale bestandsgrootte is %d MB.", maxUploadSize/(1<<20)))
			return
		}
		problemJSON(c, http.StatusBadRequest, "Upload mislukt", "Veld 'bestand' is verplicht.")
		return
	}
	defer file.Close()

	// MIME-type bepalen en valideren
	contentType := header.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = mimeFromExt(strings.ToLower(filepath.Ext(header.Filename)))
	}
	if !allowedMIMETypes[contentType] {
		problemJSON(c, http.StatusBadRequest, "Ongeldig bestandstype",
			fmt.Sprintf("MIME-type '%s' wordt niet ondersteund.", contentType))
		return
	}

	// Verplichte velden uitlezen
	naam := c.PostForm("naam")
	if naam == "" {
		naam = header.Filename
	}
	bronType := c.PostForm("brontype")
	if bronType == "" {
		bronType = "gespreksbijlage"
	}

	bronIDStr := c.PostForm("bronId")
	bronID, err := uuid.Parse(bronIDStr)
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig bronId",
			"bronId is verplicht en moet een geldig UUID zijn.")
		return
	}

	bronUrn := c.PostForm("bronUrn")
	bronUrl := c.PostForm("bronUrl")

	// Genereer storage key: <brontype>/<docId><extensie>
	docID := uuid.New()
	ext := strings.ToLower(filepath.Ext(header.Filename))
	bucketKey := fmt.Sprintf("%s/%s%s", bronType, docID.String(), ext)

	// Upload naar object storage
	if err := dh.Storage.Upload(c.Request.Context(), bucketKey, file, header.Size, contentType); err != nil {
		problemJSON(c, http.StatusInternalServerError, "Opslag mislukt",
			"Bestand kon niet worden opgeslagen in object storage.")
		return
	}

	// Metadata opslaan in database
	now := time.Now().UTC()
	doc := model.Document{
		ID:           docID,
		Naam:         naam,
		BronType:     bronType,
		BronID:       bronID,
		BronUrn:      bronUrn,
		BronUrl:      bronUrl,
		ContentType:  contentType,
		Grootte:      header.Size,
		BucketKey:    bucketKey,
		OpgeslagenOp: now,
	}

	_, err = dh.DB.NewInsert().Model(&doc).Exec(c.Request.Context())
	if err != nil {
		// Rollback: verwijder bestand uit storage als DB-insert faalt
		_ = dh.Storage.Delete(c.Request.Context(), bucketKey)
		problemJSON(c, http.StatusInternalServerError, "Opslag mislukt",
			"Metadata kon niet worden opgeslagen.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"bestandId":    doc.ID,
		"opgeslagenOp": doc.OpgeslagenOp,
		"bestandUrn":   fmt.Sprintf("urn:gesprekken:document:%s", doc.ID),
		"downloadUrl":  fmt.Sprintf("/v1/documenten/%s/download", doc.ID),
	})
}

// DownloadDocument streamt een bestand uit object storage met het juiste
// content-type en filename header.
//
//	GET /v1/documenten/:id/download
func (dh *DocumentHandler) DownloadDocument(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven ID is geen geldig UUID.")
		return
	}

	doc := new(model.Document)
	err = dh.DB.NewSelect().Model(doc).Where("doc.id = ?", id).Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Document niet gevonden.")
		return
	}

	obj, err := dh.Storage.Download(c.Request.Context(), doc.BucketKey)
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Download mislukt",
			"Bestand kon niet worden opgehaald uit object storage.")
		return
	}
	defer obj.Close()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.Naam))
	c.DataFromReader(http.StatusOK, doc.Grootte, doc.ContentType, obj, nil)
}

// GetDocumentMetadata retourneert de metadata van een document.
//
//	GET /v1/documenten/:id
func (dh *DocumentHandler) GetDocumentMetadata(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven ID is geen geldig UUID.")
		return
	}

	doc := new(model.Document)
	err = dh.DB.NewSelect().Model(doc).Where("doc.id = ?", id).Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Document niet gevonden.")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bestandId":    doc.ID,
		"naam":         doc.Naam,
		"brontype":     doc.BronType,
		"bronId":       doc.BronID,
		"bronUrn":      doc.BronUrn,
		"bronUrl":      doc.BronUrl,
		"contentType":  doc.ContentType,
		"grootte":      doc.Grootte,
		"opgeslagenOp": doc.OpgeslagenOp,
		"bijdrageId":   doc.BijdrageID,
		"bestandUrn":   fmt.Sprintf("urn:gesprekken:document:%s", doc.ID),
		"downloadUrl":  fmt.Sprintf("/v1/documenten/%s/download", doc.ID),
	})
}

// mimeFromExt geeft het MIME-type terug op basis van de bestandsextensie.
func mimeFromExt(ext string) string {
	types := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
	}
	if t, ok := types[ext]; ok {
		return t
	}
	return "application/octet-stream"
}
