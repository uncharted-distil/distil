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
    <facet-bars
      :id="id"
      :data.prop="facetData"
      :selection.prop="selection"
      :subselection.prop="subSelection"
      :disabled.prop="!enableHighlighting"
      @facet-element-updated="updateSelection"
    >
      <div slot="header-label" :class="headerClass">
        <span>{{ summary.label.toUpperCase() }}</span>
        <importance-bars v-if="enableImportance" :importance="importance" />
        <div class="d-flex align-items-center my-1">
          <toggle-explore :variable="summary.key" />
          <color-scale-drop-down
            v-if="geoEnabled"
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
        v-if="facetData.values.length > 0"
        target="facet-bars-value"
        title="${tooltip}"
      />

      <div v-else slot="content" />

      <div
        v-if="facetData.values.length > 0"
        slot="footer"
        class="facet-footer-container"
      >
        <facet-plugin-zoom-bar
          min-bar-width="8"
          auto-hide="true"
          round-caps="true"
        />
        <div
          v-if="displayFooter"
          v-child="computeCustomHTML()"
          class="facet-footer-custom-html d-flex justify-content-between"
        />
      </div>
      <div v-else slot="footer" class="facet-footer-container">
        No Data Available
      </div>
    </facet-bars>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted.software/facets-core";
import "@uncharted.software/facets-plugins";
import { FacetBarsData } from "@uncharted.software/facets-core/dist/types/facet-bars/FacetBars";
import ToggleExplore from "../ToggleExplore.vue";
import ColorScaleDropDown from "../ColorScaleDropDown.vue";
import TypeChangeMenu from "../TypeChangeMenu.vue";
import ImportanceBars from "../ImportanceBars.vue";
import { Highlight, RowSelection, VariableSummary } from "../../store/dataset";
import {
  getSubSelectionValues,
  hasBaseline,
  facetTypeChangeState,
  generateFacetLinearStyle,
} from "../../util/facets";
import { getters as routeGetters } from "../../store/route/module";
import _ from "lodash";
import { DISTIL_ROLES } from "../../util/types";
import { ColorScaleNames } from "../../util/color";
import { EventList } from "../../util/events";

