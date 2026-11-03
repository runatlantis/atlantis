import DefaultTheme from "vitepress/theme";
import { defineAsyncComponent, h } from 'vue';
import DocumentationVersion from '../components/DocumentationVersion.vue';
import SharedSidebar from '../components/SharedSidebar.vue';

export default {
  ...DefaultTheme,
  Layout() {
    return h(DefaultTheme.Layout, null, {
      'layout-top': () => h(defineAsyncComponent(() => import('../components/Banner.vue'))),
      'nav-bar-content-after': () => h(DocumentationVersion, { navbar: true }),
      'doc-before': () => h(DocumentationVersion),
      'sidebar-nav-after': () => h(SharedSidebar)
    });
  }
};
