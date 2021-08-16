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
  <div>
    <component :is="comp" v-html="cssStyle" />
    <facet-terms
      :id="id"
      :data.prop="facetData"
      action-buttons="0"
      :selection.prop="selection"
      :subselection.prop="subSelection"
      :disabled.prop="!enableHighlighting"
      @facet-element-updated="updateSelection"
    >
      <div slot="header-label" :class="headerClass" class="d-flex">
        <i :class="getGroupIcon(summary) + ' facet-header-icon'" />
        <span>{{ summary.label.toUpperCase() }}</span>
        <importance-bars v-if="enableImportance" :importance="importance" />
        <div class="facet-header-dropdown d-flex align-items-center">
          <color-scale-drop-down
            v-if="geoEnabled"
            :is-toggle="colorScaleToggle"
            :variable-summary="summary"
            is-facet-scale
            class="mr-1"
          />
          <type-change-menu
            v-if="facetEnableTypeChanges"
            :dataset="summary.dataset"
            :field="summary.key"
            :expand-collapse="expandCollapse"
            @type-change="onTypeChange"
          />
        </div>
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
          v-if="displayFooter"
          v-child="computeCustomHTML()"
          class="facet-footer-custom-html"
        />
      </div>
    </facet-terms>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";

import "@uncharted.software/facets-core";
import { FacetTermsData } from "@uncharted.software/facets-core/dist/types/facet-terms/FacetTerms";

import TypeChangeMenu from "../TypeChangeMenu.vue";
import ImportanceBars from "../ImportanceBars.vue";
import ColorScaleDropDown from "../ColorScaleDropDown.vue";
import { Highlight, RowSelection, VariableSummary } from "../../store/dataset";
import {
  getCategoricalChunkSize,
  getGroupIcon,
  getSubSelectionValues,
  hasBaseline,
  viewMoreData,
  viewLessData,
  facetTypeChangeState,
  generateFacetDiscreteStyle,
} from "../../util/facets";
import { DISTIL_ROLES } from "../../util/types";
import { getters as routeGetters } from "../../store/route/module";
import { ColorScaleNames } from "../../util/color";
import { EventList } from "../../util/events";
export default Vue.extend({
  name: "FacetCategorical",

  components: {
    TypeChangeMenu,
    ImportanceBars,
    ColorScaleDropDown,
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
    enableImportance: { type: Boolean as () => boolean, default: true },
    geoEnabled: { type: Boolean as () => boolean, default: false },
    colorScaleToggle: { type: Boolean as () => boolean, default: false },
    include: { type: Boolean as () => boolean, default: true },
  },

  data() {
    return {
      baseNumToDisplay: getCategoricalChunkSize(this.summary.type),
      moreNumToDisplay: 0,
    };
  },

  computed: {
    comp(): string {
      return "style";
    },
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
    hasColorScale(): boolean {
      return (
        routeGetters.getColorScaleVariable(this.$store) === this.summary.key
      );
    },
    colorScale(): ColorScaleNames {
      return routeGetters.getColorScale(this.$store);
    },
    id(): string {
      return "_" + this.summary.key.replace(/([\/,:!?_])/g, "");
    },
    cssStyle(): string {
      return this.hasColorScale
        ? generateFacetDiscreteStyle(
            this.id,
            "facet-terms-value-bar-0",
            this.summary,
            this.colorScale
          )
        : "";
    },
    subSelection(): number[][] {
      return getSubSelectionValues(
        this.summary,
        this.rowSelection,
        this.max,
        this.include
      );
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
      return !!this.html;
    },
  },

  methods: {
    onTypeChange() {
      this.$emit(EventList.VARIABLES.TYPE_CHANGE);
    },
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
          highlight.key === this.summary.key
            ? [...acc, ...highlight.value]
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
  padding: 6px 4px 5px 5px;
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
