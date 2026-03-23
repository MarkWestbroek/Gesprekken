package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// Gesprek is het root-object: een conversatie met een onderwerp en tijdsperiode.
type Gesprek struct {
	bun.BaseModel `bun:"table:gesprekken,alias:g"`

	ID        uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"  json:"id"`
	Onderwerp string     `bun:"onderwerp,notnull"                          json:"onderwerp"`
	Aanvang   time.Time  `bun:"aanvang,notnull,type:timestamptz"           json:"aanvang"`
	Einde     *time.Time `bun:"einde,type:timestamptz"                     json:"einde,omitempty"`

	// Relaties
	Deelnames []GesprekDeelname  `bun:"rel:has-many,join:id=gesprek_id" json:"deelnames,omitempty"`
	Bijdragen []Gespreksbijdrage `bun:"rel:has-many,join:id=gesprek_id" json:"bijdragen,omitempty"`
}

// Gespreksdeelnemer is een persoon/actor die kan deelnemen aan gesprekken.
// Referentie is een URN die de deelnemer uniek identificeert buiten dit systeem.
type Gespreksdeelnemer struct {
	bun.BaseModel `bun:"table:gespreksdeelnemers,alias:gd"`

	ID         uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Naam       string    `bun:"naam,notnull"                               json:"naam"`
	Referentie string    `bun:"referentie,notnull"                         json:"referentie"`
}

// GesprekDeelname is de associatieklasse tussen Gesprek en Gespreksdeelnemer.
// Het legt de materiële meer-op-meer relatie vast inclusief aanvang en optioneel einde.
type GesprekDeelname struct {
	bun.BaseModel `bun:"table:gesprek_deelnames,alias:gdn"`

	ID          uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	GesprekID   uuid.UUID  `bun:"gesprek_id,notnull,type:uuid"              json:"gesprekId"`
	DeelnemerID uuid.UUID  `bun:"deelnemer_id,notnull,type:uuid"            json:"deelnemerId"`
	Aanvang     time.Time  `bun:"aanvang,notnull,type:timestamptz"          json:"aanvang"`
	Einde       *time.Time `bun:"einde,type:timestamptz"                    json:"einde,omitempty"`

	// Navigatie
	Gesprek   *Gesprek           `bun:"rel:belongs-to,join:gesprek_id=id"   json:"-"`
	Deelnemer *Gespreksdeelnemer `bun:"rel:belongs-to,join:deelnemer_id=id" json:"deelnemer,omitempty"`
}

// Gespreksbijdrage is een bijdrage (bericht) binnen een gesprek, geleverd door één deelnemer.
// Tekst bevat markdown-content.
type Gespreksbijdrage struct {
	bun.BaseModel `bun:"table:gespreksbijdragen,alias:gb"`

	ID          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	GesprekID   uuid.UUID `bun:"gesprek_id,notnull,type:uuid"              json:"gesprekId"`
	BijdragerID uuid.UUID `bun:"bijdrager_id,notnull,type:uuid"           json:"bijdragerId"`
	Geleverd    time.Time `bun:"geleverd,notnull,type:timestamptz"         json:"geleverd"`
	Tekst       string    `bun:"tekst,notnull"                             json:"tekst"`

	// Navigatie
	Gesprek   *Gesprek           `bun:"rel:belongs-to,join:gesprek_id=id"  json:"-"`
	Bijdrager *Gespreksdeelnemer `bun:"rel:belongs-to,join:bijdrager_id=id" json:"bijdrager,omitempty"`
	Lezingen  []BijdrageLezing   `bun:"rel:has-many,join:id=bijdrage_id"   json:"lezingen,omitempty"`
}

// BijdrageLezing registreert dat een deelnemer (niet de bijdrager) een bijdrage heeft gelezen.
type BijdrageLezing struct {
	bun.BaseModel `bun:"table:bijdrage_lezingen,alias:bl"`

	ID         uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	BijdrageID uuid.UUID `bun:"bijdrage_id,notnull,type:uuid"             json:"bijdrageId"`
	LezerID    uuid.UUID `bun:"lezer_id,notnull,type:uuid"                json:"lezerId"`
	GelezenOp  time.Time `bun:"gelezen_op,notnull,type:timestamptz"       json:"gelezenOp"`

	// Navigatie
	Bijdrage *Gespreksbijdrage  `bun:"rel:belongs-to,join:bijdrage_id=id" json:"-"`
	Lezer    *Gespreksdeelnemer `bun:"rel:belongs-to,join:lezer_id=id"    json:"lezer,omitempty"`
}
