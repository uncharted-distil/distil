<template>
  <facet-terms
    :data.prop="facetData"
    action-buttons="0"
    :selection.prop="selection"
    :subselection.prop="subSelection"
    :disabled.prop="!enableHighlighting"
    @facet-element-updated="updateSelection"
  >
    <div slot="header-label" :class="headerClass">
      <i :class="getGroupIcon(summary) + ' facet-header-icon'" />
      <span>{{ summary.label.toUpperCase() }}</span>
      <type-change-menu
        v-if="facetEnableTypeChanges"
        class="facet-header-dropdown"
        :dataset="summary.dataset"
        :field="summary.key"
        :expand-collapse="expandCollapse"
        :expand="expand"
      />
    </div>
    <facet-template target="facet-terms-value">
      <div slot="content" class="w-100 pl-2">
        <div slot="bar" class="facet-sparkline-display">${metadata}</div>
      </div>
      <div slot="footer" class="w-100 pl-2 facet-sparkline-axis-footer">
        ${label}
      </div>
    </facet-template>
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
      />
    </div>
  </facet-terms>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted.software/facets-core";
import { FacetTermsData } from "@uncharted.software/facets-core/dist/types/facet-terms/FacetTerms";

import TypeChangeMenu from "../TypeChangeMenu.vue";
import SparklinePreview from "../SparklinePreview.vue";
import {
  Highlight,
  RowSelection,
  TimeseriesGrouping,
  VariableSummary,
} from "../../store/dataset";
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
import { IMAGE_TYPE } from "../../util/types";

export default Vue.extend({
  name: "facet-sparklines",

  components: {
    TypeChangeMenu,
    SparklinePreview,
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
    enabledTypeChanges: Array as () => string[],
    enableHighlighting: Boolean as () => boolean,
    expandCollapse: Function as () => Function,
    highlight: Object as () => Highlight,
    grouping: Object as () => TimeseriesGrouping,
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function,
    ],
    instanceName: String as () => string,
    rowSelection: Object as () => RowSelection,
    summary: Object as () => VariableSummary,
    expand: Boolean as () => Boolean,
  },

  data() {
    return {
      baseNumToDisplay: getCategoricalChunkSize(this.summary.type),
      moreNumToDisplay: 0,
    };
  },

  computed: {
    facetData(): FacetTermsData {
      let values = [];
      if (hasBaseline(this.summary) && this.grouping) {
        values = this.getFacetValues();
      }
      return {
        label: this.summary.label.toUpperCase(),
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
      if (!this.isHighlightedGroup(this.highlight, this.summary.key)) {
        return null;
      }
      const highlightValue = this.getHighlightValue(this.highlight);
      if (!highlightValue) {
        return null;
      }
      const highlightAsSelection = this.summary.baseline.buckets.reduce(
        (acc, val, ind) => {
          if (val.key === highlightValue) acc[ind] = true;
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
    getFacetValues(): {
      ratio: number;
      label: string;
      value: number;
      metadata: {};
    }[] {
      const summary = this.summary;
      const buckets = summary.baseline.buckets;
      // Use exemplars to fetch a timeseries that is representative of this bucket.  When clusters
      // are applied there are multiple timeseries associated with each pattern, and the examplar array
      // returns the ID of one example for each.
      const exemplars = summary.baseline.exemplars;
      const facetData = [];
      for (let i = 0; i < this.numToDisplay; ++i) {
        facetData.push({
          ratio: buckets[i].count / this.max,
          label: buckets[i].key,
          value: buckets[i].count,
          metadata: this.getSparkline(exemplars[i]),
        });
      }
      return facetData;
    },
    getSparkline(sparklineId: string) {
      const sp = new SparklinePreview({
        store: this.$store,
        router: this.$router,
        propsData: {
          facetView: true,
          timeseriesCol: this.grouping.idCol,
          timeseriesId: sparklineId,
          truthDataset: this.summary.dataset,
          xCol: this.grouping.xCol,
          yCol: this.grouping.yCol,
        },
      });
      sp.$mount();
      return sp.$el;
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
    getHighlightValue(highlight: Highlight): any {
      if (highlight && highlight.value) {
        return highlight.value;
      }
      return null;
    },
    isHighlightedInstance(highlight: Highlight): boolean {
      return highlight && highlight.context === this.instanceName;
    },
    isHighlightedGroup(highlight: Highlight, key: string): boolean {
      return this.isHighlightedInstance(highlight) && highlight.key === key;
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
.facet-sparkline-display {
  width: 100%;
  height: 40px;
}

.facet-sparkline-axis-footer {
  font-family: "IBM Plex Sans", sans-serif;
  font-size: 12px;
  line-height: 16px;
  letter-spacing: 0.02em;
  color: rgb(26, 27, 28);
  padding-left: 0.25rem;
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