export default Vue.extend({
  name: "FacetNumerical",

  components: {
    TypeChangeMenu,
    ImportanceBars,
    ToggleExplore,
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
    highlights: {
      type: Array as () => Highlight[],
      default: () => [] as Highlight[],
    },
    enableHighlighting: Boolean as () => boolean,
    instanceName: String as () => string,
    rowSelection: Object as () => RowSelection,
    importance: Number as () => number,
    enableImportance: { type: Boolean as () => boolean, default: true },
    geoEnabled: { type: Boolean as () => boolean, default: false },
    include: { type: Boolean as () => boolean, default: true },
    typeChangeEvent: { type: String as () => string, default: "" },
  },

  computed: {
    id(): string {
      return "_" + this.summary.key.replace(/([\/,:!?_])/g, "");
    },
    comp(): string {
      return "style";
    },
    hasColorScale(): boolean {
      return (
        routeGetters.getColorScaleVariable(this.$store) === this.summary.key
      );
    },
    colorScale(): ColorScaleNames {
      return routeGetters.getColorScale(this.$store);
    },
    cssStyle(): string {
      const highlight = this.highlights.find((h) => h.key === this.summary.key);
      return this.hasColorScale
        ? generateFacetLinearStyle(
            this.id,
            "facet-bars-value-bar-0",
            this.summary,
            this.colorScale,
            highlight
          )
        : "";
    },
    maxBucketCount(): number {
      if (hasBaseline(this.summary)) {
        const buckets = this.summary.baseline.buckets;
        return buckets.reduce((max, bucket) => Math.max(max, bucket.count), 0);
      }
      return 0;
    },
    facetData(): FacetBarsData {
      const summary = this.summary;
      const values = [];
      if (hasBaseline(summary)) {
        const buckets = summary.baseline.buckets;
        const bucketSize = buckets[1]?.key
          ? parseFloat(buckets[1].key) - parseFloat(buckets[0].key)
          : 0;
        for (let i = 0, n = buckets.length; i < n; ++i) {
          const count = buckets[i].count;
          const key = parseFloat(buckets[i].key);
          const visualWeight = 0.045;
          const ratio = !count
            ? count / this.maxBucketCount
            : Math.min(count / this.maxBucketCount + visualWeight, 1.0);

          values.push({
            ratio: ratio,
            label: key,
            tooltip: `Range:\t\t${key}-${
              key + bucketSize
            }\nTotal Count:\t${count}`,
          });
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
      return getSubSelectionValues(
        this.summary,
        this.rowSelection,
        this.maxBucketCount,
        this.include
      );
    },
    selection(): number[] {
      if (!this.enableHighlighting || !this.isHighlightedGroup()) {
        return null;
      }
      const highlightValues = this.getHighlightValues();
      const buckets = this.summary.baseline.buckets;

      // map the values used for the latest highlight filter back to the buckets
      // so we can show the latest selection made.
      const highlightAsSelection = buckets.reduce((acc, val, ind) => {
        const key = _.toNumber(val.key);
        if (
          highlightValues[0]?.from === key ||
          highlightValues[0]?.to === key
        ) {
          acc.push(ind);
        }
        return acc;
      }, []);

      // we can over shoot the bucket mapping if it's set to top end of
      // the range as the buckets are keyed to minimum bucket value
      // so in the event we've only mapped back one index, we set the
      // second index to the bucket length.
      if (highlightAsSelection.length === 1) {
        highlightAsSelection.push(buckets.length);
      }
      return highlightAsSelection.length > 0 ? highlightAsSelection : null;
    },
    displayFooter(): boolean {
      return !!this.html && this.summary.distilRole != DISTIL_ROLES.Augmented;
    },
  },

  methods: {
    onTypeChange() {
      this.$emit(EventList.VARIABLES.TYPE_CHANGE);
    },
    getHighlightValues(): { from: number; to: number }[] {
      return this.highlights.reduce(
        (acc, highlight) =>
          highlight.key === this.summary.key &&
          highlight.value &&
          highlight.value.to &&
          highlight.value.from
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
    getRange(facet): { from: number; to: number; type: string } {
      if (!facet.selection) {
        return null;
      }
      if (facet.selection.length < 2) {
        return null;
      }

      const values = this.facetData.values;
      const minIndex = facet.selection[0];
      const maxIndex = facet.selection[1];

      const lowerBound = _.toNumber(values[minIndex].label);
      let upperBound;

      if (this.summary.baseline.buckets.length > maxIndex) {
        upperBound = _.toNumber(values[maxIndex].label);
      } else {
        const maxBasis = _.toNumber(values[maxIndex - 1].label);
        const offset = maxBasis - _.toNumber(values[maxIndex - 2].label);
        upperBound = maxBasis + offset;
      }

      return {
        from: lowerBound,
        to: upperBound,
        type: this.summary.type,
      };
    },
    updateSelection(event) {
      if (!this.enableHighlighting) return;
      const facet = event.currentTarget;
      if (
        event.detail.changedProperties.get("selection") !== undefined &&
        !_.isEqual(facet.selection, this.selection)
      ) {
        this.$emit(
          EventList.FACETS.RANGE_CHANGE_EVENT,
          this.instanceName,
          this.summary.key,
          this.getRange(facet),
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
  },
});
</script>

<style scoped>
.facet-footer-container {
  min-height: 12px;
  padding: 6px 4px 5px 5px;
  font-family: "IBM Plex Sans", sans-serif;
  font-size: 12px;
  font-weight: 600;
  line-height: 16px;
}

.facet-footer-custom-html {
  margin-top: 6px;
}

.facet-header-container {
  color: rgba(0, 0, 0, 0.54);
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
