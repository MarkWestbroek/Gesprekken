# Gesprekken Frontend

Een eenvoudige chat-frontend gebouwd met **React** en **Vite** die communiceert met de Gesprekken REST API.

## Functionaliteit

1. **Deelnemer kiezen** - kies een gespreksdeelnemer om als die persoon te chatten (vergelijkbaar met inloggen). Open de app in een tweede tabblad om als een andere gebruiker te chatten.
2. **Gesprek kiezen** - bekijk alle gesprekken waar je aan deelneemt, met het onderwerp, de andere deelnemers en de datum.
3. **Chatten** - messenger-stijl chatvenster:
   - Eigen berichten rechts (groen), berichten van anderen links (wit)
   - Automatisch scrollen naar het nieuwste bericht
   - Polling elke 3 seconden voor nieuwe berichten
   - Enter verstuurt, Shift+Enter voor een nieuwe regel
   - Leesbevestigingen (✓✓) onder eigen berichten
4. **Collega erbij** - medewerkers (type `interne_actor`) zien een "+ Collega" knop in de chat-header waarmee ze een andere interne actor aan het gesprek kunnen toevoegen. Een modaal venster toont alleen collega's die nog niet deelnemen.

## Rich text opmaak

Berichten ondersteunen WhatsApp-achtige opmaak:

| Invoer | Resultaat |
|---|---|
| `*vet*` | **vet** |
| `_cursief_` | _cursief_ |
| `~doorhalen~` | ~~doorhalen~~ |
| `` `code` `` | `code` |
| `- item` | opsomming |

## Projectstructuur

```
frontend/
|- index.html              # HTML entry point
|- vite.config.js          # Vite config met proxy naar Go API
|- package.json
+- src/
    |- main.jsx            # React DOM render
    |- App.jsx             # Root-component met scherm-navigatie
    |- App.css             # Alle component-stijlen
    |- index.css           # Globale stijlen en CSS variabelen
    |- api.js              # API client (fetch wrapper)
    |- formatMessage.js    # WhatsApp-achtige tekst naar HTML formatter
    +- components/
        |- UserPicker.jsx      # Scherm 1: deelnemer kiezen
        |- GesprekkenList.jsx  # Scherm 2: gesprekkenlijst
        +- ChatView.jsx       # Scherm 3: chatvenster
```

## Starten

Zorg dat de Go API draait op poort 8080, en start dan de frontend:

```bash
cd frontend
npm install
npm run dev
```

De app is bereikbaar op http://localhost:5173.

Alle `/v1/*` requests worden door de Vite dev-server doorgestuurd naar `http://localhost:8080` (proxy).

## API endpoints die gebruikt worden

| Methode | Endpoint | Doel |
|---|---|---|
| GET | `/v1/gespreksdeelnemers` | Lijst van alle deelnemers (inclusief type) |
| GET | `/v1/gesprekken` | Lijst van alle gesprekken (met deelnames) |
| GET | `/v1/gesprekken/:id/bijdragen` | Bijdragen (berichten) van een gesprek |
| POST | `/v1/gesprekken/:id/bijdragen` | Nieuw bericht versturen |
| POST | `/v1/gesprekken/:id/bijdragen/:bijdrageId/lezingen` | Leesbevestiging registreren |
| GET | `/v1/deelnemertypen` | Lijst van deelnemertypen (voor collega-filter) |
| POST | `/v1/gesprekken/:id/deelnames` | Collega toevoegen aan een gesprek |
