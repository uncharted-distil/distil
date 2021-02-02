<template>
  <div class="facet-timeseries">
    <facet-sparklines
      :summary="summary"
      :highlight="highlight"
      :row-selection="rowSelection"
      :enabled-type-changes="enabledTypeChanges"
      :enable-highlighting="
        Boolean(enableHighlighting) && enableHighlighting[0]
      "
      :ignore-highlights="Boolean(ignoreHighlights) && ignoreHighlights[0]"
      :instance-name="instanceName"
      :html="customHtml"
      :expand-collapse="expandCollapse"
      :variable="variable"
      :expand="expand"
      @html-appended="onHtmlAppend"
      @numerical-click="onNumericalClick"
      @categorical-click="onCategoricalClick"
      @facet-click="onFacetClick"
      @range-change="onRangeChange"
    />
    <component
      :is="facetType"
      v-if="!!timelineSummary && expand"
      :summary="timelineSummary"
      :highlight="highlight"
      :row-selection="rowSelection"
      :enabled-type-changes="enabledTypeChanges"
      :instance-name="instanceName"
      :enable-highlighting="
        Boolean(enableHighlighting) && enableHighlighting[1]
      "
      :ignore-highlights="Boolean(ignoreHighlights) && ignoreHighlights[1]"
      :html="footerHtml"
      @numerical-click="onHistogramNumericalClick"
      @categorical-click="onHistogramCategoricalClick"
      @range-change="onHistogramRangeChange"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import FacetDateTime from "./FacetDateTime.vue";
import FacetNumerical from "./FacetNumerical.vue";
import FacetSparklines from "./FacetSparklines.vue";
import { getters as datasetGetters } from "../../store/dataset/module";
import {
  Variable,
  VariableSummary,
  Highlight,
  RowSelection,
  NUMERICAL_SUMMARY,
  TimeseriesGrouping,
} from "../../store/dataset";
import { EXPAND_ACTION_TYPE, COLLAPSE_ACTION_TYPE } from "../../util/types";

/**
 * Timeseries Facet.
 * @param {Boolean} [expanded=false] - To display the facet expanded; Collapsed by default.
 */
export default Vue.extend({
  name: "facet-timeseries",

  components: {
    FacetSparklines,
    FacetDateTime,
    FacetNumerical,
  },

  props: {
    summary: Object as () => VariableSummary,
    highlight: Object as () => Highlight,
    rowSelection: Object as () => RowSelection,
    instanceName: String as () => string,
    enabledTypeChanges: Array as () => string[],
    enableHighlighting: Array as () => boolean[],
    ignoreHighlights: Array as () => boolean[],
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function,
    ],
    expanded: { type: Boolean, default: false },
  },

  data() {
    return {
      customHtml: this.html,
      footerHtml: undefined,
      expand: this.expanded,
    };
  },

  computed: {
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    variable(): Variable {
      return this.variables.find((v) => v.key === this.summary.key);
    },

    grouping(): TimeseriesGrouping {
      if (!this.variable || !this.variable.grouping) {
        return null;
      }
      return this.variable.grouping as TimeseriesGrouping;
    },

    timelineSummary(): VariableSummary {
      if (this.summary.pending) {
        return null;
      }

      const summaryVar = this.variables.find((v) => v.key === this.summary.key);

      if (!summaryVar || !this.grouping || !this.variable) {
        return null;
      }
      return {
        label: this.grouping.xCol,
        key: this.grouping.xCol,
        dataset: this.summary.dataset,
        description: this.summary.description,
        type: NUMERICAL_SUMMARY,
        varType: this.summary.timelineType,
        baseline: this.summary.timelineBaseline,
        filtered: this.summary.timeline,
      };
    },
    facetType(): string {
      if (this.timelineSummary.varType === "dateTime") {
        return "facet-date-time";
      } else {
        return "facet-numerical";
      }
    },
  },

  methods: {
    expandCollapse(action) {
      if (action === EXPAND_ACTION_TYPE) {
        this.expand = true;
      } else if (action === COLLAPSE_ACTION_TYPE) {
        this.expand = false;
      }
    },
    onCategoricalClick(...args) {
      this.$emit("categorical-click", ...args);
    },
    onFacetClick(...args) {
      this.$emit("facet-click", ...args);
    },
    onNumericalClick(...args) {
      this.$emit("numerical-click", ...args);
    },
    onRangeChange(...args) {
      this.$emit("range-change", ...args);
    },
    onHistogramCategoricalClick(...args) {
      this.$emit("histogram-categorical-click", ...args);
    },
    onHistogramNumericalClick(...args) {
      this.$emit("histogram-numerical-click", ...args);
    },
    onHistogramRangeChange(...args) {
      this.$emit("histogram-range-change", ...args);
    },
    onHtmlAppend(html: HTMLDivElement) {
      // Once html is rendered in top facets, move the element to the bottom facets
      // So that custom html are rendered at the bottom of the coumpound facets
      this.footerHtml = () => html;
    },
  },
  watch: {
    expand() {
      if (!this.expand) {
        this.customHtml = () => this.footerHtml;
      }
    },
  },
});
</script>

<style>
.facet-timeseries .facets-group .group-facet-container {
  max-height: 150px !important;
}
.facet-timeseries .facets-root:first-child {
  margin-bottom: 1px;
}
</style>
