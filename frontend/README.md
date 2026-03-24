# Gesprekken Frontend

Een eenvoudige chat-frontend gebouwd met **React** en **Vite** die communiceert met de Gesprekken REST API.

## Functionaliteit

1. **Deelnemer kiezen** - kies een gespreksdeelnemer om als die persoon te chatten (vergelijkbaar met inloggen). Open de app in een tweede tabblad om als een andere gebruiker te chatten.
2. **Gesprek kiezen** - bekijk alle gesprekken waar je aan deelneemt, met het onderwerp, de andere deelnemers en de datum.
3. **Chatten** - messenger-stijl chatvenster:
   - Eigen berichten rechts (groen), berichten van anderen links (wit)
    - Automatisch scrollen naar het nieuwste bericht zolang je onderaan zit
    - Als je omhoog scrolt, blijft de scrollpositie staan en verschijnt rechts-onder een badge voor nieuwe berichten
   - Polling elke 3 seconden voor nieuwe berichten
   - Enter verstuurt, Shift+Enter voor een nieuwe regel
    - Leesbevestigingen alleen voor berichten die echt zichtbaar zijn in het chatvenster
    - Status onder eigen berichten: `✓` = verzonden/opgeslagen, `✓✓` = gelezen; grijs zolang niet alle andere deelnemers hebben gelezen, accentkleur zodra iedereen het gelezen heeft
4. **Collega erbij** - medewerkers (type `interne_actor`) zien een "+ Collega" knop in de chat-header waarmee ze een andere interne actor aan het gesprek kunnen toevoegen. Een modaal venster toont alleen collega's die nog niet deelnemen.

## Laatste UX-aanpassingen chat

- Auto-scroll forceert de gebruiker niet meer terug naar beneden tijdens polling.
- Nieuwe berichten worden rechts-onder aangekondigd met een knop zodra je omhoog bent gescrold.
- Leesbevestigingen worden alleen geregistreerd voor berichten die zichtbaar zijn in het chatvenster.
- Berichtstatus gebruikt nu WhatsApp-achtige vinkjes: `✓` voor verzonden en `✓✓` voor gelezen.
- Dubbele vinkjes blijven grijs totdat alle andere deelnemers het bericht hebben gelezen.

## Leesstatus en scrollgedrag

- De frontend verstuurt geen aparte bezorgbevestiging vanuit clients of devices. Een enkel vinkje (`✓`) betekent daarom: het bericht is succesvol opgeslagen via de API.
- Dubbele vinkjes (`✓✓`) betekenen: ten minste één andere deelnemer heeft het bericht gelezen.
- De dubbele vinkjes blijven grijs totdat alle andere actieve deelnemers in het gesprek het bericht hebben gelezen.
- Leesbevestigingen worden pas geregistreerd wanneer een bericht zichtbaar is binnen de viewport van het chatvenster.
- Tijdens polling blijft de scrollpositie behouden als je niet onderaan staat. Nieuwe berichten worden dan niet in beeld geforceerd, maar aangekondigd met een knop rechts-onder om naar het nieuwste bericht te springen.

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
