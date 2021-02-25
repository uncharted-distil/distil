<template>
  <facet-terms
    :data.prop="facetData"
    action-buttons="0"
    multiselect="false"
    :selection.prop="selection"
    :subselection.prop="subSelection"
    :disabled.prop="!enableHighlighting"
    @facet-element-updated="updateSelection"
  >
    <div slot="header-label" :class="headerClass">
      <i :class="getGroupIcon(summary) + ' facet-header-icon'"></i>
      <span>{{ summary.label.toUpperCase() }}</span>
      <importance-bars
        v-if="importance"
        :importance="importance"
      ></importance-bars>
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
        <div class="facet-footer-more-section">
          <div class="facet-footer-more-count">
            <span v-if="facetMoreCount > 0">{{ facetMoreCount }} more</span>
          </div>
          <div class="facet-footer-more-controls">
            <span v-if="hasLess" @click="viewLess"> show less</span>
            <span v-if="hasMore" @click="viewMore"> show more</span>
          </div>
        </div>
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

import "@uncharted.software/facets-core";
import { FacetTermsData } from "@uncharted.software/facets-core/dist/types/facet-terms/FacetTerms";

import TypeChangeMenu from "../TypeChangeMenu.vue";
import ImportanceBars from "../ImportanceBars.vue";
import { Highlight, RowSelection, VariableSummary } from "../../store/dataset";
import {
  getCategoricalChunkSize,
  getGroupIcon,
  getSubSelectionValues,
  hasBaseline,
  viewMoreData,
  viewLessData,
  facetTypeChangeState,
} from "../../util/facets";
import _ from "lodash";
import { getVariableImportance } from "../../util/data";

export default Vue.extend({
  name: "facet-categorical",

  components: {
    TypeChangeMenu,
    ImportanceBars,
  },

  directives: {
    child(el, binding): void {
      el.innerHTML = "";
      if (binding.value) {
        el.appendChild(binding.value);
      }
    },
  },

  props: {
    summary: Object as () => VariableSummary,
    enabledTypeChanges: Array as () => string[],
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function,
    ],
    expandCollapse: Function as () => Function,
    highlights: Array as () => Highlight[],
    enableHighlighting: Boolean as () => boolean,
    instanceName: String as () => string,
    rowSelection: Object as () => RowSelection,
    importance: Number as () => number,
  },

  data() {
    return {
      baseNumToDisplay: getCategoricalChunkSize(this.summary.type),
      moreNumToDisplay: 0,
    };
  },

  computed: {
    facetData(): FacetTermsData {
      const values = [];
      const summary = this.summary;
      if (this.hasBaseline) {
        const buckets = summary.baseline.buckets;
        for (let i = 0; i < this.numToDisplay; ++i) {
          values.push(this.getBucketData(buckets[i]));
        }
      }
      return {
        label: summary.label.toUpperCase(),
        values,
      };
    },
    facetEnableTypeChanges(): boolean {
      return facetTypeChangeState(
        this.summary.dataset,
        this.summary.key,
        this.enabledTypeChanges
      );
    },
    headerClass(): string {
      return this.facetEnableTypeChanges
        ? "facet-header-container"
        : "facet-header-container-no-scroll";
    },
    subSelection(): number[][] {
      return getSubSelectionValues(this.summary, this.rowSelection, this.max);
    },
    selection(): {} {
      if (!this.enableHighlighting || !this.isHighlightedGroup()) {
        return null;
      }

      const highlightValues = this.getHighlightValues();
      const highlightAsSelection = this.summary.baseline.buckets.reduce(
        (acc, val, ind) => {
          if (highlightValues.includes(val.key)) acc[ind] = true;
          return acc;
        },
        {}
      );
      return highlightAsSelection;
    },
    facetValueCount(): number {
      return this.hasBaseline ? this.summary.baseline.buckets.length : 0;
    },
    facetDisplayMore(): boolean {
      const chunkSize = getCategoricalChunkSize(this.summary.type);
      return this.facetValueCount > chunkSize;
    },
    facetMoreCount(): number {
      return this.facetValueCount - this.numToDisplay;
    },
    numToDisplay(): number {
      return this.hasBaseline && this.facetValueCount < this.baseNumToDisplay
        ? this.facetValueCount
        : this.baseNumToDisplay + this.moreNumToDisplay;
    },
    max(): number {
      return this.hasBaseline ? this.summary.baseline.extrema.max : 0;
    },
    hasBaseline(): boolean {
      return hasBaseline(this.summary);
    },
    hasMore(): boolean {
      return this.numToDisplay < this.facetValueCount;
    },
    hasLess(): boolean {
      return this.moreNumToDisplay > 0;
    },
  },

  methods: {
    getBucketData(bucket): { ratio: number; label: string; value: number } {
      return {
        ratio: bucket.count / this.max,
        label: bucket.key,
        value: bucket.count,
      };
    },
    viewMore() {
      this.moreNumToDisplay = viewMoreData(
        this.moreNumToDisplay,
        this.facetMoreCount,
        this.baseNumToDisplay,
        this.facetValueCount
      );
    },
    viewLess() {
      this.moreNumToDisplay = viewLessData(
        this.moreNumToDisplay,
        this.facetMoreCount,
        this.baseNumToDisplay,
        this.facetValueCount
      );
    },
    getHighlightValues(): string[] {
      return this.highlights.reduce(
        (acc, highlight) =>
          typeof highlight.value === "string" ? [...acc, highlight.value] : acc,
        []
      );
    },
    isHighlightedGroup(): boolean {
      return this.highlights.reduce(
        (acc, highlight) =>
          (highlight.key === this.summary.key &&
            highlight.context === this.instanceName) ||
          acc,
        false
      );
    },
    updateSelection(event) {
      if (!this.enableHighlighting) return;
      const facet = event.currentTarget;
      if (
        event.detail.changedProperties.get("selection") !== undefined &&
        !_.isEqual(facet.selection, this.selection)
      ) {
        let value = null;
        if (facet.selection) {
          if (this.selection) {
            const oldKey = Object.keys(this.selection)[0];
            const incomingKeys = Object.keys(facet.selection);
            const newKey = incomingKeys.filter(
              (iKey) => oldKey.indexOf(iKey) < 0
            )[0];
            value = this.facetData.values[newKey].label;
          } else {
            const newKey = Object.keys(facet.selection)[0];
            value = this.facetData.values[newKey].label;
          }
        }
        this.$emit(
          "facet-click",
          this.instanceName,
          this.summary.key,
          value,
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
              key: this.summary.key,
              type: "categorical",
            })
          : this.html;
      }
      return null;
    },
    getGroupIcon,
  },
});
</script>

<style scoped>
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

.facet-footer-more-section {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  justify-content: flex-start;
  align-content: stretch;
  align-items: flex-start;
}

.facet-footer-more-count {
  order: 0;
  flex: 1 1 auto;
  align-self: auto;
}

.facet-footer-more-controls {
  order: 0;
  flex: 0 1 auto;
  align-self: auto;
}

.facet-footer-more-controls > span {
  cursor: pointer;
}

.facet-footer-custom-html {
  margin-top: 6px;
}

.facet-header-container {
  color: rgba(0, 0, 0, 0.54);
  display: flex;
  align-items: center;
  overflow-y: scroll !important;
}

.facet-header-container-no-scroll {
  color: rgba(0, 0, 0, 0.54);
  overflow: auto;
}

.facet-header-container .dropdown-menu {
  max-height: 200px;
  overflow-y: auto;
}
</style>
