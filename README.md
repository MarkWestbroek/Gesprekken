# Gesprekken API

REST API voor het registreren en raadplegen van **gesprekken**, de **deelnemers** daaraan en hun **bijdragen**. Gebouwd in Go met Gin en Bun (PostgreSQL), conform de [NLGov REST API Design Rules](https://logius-standaarden.github.io/API-Design-Rules/).

---

## 1. Functionele beschrijving

### Domein

Het systeem beheert drie kernbegrippen:

| Begrip | Omschrijving |
|---|---|
| **Gesprek** | Een conversatie met een onderwerp, een aanvangstijdstip en een optioneel eindtijdstip. Dit is het root-object. |
| **Gespreksdeelnemer** | Een persoon of actor, geïdentificeerd door een naam en een externe URN-referentie. Bestaat onafhankelijk van gesprekken. |
| **Gespreksbijdrage** | Een bericht (markdown-tekst) dat door één deelnemer op een bepaald moment binnen een gesprek wordt geleverd. |

### Relaties

```
Gespreksdeelnemer ◄──── GesprekDeelname ────► Gesprek
       (M)            aanvang / einde           (N)
        │
        │ bijdrager (1)
        ▼
  Gespreksbijdrage ◄───────────────────────── Gesprek
        │                  (0..*)
        │
        ▼ gelezen door (1..*)
  BijdrageLezing ──────► Gespreksdeelnemer
```

- **Gesprek ↔ Gespreksdeelnemer**: materiële meer-op-meer relatie via de associatieklasse `GesprekDeelname`. Zowel het moment van aanvang (verplicht) als het moment van einde (optioneel) van de deelname worden vastgelegd.
- **Gesprek → Gespreksbijdrage**: een gesprek bevat nul of meer bijdragen (1:N).
- **Gespreksbijdrage → Gespreksdeelnemer**: elke bijdrage heeft precies één bijdrager.
- **Gespreksbijdrage → BijdrageLezing**: een bijdrage kan door één of meer andere deelnemers (niet de bijdrager) worden gelezen. Per lezing wordt het tijdstip vastgelegd.

### Markdown als tekst-formaat

De tekst van een bijdrage is `string` in het datamodel. In de OAS-specificatie is `contentMediaType: "text/markdown"` gebruikt (een OAS 3.1 feature) om aan te geven dat de inhoud Markdown-geformateerd is. Dit is de aanbevolen aanpak: `string` als type, met metadata over het mediatype.

---

## 2. Architectuur

### Projectstructuur

```
Gesprekken/
├── main.go                 # Entrypoint: configuratie, database, Gin-server
├── .env                    # Environment variabelen (database, poort, debug)
├── go.mod / go.sum         # Go module definitie en dependency sums
├── openapi.json            # OAS 3.1 specificatie
│
├── model/
│   └── models.go           # Bun ORM structs (5 entiteiten)
│
├── dbsetup/
│   └── setup.go            # Database aanmaken, verbinden, tabellen migreren
│
├── handlers/
│   └── handlers.go         # Gin handlers (CRUD per resource)
│
└── routes/
    └── routes.go           # Route-registratie, CORS, versie- en security-headers
```

### Technologie-stack

| Laag | Technologie | Versie |
|---|---|---|
| Webframework | [Gin](https://github.com/gin-gonic/gin) | v1.12 |
| ORM | [Bun](https://bun.uptrace.dev/) | v1.2 |
| Database | PostgreSQL | 14+ |
| UUID | [google/uuid](https://github.com/google/uuid) | v1.6 |
| Configuratie | [godotenv](https://github.com/joho/godotenv) | v1.5 |

### Lagen

```
HTTP Request
     │
     ▼
  routes/       ← Middleware (CORS, API-Version, Security headers)
     │             Route-definitie (/v1/...)
     ▼
  handlers/     ← Request binding, validatie, Bun queries, response
     │
     ▼
  model/        ← Go structs met Bun tags (tabel-mapping, relaties)
     │
     ▼
  dbsetup/      ← Eenmalig: database + tabellen aanmaken
     │
     ▼
  PostgreSQL
```

---

## 3. Datamodel

### Database-tabellen

| Tabel | PK | Belangrijke kolommen |
|---|---|---|
| `gesprekken` | `id` (UUID) | `onderwerp`, `aanvang` (timestamptz), `einde` (timestamptz, nullable) |
| `gespreksdeelnemers` | `id` (UUID) | `naam`, `referentie` (URN) |
| `gesprek_deelnames` | `id` (UUID) | `gesprek_id` (FK), `deelnemer_id` (FK), `aanvang`, `einde` (nullable) |
| `gespreksbijdragen` | `id` (UUID) | `gesprek_id` (FK), `bijdrager_id` (FK), `geleverd`, `tekst` |
| `bijdrage_lezingen` | `id` (UUID) | `bijdrage_id` (FK), `lezer_id` (FK), `gelezen_op` |

Alle ID's zijn UUID's (gegenereerd door PostgreSQL via `gen_random_uuid()`). Alle tijdstempels zijn `timestamptz` (UTC in responses conform ADR).

---

## 4. API-endpoints

Basis-URL: `http://localhost:8080/v1`

### Gesprekken

| Methode | Pad | Beschrijving |
|---|---|---|
| `GET` | `/gesprekken` | Alle gesprekken ophalen (met deelnames) |
| `POST` | `/gesprekken` | Nieuw gesprek aanmaken |
| `GET` | `/gesprekken/{id}` | Enkel gesprek met deelnames, bijdragen en lezingen |
| `DELETE` | `/gesprekken/{id}` | Gesprek verwijderen |

### Gespreksdeelnemers

| Methode | Pad | Beschrijving |
|---|---|---|
| `GET` | `/gespreksdeelnemers` | Alle deelnemers ophalen |
| `POST` | `/gespreksdeelnemers` | Nieuwe deelnemer registreren |
| `GET` | `/gespreksdeelnemers/{id}` | Enkele deelnemer ophalen |
| `DELETE` | `/gespreksdeelnemers/{id}` | Deelnemer verwijderen |

### Deelnames (genest onder gesprekken)

| Methode | Pad | Beschrijving |
|---|---|---|
| `GET` | `/gesprekken/{id}/deelnames` | Deelnames van een gesprek |
| `POST` | `/gesprekken/{id}/deelnames` | Deelnemer aan gesprek toevoegen |
| `PATCH` | `/gesprekken/{id}/deelnames/{deelnameId}` | Deelname bijwerken (bijv. einde registreren) |

### Bijdragen (genest onder gesprekken)

| Methode | Pad | Beschrijving |
|---|---|---|
| `GET` | `/gesprekken/{id}/bijdragen` | Bijdragen binnen een gesprek |
| `POST` | `/gesprekken/{id}/bijdragen` | Bijdrage toevoegen |
| `GET` | `/gesprekken/{id}/bijdragen/{bijdrageId}` | Enkele bijdrage met lezingen |

### Lezingen (genest onder bijdragen)

| Methode | Pad | Beschrijving |
|---|---|---|
| `GET` | `/gesprekken/{id}/bijdragen/{bijdrageId}/lezingen` | Lezingen van een bijdrage |
| `POST` | `/gesprekken/{id}/bijdragen/{bijdrageId}/lezingen` | Lezing registreren |

### OpenAPI-specificatie

| Methode | Pad | Beschrijving |
|---|---|---|
| `GET` | `/openapi.json` | OAS 3.1 document (conform ADR `/core/publish-openapi`) |

---

## 5. NL API Design Rules conformiteit

| Regel | Implementatie |
|---|---|
| `/core/naming-resources` | Zelfstandige naamwoorden: gesprekken, gespreksdeelnemers, bijdragen |
| `/core/naming-collections` | Meervoud: `/gesprekken`, `/gespreksdeelnemers` |
| `/core/path-segments-kebab-case` | Alle padsegmenten in lowercase (geen koppelstreep nodig bij deze namen) |
| `/core/no-trailing-slash` | Geen trailing slashes in route-definities |
| `/core/uri-version` | Major versie in URI: `/v1/` |
| `/core/version-header` | `API-Version: 1.0.0` header in elk response |
| `/core/date-time/format` | Alle datum/tijd-velden: `type: string, format: date-time` |
| `/core/nested-child` | Child resources genest: `/gesprekken/{id}/bijdragen` |
| `/core/http-methods` | Alleen standaard HTTP-methoden: GET, POST, PATCH, DELETE |
| `/core/error-handling/problem-details` | Foutresponses in `application/problem+json` (RFC 9457) |
| `/core/error-handling/invalid-input` | Statuscode 400 bij ongeldige invoer |
| `/core/doc-openapi` | OAS 3.1 document aanwezig |
| `/core/publish-openapi` | Beschikbaar op `/v1/openapi.json` |
| `/core/interface-language` | Interface in het Nederlands |
| `/core/transport/security-headers` | Verplichte security headers (CSP, HSTS, X-Frame-Options, etc.) |
| `/core/transport/cors` | CORS headers aanwezig |
| `/core/query-keys-camel-case` | JSON-attributen in camelCase |

---

## 6. Configuratie

Alle configuratie gaat via environment variabelen (`.env`):

| Variabele | Standaard | Beschrijving |
|---|---|---|
| `DATABASE_URL` | — | PostgreSQL connection string |
| `DATABASE_ADMIN_URL` | — | Admin-connectie (voor `AUTO_CREATE_DATABASE`) |
| `AUTO_CREATE_DATABASE` | `false` | Database automatisch aanmaken bij opstarten |
| `PORT` | `8080` | HTTP-poort |
| `GIN_MODE` | `debug` | Gin modus (`debug` / `release`) |
| `BUNDEBUG` | `0` | `1` = uitgebreide SQL-logging |

---

## 7. Gebruiksinstructies

### Vereisten

- Go 1.22+
- PostgreSQL 14+ draaiend op `localhost:5432`
- Gebruiker `postgres` met wachtwoord `1234`

### Starten

```bash
cd d:\Git\CG\Gesprekken

# Dependencies ophalen (eenmalig of na wijzigingen)
go mod tidy

# Server starten
go run main.go
```

De applicatie:
1. Laadt `.env`
2. Maakt de database `gesprekken_db` aan (als `AUTO_CREATE_DATABASE=true`)
3. Maakt alle tabellen aan (`IF NOT EXISTS`)
4. Start de Gin HTTP-server op poort 8080

### Starten vanuit VS Code

Er is een launch-config aanwezig voor debuggen vanuit VS Code:

- `.vscode/launch.json`
- Configuratie: `Gesprekken API (Gin debug)`

Deze start de server met:

- `GIN_MODE=debug`
- `BUNDEBUG=1`
- `.env` geladen via `envFile`

Gebruik in VS Code de Run and Debug view en kies `Gesprekken API (Gin debug)`.

### Voorbeeld: een gesprek aanmaken

```bash
# Deelnemers registreren
curl -X POST http://localhost:8080/v1/gespreksdeelnemers \
  -H "Content-Type: application/json" \
  -d '{"naam": "Alice", "referentie": "urn:nl:bnr:001"}'

curl -X POST http://localhost:8080/v1/gespreksdeelnemers \
  -H "Content-Type: application/json" \
  -d '{"naam": "Bob", "referentie": "urn:nl:bnr:002"}'

# Gesprek aanmaken
curl -X POST http://localhost:8080/v1/gesprekken \
  -H "Content-Type: application/json" \
  -d '{"onderwerp": "Projectplanning Q2", "aanvang": "2026-03-23T10:00:00Z"}'

# Deelnemers toevoegen aan gesprek (vervang {gesprekId} en {deelnemerId})
curl -X POST http://localhost:8080/v1/gesprekken/{gesprekId}/deelnames \
  -H "Content-Type: application/json" \
  -d '{"deelnemerId": "{aliceId}", "aanvang": "2026-03-23T10:00:00Z"}'

# Bijdrage leveren
curl -X POST http://localhost:8080/v1/gesprekken/{gesprekId}/bijdragen \
  -H "Content-Type: application/json" \
  -d '{"bijdragerId": "{aliceId}", "geleverd": "2026-03-23T10:05:00Z", "tekst": "# Voorstel\n\nLaten we beginnen met de **roadmap**."}'

# Lezing registreren
curl -X POST http://localhost:8080/v1/gesprekken/{gesprekId}/bijdragen/{bijdrageId}/lezingen \
  -H "Content-Type: application/json" \
  -d '{"lezerId": "{bobId}", "gelezenOp": "2026-03-23T10:06:00Z"}'
```

### OpenAPI-specificatie bekijken

De OAS 3.1 specificatie is beschikbaar op:

```
http://localhost:8080/v1/openapi.json
```

Dit JSON-bestand kan direct worden geopend in [Swagger Editor](https://editor.swagger.io/) of [Redoc](https://redocly.github.io/redoc/).

### Postman bootstrap-requests

Er is ook een kleine Postman-collectie toegevoegd:

- `postman/gesprekken-bootstrap.postman_collection.json`

Deze bevat twee folders met bootstrap-requests:

1. `Bootstrap deelnemers`
2. `Bootstrap gesprek`

De collectie bewaart IDs automatisch in collectievariabelen, zodat de requests op elkaar kunnen voortbouwen.

Aanbevolen uitvoervolgorde:

1. `Maak deelnemer medewerker`
2. `Maak deelnemer inwoner`
3. `Controleer deelnemerslijst`
4. `Maak eerste gesprek`
5. `Voeg medewerker toe aan gesprek`
6. `Voeg inwoner toe aan gesprek`
7. `Maak eerste bijdrage van medewerker`
8. `Registreer lezing door inwoner`
9. `Controleer gesprekdetail`
