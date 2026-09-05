# Website translations

The docs on [runatlantis.io](https://www.runatlantis.io) are machine-translated and
reviewed by maintainers before merge. English is the source of truth; a translated
page is a convenience, never an authority.

Currently translated: **Spanish (`es`)** — matching the locales the server itself
ships (see `--language` and `server/i18n/locales/`).

## How it works

`.github/workflows/translate-docs.yml` runs when English docs change on `main`:

1. Stages a throwaway copy of the English pages under `.lingo-src/en/`. The
   translator finds sources by substituting the source locale into a `[locale]`
   path, and staging keeps the real English tree exactly where it is.
2. Runs [Lingo.dev](https://github.com/lingodotdev/lingo.dev) (Apache-2.0) with the
   config in `i18n.json`. It parses markdown into an AST and skips code nodes, so
   fenced blocks and inline code are never sent for translation. `i18n.lock`
   records a checksum per content section, so an edit to one paragraph
   re-translates that paragraph and nothing else.
3. Copies the results to `runatlantis.io/<locale>/`.
4. Runs `scripts/localize-links.mjs` and `scripts/check-translation-integrity.mjs`.
5. Builds the site, then opens a PR labelled `docs`. It never pushes to `main`.

### Link handling

A locale is translated page by page, and VitePress fails the build on dead relative
links. `scripts/localize-links.mjs` rewrites any relative link or image whose
localized target doesn't exist yet so it points back into the English tree. Readers
land on the English page instead of a 404, and the nav and sidebar do the same
thing via the fallback in `runatlantis.io/.vitepress/sidebars.ts`.

### Integrity checking

`scripts/check-translation-integrity.mjs` compares every translated page against its
English source and fails if any of these drifted:

- fenced code blocks (count and exact content)
- inline code spans
- link targets, compared after resolution so the rewrite above is allowed but an
  arbitrary retarget is not
- VitePress container keywords (`:::tip`, `:::warning`, …)
- `<Badge>` markup
- heading count
- frontmatter keys, and the values VitePress interprets (`layout`, `aside`, …)

It runs in `website.yml` too, so hand-edits to a locale are checked the same way.

## Running it locally

```bash
# Translate (needs a provider API key — see the provider block in i18n.json)
export OPENAI_API_KEY=...
npx lingo.dev@0.138.4 run

# Fix up links in a partially translated locale, then verify
node scripts/localize-links.mjs es
node scripts/check-translation-integrity.mjs es

# The real gate
npm run website:build
```

## Adding a locale

1. Add the code to `locale.targets` in `i18n.json`.
2. Add a locale entry in `runatlantis.io/.vitepress/config.ts` with its UI strings.
3. Add label maps in `navbars.ts` and `sidebars.ts`.
4. Run the workflow with `workflow_dispatch`.

Add one locale at a time. Each one multiplies the link-check surface and, more
importantly, the review burden — a locale nobody can review is worse than no locale.

## Reviewing a translation PR

Focus on what machine translation gets wrong in reference docs:

- flag names, environment variables, YAML keys and file paths left in English
- command output and code samples untouched (the checker enforces this, but read them)
- `security.md`, `server-configuration.md` and anything about credentials deserve a
  careful read rather than a skim

## Changing the model or provider

The `provider` block in `i18n.json` accepts `openai`, `anthropic`, `google`,
`mistral`, `openrouter` and `ollama`, each reading its usual `*_API_KEY` environment
variable. Remove the block entirely to use Lingo.dev's hosted engine instead, which
requires an account.

The key lives in the `docs-translation` GitHub environment, not as a repository
secret, so only the translate job can read it. It belongs to a dedicated service
account and project on the provider side with a spend cap, so it can be rotated
without touching anything else.

The current model is a mini tier deliberately. Translating a 200-word section with
code already stripped out does not benefit from a frontier model, and the integrity
checks below are what actually guarantee correctness. `{source}` and `{target}` in
the prompt are substituted with locale codes by the CLI.
