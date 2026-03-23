import { useState, useEffect } from 'react';
import { listDeelnemers } from '../api';

/**
 * Startscherm: toont alle gespreksdeelnemers en laat de gebruiker
 * er een kiezen om als die persoon "in te loggen".
 *
 * Props:
 *   onSelect(deelnemer) – callback als een deelnemer gekozen wordt
 */
export default function UserPicker({ onSelect }) {
  const [deelnemers, setDeelnemers] = useState([]);
  const [error, setError] = useState(null);

  // Laad de deelnemerslijst bij eerste render
  useEffect(() => {
    listDeelnemers()
      .then(setDeelnemers)
      .catch((e) => setError(e.message));
  }, []);

  return (
    <div className="screen user-picker">
      <h2>Kies een deelnemer</h2>
      {error && <p className="error">{error}</p>}
      <ul className="user-list">
        {deelnemers.map((d) => (
          <li key={d.id}>
            <button onClick={() => onSelect(d)}>
              <span className="user-name">{d.naam}</span>
              <span className="user-ref">{d.referentie}</span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}
