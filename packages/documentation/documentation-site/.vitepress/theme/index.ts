import type {Theme} from "vitepress";
import DefaultTheme from "vitepress/theme";
import SiteLayout from "./SiteLayout.vue";
import ThemeDiagram from "./ThemeDiagram.vue";
import "../../content/theme.css";
import "./style.css";

export default {
  extends: DefaultTheme,
  Layout: SiteLayout,
  enhanceApp({app}) {
    app.component("ThemeDiagram", ThemeDiagram);
  },
} satisfies Theme;
