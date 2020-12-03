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
        v-if="highlightAsAFilter"
        is-highlight
        :filter="highlightAsAFilter"
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
import { createFilterFromHighlight } from "../util/highlights";

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

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    highlightAsAFilter(): Filter {
      if (!this.highlight || !this.highlight.value) return;
      return createFilterFromHighlight(this.highlight, INCLUDE_FILTER);
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
