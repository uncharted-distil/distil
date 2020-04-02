<template>
  <facet-terms :data.prop="facetData" action-buttons="0">
    <div slot="header-label" class="facet-header-container">
      <i :class="getGroupIcon(summary) + ' facet-header-icon'"></i>
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
      <div v-if="facetDisplayMore" class="facet-footer-more">
        <span v-if="facetMoreCount > 0">{{ facetMoreCount }} more</span>
      </div>
      <div
        v-if="this.html"
        v-child="computeCustomHTML()"
        class="facet-footer-custom-html"
      ></div>
    </div>
  </facet-terms>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted/facets-core";
import { FacetTermsData } from "@uncharted/facets-core/dist/types/facet-terms/FacetTerms";

import TypeChangeMenu from "../TypeChangeMenu";
import { VariableSummary } from "../../store/dataset";
import { getCategoricalChunkSize, getGroupIcon } from "../../util/facets";
import _ from "lodash";

export default Vue.extend({
  name: "facet-categorical",

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

  data() {
    return {
      numToDisplay: getCategoricalChunkSize(this.summary.type)
    };
  },

  computed: {
    facetData(): FacetTermsData {
      const summary = this.summary;
      const numToDisplay = this.numToDisplay;
      const values = [];
      if (summary.baseline.buckets.length) {
        const buckets = summary.baseline.buckets;
        const maxCount = summary.baseline.extrema.max;
        for (
          let i = 0, n = Math.min(buckets.length, numToDisplay);
          i < n;
          ++i
        ) {
          values.push({
            ratio: buckets[i].count / maxCount,
            label: buckets[i].key,
            value: buckets[i].count
          });
        }
      }
      return {
        label: summary.label.toUpperCase(),
        values
      };
    },

    facetValueCount(): number {
      return this.summary.baseline.buckets.length;
    },

    facetEnableTypeChanges(): boolean {
      const key = `${this.summary.dataset}:${this.summary.key}`;
      return Boolean(this.enabledTypeChanges.find(e => e === key));
    },

    facetDisplayMore(): boolean {
      const chunkSize = getCategoricalChunkSize(this.summary.type);
      return this.facetValueCount > chunkSize;
    },

    facetMoreCount(): number {
      return this.facetValueCount - this.numToDisplay;
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
    },

    getGroupIcon
  }
});
</script>

<style scoped>
.facet-header-container {
  display: flex;
  align-items: center;
}

.facet-header-icon {
  margin-right: 6px;
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

.facet-footer-more {
  margin-bottom: 4px;
}

.facet-footer-custom-html {
  margin-top: 6px;
}
</style>
