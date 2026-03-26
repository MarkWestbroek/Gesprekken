# Chat Bijlagen — Frontend

Bijlage-functionaliteit geïntegreerd in de bestaande ChatView composer.

## Gebruik

De chat composer bevat een 📎 (paperclip) knop links van het tekstveld. Hiermee kan de gebruiker:

1. **Bestand selecteren** — opent native file picker (meerdere bestanden mogelijk)
2. **Afbeelding plakken** — Ctrl+V / ⌘V plakt een afbeelding uit het clipboard (bijv. screenshot)
3. **Preview** — geselecteerde bestanden verschijnen als chips onder het tekstveld
4. **Verwijderen** — klik ✕ op een chip om het bestand te verwijderen vóór verzending
5. **Versturen** — bestanden worden geüpload, daarna wordt het bericht aangemaakt met bijlage-referenties

## Upload flow

```
[Gebruiker selecteert bestand of plakt afbeelding]
        ↓
[Client-side validatie: grootte + MIME-type]
        ↓
[Bestand verschijnt als pending chip]
        ↓
[Gebruiker klikt "Verstuur"]
        ↓
[POST /v1/documenten per bestand → bestandId]
        ↓
[POST /v1/gesprekken/:id/bijdragen met bijlageIds]
        ↓
[Bericht + bijlagen verschijnen in chat]
```

## Bijlagen in berichten

- **Afbeeldingen** (JPEG/PNG/GIF/WebP) → inline thumbnail preview, klikbaar naar volledige download
- **Documenten** (PDF/DOC/XLSX/etc.) → downloadlink met bestandsnaam en grootte

## Validatie (client-side)

| Regel | Waarde |
|---|---|
| Max bestandsgrootte | 25 MB |
| Toegestane types | Afbeeldingen, PDF, Office documenten, platte tekst |

Foutmeldingen verschijnen als rood foutbericht in de chat.

## Componenten

| Bestand | Beschrijving |
|---|---|
| `src/api.js` | `uploadDocument()`, `documentDownloadUrl()`, `MAX_FILE_SIZE`, `ALLOWED_MIME_TYPES` |
| `src/components/ChatView.jsx` | Composer integratie + berichtweergave met bijlagen |
| `src/App.css` | Styling: `.message-bijlagen`, `.pending-files`, `.btn-attach` |

## Afhankelijkheden

Geen extra npm packages — gebruikt native `FormData` en `fetch`.
