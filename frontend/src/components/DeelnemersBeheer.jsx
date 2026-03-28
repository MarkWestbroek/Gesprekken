import { useState, useEffect } from 'react';
import {
  listDeelnemers,
  listDeelnemertypen,
  createDeelnemer,
  updateDeelnemer,
  deleteDeelnemer,
} from '../api';

/**
 * Beheercherm voor gespreksdeelnemers: een inline-bewerkbare tabel
 * met paginering. Medewerkers kunnen deelnemers toevoegen, wijzigen
 * en verwijderen.
 *
 * Elke rij heeft een "Bewerk" / "Bewaar"+"Annuleer" toggle.
 * Onderaan de tabel kan een nieuwe deelnemer worden toegevoegd.
 *
 * Props:
 *   onBack – callback om terug te navigeren
 */
const PAGE_SIZE = 10;

export default function DeelnemersBeheer({ onBack }) {
  const [deelnemers, setDeelnemers] = useState([]);
  const [typen, setTypen] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  // Paginering
  const [page, setPage] = useState(0);

  // Inline editing: ID van de rij die bewerkt wordt + buffervelden
  const [editId, setEditId] = useState(null);
  const [editNaam, setEditNaam] = useState('');
  const [editRef, setEditRef] = useState('');
  const [editTypeId, setEditTypeId] = useState('');

  // Nieuwe deelnemer toevoegen (inline onderaan de tabel)
  const [showNew, setShowNew] = useState(false);
  const [newNaam, setNewNaam] = useState('');
  const [newRef, setNewRef] = useState('');
  const [newTypeId, setNewTypeId] = useState('');

  const [saving, setSaving] = useState(false);

  // Laad deelnemers en typen bij openen
  useEffect(() => {
    Promise.all([listDeelnemers(), listDeelnemertypen()])
      .then(([d, t]) => {
        setDeelnemers(d);
        setTypen(t);
        if (t.length > 0) setNewTypeId(t[0].id);
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false));
  }, []);

  const totalPages = Math.max(1, Math.ceil(deelnemers.length / PAGE_SIZE));
  const paged = deelnemers.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

  // ─── Inline bewerken ───

  const startEdit = (d) => {
    setEditId(d.id);
    setEditNaam(d.naam);
    setEditRef(d.referentie);
    setEditTypeId(d.typeId);
    setError(null);
  };

  const cancelEdit = () => {
    setEditId(null);
  };

  const saveEdit = async () => {
    if (!editNaam.trim() || !editRef.trim() || !editTypeId) return;
    setSaving(true);
    setError(null);
    try {
      const updated = await updateDeelnemer(editId, editNaam.trim(), editRef.trim(), editTypeId);
      setDeelnemers((prev) => prev.map((d) => (d.id === editId ? updated : d)));
      setEditId(null);
    } catch (e) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  // ─── Verwijderen ───

  const handleDelete = async (id) => {
    setSaving(true);
    setError(null);
    try {
      await deleteDeelnemer(id);
      setDeelnemers((prev) => {
        const next = prev.filter((d) => d.id !== id);
        // Corrigeer pagina als die nu leeg is
        const newTotalPages = Math.max(1, Math.ceil(next.length / PAGE_SIZE));
        if (page >= newTotalPages) setPage(newTotalPages - 1);
        return next;
      });
      if (editId === id) setEditId(null);
    } catch (e) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  // ─── Nieuw toevoegen ───

  const handleAdd = async () => {
    if (!newNaam.trim() || !newRef.trim() || !newTypeId) return;
    setSaving(true);
    setError(null);
    try {
      const created = await createDeelnemer(newNaam.trim(), newRef.trim(), newTypeId);
      // Het response bevat geen type-relatie; voeg die handmatig toe
      created.type = typen.find((t) => t.id === created.typeId) || null;
      setDeelnemers((prev) => [...prev, created]);
      setNewNaam('');
      setNewRef('');
      setNewTypeId(typen[0]?.id || '');
      setShowNew(false);
      // Spring naar laatste pagina zodat de nieuwe rij zichtbaar is
      setPage(Math.ceil((deelnemers.length + 1) / PAGE_SIZE) - 1);
    } catch (e) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  };

  /** Helper: geef de weergavenaam van een type terug op basis van ID */
  const typeNaam = (typeId) => typen.find((t) => t.id === typeId)?.naam || '—';

  if (loading) return <div className="screen"><p className="empty">Laden…</p></div>;

  return (
    <div className="screen beheer-screen">
      <header className="list-header">
        <div className="header-brand">
          <img className="header-brandmark" src="/cg-brandmark.svg" alt="Common Ground" />
          <h2>Deelnemers beheren</h2>
        </div>
        <button className="btn-switch" onClick={onBack}>← Terug</button>
      </header>

      {error && <p className="error">{error}</p>}

      <div className="table-wrapper">
        <table className="beheer-table">
          <thead>
            <tr>
              <th>Naam</th>
              <th>Referentie</th>
              <th>Type</th>
              <th className="col-actions">Acties</th>
            </tr>
          </thead>
          <tbody>
            {paged.map((d) =>
              editId === d.id ? (
                /* ─── Bewerkregel ─── */
                <tr key={d.id} className="editing-row">
                  <td>
                    <input
                      value={editNaam}
                      onChange={(e) => setEditNaam(e.target.value)}
                      disabled={saving}
                    />
                  </td>
                  <td>
                    <input
                      value={editRef}
                      onChange={(e) => setEditRef(e.target.value)}
                      disabled={saving}
                    />
                  </td>
                  <td>
                    <select
                      value={editTypeId}
                      onChange={(e) => setEditTypeId(e.target.value)}
                      disabled={saving}
                    >
                      {typen.map((t) => (
                        <option key={t.id} value={t.id}>{t.naam}</option>
                      ))}
                    </select>
                  </td>
                  <td className="col-actions">
                    <button className="btn-save" onClick={saveEdit} disabled={saving} aria-label="Wijziging opslaan">
                      ✓
                    </button>
                    <button className="btn-cancel" onClick={cancelEdit} disabled={saving} aria-label="Bewerking annuleren">
                      ✕
                    </button>
                  </td>
                </tr>
              ) : (
                /* ─── Leesregel ─── */
                <tr key={d.id}>
                  <td>{d.naam}</td>
                  <td className="cell-ref">{d.referentie}</td>
                  <td>{d.type?.naam || typeNaam(d.typeId)}</td>
                  <td className="col-actions">
                    <button className="btn-edit" onClick={() => startEdit(d)} disabled={saving} aria-label={`Bewerk ${d.naam}`}>
                      ✎
                    </button>
                    <button className="btn-delete" onClick={() => handleDelete(d.id)} disabled={saving} aria-label={`Verwijder ${d.naam}`}>
                      🗑
                    </button>
                  </td>
                </tr>
              )
            )}

            {/* ─── Nieuw-toevoegen rij ─── */}
            {showNew && (
              <tr className="editing-row new-row">
                <td>
                  <input
                    placeholder="Naam"
                    value={newNaam}
                    onChange={(e) => setNewNaam(e.target.value)}
                    disabled={saving}
                  />
                </td>
                <td>
                  <input
                    placeholder="urn:..."
                    value={newRef}
                    onChange={(e) => setNewRef(e.target.value)}
                    disabled={saving}
                  />
                </td>
                <td>
                  <select
                    value={newTypeId}
                    onChange={(e) => setNewTypeId(e.target.value)}
                    disabled={saving}
                  >
                    {typen.map((t) => (
                      <option key={t.id} value={t.id}>{t.naam}</option>
                    ))}
                  </select>
                </td>
                <td className="col-actions">
                  <button className="btn-save" onClick={handleAdd} disabled={saving} aria-label="Deelnemer toevoegen">
                    ✓
                  </button>
                  <button className="btn-cancel" onClick={() => setShowNew(false)} disabled={saving} aria-label="Annuleer toevoegen">
                    ✕
                  </button>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* ─── Onderste balk: paginering + toevoegen ─── */}
      <div className="beheer-footer">
        <button
          className="btn-add"
          onClick={() => { setShowNew(true); setError(null); }}
          disabled={showNew || saving}
        >
          + Nieuwe deelnemer
        </button>

        <div className="pagination">
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
          >
            ‹
          </button>
          <span>{page + 1} / {totalPages}</span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
          >
            ›
          </button>
        </div>
      </div>
    </div>
  );
}
