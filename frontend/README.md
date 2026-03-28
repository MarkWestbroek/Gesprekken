# Gesprekken Frontend

Een eenvoudige chat-frontend gebouwd met **React** en **Vite** die communiceert met de Gesprekken REST API.

## Functionaliteit

1. **Deelnemer kiezen** - kies een gespreksdeelnemer om als die persoon te chatten (vergelijkbaar met inloggen). Open de app in een tweede tabblad om als een andere gebruiker te chatten.
2. **Gesprek kiezen** - bekijk alle gesprekken waar je aan deelneemt, met het onderwerp, de andere deelnemers en de datum.
3. **Chatten** - messenger-stijl chatvenster:
   - Eigen berichten rechts (lichtblauw), berichten van anderen links (wit)
    - Automatisch scrollen naar het nieuwste bericht zolang je onderaan zit
    - Als je omhoog scrolt, blijft de scrollpositie staan en verschijnt rechts-onder een badge voor nieuwe berichten
   - Polling elke 3 seconden voor nieuwe berichten
   - Enter verstuurt, Shift+Enter voor een nieuwe regel
    - Leesbevestigingen alleen voor berichten die echt zichtbaar zijn in het chatvenster
    - Status onder eigen berichten: `✓` = verzonden/opgeslagen, `✓✓` = gelezen; grijs zolang niet alle andere deelnemers hebben gelezen, donkerblauw zodra iedereen het gelezen heeft
4. **Collega erbij** - medewerkers (type `interne_actor`) zien een "+ Collega" knop in de chat-header waarmee ze een andere interne actor aan het gesprek kunnen toevoegen. Een modaal venster toont alleen collega's die nog niet deelnemen.
5. **Emoji-kiezer** - een beperkte set zakelijke emoji's (40 stuks) kan via het 😊-knopje worden ingevoegd in berichten.
6. **Reageren** - reply-to op een bericht toont de originele tekst als quote-blok.
7. **Links en previews** - `http(s)://` links in tekst worden klikbaar gemaakt; bij het eerste gevonden URL in een bericht tonen we een compacte preview-kaart (host + URL). Tijdens typen tonen we ook een **voorvertoning** boven de composer.
8. **ASCII-smileys** - klassieke notatie zoals `:)`, `:-)`, `;)`, `;-)`, `:(`, `:-(`, `:D`, `:-D`, `:P`, `:-P` wordt automatisch omgezet naar emoji, zowel **tijdens typen** als in de berichtweergave.

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
|- index.html              # HTML entry point (lang="nl")
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
        |- ChatView.jsx        # Scherm 3: chatvenster
        |- EmojiPicker.jsx     # Emoji-kiezer (zakelijke set)
        +- DeelnemersBeheer.jsx # Scherm 4: CRUD deelnemers
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

## Huisstijl en visueel ontwerp

De frontend volgt de visuele identiteit van **Common Ground** (commonground.nl):

| Element | Keuze | Reden |
|---|---|---|
| Primaire kleur | `#009ee3` (CG-blauw) | Herkenbare huiskleur Common Ground |
| Donker accent | `#01689b` (CG-donkerblauw) | Gradient headers, primaire-actie contrast |
| Geel accent | `#ffbc2c` (CG-geel) | Knoppen, badges, datumscheiders — warm accent naast het koele blauw |
| Navy | `#1d3f72` | Koppen en woordmerk-tekst |
| Koppen | **Montserrat** 600–800 | Krachtig, zakelijk lettertype passend bij overheidsuitstraling |
| Broodtekst | **Source Sans 3** 400–700 | Goede leesbaarheid op scherm, hoge x-hoogte |
| Beeldmerk | `public/cg-brandmark.svg` | Officiële Common Ground SVG aangeleverd door de gebruiker |

Kleurvariabelen staan in `src/index.css` onder `:root`. Wijzig ze daar om het thema aan te passen.

## NL Design System richtlijnen

