<script setup>
import { computed } from 'vue';
import { useData } from 'vitepress';

const props = defineProps({ items: Array, nested: Boolean });
const { theme } = useData();
const links = computed(() => props.items ?? theme.value.sharedSidebar ?? []);
const href = (link) => link.endsWith('/') || link.endsWith('.html') ? link : `${link}.html`;
</script>

<template>
  <div v-if="links.length" class="shared-sidebar" :class="{ nested }">
    <template v-for="item in links" :key="item.text">
      <details v-if="item.items?.length" :open="item.collapsed !== true">
        <summary>
          <a v-if="item.link" :href="href(item.link)" target="_self">{{ item.text }}</a>
          <span v-else>{{ item.text }}</span>
        </summary>
        <SharedSidebar :items="item.items" nested />
      </details>
      <a v-else-if="item.link" :href="href(item.link)" target="_self">{{ item.text }}</a>
    </template>
  </div>
</template>

<style scoped>
.shared-sidebar > details { padding: 16px 0; border-top: 1px solid var(--vp-c-divider); }
.shared-sidebar summary { display: flex; align-items: center; justify-content: space-between; gap: 8px; cursor: pointer; list-style: none; font-weight: 700; }
.shared-sidebar summary::-webkit-details-marker { display: none; }
.shared-sidebar summary::after { content: '›'; transform: rotate(90deg); color: var(--vp-c-text-2); }
.shared-sidebar details:not([open]) > summary::after { transform: none; }
.shared-sidebar a, .shared-sidebar span { display: block; padding: 4px 0; font-size: 14px; line-height: 24px; }
.shared-sidebar a:hover { color: var(--vp-c-brand-1); }
.shared-sidebar.nested { margin-top: 4px; color: var(--vp-c-text-2); }
.shared-sidebar.nested > details { padding: 0; border: 0; }
.shared-sidebar.nested summary { font-weight: 500; }
.shared-sidebar.nested .nested { padding-left: 16px; border-left: 1px solid var(--vp-c-divider); }
</style>
