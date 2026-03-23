import { useState, useEffect, useRef } from 'react';
import { listBijdragen, createBijdrage, createLezing, listDeelnemers, createDeelname } from '../api';
import { formatMessage } from '../formatMessage';

/**
 * Chatscherm: toont de bijdragen (berichten) van een gesprek in
 * messenger-stijl. Eigen berichten staan rechts (groen), berichten
 * van anderen links (wit).
 *
 * Berichten worden elke 3 seconden opnieuw opgehaald (polling).
 * De gebruiker kan een nieuw bericht typen met WhatsApp-achtige
 * opmaak (*vet*, _cursief_, ~doorhalen~, `code`, - opsomming).
 *
 * Leesbevestigingen (✓✓) worden automatisch verzonden bij het
 * openen van berichten en getoond onder eigen berichten.
 *
 * Interne actoren (medewerkers) zien een "+ Collega" knop waarmee
 * ze andere interne actoren aan het gesprek kunnen toevoegen via
 * een modaal keuzelijst.
 *
 * Props:
 *   user     – de actief gekozen gespreksdeelnemer (inclusief type)
 *   gesprek  – het geselecteerde gesprek-object (inclusief deelnames)
 *   onBack   – callback om terug te gaan naar de gesprekkenlijst
 */
export default function ChatView({ user, gesprek, onBack }) {
  const [bijdragen, setBijdragen] = useState([]);
  const [tekst, setTekst] = useState('');
  const [sending, setSending] = useState(false);
  const [error, setError] = useState(null);
  const endRef = useRef(null);
  const inputRef = useRef(null);
  // Bijhouden welke bijdrage-IDs al als gelezen geregistreerd zijn in deze sessie,
  // zodat we de API niet onnodig vaker aanroepen dan nodig.
  const markedRef = useRef(new Set());

  // Collega-erbij modal (alleen voor interne_actor)
  const [showCollegaModal, setShowCollegaModal] = useState(false);
  const [beschikbareCollega, setBeschikbareCollega] = useState([]);
  const [collegaLoading, setCollegaLoading] = useState(false);

  const isInterneActor = user.type?.code === 'interne_actor';

  /**
   * Registreer lezingen voor alle berichten die:
   *   - niet van de huidige gebruiker zelf zijn, én
   *   - nog geen lezing hebben van de huidige gebruiker.
   * Fouten worden genegeerd (fire-and-forget).
   */
  const markAsRead = (geladen) => {
    geladen.forEach((b) => {
      if (
        b.bijdragerId !== user.id &&
        !markedRef.current.has(b.id) &&
        !b.lezingen?.some((l) => l.lezerId === user.id)
      ) {
        markedRef.current.add(b.id);
        createLezing(gesprek.id, b.id, user.id).catch(() => {
          // Verwijder uit de set zodat het later opnieuw geprobeerd kan worden
          markedRef.current.delete(b.id);
        });
      }
    });
  };

  /** Haal bijdragen op van de API en markeer ongelezen berichten */
  const loadBijdragen = () =>
    listBijdragen(gesprek.id)
      .then((data) => {
        setBijdragen(data);
        markAsRead(data);
      })
      .catch((e) => setError(e.message));

  // Laad bijdragen bij openen en poll elke 3 seconden voor nieuwe berichten
  useEffect(() => {
    loadBijdragen();
    const interval = setInterval(loadBijdragen, 3000);
    return () => clearInterval(interval);
  }, [gesprek.id]);

  // Scroll automatisch naar het laatste bericht bij nieuwe berichten
  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [bijdragen]);

  /** Verstuur het bericht en laad de lijst opnieuw */
  const handleSend = async () => {
    const trimmed = tekst.trim();
    if (!trimmed || sending) return;
    setSending(true);
    setError(null);
    try {
      await createBijdrage(gesprek.id, user.id, trimmed);
      setTekst('');
      await loadBijdragen();
      inputRef.current?.focus();
    } catch (e) {
      setError(e.message);
    } finally {
      setSending(false);
    }
  };

  /** Enter verstuurt, Shift+Enter maakt een nieuwe regel */
  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  /** Formatteer tijdstip als HH:MM */
  const formatTime = (iso) =>
    new Date(iso).toLocaleTimeString('nl-NL', {
      hour: '2-digit',
      minute: '2-digit',
    });

  /** Formatteer datum als "maandag 23 maart" */
  const formatDate = (iso) =>
    new Date(iso).toLocaleDateString('nl-NL', {
      weekday: 'long',
      day: 'numeric',
      month: 'long',
    });

  // Bijhouden van de laatst getoonde datum voor datumseparatoren
  let lastDate = '';

  /**
   * Open het collega-modal: haal alle deelnemers op en filter
   * op interne_actors die nog niet in dit gesprek zitten.
   */
  const openCollegaModal = async () => {
    setCollegaLoading(true);
    setShowCollegaModal(true);
    try {
      const alleDeelnemers = await listDeelnemers();
      // Verzamel IDs van deelnemers die al in dit gesprek zitten
      const alInGesprek = new Set(
        (gesprek.deelnames || []).map((d) => d.deelnemerId)
      );
      // Filter: alleen interne actors die er nog niet in zitten
      const beschikbaar = alleDeelnemers.filter(
        (d) => d.type?.code === 'interne_actor' && !alInGesprek.has(d.id)
      );
      setBeschikbareCollega(beschikbaar);
    } catch (e) {
      setError(e.message);
      setShowCollegaModal(false);
    } finally {
      setCollegaLoading(false);
    }
  };

  /** Voeg een collega toe aan het gesprek en sluit het modal */
  const handleAddCollega = async (deelnemerId) => {
    setCollegaLoading(true);
    try {
      await createDeelname(gesprek.id, deelnemerId);
      // Verwijder uit de beschikbare lijst
      setBeschikbareCollega((prev) => prev.filter((d) => d.id !== deelnemerId));
    } catch (e) {
      setError(e.message);
    } finally {
      setCollegaLoading(false);
    }
  };

  return (
    <div className="screen chat-view">
      <header className="chat-header">
        <button className="btn-back" onClick={onBack}>
          ←
        </button>
        <div className="chat-title">
          <h2>{gesprek.onderwerp}</h2>
          <span className="chat-subtitle">{user.naam}</span>
        </div>
        {isInterneActor && (
          <button className="btn-collega" onClick={openCollegaModal}>
            + Collega
          </button>
        )}
      </header>

      {/* Modal: collega toevoegen aan gesprek */}
      {showCollegaModal && (
        <div className="modal-overlay" onClick={() => setShowCollegaModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>Collega erbij</h3>
            {collegaLoading && <p className="empty">Laden…</p>}
            {!collegaLoading && beschikbareCollega.length === 0 && (
              <p className="empty">Geen beschikbare collega's gevonden.</p>
            )}
            <ul className="collega-list">
              {beschikbareCollega.map((d) => (
                <li key={d.id}>
                  <button onClick={() => handleAddCollega(d.id)} disabled={collegaLoading}>
                    <span className="user-name">{d.naam}</span>
                    <span className="user-ref">{d.referentie}</span>
                  </button>
                </li>
              ))}
            </ul>
            <button className="btn-close-modal" onClick={() => setShowCollegaModal(false)}>
              Sluiten
            </button>
          </div>
        </div>
      )}

      <div className="messages">
        {bijdragen.map((b) => {
          const isMine = b.bijdragerId === user.id;
          const dateStr = formatDate(b.geleverd);
          let dateSeparator = null;
          if (dateStr !== lastDate) {
            lastDate = dateStr;
            dateSeparator = <div className="date-separator">{dateStr}</div>;
          }

          return (
            <div key={b.id}>
              {dateSeparator}
              <div className={`message ${isMine ? 'mine' : 'theirs'}`}>
                {!isMine && (
                  <span className="message-sender">
                    {b.bijdrager?.naam || 'Onbekend'}
                  </span>
                )}
                <div
                  className="message-text"
                  dangerouslySetInnerHTML={{ __html: formatMessage(b.tekst) }}
                />
                <div className="message-footer">
                  <span className="message-time">{formatTime(b.geleverd)}</span>
                  {/* Toon leesbevestigingen alleen onder eigen berichten */}
                  {isMine && b.lezingen && b.lezingen.length > 0 && (
                    <span className="message-read">
                      ✓✓ {b.lezingen.map((l) => l.lezer?.naam || 'Onbekend').join(', ')}
                    </span>
                  )}
                </div>
              </div>
            </div>
          );
        })}
        <div ref={endRef} />
      </div>

      {error && <p className="error chat-error">{error}</p>}

      <div className="chat-input">
        <textarea
          ref={inputRef}
          value={tekst}
          onChange={(e) => setTekst(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Schrijf een bericht… (Shift+Enter voor nieuwe regel)"
          rows={2}
          disabled={sending}
        />
        <button
          className="btn-send"
          onClick={handleSend}
          disabled={!tekst.trim() || sending}
        >
          Verstuur
        </button>
      </div>

      <div className="format-hint">
        <code>*vet*</code> <code>_cursief_</code> <code>~doorhalen~</code>{' '}
        <code>`code`</code> <code>- opsomming</code>
      </div>
    </div>
  );
}
