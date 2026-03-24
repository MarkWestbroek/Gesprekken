# Documenten API — Bijlagen upload/download

REST endpoints voor het uploaden, downloaden en ophalen van metadata van bestandsbijlagen.

## Endpoints

### `POST /v1/documenten` — Bestand uploaden

**Content-Type:** `multipart/form-data`

| Veld | Type | Verplicht | Beschrijving |
|---|---|---|---|
| `bestand` | file | Ja | Het te uploaden bestand |
| `naam` | string | Nee | Naam (default: originele bestandsnaam) |
| `brontype` | string | Nee | Type bron, default `gespreksbijlage` |
| `bronId` | UUID | Ja | ID van de bronentiteit (bijv. gesprekId) |
| `bronUrn` | string | Nee | URN van de bron |
| `bronUrl` | string | Nee | Relatieve URL van de bron |

**Response (201 Created):**
```json
{
  "bestandId": "a1b2c3d4-...",
  "opgeslagenOp": "2026-03-24T12:00:00Z",
  "bestandUrn": "urn:gesprekken:document:a1b2c3d4-...",
  "downloadUrl": "/v1/documenten/a1b2c3d4-.../download"
}
```

### `GET /v1/documenten/{bestandId}` — Metadata ophalen

**Response (200 OK):**
```json
{
  "bestandId": "a1b2c3d4-...",
  "naam": "rapport.pdf",
  "brontype": "gespreksbijlage",
  "bronId": "...",
  "bronUrn": "urn:gesprekken:gespreksbijlage:...",
  "bronUrl": "/v1/gesprekken/...",
  "contentType": "application/pdf",
  "grootte": 102400,
  "opgeslagenOp": "2026-03-24T12:00:00Z",
  "bestandUrn": "urn:gesprekken:document:a1b2c3d4-...",
  "downloadUrl": "/v1/documenten/a1b2c3d4-.../download"
}
```

### `GET /v1/documenten/{bestandId}/download` — Bestand downloaden

Streamt het bestand met het juiste `Content-Type` en `Content-Disposition` header.

## Limieten en validatie

| Regel | Waarde |
|---|---|
| Max bestandsgrootte | 25 MB |
| Toegestane types | JPEG, PNG, GIF, WebP, PDF, DOC, DOCX, XLSX, PPTX, TXT |

Foutresponses gebruiken RFC 9457 ProblemDetail format.

## Bijlagen koppelen aan bijdragen

Bij het aanmaken van een bijdrage (`POST /v1/gesprekken/{id}/bijdragen`) kan een optioneel veld `bijlageIds` (array van UUIDs) worden meegegeven. Dit koppelt eerder geüploade documenten aan het bericht.

```json
{
  "bijdragerId": "...",
  "geleverd": "2026-03-24T12:00:00Z",
  "tekst": "Zie bijlage",
  "bijlageIds": ["a1b2c3d4-..."]
}
```

## Auth

Dezelfde authenticatie als de overige API-endpoints (momenteel geen auth in development).

## Tests

```bash
go test ./handlers/ -v
```
