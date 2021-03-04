<template>
  <div class="search-input">
    <header>Filters</header>
    <main>
      <filter-badge
        v-for="(filter, index) in filters"
        :key="index"
        :filter="filter"
      />
      <filter-badge
        v-for="(highlight, index) in highlightsAsFilters"
        :key="index"
        :filter="highlight"
      />
    </main>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import FilterBadge from "../components/FilterBadge.vue";
import { Highlight } from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { Filter, INCLUDE_FILTER } from "../util/filters";
import { createFiltersFromHighlights } from "../util/highlights";

export default Vue.extend({
  name: "SearchInput",

  components: {
    FilterBadge,
  },

  computed: {
    filters(): Filter[] {
      return routeGetters
        .getDecodedFilters(this.$store)
        .filter((f) => f.type !== "row");
    },

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },

    highlightsAsFilters(): Filter[] {
      if (!this.highlights || this.highlights.length < 1) {
        return null;
      }
      return createFiltersFromHighlights(this.highlights, INCLUDE_FILTER);
    },
  },
});
</script>

<style scoped>
header {
  font-weight: bold;
}

main {
  background-color: var(--gray-300);
  border: 1px solid var(--gray-500);
  border-radius: 0.2rem;
  display: flex;
  flex-shrink: 0; /* To avoid it to collapse and have the badges overflow. */
  flex-wrap: wrap;
  min-height: 2.5rem;
  padding: 0.2rem;
}
</style>
