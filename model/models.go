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

// Deelnemertype is een opzoektabel die het type deelnemer classificeert.
// Voorbeelden: "interne_actor" (medewerker) en "partij" (externe deelnemer).
type Deelnemertype struct {
	bun.BaseModel `bun:"table:deelnemertypen,alias:dt"`

	ID   uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Code string    `bun:"code,notnull,unique"                       json:"code"`
	Naam string    `bun:"naam,notnull"                               json:"naam"`
}

// Gespreksdeelnemer is een persoon/actor die kan deelnemen aan gesprekken.
// Referentie is een URN die de deelnemer uniek identificeert buiten dit systeem.
// TypeID verwijst naar het Deelnemertype (interne_actor of partij).
type Gespreksdeelnemer struct {
	bun.BaseModel `bun:"table:gespreksdeelnemers,alias:gd"`

	ID         uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Naam       string    `bun:"naam,notnull"                               json:"naam"`
	Referentie string    `bun:"referentie,notnull"                         json:"referentie"`
	TypeID     uuid.UUID `bun:"type_id,notnull,type:uuid"                  json:"typeId"`

	// Navigatie
	Type *Deelnemertype `bun:"rel:belongs-to,join:type_id=id" json:"type,omitempty"`
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

	ID              uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	GesprekID       uuid.UUID  `bun:"gesprek_id,notnull,type:uuid"              json:"gesprekId"`
	BijdragerID     uuid.UUID  `bun:"bijdrager_id,notnull,type:uuid"           json:"bijdragerId"`
	Geleverd        time.Time  `bun:"geleverd,notnull,type:timestamptz"         json:"geleverd"`
	Tekst           string     `bun:"tekst,notnull"                             json:"tekst"`
	ReactieOpID     *uuid.UUID `bun:"reactie_op_id,type:uuid"                   json:"reactieOpId,omitempty"`
	LaatstBewerktOp *time.Time `bun:"laatst_bewerkt_op,type:timestamptz"        json:"laatstBewerktOp,omitempty"`
	Teruggetrokken  bool       `bun:"teruggetrokken,notnull,default:false"      json:"teruggetrokken"`

	// Navigatie
	Gesprek   *Gesprek           `bun:"rel:belongs-to,join:gesprek_id=id"    json:"-"`
	Bijdrager *Gespreksdeelnemer `bun:"rel:belongs-to,join:bijdrager_id=id"  json:"bijdrager,omitempty"`
	ReactieOp *Gespreksbijdrage  `bun:"rel:belongs-to,join:reactie_op_id=id" json:"reactieOp,omitempty"`
	Lezingen  []BijdrageLezing   `bun:"rel:has-many,join:id=bijdrage_id"     json:"lezingen,omitempty"`
	Bijlagen  []Document         `bun:"rel:has-many,join:id=bijdrage_id"     json:"bijlagen,omitempty"`
}

// Document bevat metadata voor een geüpload bestand in object storage.
type Document struct {
	bun.BaseModel `bun:"table:documenten,alias:doc"`

	ID           uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"bestandId"`
	Naam         string     `bun:"naam,notnull"                               json:"naam"`
	BronType     string     `bun:"bron_type,notnull"                          json:"brontype"`
	BronID       uuid.UUID  `bun:"bron_id,notnull,type:uuid"                  json:"bronId"`
	BronUrn      string     `bun:"bron_urn,notnull"                           json:"bronUrn"`
	BronUrl      string     `bun:"bron_url,notnull"                           json:"bronUrl"`
	ContentType  string     `bun:"content_type,notnull"                       json:"contentType"`
	Grootte      int64      `bun:"grootte,notnull"                            json:"grootte"`
	BucketKey    string     `bun:"bucket_key,notnull"                         json:"-"`
	OpgeslagenOp time.Time  `bun:"opgeslagen_op,notnull,type:timestamptz"     json:"opgeslagenOp"`
	BijdrageID   *uuid.UUID `bun:"bijdrage_id,type:uuid"                     json:"bijdrageId,omitempty"`

	// Navigatie
	Bijdrage *Gespreksbijdrage `bun:"rel:belongs-to,join:bijdrage_id=id" json:"-"`
}

// GespreksbijdrageVersie bewaart eerdere versies van een bewerkte bijdrage.
// Bij elke bewerking wordt de oude tekst hier opgeslagen als audit trail.
type GespreksbijdrageVersie struct {
	bun.BaseModel `bun:"table:gespreksbijdrage_versies,alias:gbv"`

	ID          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	BijdrageID  uuid.UUID `bun:"bijdrage_id,notnull,type:uuid"             json:"bijdrageId"`
	Versie      int       `bun:"versie,notnull"                            json:"versie"`
	Tekst       string    `bun:"tekst,notnull"                             json:"tekst"`
	GewijzigdOp time.Time `bun:"gewijzigd_op,notnull,type:timestamptz"     json:"gewijzigdOp"`

	// Navigatie
	Bijdrage *Gespreksbijdrage `bun:"rel:belongs-to,join:bijdrage_id=id" json:"-"`
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
