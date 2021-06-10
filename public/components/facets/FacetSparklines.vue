<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

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
        v-if="displayFooter"
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
  Variable,
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
import { DISTIL_ROLES } from "../../util/types";
import { EventList } from "../../util/events";
export default Vue.extend({
  name: "FacetSparklines",

  components: {
    TypeChangeMenu,
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
    highlights: Array as () => Highlight[],
    variable: Object as () => Variable,
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function,
    ],
    instanceName: String as () => string,
    rowSelection: Object as () => RowSelection,
    summary: Object as () => VariableSummary,
    expand: Boolean as () => boolean,
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
      if (hasBaseline(this.summary)) {
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
    displayFooter(): boolean {
      return !!this.html && this.summary.distilRole != DISTIL_ROLES.Augmented;
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
      if (!this.variable) {
        return;
      }
      const grouping = this.variable.grouping as TimeseriesGrouping;
      const sp = new SparklinePreview({
        store: this.$store,
        router: this.$router,
        propsData: {
          facetView: true,
          variableKey: this.variable.key,
          timeseriesId: sparklineId,
          truthDataset: this.summary.dataset,
          xCol: grouping.xCol,
          yCol: grouping.yCol,
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
    getHighlightValues(): string[] {
      return this.highlights.reduce(
        (acc, highlight) =>
          highlight.key === this.summary.key &&
          typeof highlight.value === "string"
            ? [...acc, highlight.value]
            : acc,
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
        const values = [];
        if (facet.selection) {
          const incomingKeys = Object.keys(facet.selection);
          incomingKeys.forEach((ik) =>
            values.push(this.facetData.values[ik].label)
          );
        }
        this.$emit(
          EventList.FACETS.CLICK_EVENT,
          this.instanceName,
          this.summary.key,
          values,
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
