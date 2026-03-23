import { useState } from 'react';
import UserPicker from './components/UserPicker';
import GesprekkenList from './components/GesprekkenList';
import ChatView from './components/ChatView';
import './App.css';

/**
 * Root-component met een eenvoudige scherm-navigatie:
 *
 *   1. UserPicker     – kies een deelnemer ("inloggen")
 *   2. GesprekkenList  – kies een gesprek waar je aan deelneemt
 *   3. ChatView        – chat in het gekozen gesprek
 *
 * State:
 *   user    – de gekozen gespreksdeelnemer, of null
 *   gesprek – het gekozen gesprek, of null
 */
export default function App() {
  const [user, setUser] = useState(null);
  const [gesprek, setGesprek] = useState(null);

  // Stap 1: nog geen gebruiker gekozen → toon de deelnemerkiezer
  if (!user) {
    return <UserPicker onSelect={setUser} />;
  }

  // Stap 2: nog geen gesprek gekozen → toon de gesprekkenlijst
  if (!gesprek) {
    return (
      <GesprekkenList
        user={user}
        onSelect={setGesprek}
        onSwitchUser={() => setUser(null)}
      />
    );
  }

  // Stap 3: gesprek gekozen → toon het chatvenster
  return (
    <ChatView
      user={user}
      gesprek={gesprek}
      onBack={() => setGesprek(null)}
    />
  );
}
