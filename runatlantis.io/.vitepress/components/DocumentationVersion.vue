<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useData } from 'vitepress';
import { versionTarget } from '../versioning.mjs';

const props = defineProps({ navbar: Boolean });
const { theme, page } = useData();
const versions = computed(() => theme.value.docVersions ?? []);
const options = computed(() => [...versions.value].sort((a, b) => Number(b.version === 'main') - Number(a.version === 'main')));
const latest = computed(() => versions.value.find((release) => release.version !== 'main')?.version);
const current = computed(() => theme.value.docVersion || latest.value);
const versionedPage = computed(() => /^(?:es\/)?(docs|guide)(\/|\.md$)/.test(page.value.relativePath));
const loading = ref(false);
const ready = ref(false);
const notice = ref('');
let manifest;

onMounted(() => {
  if (!versions.value.length) return;
  ready.value = true;
  if (props.navbar) {
    window.addEventListener('pageshow', restoreSelector);
    return;
  }
  const query = new URLSearchParams(window.location.search);
  if (query.has('missing-page')) notice.value = 'That page is unavailable in this release. Showing the documentation index.';
  else if (query.has('missing-section')) notice.value = `The section “${query.get('missing-section')}” is unavailable in this release.`;
});

onUnmounted(() => {
  if (props.navbar) window.removeEventListener('pageshow', restoreSelector);
});

function restoreSelector() {
  loading.value = false;
  const select = document.getElementById('documentation-version');
  if (select) select.value = current.value;
}

function changeVersion(event) {
  navigate(event.target.value);
}

async function navigate(version) {
  loading.value = true;
  try {
    if (!manifest) {
      const response = await fetch('/doc-versions.json', { cache: 'no-cache' });
      if (!response.ok) throw new Error('Documentation versions could not be loaded. Please try again.');
      manifest = await response.json();
    }
    const target = versionTarget(page.value.relativePath, version, manifest);
    const url = new URL(target.path, window.location.origin);
    if (target.missingPage) url.searchParams.set('missing-page', page.value.relativePath);
    const response = await fetch(url);
    if (!response.ok) throw new Error('Documentation could not be loaded. Please try again.');
    if (window.location.hash && !target.missingPage) {
      let id;
      try { id = decodeURIComponent(window.location.hash.slice(1)); }
      catch {
        window.location.assign(url.href);
        return;
      }
      const document = new DOMParser().parseFromString(await response.text(), 'text/html');
      if (document.getElementById(id)) url.hash = window.location.hash;
      else url.searchParams.set('missing-section', id);
    }
    window.location.assign(url.href);
  } catch (error) {
    notice.value = error.message;
    document.getElementById('documentation-version').value = current.value;
    loading.value = false;
  }
}
</script>

<template>
  <div v-if="versions.length && navbar && versionedPage" class="documentation-version-nav">
    <label for="documentation-version" class="visually-hidden">Documentation version</label>
    <div class="version-select">
      <select id="documentation-version" :value="current" :disabled="!ready || loading" title="Documentation version" @change="changeVersion">
        <option v-for="release in options" :key="release.version" :value="release.version">{{ release.version === 'main' ? 'Main (next)' : release.version + (release.version === latest ? ' (latest)' : '') }}</option>
      </select>
    </div>
    <p v-if="notice" class="version-error" role="alert">{{ notice }}</p>
  </div>
  <aside v-else-if="versions.length && !navbar && versionedPage && (current === 'main' || notice)" class="documentation-version" aria-label="Documentation notice">
    <p v-if="current === 'main'">Development documentation from main. Features may be unreleased.</p>
    <p v-if="notice" role="status">{{ notice }}</p>
  </aside>
</template>

<style scoped>
.documentation-version { margin-bottom: 24px; padding: 16px; border: 1px solid var(--vp-c-divider); border-radius: 8px; background: var(--vp-c-bg-soft); font-size: 14px; }
.documentation-version-nav { position: relative; flex-shrink: 0; font-size: 14px; }
.version-select { position: relative; }
.version-select::after { position: absolute; top: 12px; right: 12px; width: 6px; height: 6px; border-right: 1.5px solid var(--vp-c-text-2); border-bottom: 1.5px solid var(--vp-c-text-2); transform: rotate(45deg); pointer-events: none; content: ''; }
select { width: 156px; padding: 5px 28px 5px 10px; border: 1px solid var(--vp-c-divider); border-radius: 6px; background: var(--vp-c-bg-soft); color: var(--vp-c-text-1); font-size: 14px; line-height: 22px; cursor: pointer; }
select:focus-visible { outline: 2px solid var(--vp-c-brand-1); outline-offset: 2px; }
p { margin: 0; line-height: 1.6; }
.documentation-version p + p { margin-top: 8px; }
.version-error { position: absolute; top: calc(100% + 8px); right: 0; padding: 12px; width: min(300px, 80vw); border: 1px solid var(--vp-c-divider); border-radius: 6px; background: var(--vp-c-bg); white-space: normal; }
@media (max-width: 767px) { select { width: 140px; } }
@media (max-width: 359px) { select { width: 124px; padding-left: 6px; font-size: 12px; } }
@media (min-width: 768px) {
  :global(.VPNavBar:has(.documentation-version-nav) .VPNavBarMenuLink:not([href$="/docs"]):not([href$="/docs/"]):not([href$="/docs.html"])) { display: none; }
  :global(.VPNavBar:has(.documentation-version-nav) .VPNavBarTranslations),
  :global(.VPNavBar:has(.documentation-version-nav) .VPNavBarAppearance),
  :global(.VPNavBar:has(.documentation-version-nav) .VPNavBarSocialLinks),
  :global(.VPNavBar:has(.documentation-version-nav) .VPNavBarExtra) { order: 1; }
}
@media (min-width: 768px) and (max-width: 959px) {
  select { width: 140px; }
  :global(.VPNavBar:has(.documentation-version-nav) .VPNavBarSearch) { padding-left: 12px; }
  :global(.VPNavBar:has(.documentation-version-nav) .DocSearch-Button-Placeholder),
  :global(.VPNavBar:has(.documentation-version-nav) .DocSearch-Button-Keys) { display: none; }
}
</style>
