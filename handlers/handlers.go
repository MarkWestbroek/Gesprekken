package handlers

import (
	"net/http"
	"time"

	"gesprekken/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ProblemDetail implementeert RFC 9457 (Problem Details for HTTP APIs)
// conform NL API Design Rules /core/error-handling/problem-details.
type ProblemDetail struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func problemJSON(c *gin.Context, status int, title, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, ProblemDetail{Status: status, Title: title, Detail: detail})
}

// Handler houdt de database-referentie voor alle handlers.
type Handler struct {
	DB *bun.DB
}

func New(db *bun.DB) *Handler {
	return &Handler{DB: db}
}

// ──────────────────── Gesprekken ────────────────────

func (h *Handler) ListGesprekken(c *gin.Context) {
	var gesprekken []model.Gesprek
	err := h.DB.NewSelect().Model(&gesprekken).
		Relation("Deelnames").
		Relation("Deelnames.Deelnemer").
		OrderExpr("g.aanvang DESC").
		Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Interne fout", err.Error())
		return
	}
	c.JSON(http.StatusOK, gesprekken)
}

func (h *Handler) GetGesprek(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven ID is geen geldig UUID.")
		return
	}

	gesprek := new(model.Gesprek)
	err = h.DB.NewSelect().Model(gesprek).
		Where("g.id = ?", id).
		Relation("Deelnames").
		Relation("Deelnames.Deelnemer").
		Relation("Bijdragen").
		Relation("Bijdragen.Bijdrager").
		Relation("Bijdragen.Lezingen").
		Relation("Bijdragen.Lezingen.Lezer").
		Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Gesprek niet gevonden.")
		return
	}
	c.JSON(http.StatusOK, gesprek)
}

type CreateGesprekInput struct {
	Onderwerp string     `json:"onderwerp" binding:"required"`
	Aanvang   time.Time  `json:"aanvang"   binding:"required"`
	Einde     *time.Time `json:"einde"`
}

func (h *Handler) CreateGesprek(c *gin.Context) {
	var input CreateGesprekInput
	if err := c.ShouldBindJSON(&input); err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldige invoer", err.Error())
		return
	}

	gesprek := model.Gesprek{
		Onderwerp: input.Onderwerp,
		Aanvang:   input.Aanvang,
		Einde:     input.Einde,
	}
	_, err := h.DB.NewInsert().Model(&gesprek).Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Aanmaken mislukt", err.Error())
		return
	}
	c.JSON(http.StatusCreated, gesprek)
}

func (h *Handler) DeleteGesprek(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven ID is geen geldig UUID.")
		return
	}
	res, err := h.DB.NewDelete().Model((*model.Gesprek)(nil)).Where("id = ?", id).Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Verwijderen mislukt", err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Gesprek niet gevonden.")
		return
	}
	c.Status(http.StatusNoContent)
}

// ──────────────────── Deelnemertypen (opzoektabel) ────────────────────

// ListDeelnemertypen retourneert alle deelnemertypen (GET /deelnemertypen).
// Dit is een read-only opzoektabel met waarden als "interne_actor" en "partij".
func (h *Handler) ListDeelnemertypen(c *gin.Context) {
	var typen []model.Deelnemertype
	err := h.DB.NewSelect().Model(&typen).OrderExpr("dt.naam ASC").Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Interne fout", err.Error())
		return
	}
	c.JSON(http.StatusOK, typen)
}

// GetDeelnemertype retourneert een enkel deelnemertype op basis van UUID.
func (h *Handler) GetDeelnemertype(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven ID is geen geldig UUID.")
		return
	}
	dt := new(model.Deelnemertype)
	err = h.DB.NewSelect().Model(dt).Where("dt.id = ?", id).Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Deelnemertype niet gevonden.")
		return
	}
	c.JSON(http.StatusOK, dt)
}

// ──────────────────── Gespreksdeelnemers ────────────────────

// ListGespreksdeelnemers retourneert alle gespreksdeelnemers inclusief
// hun deelnemertype (Relation "Type"), gesorteerd op naam.
func (h *Handler) ListGespreksdeelnemers(c *gin.Context) {
	var deelnemers []model.Gespreksdeelnemer
	err := h.DB.NewSelect().Model(&deelnemers).
		Relation("Type").
		OrderExpr("gd.naam ASC").
		Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Interne fout", err.Error())
		return
	}
	c.JSON(http.StatusOK, deelnemers)
}

