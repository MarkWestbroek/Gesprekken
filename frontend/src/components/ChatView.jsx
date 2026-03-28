import { useState, useEffect, useRef } from 'react';
import {
  listBijdragen, createBijdrage, createLezing, listDeelnemers, createDeelname,
  uploadDocument, documentDownloadUrl, MAX_FILE_SIZE, ALLOWED_MIME_TYPES,
} from '../api';
import { formatMessage, normalizeAsciiSmileys } from '../formatMessage';
import EmojiPicker from './EmojiPicker';

function PaperclipIcon() {
  return (
    <svg className="composer-icon" viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <path
        d="M8.5 12.5L15.9 5.1a3.25 3.25 0 1 1 4.6 4.6l-9.2 9.2a5.25 5.25 0 1 1-7.4-7.4l9.6-9.6"
        fill="none"
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="1.9"
      />
    </svg>
  );
}

const URL_REGEX = /https?:\/\/[^\s]+/g;

function extractUrls(text) {
  return (text.match(URL_REGEX) || []).map((raw) => raw.replace(/[.,!?;:)\]]*$/, ''));
}

function formatDateLabel(iso) {
  const d = new Date(iso);
  const date = new Date(d.getFullYear(), d.getMonth(), d.getDate());
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const diffInDays = Math.round((today - date) / 86400000);

  if (diffInDays === 1) return 'Gisteren';

  return d.toLocaleDateString('nl-NL', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
  });
}

function toHostname(url) {
  try {
    return new URL(url).hostname;
  } catch {
    return '';
  }
}

