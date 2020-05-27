<template>
  <facet-bars
    :data.prop="facetData"
    :selection.prop="selection"
    :subselection.prop="subSelection"
    @facet-element-updated="updateSelection"
  >
    <div slot="header-label" :class="headerClass">
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
      <facet-plugin-scrollbar
        min-bar-width="8"
        auto-hide="true"
        round-caps="true"
      ></facet-plugin-scrollbar>
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
import "@uncharted/facets-plugins";
import { FacetBarsData } from "@uncharted/facets-core/dist/types/facet-bars/FacetBars";

import TypeChangeMenu from "../TypeChangeMenu";
import { Highlight, VariableSummary } from "../../store/dataset";
import { getSubSelectionValues } from "../../util/facets";
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
    expandCollapse: Function as () => Function,
    highlight: Object as () => Highlight,
    enableHighlighting: Boolean as () => boolean,
    instanceName: String as () => string
  },

  computed: {
    max(): number {
      if (this.summary.baseline.buckets.length) {
        const buckets = this.summary.baseline.buckets;
        // seems to be incorrect compute based on the current buckets
        // const maxCount = summary.baseline.extrema.max;
        return buckets.reduce((max, bucket) => Math.max(max, bucket.count), 0);
      }
      return 0;
    },
    facetData(): FacetBarsData {
      const summary = this.summary;
      const values = [];
      if (summary.baseline.buckets.length) {
        const buckets = summary.baseline.buckets;
        for (let i = 0, n = buckets.length; i < n; ++i) {
          values.push({
            ratio: buckets[i].count / this.max,
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
    },
    headerClass(): string {
      return this.facetEnableTypeChanges
        ? "facet-header-container"
        : "facet-header-container-no-scroll";
    },
    subSelection(): number[] {
      return getSubSelectionValues(this.summary, this.max);
    },
    selection(): number[] {
      if (!this.isHighlightedGroup(this.highlight, this.summary.key)) {
        return null;
      }
      const highlightValue = this.getHighlightValue(this.highlight);
      if (!highlightValue) {
        return null;
      }
      const highlightAsSelection = this.summary.baseline.buckets.reduce(
        (acc, val, ind) => {
          const key = _.toNumber(val.key);
          if (key === highlightValue.from || key === highlightValue.to) {
            acc.push(ind);
          }
          return acc;
        },
        []
      );
      return highlightAsSelection.length > 0 ? highlightAsSelection : null;
    }
  },

  methods: {
    getHighlightValue(highlight: Highlight): any {
      if (highlight && highlight.value) {
        return highlight.value;
      }
      return null;
    },
    isHighlightedInstance(highlight: Highlight): boolean {
      return highlight && highlight.context === this.instanceName;
    },
    isHighlightedGroup(highlight: Highlight, colName: string): boolean {
      return this.isHighlightedInstance(highlight) && highlight.key === colName;
    },
    updateSelection(event) {
      if (!this.enableHighlighting) return;
      const facet = event.currentTarget;
      if (
        event.detail.changedProperties.get("selection") !== undefined &&
        !_.isEqual(facet.selection, this.selection)
      ) {
        const range =
          facet.selection && facet.selection.length > 1
            ? {
                from: _.toNumber(
                  this.facetData.values[facet.selection[0]].label
                ),
                to: _.toNumber(this.facetData.values[facet.selection[1]].label),
                type: this.summary.type
              }
            : null;
        this.$emit(
          "range-change",
          this.instanceName,
          this.summary.key,
          range,
          this.summary.dataset
        );
      }
    },
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

.facet-header-container {
  overflow-y: scroll !important;
}

.facet-header-container-no-scroll {
  overflow: auto;
}

.facet-header-container .dropdown-menu {
  max-height: 200px;
  overflow-y: auto;
}
</style>
