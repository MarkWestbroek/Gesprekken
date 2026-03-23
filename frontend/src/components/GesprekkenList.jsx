import { useState, useEffect } from 'react';
import { listGesprekken } from '../api';

/**
 * Overzichtsscherm: toont alle gesprekken waar de huidige gebruiker
 * deelnemer aan is. Elke kaart toont het onderwerp, de andere
 * deelnemers, en de aanvangsdatum.
 *
 * Props:
 *   user          – de actief gekozen gespreksdeelnemer
 *   onSelect(g)   – callback als een gesprek aangeklikt wordt
 *   onSwitchUser  – callback om terug te gaan naar de UserPicker
 *   onBeheer      – callback om het beheerscherm te openen
 */
export default function GesprekkenList({ user, onSelect, onSwitchUser, onBeheer }) {
  const [gesprekken, setGesprekken] = useState([]);
  const [error, setError] = useState(null);

  // Haal gesprekken op en filter op deelname van de huidige gebruiker
  useEffect(() => {
    listGesprekken()
      .then((all) => {
        // Toon alleen gesprekken waar de gebruiker aan deelneemt
        const mine = all.filter((g) =>
          g.deelnames?.some((d) => d.deelnemerId === user.id)
        );
        setGesprekken(mine);
      })
      .catch((e) => setError(e.message));
  }, [user.id]);

  return (
    <div className="screen gesprekken-list">
      <header className="list-header">
        <h2>Gesprekken</h2>
        <div className="header-right">
          <span className="current-user">{user.naam}</span>
          <button className="btn-switch" onClick={onSwitchUser}>
            Wissel gebruiker
          </button>
          <button className="btn-switch" onClick={onBeheer}>
            ⚙ Beheer
          </button>
        </div>
      </header>
      {error && <p className="error">{error}</p>}
      {gesprekken.length === 0 && !error && (
        <p className="empty">Geen gesprekken gevonden.</p>
      )}
      <ul className="gesprek-items">
        {gesprekken.map((g) => {
          // Toon de namen van de andere deelnemers bij elk gesprek
          const others = g.deelnames
            ?.filter((d) => d.deelnemerId !== user.id)
            .map((d) => d.deelnemer?.naam || 'Onbekend')
            .join(', ');
          return (
            <li key={g.id}>
              <button onClick={() => onSelect(g)}>
                <span className="gesprek-onderwerp">{g.onderwerp}</span>
                <span className="gesprek-meta">
                  met {others || '—'} · {new Date(g.aanvang).toLocaleDateString('nl-NL')}
                </span>
              </button>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
