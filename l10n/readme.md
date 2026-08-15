# 🔤 Localization

This directory contains everything related to client-side localization of error pages. The browser automatically
detects the user's language via `navigator.languages` and replaces English strings with translated ones - no server
involvement required.

Eep defaults to browser detection with `locale: auto`. An explicit plugin locale appends a call to
`window.l10n.setLocale(...)`; English locales omit the script entirely. Localization applies only to
HTML responses.

## How it works end-to-end

```
locales.json (source of truth - edit this)
     │
generate/localize.go (code generator)
     │
     ├── localize(.min).js (ignored build artifact embedded into eep)
     ├── supported_locales.go (ignored locale-validation build artifact)
     └── playground.html (ignored standalone page for manually testing all tokens and languages)
```

Mise and Docker builds invoke the generator automatically. Generated outputs are intentionally not committed.

### Runtime flow in the browser

1. The page loads with English text in elements that carry the `data-l10n` attribute, e.g.:
   ```html
   <span data-l10n>Not Found</span>
   <p data-l10n>The server can not find the requested page</p>
   ```
2. The inline `<script>{{ l10nScript }}</script>` block contains the `localize.min.js` script
3. On `DOMContentLoaded` the script reads `navigator.languages`, finds translations for each `[data-l10n]` element,
   and replaces the element's `textContent` in-place
4. A `MutationObserver` handles elements added dynamically after initial load
5. English (`en`/`en-*`) is the passthrough - the original text is kept as-is
6. BCP 47 resolution: `zh-TW` tries `zh-tw` first, then falls back to `zh`

### `window.l10n` public API

The script exposes a frozen object on `window.l10n`:

| Method                           | Description                                                                         |
| -------------------------------- | ----------------------------------------------------------------------------------- |
| `l10n.setLocale(locale)`         | Override the locale (string or array of strings); re-localizes the page immediately |
| `l10n.translate(token)`          | Return the translated token, or `null` when no translation is available             |
| `l10n.localizeDocument(locale?)` | Re-localize all `[data-l10n]` elements, optionally changing the language            |

## `locales.json` structure

The file is a JSON object. Every top-level key is an English source string (called a **token**). The value is an object
mapping [BCP 47](https://www.rfc-editor.org/rfc/rfc5646) language codes to translated strings.

```json
{
  "$schema": "locales.schema.json",
  "Not Found": {
    "de": "Nicht gefunden",
    "es": "No encontrado",
    "fr": "Introuvable",
    "ru": "Страница не найдена",
    "zh": "未找到"
  },
  "The server can not find the requested page": {
    "de": "Der Server kann die angeforderte Seite nicht finden",
    "fr": "Le serveur ne peut trouver la page demandée",
    "ru": "Сервер не смог найти запрашиваемую страницу",
    "zh": "服务器找不到请求的页面"
  }
}
```

### Rules

- Keys are the English strings exactly as they appear in `data-l10n` element text content
- Token matching is case-insensitive and strips all non-alphanumeric characters, so `"Not Found"`, `"not found"`,
  and `"NOT FOUND"` all resolve to the same token
- Language codes must be valid BCP 47 codes
- Every key must have at least one translation

## Running the generator

Normal builds generate localization automatically. To generate the artifacts manually from the **project root**:

```sh
mise run l10n-generate
```

## How to add a new language

1. Open `locales.json`. Add your language code and translation to **every** token. A token without your language
   code silently falls back to English
2. Run `mise run l10n-generate`
3. Open `l10n/playground.html` in a browser. Your new language button should appear. Click it and verify every string
   translates correctly
4. Send a PR with the source changes and add yourself to the list of translators below. Generated files are ignored.

## 👍 Translators

- 🇫🇷 French by [@jvin042](https://github.com/jvin042)
- 🇵🇹 Portuguese by [@fabtrompet](https://github.com/fabtrompet)
- 🇳🇱 Dutch by [@SchoNie](https://github.com/SchoNie)
- 🇩🇪 German by [@mschoeffmann](https://github.com/mschoeffmann)
- 🇪🇸 Spanish by [@Runig006](https://github.com/Runig006)
- 🇨🇳 Chinese by [@CDN18](https://github.com/CDN18)
- 🇮🇩 Indonesian by [@getwisp](https://github.com/getwisp)
- 🇵🇱 Polish by [@wielorzeczownik](https://github.com/wielorzeczownik)
- 🇰🇷 Korean by [@NavyStack](https://github.com/NavyStack)
- 🇭🇺 Hungarian by [@oszto90](https://github.com/oszto90)
- 🇳🇴 Norwegian by [@EliasTors](https://github.com/EliasTors)
- 🇷🇴 Romanian by [@pasarenicu](https://github.com/pasarenicu)
- 🇮🇹 Italian by [@Vigno04](https://github.com/Vigno04)
