import DefaultTheme from "vitepress/theme";
import { defineAsyncComponent, h } from 'vue';
import "./index.scss";
import "./palette.scss";

export default {
  ...DefaultTheme,
  Layout() {
    return h(DefaultTheme.Layout, null, {
      'layout-top': () => h(defineAsyncComponent(() => import('../components/Banner.vue')))
    });
  }
};
