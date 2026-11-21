# Translating PND

Translations live in `web/src/locales/`. English (`en.json`) defines the key set, and Traditional Chinese (`zh-TW.json`) is included.

1. Copy `en.json` to a BCP 47 locale name such as `de.json`.
2. Translate values while preserving keys and technical identifiers such as Node ID, IPv4, and GiB.
3. Import the new file and add it to the locale map and language selector in `web/src/main.ts`.
4. Run `pnpm run check:locales` and `pnpm run build`.
5. Check forms, contextual help, validation messages, and narrow-screen layouts in the new language.

All locales share stable semantic keys. Missing runtime values fall back to English; repository checks require complete key coverage. User language selection is stored locally in the browser. First use follows the browser's language for English and Chinese. Theme preference is independent of language.

Submit translations with the locale file and small registration change. Project documentation remains in English. Translation resources may contain the target language's characters and punctuation.

The static experience has its own small key set in `demo/src/locales/`. Add the corresponding locale and register it in `demo/src/main.ts` when translating the demo. `pnpm run check:locales` checks both directories.
