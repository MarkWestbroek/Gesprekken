/**
 * Formatter voor berichttekst met WhatsApp-achtige opmaak.
 *
 * Ondersteunde opmaak:
 *   *vet*           →  <strong>vet</strong>
 *   _cursief_       →  <em>cursief</em>
 *   ~doorhalen~     →  <del>doorhalen</del>
 *   `code`          →  <code>code</code>
 *   - opsomming     →  <ul><li>opsomming</li></ul>
 *   • opsomming     →  (idem, met bullet-teken)
 *   Nieuwe regels   →  <br />
 *
 * Veiligheid: HTML wordt eerst geëscaped zodat gebruikersinvoer
 * niet als echte HTML geïnterpreteerd wordt (XSS-preventie).
 */
export function formatMessage(text) {
  // Stap 1: HTML-escaping om XSS te voorkomen
  let html = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');

  // Stap 2: Inline opmaak toepassen
  // Volgorde is belangrijk: `code` eerst, zodat * en _ binnen code niet omgezet worden
  html = html.replace(/`([^`]+)`/g, '<code>$1</code>');
  html = html.replace(/\*([^*]+)\*/g, '<strong>$1</strong>');
  html = html.replace(/(?<!\w)_([^_]+)_(?!\w)/g, '<em>$1</em>');
  html = html.replace(/~([^~]+)~/g, '<del>$1</del>');

  // Stap 3: Opsommingslijsten – regels die beginnen met • of - worden <li>'s
  const lines = html.split('\n');
  const result = [];
  let inList = false;

  for (const line of lines) {
    const bulletMatch = line.match(/^\s*[•\-]\s+(.*)/);
    if (bulletMatch) {
      if (!inList) {
        result.push('<ul>');
        inList = true;
      }
      result.push(`<li>${bulletMatch[1]}</li>`);
    } else {
      if (inList) {
        result.push('</ul>');
        inList = false;
      }
      result.push(line);
    }
  }
  if (inList) result.push('</ul>');

  // Stap 4: Resterende newlines omzetten naar <br />, maar niet binnen lijsten
  return result
    .join('\n')
    .replace(/(?<!<\/ul>|<ul>|<\/li>|<li>[^<]*)\n(?!<ul>|<\/ul>|<li>|<\/li>)/g, '<br />');
}
