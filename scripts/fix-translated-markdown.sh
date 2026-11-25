#!/usr/bin/env bash
# Repair the mechanical markdown problems the translator leaves behind.
#
# Website Check runs markdownlint on every locale. Translated tables rarely keep
# the source's column widths, which fails MD060, and the model occasionally
# leaves a stray space inside link text or doubles a blank line. Findings that
# `--fix` cannot repair are left for the integrity step, which retries or
# quarantines the page instead of failing the run.
#
# Usage: scripts/fix-translated-markdown.sh <locale> [...more locales]
set -euo pipefail

# The translator's markdown serializer writes a link whose only content is an
# image as five lines: "[", blank, the image, blank, "](url)". That trips MD034
# and renders a stray bracket. Fold it back onto one line, which is how the
# English source writes it.
for locale in "$@"; do
  # shellcheck disable=SC2016 # $1 is perl's capture group, not a shell variable.
  find "runatlantis.io/${locale}" -name '*.md' -print0 \
    | xargs -0 perl -0pi -e 's/\[\n\n(!\[[^\]]*\]\([^)]*\))\n\n\]\(/[$1](/g'
done

node scripts/align-tables.mjs "$@"

# Quoted so markdownlint-cli2 expands `**` itself. Unquoted, bash treats `**`
# as `*` and the top-level <locale>/*.md pages are never linted.
globs=()
for locale in "$@"; do globs+=("runatlantis.io/${locale}/**/*.md"); done
npx --yes markdownlint-cli2@0.20.0 --fix --config .markdownlint.yaml "${globs[@]}" || true
