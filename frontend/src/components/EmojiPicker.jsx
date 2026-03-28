import { useState, useRef, useEffect } from 'react';

/**
 * Compacte emoji-kiezer met een beperkte, zakelijke set emoticons
 * passend bij een overheidstoepassing. Geen exotische of informele
 * emoji's — alleen veelgebruikte uitdrukkingen, hand­gebaren en
 * objecten die in professionele communicatie gangbaar zijn.
 *
 * Props:
 *   onSelect(emoji) – callback wanneer een emoji wordt gekozen
 */

const EMOJI_CATEGORIES = [
  {
    label: 'Veelgebruikt',
    emojis: ['👍', '👎', '👋', '🙏', '👏', '🤝', '💪', '✅', '❌', '❓'],
  },
  {
    label: 'Reacties',
    emojis: ['😊', '😄', '🙂', '😉', '🤔', '😮', '😅', '😢', '😬', '🫡'],
  },
  {
    label: 'Objecten',
    emojis: ['📎', '📄', '📅', '📌', '📝', '💡', '🔔', '⏰', '🏠', '🔗'],
  },
  {
    label: 'Symbolen',
    emojis: ['⭐', '❤️', '🔥', '⚠️', 'ℹ️', '🎉', '➡️', '⬅️', '🔄', '✨'],
  },
];

export default function EmojiPicker({ onSelect }) {
  const [open, setOpen] = useState(false);
  const ref = useRef(null);

  // Sluit bij klik buiten het paneel
  useEffect(() => {
    if (!open) return;
    const handler = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [open]);

  // Sluit bij Escape-toets
  useEffect(() => {
    if (!open) return;
    const handler = (e) => { if (e.key === 'Escape') setOpen(false); };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [open]);

  return (
    <div className="emoji-picker-wrapper" ref={ref}>
      <button
        type="button"
        className="btn-emoji-toggle"
        onClick={() => setOpen((v) => !v)}
        aria-label="Emoji invoegen"
        aria-expanded={open}
        title="Emoji invoegen"
      >
        <span className="emoji-toggle-glyph" aria-hidden="true">😊</span>
      </button>

      {open && (
        <div className="emoji-panel" role="dialog" aria-label="Emojikiezer">
          {EMOJI_CATEGORIES.map((cat) => (
            <div key={cat.label} className="emoji-category">
              <span className="emoji-category-label">{cat.label}</span>
              <div className="emoji-grid">
                {cat.emojis.map((emoji) => (
                  <button
                    key={emoji}
                    type="button"
                    className="emoji-btn"
                    onClick={() => {
                      onSelect(emoji);
                      setOpen(false);
                    }}
                    aria-label={emoji}
                  >
                    {emoji}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
