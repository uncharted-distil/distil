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
      :data.prop="imgData"
      action-buttons="0"
      :selection.prop="selection"
      :subselection.prop="subSelection"
      :disabled.prop="!enableHighlighting"
      class="facet-image"
      @facet-element-updated="updateSelection"
    >
      <div slot="header-label" :class="headerClass">
        <div class="d-flex align-items-center justify-content-between">
          <div>
            <i :class="getGroupIcon(summary) + ' facet-header-icon'" />
            {{ summary.label.toUpperCase() }}
          </div>
          <button-training-target
            :variable="summary.key"
            :dataset-name="datasetName"
            :active-variables="activeVariables"
          />
        </div>
        <div class="d-flex align-items-center my-1">
          <button-explore :variable="summary.key" />
          <color-scale-drop-down
            v-if="geoEnabled && isClustering"
            :variable-summary="summary"
            is-facet-scale
            class="mr-1"
          />
          <type-change-menu
            v-if="facetEnableTypeChanges"
            :dataset="summary.dataset"
            :field="summary.key"
            :expand-collapse="expandCollapse"
            :type-change-event="typeChangeEvent"
            @type-change="onTypeChange"
          />
        </div>
      </div>
      <facet-template
        target="facet-terms-value"
        class="facet-content-container"
      >
        <div slot="header" class="facet-image-preview-display">${metadata}</div>
        <div slot="label" class="facet-image-label" title="${label} ${value}">
          ${label} ${value}
        </div>
        <div slot="annotation" class="collapse-unused" />
        <div slot="value" class="collapse-unused" />
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
      </div>
    </facet-terms>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";

import "@uncharted.software/facets-core";
import { FacetTermsData } from "@uncharted.software/facets-core/dist/types/facet-terms/FacetTerms";
import { getters as routeGetters } from "../../store/route/module";
import TypeChangeMenu from "../TypeChangeMenu.vue";
import ImagePreview from "../ImagePreview.vue";
import {
  DataMode,
  Highlight,
  RowSelection,
  VariableSummary,
  Variable,
} from "../../store/dataset";
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
import ButtonExplore from "../ButtonExplore.vue";
import ButtonTrainingTarget from "../ButtonTrainingTarget.vue";
import ColorScaleDropDown from "../ColorScaleDropDown.vue";
import { EventList } from "../../util/events";
import { ColorScaleNames } from "../../util/color";
export default Vue.extend({
  name: "FacetImage",

  components: {
    TypeChangeMenu,
    ImagePreview,
    ButtonExplore,
    ButtonTrainingTarget,
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
    activeVariables: {
      type: Array as () => Variable[],
      default: () => [] as Variable[],
    },
    datasetName: { type: String as () => string, default: null },
    enabledTypeChanges: Array as () => string[],
    enableHighlighting: Boolean as () => boolean,
    expandCollapse: Function as () => Function,
    geoEnabled: { type: Boolean as () => boolean, default: false },
    highlights: Array as () => Highlight[],
    include: { type: Boolean as () => boolean, default: true },
    instanceName: String as () => string,
    rowSelection: Object as () => RowSelection,
    summary: Object as () => VariableSummary,
    typeChangeEvent: { type: String as () => string, default: "" },
  },

  data() {
    return {
      baseNumToDisplay: getCategoricalChunkSize(this.summary.type),
      moreNumToDisplay: 0,
      imgData: {} as FacetTermsData,
    };
  },

  computed: {
    comp(): string {
      return "style";
    },
    // facetData(): FacetTermsData {
    //   let values = [];
    //   if (hasBaseline(this.summary)) {
    //     values = this.getFacetValues();
    //   }
    //   return {
    //     label: this.summary.label.toUpperCase(),
    //     values,
    //   };
    // },
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
      return getSubSelectionValues(
        this.summary,
        this.rowSelection,
        this.max,
        this.include
      );
    },
    hasColorScale(): boolean {
      return (
        routeGetters.getColorScaleVariable(this.$store) === this.summary.key
      );
    },
    id(): string {
      return "_" + this.summary.key.replace(/([\/,:!?_])/g, "");
    },
    colorScale(): ColorScaleNames {
      return routeGetters.getColorScale(this.$store);
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
    isClustering(): boolean {
      return routeGetters.getDataMode(this.$store) === DataMode.Cluster;
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
      return this.hasExamplars &&
        this.summary.baseline.exemplars.length < this.baseNumToDisplay
        ? this.summary.baseline.exemplars.length
        : this.hasBaseline && this.facetValueCount < this.baseNumToDisplay
        ? this.facetValueCount
        : this.baseNumToDisplay + this.moreNumToDisplay;
    },
    max(): number {
      return this.hasBaseline ? this.summary.baseline.extrema.max : 0;
    },
    hasExamplars(): boolean {
      return !!this.summary.baseline.exemplars;
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
  watch: {
    summary(cur) {
      if (!this.imgData?.values) {
        const values = hasBaseline(cur) ? this.getFacetValues() : [];
        this.imgData = { label: cur.label.toUpperCase(), values };
      }
    },
  },
  methods: {
    onTypeChange() {
      this.$emit(EventList.VARIABLES.TYPE_CHANGE);
    },
    getFacetValues(): {
      ratio: number;
      label: string;
      value: number;
      metadata: {};
    }[] {
      const summary = this.summary;
      const buckets = summary.baseline.buckets;
      const facetData = [];
      for (let i = 0; i < this.numToDisplay; ++i) {
        const imageUrl = this.hasExamplars
          ? summary.baseline.exemplars[i]
          : buckets[i].key;
        facetData.push({
          ratio: buckets[i].count / this.max,
          label: buckets[i].key,
          value: buckets[i].count,
          metadata: this.getImagePreview(imageUrl),
        });
      }
      return facetData;
    },
    getImagePreview(imageUrl: string) {
      const ip = new ImagePreview({
        store: this.$store,
        router: this.$router,
        propsData: {
          // NOTE: there seems to be an issue with the visibility plugin used
          // when injecting this way. Cancel the visibility flagging for facets.
          preventHiding: true,
          imageUrl,
          type: this.summary.varType,
          datasetName: this.datasetName,
        },
      });
      ip.$mount();
      return ip.$el;
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
            values.push(this.imgData.values[ik].label)
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
    getGroupIcon,
  },
});
</script>

<style>
::part(facet-container-header) {
  height: auto;
}

.facet-image .facet-terms-container {
  max-height: 200px !important;
  overflow-y: auto;
  display: flex;
  flex-wrap: wrap;
}
</style>
<style scoped>
.collapse-unused {
  display: none;
}
.facet-image-label {
  max-width: 75px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.facet-content-container {
  display: inline-block;
  width: 85px;
  overflow: hidden;
}
.facet-image-preview-display {
  padding-left: 10px;
}

.facet-header-icon {
  margin-right: 6px;
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
