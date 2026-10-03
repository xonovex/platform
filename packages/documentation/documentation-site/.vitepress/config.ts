import {dirname, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {defineConfig} from "vitepress";
import {logo} from "../src/brand.js";
import {createCatalog} from "../src/catalog.js";
import {repositoryMarkdown} from "../src/markdown-renderer.js";
import {componentTypes} from "../src/model.js";
import {searchChunkName, searchChunks} from "../src/search-chunks.js";
import {categoryNames} from "../src/site.js";

const site = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const root = resolve(site, "../../..");
const catalog = createCatalog(root);
const categories = componentTypes;
const base = process.env.DOCS_BASE ?? "/";
if (!base.startsWith("/") || !base.endsWith("/"))
  throw new Error("DOCS_BASE must start and end with a slash.");

export default defineConfig({
  title: "Xonovex",
  description: "Agent tools, skills, plugins, and development workflows.",
  srcDir: "content",
  outDir: "dist",
  base,
  appearance: true,
  ignoreDeadLinks: false,
  head: [
    ["link", {rel: "icon", type: "image/svg+xml", href: base + "favicon.svg"}],
    ["link", {rel: "preconnect", href: "https://fonts.googleapis.com"}],
    [
      "link",
      {rel: "preconnect", href: "https://fonts.gstatic.com", crossorigin: ""},
    ],
    [
      "link",
      {
        rel: "stylesheet",
        href: "https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Space+Grotesk:wght@500;600;700&display=swap",
      },
    ],
  ],
  themeConfig: {
    siteTitle: "Xonovex",
    logo,
    nav: [
      {
        text: "Catalog",
        items: categories.map((type) => ({
          text: categoryNames[type],
          link: "/catalog/" + type + ".html",
        })),
      },
      {text: "Workflow", link: "/plugins/xonovex-workflow.html"},
      {text: "Documentation", link: "/guide/platform.html"},
    ],
    sidebar: [
      {
        text: "Get started",
        items: [
          {text: "Platform overview", link: "/guide/platform.html"},
          {text: "Contributing", link: "/guide/contributing.html"},
          {text: "Design", link: "/guide/design.html"},
        ],
      },
      {
        text: "Component catalogs",
        items: categories.map((type) => ({
          text: categoryNames[type],
          link: "/catalog/" + type + ".html",
        })),
      },
      {
        text: "Plugins",
        collapsed: false,
        items: catalog.plugins.map((plugin) => ({
          text: plugin.name.replace(/^xonovex-/u, ""),
          link: "/plugins/" + plugin.name + ".html",
        })),
      },
      {
        text: "Agent tools",
        collapsed: true,
        items: catalog.entries
          .filter((entry) => entry.type === "tools")
          .map((entry) => ({text: entry.name, link: entry.link + ".html"})),
      },
    ],
    search: {provider: "local"},
    socialLinks: [{icon: "github", link: catalog.repository}],
    outline: [2, 3],
    footer: {
      message: "Documentation from the Xonovex repository.",
      copyright: "Xonovex",
    },
  },
  markdown: {
    html: false,
    attrs: {disable: true},
    config: (md) => {
      repositoryMarkdown(md, root, catalog, base);
    },
  },
  vite: {
    build: {rollupOptions: {output: {manualChunks: searchChunkName}}},
    plugins: [searchChunks()],
  },
});
