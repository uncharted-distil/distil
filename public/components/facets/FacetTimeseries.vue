<template>
  <div class="facet-timeseries">
    <facet-entry
      :summary="summary"
      :highlight="highlight"
      :row-selection="rowSelection"
      :enabled-type-changes="enabledTypeChanges"
      :enable-highlighting="
        Boolean(enableHighlighting) && enableHighlighting[0]
      "
      :ignore-highlights="Boolean(ignoreHighlights) && ignoreHighlights[0]"
      :instanceName="instanceName"
      :html="customHtml"
      :expandCollapse="expandCollapse"
      @html-appended="onHtmlAppend"
      @numerical-click="onNumericalClick"
      @categorical-click="onCategoricalClick"
      @facet-click="onFacetClick"
      @range-change="onRangeChange"
    >
    </facet-entry>
    <facet-entry
      v-if="!!timelineSummary && expand"
      :summary="timelineSummary"
      :highlight="highlight"
      :row-selection="rowSelection"
      :enabled-type-changes="enabledTypeChanges"
      :instanceName="instanceName"
      :enable-highlighting="
        Boolean(enableHighlighting) && enableHighlighting[1]
      "
      :ignore-highlights="Boolean(ignoreHighlights) && ignoreHighlights[1]"
      :html="footerHtml"
      @numerical-click="onHistogramNumericalClick"
      @categorical-click="onHistogramCategoricalClick"
      @range-change="onHistogramRangeChange"
    >
    </facet-entry>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import FacetEntry from "./FacetEntry.vue";
import { getters as datasetGetters } from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import {
  Dataset,
  Variable,
  VariableSummary,
  Highlight,
  RowSelection,
  Row,
  NUMERICAL_SUMMARY
} from "../../store/dataset";
import {
  INTEGER_TYPE,
  EXPAND_ACTION_TYPE,
  COLLAPSE_ACTION_TYPE
} from "../../util/types";

export default Vue.extend({
  name: "facet-timeseries",

  components: {
    FacetEntry
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
      Function as () => Function
    ]
  },

  data() {
    return {
      customHtml: this.html,
      footerHtml: undefined,
      expand: true
    };
  },

  computed: {
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    variable(): Variable {
      return this.variables.find(v => v.colName === this.summary.key);
    },

    timelineSummary(): VariableSummary {
      if (this.summary.pending) {
        return null;
      }

      const summaryVar = this.variables.find(
        v => v.colName === this.summary.key
      );
      if (!summaryVar) {
        return null;
      }

      const grouping = this.variable.grouping;
      if (!grouping) {
        return null;
      }
      const timeVarName = grouping.properties.xCol;

      if (this.summary.pending || !this.variable) {
        return null;
      }

      return {
        label: timeVarName,
        key: timeVarName,
        dataset: this.summary.dataset,
        description: this.summary.description,
        type: NUMERICAL_SUMMARY,
        varType: this.summary.timelineType,
        baseline: this.summary.timelineBaseline,
        filtered: this.summary.timeline
      };
    }
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
    }
  },
  watch: {
    expand() {
      if (!this.expand) {
        this.customHtml = () => this.footerHtml;
      }
    }
  }
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
