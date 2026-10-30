#!/usr/bin/env bash
# Stage the website markdown under .lingo-src/ for lingo.dev.
#
# Lingo.dev locates source files by substituting the source locale into the
# [locale] placeholder of i18n.json, so the English tree is copied to
# .lingo-src/en/ and each existing translation to .lingo-src/<locale>/. The
# finished translations are copied back to runatlantis.io/<locale>/ afterwards.
#
# Usage: scripts/stage-translation-sources.sh <locale> [...more locales]
set -euo pipefail

rm -rf .lingo-src
mkdir -p .lingo-src/en

# Markdown only. The blog keeps its images and Terraform samples in per-post
# directories; those are not translated, and the link localizer already points
# a translated post back at the English copy of its assets.
# ADR wrappers load canonical records and are maintained by hand.
rsync -a --exclude='/contributing/adr/' --include='*/' --include='*.md' --exclude='*' --prune-empty-dirs \
  runatlantis.io/docs runatlantis.io/guide runatlantis.io/blog runatlantis.io/contributing \
  .lingo-src/en/
cp runatlantis.io/*.md .lingo-src/en/

# Stage the translations already on the branch too. Without them the translator
# sees no target files and re-translates every page on every run, which costs
# the full token budget and rewrites pages whose source did not change.
for locale in "$@"; do
  if [ -d "runatlantis.io/${locale}" ]; then
    cp -r "runatlantis.io/${locale}" ".lingo-src/${locale}"
  fi
done