De frontend houdt rekening met de [NL Design System-richtlijnen](https://nldesignsystem.nl/richtlijnen/) (WCAG 2.2 AA + selecte AAA):

| Richtlijn | Status | Toelichting |
|---|---|---|
| **Taalinstelling** (WCAG 3.1.1) | ✅ | `<html lang="nl">` in `index.html` |
| **Typografie** | ✅ | Leesbare fonts, `line-height: 1.5`, `font-size ≥ 16px` voor broodtekst |
| **Focus-indicatoren** (WCAG 2.4.13 AAA) | ✅ | 3px donkerblauwe outline op `:focus-visible`; witte outline op donkere achtergronden |
| **Aanwijsgebied** (WCAG 2.5.8) | ✅ | Alle knoppen minimaal 24×24px; icoon-knoppen 36×36px |
| **Kleurcontrast tekst** (WCAG 1.4.3) | ✅ | Broodtekst `#262d30` op wit → >12:1; witte tekst op donkerblauwe header → >5:1 |
| **Aria-labels** (WCAG 4.1.2) | ✅ | Alle icoon-knoppen (📎 ↩ ✕ ← ✎ 🗑 😊) hebben een `aria-label` |
| **Kleur niet als enig middel** (WCAG 1.4.1) | ✅ | Leesstatus gebruikt ✓/✓✓ symbolen naast kleur |
| **Toetsenbordnavigatie** (WCAG 2.1.1) | ✅ | Alle acties bereikbaar met Tab + Enter/Space; modals met Escape |

### Aandachtspunten voor toekomstig werk

- Contrastverhouding van witte tekst op `#009ee3` (lichte uiteinde header-gradient) is ~2,9:1 — voldoende voor grote/vette tekst maar niet voor kleine tekst. De subtitel in de chat-header gebruikt `rgba(255,255,255,.92)` i.p.v. `opacity`.
- Bij inbedding als widget: test of `prefers-reduced-motion` en `prefers-color-scheme` correct doorgegeven worden door het portaal.
- Overweeg een skip-link (`#main-content`) wanneer de chat naast andere portaal-navigatie staat.

## Widget / inbedding

De chat is ontworpen om als **scherm of widget** binnen een klant- of medewerkersportaal ingebed te worden:

- **Geen harde viewport-hoogte op schermen**: CSS `.screen` gebruikt `flex: 1` i.p.v. `min-height: 100vh`. Wanneer `#root` in een flex/grid-container van het portaal staat, neemt het de beschikbare hoogte over.
- **`max-width: 900px`**: beperkt de breedte. Pas `#root` in `index.css` aan als het portaal een andere breedte voorschrijft.
- **Geen globale stijl-overrides**: stijlen zijn gescoped via klasse-namen. Er zijn geen element-selectors behalve `body` en `*` in `index.css`.
- **Proxy-config**: de Vite dev-proxy stuurt `/v1/*` door naar `localhost:8080`. In productie moet het portaal de API-base-URL configureren.

## Emoji-kiezer

De emoji-kiezer (`EmojiPicker.jsx`) biedt een zakelijke selectie van 40 emoji's in 4 categorieën (Veelgebruikt, Reacties, Objecten, Symbolen). De set is bewust beperkt: het is een overheidstoepassing, dus geen exotische of informele emoji's.

De kiezer opent boven de composer-balk, sluit bij klik erbuiten of Escape, en voegt de emoji in op de cursorpositie.

## Linkherkenning, datumlabels en live composer-gedrag

- Datumseparatoren tonen automatisch **Gisteren** voor berichten van de vorige kalenderdag.
- URL's in berichttekst worden bij weergave omgezet naar klikbare links.
- Voor het eerste URL in een bericht tonen we een compacte preview-kaart (label + host + URL).
- Tijdens typen in de composer tonen we al een link-voorvertoning voor het eerste gevonden URL (ook wanneer de URL in een bullet-regel staat).
- ASCII-smileys worden al in het invoerveld vervangen, zodat gebruikers direct feedback krijgen vóór verzenden.
- Tijdens typen tonen we ook een **conceptweergave** van de berichtopmaak (inclusief bullets/lijsten bij `-` + Shift+Enter).

## Reageren op bijdragen

Gebruikers kunnen op een eerder bericht reageren (reply-to):

- Hover over een bericht toont de ↩ knop
- Klikken opent een reply-preview boven het invoerveld
- Het verstuurde bericht toont een quote-blok met de originele afzender en tekst
- Klikken op een quote-blok scrollt naar het originele bericht en markeert het kort
