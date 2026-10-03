<script setup lang="ts">
import {computed} from "vue";
import {useData, withBase} from "vitepress";

const props = defineProps<{light: string; dark: string; alt: string}>();
const {isDark} = useData();
const source = computed(() =>
  withBase(isDark.value ? props.dark : props.light),
);
</script>

<template>
  <a :href="source" class="theme-diagram" :aria-label="'Open diagram: ' + alt">
    <img :src="source" :alt="alt" loading="lazy" />
  </a>
</template>