/**
 * Chatscherm: toont de bijdragen (berichten) van een gesprek in
 * messenger-stijl. Eigen berichten staan rechts (lichtblauw), berichten
 * van anderen links (wit).
 *
 * Berichten worden elke 3 seconden opnieuw opgehaald (polling).
 * De gebruiker kan een nieuw bericht typen met WhatsApp-achtige
 * opmaak (*vet*, _cursief_, ~doorhalen~, `code`, - opsomming).
 *
 * Als de gebruiker omhoog is gescrold, blijft de scrollpositie
 * staan en verschijnt een badge voor nieuwe berichten.
 *
 * Leesbevestigingen worden alleen verzonden voor berichten die
 * zichtbaar zijn in de viewport. Onder eigen berichten tonen
 * vinkjes de status: ✓ = verzonden, ✓✓ = gelezen.
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
  const [newMessagesCount, setNewMessagesCount] = useState(0);
  const messagesRef = useRef(null);
  const endRef = useRef(null);
  const inputRef = useRef(null);
  const fileInputRef = useRef(null);
  const shouldScrollAfterUpdateRef = useRef(true);
  const hasLoadedOnceRef = useRef(false);
  const latestLoadRequestIdRef = useRef(0);
  const unseenMessageIdsRef = useRef(new Set());
  // Bijhouden welke bijdrage-IDs al als gelezen geregistreerd zijn in deze sessie,
  // zodat we de API niet onnodig vaker aanroepen dan nodig.
  const markedRef = useRef(new Set());

  // Pending bestanden die de gebruiker heeft geselecteerd maar nog niet verstuurd
  const [pendingFiles, setPendingFiles] = useState([]);
  const [uploadProgress, setUploadProgress] = useState(null); // 'uploading' | null

  // Reageren op een bijdrage
  const [replyTo, setReplyTo] = useState(null);

  // Collega-erbij modal (alleen voor interne_actor)
  const [showCollegaModal, setShowCollegaModal] = useState(false);
  const [beschikbareCollega, setBeschikbareCollega] = useState([]);
  const [collegaLoading, setCollegaLoading] = useState(false);

  const isInterneActor = user.type?.code === 'interne_actor';
  const draftUrls = extractUrls(tekst);
  const draftPreviewUrl = draftUrls.length > 0 ? draftUrls[0] : null;
  const draftHtml = tekst.trim() ? formatMessage(tekst) : '';

  const scrollMessagesToBottom = (behavior = 'auto') => {
    const container = messagesRef.current;
    if (!container) return;
    const top = container.scrollHeight;
    if (behavior === 'smooth' && typeof container.scrollTo === 'function') {
      container.scrollTo({ top, behavior: 'smooth' });
      return;
    }
    container.scrollTop = top;
  };

  const isNearBottom = (element, threshold = 80) => {
    if (!element) return true;
    const distanceToBottom = element.scrollHeight - element.scrollTop - element.clientHeight;
    return distanceToBottom <= threshold;
  };

  /**
   * Registreer lezingen alleen voor berichten die zichtbaar zijn in de
   * viewport van het berichtenpaneel.
   */
  const markVisibleAsRead = (geladen) => {
    const container = messagesRef.current;
    if (!container) return;

    const viewportTop = container.scrollTop;
    const viewportBottom = viewportTop + container.clientHeight;
    const visibleIds = new Set();

    container.querySelectorAll('[data-bijdrage-id]').forEach((el) => {
      const top = el.offsetTop;
      const bottom = top + el.offsetHeight;
      if (bottom > viewportTop && top < viewportBottom) {
        visibleIds.add(el.getAttribute('data-bijdrage-id'));
      }
    });

    geladen.forEach((b) => {
      if (
        visibleIds.has(String(b.id)) &&
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
  const loadBijdragen = (forceScroll = false) => {
    const requestId = ++latestLoadRequestIdRef.current;
    return listBijdragen(gesprek.id)
      .then((data) => {
        const wasNearBottom = isNearBottom(messagesRef.current);
        shouldScrollAfterUpdateRef.current = forceScroll || !hasLoadedOnceRef.current || wasNearBottom;

        setBijdragen((prev) => {
          if (requestId !== latestLoadRequestIdRef.current) {
            return prev;
          }

          if (wasNearBottom) {
            unseenMessageIdsRef.current.clear();
            setNewMessagesCount(0);
          } else if (hasLoadedOnceRef.current) {
            const previousIds = new Set(prev.map((b) => b.id));
            data.forEach((b) => {
              if (!previousIds.has(b.id)) {
                unseenMessageIdsRef.current.add(b.id);
              }
            });
            setNewMessagesCount(unseenMessageIdsRef.current.size);
          }
          hasLoadedOnceRef.current = true;
          return data;
        });
      })
        .catch((e) => setError(e.message));
      };

  // Laad bijdragen bij openen en poll elke 3 seconden voor nieuwe berichten
  useEffect(() => {
    loadBijdragen();
    const interval = setInterval(loadBijdragen, 3000);
    return () => clearInterval(interval);
  }, [gesprek.id]);

  // Bij een nieuw gesprek starten we weer onderaan.
  useEffect(() => {
    shouldScrollAfterUpdateRef.current = true;
    hasLoadedOnceRef.current = false;
    latestLoadRequestIdRef.current = 0;
    unseenMessageIdsRef.current.clear();
    setNewMessagesCount(0);
  }, [gesprek.id]);

  // Scroll alleen automatisch als de gebruiker onderaan stond bij het ophalen,
  // of als we expliciet geforceerd hebben (bijv. na verzenden).
  useEffect(() => {
    if (shouldScrollAfterUpdateRef.current) {
      scrollMessagesToBottom('auto');
      setNewMessagesCount(0);
    }
  }, [bijdragen]);

  useEffect(() => {
    markVisibleAsRead(bijdragen);
  }, [bijdragen]);

  const handleMessagesScroll = () => {
    markVisibleAsRead(bijdragen);
    if (isNearBottom(messagesRef.current)) {
      unseenMessageIdsRef.current.clear();
      setNewMessagesCount(0);
    }
  };

  const scrollToLatest = () => {
    shouldScrollAfterUpdateRef.current = true;
    unseenMessageIdsRef.current.clear();
    setNewMessagesCount(0);
    scrollMessagesToBottom('smooth');
  };

  /** Verstuur het bericht (met eventuele bijlagen) en laad de lijst opnieuw */
  const handleSend = async () => {
    const trimmed = tekst.trim();
    if ((!trimmed && pendingFiles.length === 0) || sending) return;
    setSending(true);
    setUploadProgress(pendingFiles.length > 0 ? 'uploading' : null);
    setError(null);
    try {
      // Upload alle geselecteerde bestanden
      const uploadedIds = [];
      for (const file of pendingFiles) {
        const result = await uploadDocument(file, gesprek.id);
        uploadedIds.push(result.bestandId);
      }
      // Maak bijdrage aan met eventuele bijlage-IDs
      const berichtTekst = trimmed || (uploadedIds.length > 0 ? '📎 Bijlage' : '');
      await createBijdrage(gesprek.id, user.id, berichtTekst, uploadedIds, replyTo?.id);
      setTekst('');
      setPendingFiles([]);
      setReplyTo(null);
      setUploadProgress(null);
      await loadBijdragen(true);
      inputRef.current?.focus();
    } catch (e) {
      setError(e.message);
      setUploadProgress(null);
    } finally {
      setSending(false);
    }
  };

  /** Bestand(en) selecteren via file input */
  const handleFileSelect = (e) => {
    const files = Array.from(e.target.files);
    const validFiles = [];
    for (const file of files) {
      if (file.size > MAX_FILE_SIZE) {
        setError(`"${file.name}" is te groot (max 25 MB).`);
        continue;
      }
      if (!ALLOWED_MIME_TYPES.has(file.type)) {
        setError(`"${file.name}" heeft een niet-ondersteund bestandstype.`);
        continue;
      }
      validFiles.push(file);
    }
    setPendingFiles((prev) => [...prev, ...validFiles]);
    e.target.value = ''; // Reset zodat hetzelfde bestand opnieuw gekozen kan worden
  };

  /** Verwijder een pending bestand uit de wachtrij */
  const removePendingFile = (index) => {
    setPendingFiles((prev) => prev.filter((_, i) => i !== index));
  };

  /** Plak een afbeelding vanuit het clipboard als pending bijlage */
  const handlePaste = (e) => {
    const items = Array.from(e.clipboardData?.items || []);
    const imageFiles = items
      .filter((item) => item.kind === 'file' && item.type.startsWith('image/'))
      .map((item) => item.getAsFile())
      .filter(Boolean);
    if (imageFiles.length === 0) return; // gewone tekst-paste, laat default gedrag
    e.preventDefault();
    const validFiles = [];
    for (const file of imageFiles) {
      if (file.size > MAX_FILE_SIZE) {
        setError(`Geplakte afbeelding is te groot (max 25 MB).`);
        continue;
      }
      if (!ALLOWED_MIME_TYPES.has(file.type)) {
        setError(`Geplakt bestandstype '${file.type}' wordt niet ondersteund.`);
        continue;
      }
      // Clipboard afbeeldingen hebben vaak een generieke naam; geef ze een timestamp
      const ext = file.type === 'image/png' ? '.png' : file.type === 'image/jpeg' ? '.jpg' : file.type === 'image/gif' ? '.gif' : '.webp';
      const named = new File([file], `clipboard-${Date.now()}${ext}`, { type: file.type });
      validFiles.push(named);
    }
    if (validFiles.length > 0) {
      setPendingFiles((prev) => [...prev, ...validFiles]);
    }
  };

  /** Scroll naar de oorspronkelijke bijdrage en markeer deze kort */
  const scrollToBijdrage = (id) => {
    const el = messagesRef.current?.querySelector(`[data-bijdrage-id="${id}"]`);
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' });
      el.classList.add('highlight');
      setTimeout(() => el.classList.remove('highlight'), 1500);
    }
  };

  /** Enter verstuurt, Shift+Enter maakt een nieuwe regel */
  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  /** Voeg een emoji in op de cursorpositie van het tekstveld */
  const insertEmoji = (emoji) => {
    const ta = inputRef.current;
    if (!ta) { setTekst((prev) => prev + emoji); return; }
    const start = ta.selectionStart;
    const end = ta.selectionEnd;
    const before = tekst.slice(0, start);
    const after = tekst.slice(end);
    setTekst(before + emoji + after);
    requestAnimationFrame(() => {
      ta.focus();
      const pos = start + emoji.length;
      ta.setSelectionRange(pos, pos);
    });
  };

  const handleTekstChange = (value) => {
    // Voer dezelfde smiley-normalisatie alvast in tijdens typen,
    // zodat gebruikers direct zien wat ze gaan versturen.
    setTekst(normalizeAsciiSmileys(value));
  };

  /** Formatteer tijdstip als HH:MM */
  const formatTime = (iso) =>
    new Date(iso).toLocaleTimeString('nl-NL', {
      hour: '2-digit',
      minute: '2-digit',
    });

  const otherParticipantIds = new Set(
    (gesprek.deelnames || [])
      .map((d) => d.deelnemerId)
      .filter((id) => id != null && id !== user.id)
  );

  // Bijhouden van de laatst getoonde datum voor datumseparatoren
  let lastDateKey = '';

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
        <button className="btn-back" onClick={onBack} aria-label="Terug naar gesprekkenlijst">
          ←
        </button>
        <div className="chat-title">
          <h2>{gesprek.onderwerp}</h2>
          <span className="chat-subtitle">{user.naam}</span>
        </div>
        <div className="chat-header-right">
          {isInterneActor && (
            <button className="btn-collega" onClick={openCollegaModal}>
              + Collega
            </button>
          )}
          <img className="chat-brandmark" src="/cg-brandmark.svg" alt="Common Ground" />
        </div>
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

      <div className="messages" ref={messagesRef} onScroll={handleMessagesScroll}>
        {bijdragen.map((b) => {
          const isMine = b.bijdragerId === user.id;
          const day = new Date(b.geleverd);
          const dateKey = `${day.getFullYear()}-${day.getMonth()}-${day.getDate()}`;
          const dateStr = formatDateLabel(b.geleverd);
          const urls = extractUrls(b.tekst);
          const previewUrl = urls.length > 0 ? urls[0] : null;
          let dateSeparator = null;
          if (dateKey !== lastDateKey) {
            lastDateKey = dateKey;
            dateSeparator = <div className="date-separator">{dateStr}</div>;
          }

          return (
            <div key={b.id} data-bijdrage-id={String(b.id)}>
              {dateSeparator}
              <div className={`message ${isMine ? 'mine' : 'theirs'}`}>
                {!isMine && (
                  <span className="message-sender">
                    {b.bijdrager?.naam || 'Onbekend'}
                  </span>
                )}
                {b.reactieOp && (
                  <div className="reply-quote" onClick={() => scrollToBijdrage(b.reactieOpId)}>
                    <span className="reply-quote-sender">{b.reactieOp.bijdrager?.naam || 'Onbekend'}</span>
                    <span className="reply-quote-text">
                      {b.reactieOp.tekst.length > 120 ? b.reactieOp.tekst.slice(0, 120) + '…' : b.reactieOp.tekst}
                    </span>
                  </div>
                )}
                {previewUrl && (
                  <a
                    className="message-link-preview"
                    href={previewUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <span className="message-link-preview-title">Link gedeeld</span>
                    <span className="message-link-preview-host">{toHostname(previewUrl)}</span>
                    <span className="message-link-preview-url">{previewUrl}</span>
                  </a>
                )}
                {/* Bijlagen tonen */}
                {b.bijlagen && b.bijlagen.length > 0 && (
                  <div className="message-bijlagen">
                    {b.bijlagen.map((bijlage) => {
                      const url = documentDownloadUrl(bijlage.bestandId);
                      const isImage = bijlage.contentType?.startsWith('image/');
                      return isImage ? (
                        <a key={bijlage.bestandId} href={url} target="_blank" rel="noopener noreferrer"
                           className="bijlage-image-link">
                          <img src={url} alt={bijlage.naam} className="bijlage-preview" />
                        </a>
                      ) : (
                        <a key={bijlage.bestandId} href={url} className="bijlage-file-link"
                           download={bijlage.naam}>
                          <span className="bijlage-icon">📄</span>
                          <span className="bijlage-info">
                            <span className="bijlage-naam">{bijlage.naam}</span>
                            <span className="bijlage-grootte">
                              {bijlage.grootte < 1024
                                ? `${bijlage.grootte} B`
                                : bijlage.grootte < 1048576
                                  ? `${(bijlage.grootte / 1024).toFixed(1)} KB`
                                  : `${(bijlage.grootte / 1048576).toFixed(1)} MB`}
                            </span>
                          </span>
                        </a>
                      );
                    })}
                  </div>
                )}
                <div
                  className="message-text"
                  dangerouslySetInnerHTML={{ __html: formatMessage(b.tekst) }}
                />
                <div className="message-footer">
                  <button className="btn-reply" onClick={() => setReplyTo(b)} title="Reageer" aria-label="Reageer op dit bericht">↩</button>
                  <span className="message-time">{formatTime(b.geleverd)}</span>
                  {/* Toon status alleen onder eigen berichten: ✓ verzonden, ✓✓ gelezen */}
                  {isMine && (() => {
                    const readByOthers = new Set(
                      (b.lezingen || [])
                        .map((l) => l.lezerId)
                        .filter((id) => id != null && id !== user.id)
                    );
                    const hasAnyRead = readByOthers.size > 0;
                    const allOthersRead =
                      otherParticipantIds.size > 0 &&
                      Array.from(otherParticipantIds).every((id) => readByOthers.has(id));

                    const statusClass = !hasAnyRead
                      ? 'sent'
                      : allOthersRead
                        ? 'all-read'
                        : 'partial-read';
                    const symbol = hasAnyRead ? '✓✓' : '✓';
                    const readersTitle = (b.lezingen || [])
                      .map((l) => l.lezer?.naam || 'Onbekend')
                      .join(', ');

                    return (
                      <span
                        className={`message-read ${statusClass}`}
                        title={readersTitle ? `Gelezen door: ${readersTitle}` : 'Verzonden'}
                        aria-label={readersTitle ? `Gelezen door: ${readersTitle}` : 'Verzonden'}
                      >
                        {symbol}
                      </span>
                    );
                  })()}
                </div>
              </div>
            </div>
          );
        })}
        <div ref={endRef} />
      </div>

      {newMessagesCount > 0 && (
        <div className="new-messages-bar">
          <button className="btn-new-messages" onClick={scrollToLatest}>
            {newMessagesCount === 1
              ? '1 nieuw bericht ↓'
              : `${newMessagesCount} nieuwe berichten ↓`}
          </button>
        </div>
      )}

      {error && <p className="error chat-error">{error}</p>}

      {/* Reply-preview */}
      {replyTo && (
        <div className="reply-preview">
          <div className="reply-preview-content">
            <span className="reply-preview-sender">{replyTo.bijdrager?.naam || 'Onbekend'}</span>
            <span className="reply-preview-text">
              {replyTo.tekst.length > 80 ? replyTo.tekst.slice(0, 80) + '…' : replyTo.tekst}
            </span>
          </div>
          <button className="reply-preview-close" onClick={() => setReplyTo(null)} aria-label="Annuleer reactie">✕</button>
        </div>
      )}

      {/* Geselecteerde bestanden wachtrij */}
      {pendingFiles.length > 0 && (
        <div className="pending-files">
          {pendingFiles.map((file, i) => (
            <div key={i} className="pending-file">
              <span className="pending-file-name">
                {file.type.startsWith('image/') ? '🖼️' : '📄'} {file.name}
              </span>
              <button className="pending-file-remove" onClick={() => removePendingFile(i)}
                      disabled={sending} title="Verwijderen" aria-label={`Verwijder ${file.name}`}>✕</button>
            </div>
          ))}
          {uploadProgress === 'uploading' && (
            <div className="upload-status">Bezig met uploaden…</div>
          )}
        </div>
      )}

      {draftPreviewUrl && (
        <div className="draft-link-preview" aria-live="polite">
          <span className="draft-link-preview-title">Voorvertoning link</span>
          <a href={draftPreviewUrl} target="_blank" rel="noopener noreferrer" className="draft-link-preview-url">
            {toHostname(draftPreviewUrl) || draftPreviewUrl}
          </a>
        </div>
      )}

      {draftHtml && (
        <div className="draft-message-preview" aria-live="polite">
          <span className="draft-message-preview-title">Conceptweergave</span>
          <div
            className="draft-message-preview-body message-text"
            dangerouslySetInnerHTML={{ __html: draftHtml }}
          />
        </div>
      )}

      <div className="chat-input">
        <input
          ref={fileInputRef}
          type="file"
          multiple
          accept="image/jpeg,image/png,image/gif,image/webp,application/pdf,.doc,.docx,.xlsx,.pptx,.txt"
          onChange={handleFileSelect}
          style={{ display: 'none' }}
        />
        <button
          className="btn-attach"
          onClick={() => fileInputRef.current?.click()}
          disabled={sending}
          title="Bestand bijvoegen"
          aria-label="Bestand bijvoegen"
        >
          <PaperclipIcon />
        </button>
        <EmojiPicker onSelect={insertEmoji} />
        <textarea
          ref={inputRef}
          value={tekst}
          onChange={(e) => handleTekstChange(e.target.value)}
          onKeyDown={handleKeyDown}
          onPaste={handlePaste}
          placeholder="Schrijf een bericht… (Shift+Enter voor nieuwe regel)"
          rows={2}
          disabled={sending}
        />
        <button
          className="btn-send"
          onClick={handleSend}
          disabled={(!tekst.trim() && pendingFiles.length === 0) || sending}
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
