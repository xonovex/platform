<script setup lang="ts">
import {computed} from "vue";
import DefaultTheme from "vitepress/theme";
import {useData, withBase} from "vitepress";
import data from "../../content/catalog.json";
import {componentTypes} from "../../src/model.js";
import ComponentCatalog from "./ComponentCatalog.vue";

const {frontmatter} = useData();
const category = computed(() =>
  componentTypes.find((type) => type === frontmatter.value.catalog),
);
const component = computed(() =>
  data.entries.find((entry) => entry.id === frontmatter.value.component),
);
const source = computed(() =>
  typeof frontmatter.value.sourcePath === "string"
    ? data.repository + "/blob/main/" + frontmatter.value.sourcePath
    : undefined,
);
const featured = data.entries
  .filter((entry) => entry.type === "plugins")
  .slice(0, 6);
</script>

<template>
  <DefaultTheme.Layout>
    <template #doc-before>
      <div v-if="component" class="component-meta">
        <span>{{
          component.type === "mcp-servers"
            ? "MCP server"
            : component.type.slice(0, -1)
        }}</span>
        <span v-if="component.version">v{{ component.version }}</span>
        <a
          v-if="component.plugin && component.type !== 'plugins'"
          :href="withBase('/plugins/' + component.plugin + '.html')"
          >{{ component.plugin }}</a
        >
        <span v-for="harness in component.harnesses" :key="harness">{{
          harness
        }}</span>
      </div>
    </template>
    <template #doc-after>
      <ComponentCatalog v-if="category" :key="category" :type="category" />
      <a
        v-if="source"
        :href="source"
        class="document-source"
        target="_blank"
        rel="noreferrer"
        >View the source document on GitHub</a
      >
    </template>
    <template #home-features-after>
      <section class="home-catalog">
        <div class="home-catalog-heading">
          <div>
            <span class="catalog-eyebrow">Reusable guidance</span>
            <h2>Choose a plugin for the task.</h2>
          </div>
          <a :href="withBase('/catalog/plugins.html')">View all plugins</a>
        </div>
        <div class="catalog-grid">
          <article
            v-for="entry in featured"
            :key="entry.id"
            class="catalog-card">
            <span class="catalog-kind">Plugin</span>
            <h3>
              <a :href="withBase(entry.link + '.html')">{{
                entry.name.replace(/^xonovex-/, "")
              }}</a>
            </h3>
            <p>{{ entry.description }}</p>
            <span class="catalog-version">v{{ entry.version }}</span>
          </article>
        </div>
      </section>
    </template>
  </DefaultTheme.Layout>
</template>
