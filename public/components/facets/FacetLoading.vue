<template>
  <facet-bars :data.prop="facetData">
    <div slot="footer" class="facet-footer"></div>
  </facet-bars>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted.software/facets-core";
import { FacetBarsData } from "@uncharted.software/facets-core/dist/types/facet-bars/FacetBars";

import { VariableSummary } from "../../store/dataset";

export default Vue.extend({
  name: "facet-loading",

  props: {
    summary: Object as () => VariableSummary,
  },

  computed: {
    facetData(): FacetBarsData | null {
      const summary = this.summary;
      return {
        label: summary.label.toUpperCase(),
        values: [
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
          null,
        ],
      };
    },
  },
});
</script>

<style scoped>
.facet-footer {
  height: 25px;
}
</style>
