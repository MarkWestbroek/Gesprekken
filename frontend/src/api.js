/**
 * API client voor de Gesprekken REST API.
 *
 * Alle requests gaan via het relatieve pad /v1, dat door de Vite
 * dev-server geproxied wordt naar de Go backend (localhost:8080).
 *
 * Bij een fout wordt de `detail`-tekst uit het RFC 9457 ProblemDetail
 * response gegooid als Error.
 */

/** Basispad – komt overeen met de API-versie in de Go router */
const BASE = '/v1';

/**
 * Generieke fetch-wrapper die JSON verstuurt en ontvangt.
 * Gooit een Error met de ProblemDetail-melding als de response niet ok is.
 */
async function fetchJSON(url, options) {
  const res = await fetch(BASE + url, {
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.detail || `HTTP ${res.status}`);
  }
  // 204 No Content heeft geen body
  if (res.status === 204) return null;
  return res.json();
}

/** Haal alle gespreksdeelnemers op (GET /gespreksdeelnemers). */
export function listDeelnemers() {
  return fetchJSON('/gespreksdeelnemers');
}

/** Haal alle gesprekken op inclusief deelnames (GET /gesprekken). */
export function listGesprekken() {
  return fetchJSON('/gesprekken');
}

/** Haal een enkel gesprek op met bijdragen en deelnames (GET /gesprekken/:id). */
export function getGesprek(id) {
  return fetchJSON(`/gesprekken/${encodeURIComponent(id)}`);
}

/** Haal alle bijdragen van een gesprek op (GET /gesprekken/:id/bijdragen). */
export function listBijdragen(gesprekId) {
  return fetchJSON(`/gesprekken/${encodeURIComponent(gesprekId)}/bijdragen`);
}

/**
 * Verstuur een nieuwe bijdrage (bericht) in een gesprek.
 * Het tijdstip (geleverd) wordt automatisch op nu gezet.
 * Optioneel: bijlageIds koppelt eerder geüploade documenten aan het bericht.
 */
export function createBijdrage(gesprekId, bijdragerId, tekst, bijlageIds = [], reactieOpId = null) {
  return fetchJSON(`/gesprekken/${encodeURIComponent(gesprekId)}/bijdragen`, {
    method: 'POST',
    body: JSON.stringify({
      bijdragerId,
      geleverd: new Date().toISOString(),
      tekst,
      bijlageIds: bijlageIds.length > 0 ? bijlageIds : undefined,
      reactieOpId: reactieOpId || undefined,
    }),
  });
}

/**
 * Bewerk de tekst van een bestaande bijdrage (PUT).
 * Alleen de oorspronkelijke bijdrager mag dit doen.
 */
export function updateBijdrage(gesprekId, bijdrageId, bijdragerId, tekst) {
  return fetchJSON(
    `/gesprekken/${encodeURIComponent(gesprekId)}/bijdragen/${encodeURIComponent(bijdrageId)}`,
    {
      method: 'PUT',
      body: JSON.stringify({ bijdragerId, tekst }),
    }
  );
}

/**
 * Haal de versiehistorie van een bijdrage op
 * (GET /gesprekken/:gesprekId/bijdragen/:bijdrageId/versies).
 */
export function listBijdrageVersies(gesprekId, bijdrageId) {
  return fetchJSON(
    `/gesprekken/${encodeURIComponent(gesprekId)}/bijdragen/${encodeURIComponent(bijdrageId)}/versies`
  );
}

/**
 * Trek een bijdrage terug (PATCH).
 * Het bericht wordt niet verwijderd maar gemarkeerd als teruggetrokken.
 */
export function trekBijdrageTerug(gesprekId, bijdrageId, bijdragerId) {
  return fetchJSON(
    `/gesprekken/${encodeURIComponent(gesprekId)}/bijdragen/${encodeURIComponent(bijdrageId)}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ bijdragerId, teruggetrokken: true }),
    }
  );
}

/**
 * Registreer dat een deelnemer een bijdrage heeft gelezen
 * (POST /gesprekken/:gesprekId/bijdragen/:bijdrageId/lezingen).
 * Het tijdstip (gelezenOp) wordt automatisch op nu gezet.
 */
export function createLezing(gesprekId, bijdrageId, lezerId) {
  return fetchJSON(
    `/gesprekken/${encodeURIComponent(gesprekId)}/bijdragen/${encodeURIComponent(bijdrageId)}/lezingen`,
    {
      method: 'POST',
      body: JSON.stringify({
        lezerId,
        gelezenOp: new Date().toISOString(),
      }),
    }
  );
}

/** Haal alle deelnemertypen op (GET /deelnemertypen). */
export function listDeelnemertypen() {
  return fetchJSON('/deelnemertypen');
}

/**
 * Voeg een deelnemer toe aan een gesprek (POST /gesprekken/:id/deelnames).
 * Het aanvangstijdstip wordt automatisch op nu gezet.
 */
export function createDeelname(gesprekId, deelnemerId) {
  return fetchJSON(`/gesprekken/${encodeURIComponent(gesprekId)}/deelnames`, {
    method: 'POST',
    body: JSON.stringify({
      deelnemerId,
      aanvang: new Date().toISOString(),
    }),
  });
}

/** Maak een nieuwe gespreksdeelnemer aan (POST /gespreksdeelnemers). */
export function createDeelnemer(naam, referentie, typeId) {
  return fetchJSON('/gespreksdeelnemers', {
    method: 'POST',
    body: JSON.stringify({ naam, referentie, typeId }),
  });
}

/** Werk een bestaande gespreksdeelnemer bij (PUT /gespreksdeelnemers/:id). */
export function updateDeelnemer(id, naam, referentie, typeId) {
  return fetchJSON(`/gespreksdeelnemers/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify({ naam, referentie, typeId }),
  });
}

/** Verwijder een gespreksdeelnemer (DELETE /gespreksdeelnemers/:id). */
export function deleteDeelnemer(id) {
  return fetchJSON(`/gespreksdeelnemers/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  });
}

// ──────────────── Documenten (bijlagen) ────────────────

/** Maximale bestandsgrootte in bytes (25 MB, overeenkomend met backend). */
export const MAX_FILE_SIZE = 25 * 1024 * 1024;

/** Toegestane MIME-types (overeenkomend met backend). */
export const ALLOWED_MIME_TYPES = new Set([
  'image/jpeg', 'image/png', 'image/gif', 'image/webp',
  'application/pdf',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  'text/plain',
]);

/**
 * Upload een bestand als document-bijlage.
 * Geeft { bestandId, opgeslagenOp, bestandUrn, downloadUrl } terug.
 */
export async function uploadDocument(file, bronId, brontype = 'gespreksbijlage') {
  const form = new FormData();
  form.append('bestand', file);
  form.append('naam', file.name);
  form.append('brontype', brontype);
  form.append('bronId', bronId);
  form.append('bronUrn', `urn:gesprekken:${brontype}:${bronId}`);
  form.append('bronUrl', `/v1/gesprekken/${bronId}`);

  const res = await fetch(`${BASE}/documenten`, {
    method: 'POST',
    body: form,
    // Geen Content-Type header: browser zet multipart boundary automatisch
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.detail || `HTTP ${res.status}`);
  }
  return res.json();
}

/** Genereer de download-URL voor een document. */
export function documentDownloadUrl(bestandId) {
  return `${BASE}/documenten/${encodeURIComponent(bestandId)}/download`;
}