// GetGespreksdeelnemer retourneert een enkele deelnemer inclusief type.
func (h *Handler) GetGespreksdeelnemer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven ID is geen geldig UUID.")
		return
	}
	deelnemer := new(model.Gespreksdeelnemer)
	err = h.DB.NewSelect().Model(deelnemer).
		Where("gd.id = ?", id).
		Relation("Type").
		Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Gespreksdeelnemer niet gevonden.")
		return
	}
	c.JSON(http.StatusOK, deelnemer)
}

type CreateGespreksdeelnemerInput struct {
	Naam       string    `json:"naam"       binding:"required"`
	Referentie string    `json:"referentie" binding:"required"`
	TypeID     uuid.UUID `json:"typeId"     binding:"required"`
}

func (h *Handler) CreateGespreksdeelnemer(c *gin.Context) {
	var input CreateGespreksdeelnemerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldige invoer", err.Error())
		return
	}
	deelnemer := model.Gespreksdeelnemer{
		Naam:       input.Naam,
		Referentie: input.Referentie,
		TypeID:     input.TypeID,
	}
	_, err := h.DB.NewInsert().Model(&deelnemer).Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Aanmaken mislukt", err.Error())
		return
	}
	c.JSON(http.StatusCreated, deelnemer)
}

// UpdateGespreksdeelnemer wijzigt naam, referentie en/of type van een deelnemer (PUT).
func (h *Handler) UpdateGespreksdeelnemer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven ID is geen geldig UUID.")
		return
	}
	var input CreateGespreksdeelnemerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldige invoer", err.Error())
		return
	}
	res, err := h.DB.NewUpdate().Model((*model.Gespreksdeelnemer)(nil)).
		Set("naam = ?", input.Naam).
		Set("referentie = ?", input.Referentie).
		Set("type_id = ?", input.TypeID).
		Where("id = ?", id).
		Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Bijwerken mislukt", err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Gespreksdeelnemer niet gevonden.")
		return
	}
	// Haal de bijgewerkte deelnemer op met type-relatie
	deelnemer := new(model.Gespreksdeelnemer)
	_ = h.DB.NewSelect().Model(deelnemer).Where("gd.id = ?", id).Relation("Type").Scan(c.Request.Context())
	c.JSON(http.StatusOK, deelnemer)
}

func (h *Handler) DeleteGespreksdeelnemer(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven ID is geen geldig UUID.")
		return
	}
	res, err := h.DB.NewDelete().Model((*model.Gespreksdeelnemer)(nil)).Where("id = ?", id).Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Verwijderen mislukt", err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Gespreksdeelnemer niet gevonden.")
		return
	}
	c.Status(http.StatusNoContent)
}

// ──────────────────── Deelnames (genest onder gesprekken) ────────────────────

func (h *Handler) ListDeelnames(c *gin.Context) {
	gesprekID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven gesprek-ID is geen geldig UUID.")
		return
	}
	var deelnames []model.GesprekDeelname
	err = h.DB.NewSelect().Model(&deelnames).
		Where("gdn.gesprek_id = ?", gesprekID).
		Relation("Deelnemer").
		OrderExpr("gdn.aanvang ASC").
		Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Interne fout", err.Error())
		return
	}
	c.JSON(http.StatusOK, deelnames)
}

type CreateDeelnameInput struct {
	DeelnemerID uuid.UUID  `json:"deelnemerId" binding:"required"`
	Aanvang     time.Time  `json:"aanvang"     binding:"required"`
	Einde       *time.Time `json:"einde"`
}

func (h *Handler) CreateDeelname(c *gin.Context) {
	gesprekID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven gesprek-ID is geen geldig UUID.")
		return
	}
	var input CreateDeelnameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldige invoer", err.Error())
		return
	}
	deelname := model.GesprekDeelname{
		GesprekID:   gesprekID,
		DeelnemerID: input.DeelnemerID,
		Aanvang:     input.Aanvang,
		Einde:       input.Einde,
	}
	_, err = h.DB.NewInsert().Model(&deelname).Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Aanmaken mislukt", err.Error())
		return
	}
	c.JSON(http.StatusCreated, deelname)
}

