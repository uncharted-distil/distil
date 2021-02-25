<template>
  <facet-bars
    :data.prop="facetData"
    :selection.prop="selection"
    :subselection.prop="subSelection"
    :disabled.prop="!enableHighlighting"
    @facet-element-updated="updateSelection"
  >
    <div slot="header-label" :class="headerClass">
      <span>{{ summary.label.toUpperCase() }}</span>
      <importance-bars v-if="importance" :importance="importance" />
      <type-change-menu
        v-if="facetEnableTypeChanges"
        class="facet-header-dropdown"
        :dataset="summary.dataset"
        :field="summary.key"
        :expand-collapse="expandCollapse"
      />
    </div>

    <facet-template target="facet-bars-value" title="${tooltip}" />

    <div slot="footer" class="facet-footer-container">
      <facet-plugin-zoom-bar
        min-bar-width="8"
        auto-hide="true"
        round-caps="true"
      />
      <div
        v-if="this.html"
        v-child="computeCustomHTML()"
        class="facet-footer-custom-html"
      />
    </div>
  </facet-bars>
</template>

<script lang="ts">
import Vue from "vue";

import "@uncharted.software/facets-core";
import "@uncharted.software/facets-plugins";

import TypeChangeMenu from "../TypeChangeMenu";
import ImportanceBars from "../ImportanceBars";
import { Highlight, RowSelection, VariableSummary } from "../../store/dataset";
import {
  getSubSelectionValues,
  hasBaseline,
  hasSummary,
  facetTypeChangeState,
} from "../../util/facets";
import { DATETIME_FILTER } from "../../util/filters";
import { numToDate, dateToNum } from "../../util/types";
import _ from "lodash";
import moment from "moment";

export default Vue.extend({
  name: "FacetDateTime",

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

  computed: {
    max(): number {
      if (hasBaseline(this.summary)) {
        const buckets = this.summary.baseline.buckets;
        // seems to be incorrect compute based on the current buckets
        // const maxCount = summary.baseline.extrema.max;
        return buckets.reduce((max, bucket) => Math.max(max, bucket.count), 0);
      }
      return 0;
    },
    facetData(): { label: string; values: { ratio: number; label: string }[] } {
      const summary = this.summary;
      const values = [];
      if (hasBaseline(summary)) {
        const buckets = summary.baseline.buckets;
        for (let i = 0, n = buckets.length; i < n; ++i) {
          const bucketDate = this.numToDate(buckets[i].key);
          const count = buckets[i].count;
          values.push({
            ratio: count / this.max,
            label: bucketDate,
            tooltip: `Date:\t\t${bucketDate}\nTotal Count:\t${count}`,
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
      return getSubSelectionValues(this.summary, this.rowSelection, this.max);
    },
    selection(): number[] {
      if (!this.enableHighlighting || !this.isHighlightedGroup()) {
        return null;
      }
      const highlightValues = this.getHighlightValues();

      const buckets = this.summary.baseline.buckets;
      const highlightValuesFlat = highlightValues.reduce(
        (acc, val) => (val.from && val.to ? [...acc, val.from, val.to] : acc),
        []
      );

      // map the values used for the highlight filter back to the buckets
      const highlightAsSelection = buckets.reduce((acc, val, ind) => {
        const key = _.toNumber(val.key);
        if (highlightValuesFlat.includes(key)) {
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
  },

  methods: {
    numToDate,
    dateToNum,
    getHighlightValues(): { from: number; to: number }[] {
      return this.highlights.reduce(
        (acc, highlight) =>
          highlight.value && highlight.value.to && highlight.value.from
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

      const lowerBound = this.dateToNum(values[minIndex].label);
      let upperBound;

      if (this.summary.baseline.buckets.length > maxIndex) {
        upperBound = this.dateToNum(values[maxIndex].label);
      } else {
        const maxBasis = this.dateToNum(values[maxIndex - 1].label);
        const offset = maxBasis - this.dateToNum(values[maxIndex - 2].label);
        // increase filter upperbound to definitely include the last date bucket
        upperBound = maxBasis + offset * 2;
      }
      return {
        from: lowerBound,
        to: upperBound,
        type: DATETIME_FILTER,
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
          "range-change",
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
