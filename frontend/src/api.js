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
 */
export function createBijdrage(gesprekId, bijdragerId, tekst) {
  return fetchJSON(`/gesprekken/${encodeURIComponent(gesprekId)}/bijdragen`, {
    method: 'POST',
    body: JSON.stringify({
      bijdragerId,
      geleverd: new Date().toISOString(),
      tekst,
    }),
  });
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