func (h *Handler) UpdateDeelname(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven gesprek-ID is geen geldig UUID.")
		return
	}
	deelnameID, err := uuid.Parse(c.Param("deelnameId"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven deelname-ID is geen geldig UUID.")
		return
	}

	var input struct {
		Einde *time.Time `json:"einde"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldige invoer", err.Error())
		return
	}

	res, err := h.DB.NewUpdate().Model((*model.GesprekDeelname)(nil)).
		Set("einde = ?", input.Einde).
		Where("id = ?", deelnameID).
		Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Bijwerken mislukt", err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Deelname niet gevonden.")
		return
	}
	c.Status(http.StatusNoContent)
}

// ──────────────────── Bijdragen (genest onder gesprekken) ────────────────────

func (h *Handler) ListBijdragen(c *gin.Context) {
	gesprekID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven gesprek-ID is geen geldig UUID.")
		return
	}
	var bijdragen []model.Gespreksbijdrage
	err = h.DB.NewSelect().Model(&bijdragen).
		Where("gb.gesprek_id = ?", gesprekID).
		Relation("Bijdrager").
		Relation("Lezingen").
		Relation("Lezingen.Lezer").
		OrderExpr("gb.geleverd ASC").
		Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Interne fout", err.Error())
		return
	}
	c.JSON(http.StatusOK, bijdragen)
}

func (h *Handler) GetBijdrage(c *gin.Context) {
	bijdrageID, err := uuid.Parse(c.Param("bijdrageId"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven bijdrage-ID is geen geldig UUID.")
		return
	}
	bijdrage := new(model.Gespreksbijdrage)
	err = h.DB.NewSelect().Model(bijdrage).
		Where("gb.id = ?", bijdrageID).
		Relation("Bijdrager").
		Relation("Lezingen").
		Relation("Lezingen.Lezer").
		Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusNotFound, "Niet gevonden", "Bijdrage niet gevonden.")
		return
	}
	c.JSON(http.StatusOK, bijdrage)
}

type CreateBijdrageInput struct {
	BijdragerID uuid.UUID `json:"bijdragerId" binding:"required"`
	Geleverd    time.Time `json:"geleverd"    binding:"required"`
	Tekst       string    `json:"tekst"       binding:"required"`
}

func (h *Handler) CreateBijdrage(c *gin.Context) {
	gesprekID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven gesprek-ID is geen geldig UUID.")
		return
	}
	var input CreateBijdrageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldige invoer", err.Error())
		return
	}
	bijdrage := model.Gespreksbijdrage{
		GesprekID:   gesprekID,
		BijdragerID: input.BijdragerID,
		Geleverd:    input.Geleverd,
		Tekst:       input.Tekst,
	}
	_, err = h.DB.NewInsert().Model(&bijdrage).Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Aanmaken mislukt", err.Error())
		return
	}
	c.JSON(http.StatusCreated, bijdrage)
}

// ──────────────────── Lezingen (genest onder bijdragen) ────────────────────

func (h *Handler) ListLezingen(c *gin.Context) {
	bijdrageID, err := uuid.Parse(c.Param("bijdrageId"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven bijdrage-ID is geen geldig UUID.")
		return
	}
	var lezingen []model.BijdrageLezing
	err = h.DB.NewSelect().Model(&lezingen).
		Where("bl.bijdrage_id = ?", bijdrageID).
		Relation("Lezer").
		OrderExpr("bl.gelezen_op ASC").
		Scan(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Interne fout", err.Error())
		return
	}
	c.JSON(http.StatusOK, lezingen)
}

type CreateLezingInput struct {
	LezerID   uuid.UUID `json:"lezerId"   binding:"required"`
	GelezenOp time.Time `json:"gelezenOp" binding:"required"`
}

func (h *Handler) CreateLezing(c *gin.Context) {
	bijdrageID, err := uuid.Parse(c.Param("bijdrageId"))
	if err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldig ID", "Het opgegeven bijdrage-ID is geen geldig UUID.")
		return
	}
	var input CreateLezingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		problemJSON(c, http.StatusBadRequest, "Ongeldige invoer", err.Error())
		return
	}
	lezing := model.BijdrageLezing{
		BijdrageID: bijdrageID,
		LezerID:    input.LezerID,
		GelezenOp:  input.GelezenOp,
	}
	_, err = h.DB.NewInsert().Model(&lezing).Exec(c.Request.Context())
	if err != nil {
		problemJSON(c, http.StatusInternalServerError, "Aanmaken mislukt", err.Error())
		return
	}
	c.JSON(http.StatusCreated, lezing)
}
