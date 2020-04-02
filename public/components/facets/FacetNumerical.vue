<template>
  <facet-bars :data.prop="facetData">
    <div slot="header-label" class="facet-header-container">
      <span>{{ summary.label.toUpperCase() }}</span>
      <type-change-menu
        v-if="facetEnableTypeChanges"
        class="facet-header-dropdown"
        :dataset="summary.dataset"
        :field="summary.key"
        :expandCollapse="expandCollapse"
      >
      </type-change-menu>
    </div>

    <div slot="footer" class="facet-footer-container">
      <div
        v-if="this.html"
        v-child="computeCustomHTML()"
        class="facet-footer-custom-html"
      ></div>
    </div>
  </facet-bars>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted/facets-core";
import { FacetBarsData } from "@uncharted/facets-core/dist/types/facet-bars/FacetBars";

import TypeChangeMenu from "../TypeChangeMenu";
import { VariableSummary } from "../../store/dataset";
import _ from "lodash";

export default Vue.extend({
  name: "facet-numerical",

  components: {
    TypeChangeMenu
  },

  directives: {
    child(el, binding): void {
      el.innerHTML = "";
      if (binding.value) {
        el.appendChild(binding.value);
      }
    }
  },

  props: {
    summary: Object as () => VariableSummary,
    enabledTypeChanges: Array as () => string[],
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function
    ],
    expandCollapse: Function as () => Function
  },

  computed: {
    facetData(): FacetBarsData {
      const summary = this.summary;
      const values = [];
      if (summary.baseline.buckets.length) {
        const buckets = summary.baseline.buckets;
        // seems to be incorrect compute based on the current buckets
        // const maxCount = summary.baseline.extrema.max;
        const maxCount = buckets.reduce(
          (max, bucket) => Math.max(max, bucket.count),
          0
        );
        for (let i = 0, n = buckets.length; i < n; ++i) {
          values.push({
            ratio: buckets[i].count / maxCount,
            label: buckets[i].key
          });
        }
      }
      return {
        label: summary.label.toUpperCase(),
        values
      };
    },

    facetEnableTypeChanges(): boolean {
      const key = `${this.summary.dataset}:${this.summary.key}`;
      return Boolean(this.enabledTypeChanges.find(e => e === key));
    }
  },

  methods: {
    computeCustomHTML(): HTMLElement | null {
      // hack to get the custom html buttons showing up
      // changing this would mean to change how the instantiation of the facets works
      // right now they are wrapped by other components like
      // available-target-variables, available-training-variables, etc
      // those components inject HTML into the facets through their `html` function
      // we might want to change that in the future though
      if (this.html) {
        return _.isFunction(this.html)
          ? this.html({
              colName: this.summary.key
            })
          : this.html;
      }
      return null;
    }
  }
});
</script>

<style scoped>
.facet-header-container {
  display: flex;
  align-items: center;
}

.facet-header-dropdown {
  position: absolute;
  right: 12px;
}

.facet-footer-container {
  min-height: 12px;
  padding: 6px 12px 5px;
  font-family: "IBM Plex Sans", sans-serif;
  font-size: 12px;
  font-weight: 600;
  line-height: 16px;
}

.facet-footer-custom-html {
  margin-top: 6px;
}
</style>
