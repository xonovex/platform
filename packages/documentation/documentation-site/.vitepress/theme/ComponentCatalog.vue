<script setup lang="ts">
import {computed, ref} from "vue";
import {withBase} from "vitepress";
import catalog from "../../content/catalog.json";
import {filterComponents} from "../../src/filter.js";
import type {ComponentType} from "../../src/model.js";

const props = defineProps<{type: ComponentType}>();
const query = ref("");
const plugin = ref("");
const harness = ref("");
const candidates = computed(() =>
  catalog.entries.filter((entry) => entry.type === props.type),
);
const plugins = computed(() =>
  [
    ...new Set(candidates.value.map((entry) => entry.plugin).filter(Boolean)),
  ].toSorted(),
);
const harnesses = computed(() =>
  [...new Set(candidates.value.flatMap((entry) => entry.harnesses))].toSorted(),
);
const matches = computed(() =>
  filterComponents(catalog.entries, {
    type: props.type,
    query: query.value,
    plugin: plugin.value,
    harness: harness.value,
  }),
);
const reset = (): void => {
  query.value = "";
  plugin.value = "";
  harness.value = "";
};
</script>

<template>
  <section class="component-catalog" aria-label="Component catalog">
    <div class="catalog-filters">
      <label class="catalog-search">
        <span>Search this catalog</span>
        <input
          v-model="query"
          type="search"
          placeholder="Search names and descriptions"
          autocomplete="off" />
      </label>
      <label v-if="plugins.length > 1 && type !== 'plugins'">
        <span>Plugin</span>
        <select v-model="plugin">
          <option value="">All plugins</option>
          <option v-for="name in plugins" :key="name" :value="name">
            {{ name }}
          </option>
        </select>
      </label>
      <label v-if="harnesses.length">
        <span>Harness</span>
        <select v-model="harness">
          <option value="">All harnesses</option>
          <option v-for="name in harnesses" :key="name" :value="name">
            {{ name }}
          </option>
        </select>
      </label>
    </div>
    <div class="catalog-results">
      <p role="status" aria-live="polite" aria-atomic="true">
        {{ matches.length }} matching
        {{ matches.length === 1 ? "component" : "components" }}
      </p>
      <button v-if="query || plugin || harness" type="button" @click="reset">
        Clear filters
      </button>
    </div>
    <div v-if="matches.length" class="catalog-grid">
      <article v-for="entry in matches" :key="entry.id" class="catalog-card">
        <div class="catalog-card-heading">
          <span class="catalog-kind">{{
            entry.type === "mcp-servers"
              ? "MCP server"
              : entry.type.slice(0, -1)
          }}</span
          ><span v-if="entry.version" class="catalog-version"
            >v{{ entry.version }}</span
          >
        </div>
        <h2>
          <a :href="withBase(entry.link + '.html')">{{ entry.name }}</a>
        </h2>
        <p>{{ entry.description }}</p>
        <a
          v-if="entry.plugin && type !== 'plugins'"
          class="catalog-owner"
          :href="withBase('/plugins/' + entry.plugin + '.html')"
          >{{ entry.plugin }}</a
        >
        <ul
          v-if="entry.harnesses.length"
          class="catalog-harnesses"
          aria-label="Supported harnesses">
          <li v-for="name in entry.harnesses" :key="name">
            {{ name }}
          </li>
        </ul>
      </article>
    </div>
    <div v-else class="catalog-empty">
      <h2>
        {{
          candidates.length
            ? "No components match these filters."
            : "No components are declared yet."
        }}
      </h2>
      <p v-if="candidates.length">
        Try another search or clear the filters to see the full catalog.
      </p>
      <p v-else>
        Components appear here when a plugin declares them in the repository.
      </p>
      <button v-if="candidates.length" type="button" @click="reset">
        Show all components
      </button>
    </div>
  </section>
</template>
