<template>
  <div
    class="result-group"
    :class="{ 'result-group-selected': isSelected }"
    @click="onClick()"
  >
    <header class="result-group-title">
      <h5 v-if="modelName">
        {{ modelName }}
      </h5>
      <b>
        {{ name }} <sup>{{ solutionIndex }}</sup>
      </b>

      <template v-if="!isErrored && !isSelected">
        <div
          class="pull-right pl-2 solution-button"
          @click.stop="minimized = !minimized"
        >
          <i
            class="fa"
            :class="{
              'fa-angle-down': !isMaximized,
              'fa-angle-up': isMaximized,
            }"
          />
        </div>
        <!--
        <div class="pull-right">|</div>
        -->
      </template>
      <!--
      <div class="pull-right pr-2 solution-button" @click.stop="onDelete"><i class="fa fa-trash"></i></div>
      -->
      <template v-if="isPending">
        <b-badge variant="info">
          {{ progressLabel }}
        </b-badge>
        <b-progress
          :value="percentComplete"
          variant="outline-secondary"
          striped
          :animated="true"
        />
      </template>
      <template v-if="isCompleted">
        <b-badge
          v-for="score in scores"
          :key="`${score.metric}-${solutionId}`"
          variant="info"
        >
          {{ score.label }}: {{ score.value.toFixed(2) }}
        </b-badge>
        &nbsp;
      </template>
      <template v-if="hasExplanations">
        <b-badge variant="info">Explanations</b-badge>
      </template>
      <template v-if="isErrored">
        <b-badge variant="danger">ERROR</b-badge>
      </template>
      <template v-if="isCancelled">
        <b-badge variant="secondary">CANCELLED</b-badge>
      </template>
    </header>

    <div v-if="isMaximized" class="result-group-body">
      <template v-if="isCompleted">
        <div
          v-for="summary in predictedSummaries"
          :key="summary.key"
          ref="predicted-summaries"
        >
          <component
            :is="getFacetByType(summary.type)"
            enable-highlighting
            :summary="summary"
            :highlight="highlight"
            :enabled-type-changes="[]"
            :row-selection="rowSelection"
            :instance-name="predictedInstanceName"
            :style="facetColors"
            @numerical-click="onResultNumericalClick"
            @range-change="onResultRangeChange"
            @facet-click="onResultCategoricalClick"
          />
        </div>

        <div class="residual-group-container">
          <component
            :is="getFacetByType(summary.type)"
            v-for="summary in residualSummaries"
            :key="summary.key"
            class="residual-container"
            show-origin
            enable-highlighting
            :summary="summary"
            :highlight="highlight"
            :enabled-type-changes="[]"
            :row-selection="rowSelection"
            :instance-name="residualInstanceName"
            :deemphasis="residualThreshold"
            :style="errorColor"
            @numerical-click="onResidualNumericalClick"
            @range-change="onResidualRangeChange"
            @facet-click="onResultCategoricalClick"
          />
        </div>

        <component
          :is="getFacetByType(summary.type)"
          v-for="summary in correctnessSummaries"
          :key="summary.key"
          enable-highlighting
          :summary="summary"
          :highlight="highlight"
          :enabled-type-changes="[]"
          :row-selection="rowSelection"
          :instance-name="correctnessInstanceName"
          :style="errorColor"
          @facet-click="onCorrectnessCategoricalClick"
        />

        <component
          :is="getFacetByType(summary.type)"
          v-for="summary in confidenceSummaries"
          :key="summary.key"
          enable-highlighting
          :summary="summary"
          :highlight="highlight"
          :enabled-type-changes="[]"
          :row-selection="rowSelection"
          :instance-name="confidenceInstanceName"
          :style="facetColors"
          @range-change="onConfidenceRangeChange"
          @facet-click="onConfidenceClick"
        />
      </template>
    </div>
  </div>
</template>

<script lang="ts">
// Component that contains a histogram of regression predictions, a histogram of the
// of prediction-truth residuals, and scoring information.

import Vue from "vue";
import FacetNumerical from "../components/facets/FacetNumerical.vue";
import FacetCategorical from "../components/facets/FacetCategorical.vue";
import {
  Extrema,
  VariableSummary,
  RowSelection,
  Highlight,
} from "../store/dataset/index";
import {
  SOLUTION_CANCELLED,
  SOLUTION_COMPLETED,
  SOLUTION_ERRORED,
} from "../store/requests/index";
import { getters as routeGetters } from "../store/route/module";
import {
  getFacetByType,
  applyColor,
  FACET_COLOR_SELECT,
  FACET_COLOR_FILTERED,
  FACET_COLOR_ERROR,
} from "../util/facets";
import {
  getSolutionIndex,
  getSolutionById,
  isTopSolutionByScore,
  SOLUTION_PROGRESS,
  SOLUTION_LABELS,
} from "../util/solutions";
import { getModelNameByFittedSolutionId } from "../util/models";
import { overlayRouteEntry } from "../util/routes";
import { updateHighlight, clearHighlight } from "../util/highlights";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import _ from "lodash";
import store from "../store/store";

export default Vue.extend({
  name: "ResultGroup",

  components: {
    FacetNumerical,
    FacetCategorical,
  },

  props: {
    name: String as () => string,
    timestamp: String as () => string,
    requestId: String as () => string,
    solutionId: String as () => string,
    singleSolution: Boolean as () => boolean,
    scores: Array as () => number[],
    targetSummary: Object as () => VariableSummary,
    predictedSummary: Object as () => VariableSummary,
    residualsSummary: Object as () => VariableSummary,
    correctnessSummary: Object as () => VariableSummary,
    confidenceSummary: Object as () => VariableSummary,
  },

  data() {
    return {
      minimized: null,
      openDeleteModal: false,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    errorColor(): string {
      return applyColor([FACET_COLOR_ERROR]);
    },
    facetColors(): string {
      return applyColor([
        null,
        !!this.rowSelection ? FACET_COLOR_SELECT : null,
        null,
        FACET_COLOR_FILTERED,
      ]);
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    /**
     * Name of the model if it exist.
     * @return {String}
     */
    modelName(): string {
      // Find the fitted solution ID.
      const fittedSolutionId = getSolutionById(
        store.state.requestsModule.solutions,
        this.solutionId
      )?.fittedSolutionId;
      if (_.isEmpty(fittedSolutionId)) {
        return;
      }

      // Retreive the model name from the fitted solution.
      const name = getModelNameByFittedSolutionId(fittedSolutionId);

      // Return the name if it exist, null otherwise.
      return _.isNil(name) ? null : name;
    },

    predictedInstanceName(): string {
      return `predicted-result-facet-${this.solutionId}`;
    },

    residualInstanceName(): string {
      return `residual-result-facet-${this.solutionId}`;
    },

    correctnessInstanceName(): string {
      return `correctness-result-facet-${this.solutionId}`;
    },
    confidenceInstanceName(): string {
      return `confidence-result-facet-${this.solutionId}`;
    },

    routeSolutionId(): string {
      return routeGetters.getRouteSolutionId(store);
    },

    solutionStatus(): string {
      const solution = getSolutionById(
        store.state.requestsModule.solutions,
        this.solutionId
      );
      if (solution) {
        return solution.progress;
      }
      return "unknown";
    },

    progressLabel(): string {
      return SOLUTION_LABELS[this.solutionStatus];
    },

    percentComplete(): number {
      return SOLUTION_PROGRESS[this.solutionStatus];
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    solutionIndex(): number {
      return getSolutionIndex(this.solutionId);
    },

    predictedSummaries(): VariableSummary[] {
      return this.predictedSummary ? [this.predictedSummary] : [];
    },

    correctnessSummaries(): VariableSummary[] {
      return this.correctnessSummary ? [this.correctnessSummary] : [];
    },

    confidenceSummaries(): VariableSummary[] {
      return this.confidenceSummary ? [this.confidenceSummary] : [];
    },

    residualSummaries(): VariableSummary[] {
      return this.residualsSummary ? [this.residualsSummary] : [];
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    residualThreshold(): Extrema {
      return {
        min: _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store)),
        max: _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store)),
      };
    },

    isPending(): boolean {
      return (
        this.solutionStatus !== SOLUTION_COMPLETED &&
        this.solutionStatus !== SOLUTION_ERRORED &&
        this.solutionStatus !== SOLUTION_CANCELLED
      );
    },

    isCompleted(): boolean {
      return this.solutionStatus === SOLUTION_COMPLETED;
    },

    isErrored(): boolean {
      return this.solutionStatus === SOLUTION_ERRORED;
    },

    isCancelled(): boolean {
      return this.solutionStatus === SOLUTION_CANCELLED;
    },

    isBad(): boolean {
      const solution = getSolutionById(
        store.state.requestsModule.solutions,
        this.solutionId
      );
      if (solution) {
        return solution.isBad;
      }
      return false;
    },

    isMinimized(): boolean {
      return this.minimized !== null ? this.minimized : !this.isTopN;
    },

    isMaximized(): boolean {
      return (
        this.routeSolutionId === this.solutionId ||
        (!this.isMinimized && !this.isErrored)
      );
    },

    isSelected(): boolean {
      return (
        !this.singleSolution &&
        this.predictedSummary &&
        this.solutionId === this.routeSolutionId
      );
    },

    isTopN(): boolean {
      return isTopSolutionByScore(
        store.state.requestsModule.solutions,
        this.requestId,
        this.solutionId,
        3
      );
    },

    hasExplanations(): boolean {
      // waiting for explanation enum to be added to results summaries
      return !!this.predictedSummary?.weighted;
    },
  },

  methods: {
    getFacetByType: getFacetByType,
    onResultCategoricalClick(
      context: string,
      key: string,
      value: string,
      dataset: string
    ) {
      if (key && value) {
        // extract the var name from the key
        updateHighlight(this.$router, {
          context: context,
          dataset: dataset,
          key: key,
          value: value,
        });
      } else {
        clearHighlight(this.$router);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value },
      });
    },

    onCorrectnessCategoricalClick(
      context: string,
      key: string,
      value: string,
      dataset: string
    ) {
      if (key && value) {
        // extract the var name from the key
        updateHighlight(this.$router, {
          context: context,
          dataset: dataset,
          key: key,
          value: value,
        });
      } else {
        clearHighlight(this.$router);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value },
      });
    },

    onResultNumericalClick(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      if (!this.highlight || this.highlight.key !== key) {
        updateHighlight(this.$router, {
          context: context,
          dataset: dataset,
          key: key,
          value: value,
        });
      }
    },

    onResultRangeChange(
      context: string,
      key: string,
      value: { from: { label: string[] }; to: { label: string[] } },
      dataset: string
    ) {
      updateHighlight(this.$router, {
        context: context,
        dataset: dataset,
        key: key,
        value: value,
      });
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value },
      });
      this.$emit("range-change", key, value);
    },

    onResidualNumericalClick(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      if (!this.highlight || this.highlight.key !== key) {
        updateHighlight(this.$router, {
          context: context,
          dataset: dataset,
          key: key,
          value: value,
        });
      }
    },

    onResidualRangeChange(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      updateHighlight(this.$router, {
        context: context,
        dataset: dataset,
        key: key,
        value: value,
      });
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value },
      });
      this.$emit("range-change", key, value);
    },

    onConfidenceClick(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      if (!this.highlight || this.highlight.key !== key) {
        updateHighlight(this.$router, {
          context: context,
          dataset: dataset,
          key: key,
          value: value,
        });
      }
    },

    onConfidenceRangeChange(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      updateHighlight(this.$router, {
        context: context,
        dataset: dataset,
        key: key,
        value: value,
      });
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value },
      });
      this.$emit("range-change", key, value);
    },

    onClick() {
      if (this.predictedSummary && this.routeSolutionId !== this.solutionId) {
        appActions.logUserEvent(this.$store, {
          feature: Feature.SELECT_MODEL,
          activity: Activity.MODEL_SELECTION,
          subActivity: SubActivity.MODEL_EXPLANATION,
          details: { solutionId: this.solutionId },
        });
        const entry = overlayRouteEntry(this.$route, {
          solutionId: this.solutionId,
          highlights: null,
        });
        this.$router.push(entry).catch((err) => console.warn(err));
      }
    },

    onDelete() {
      this.openDeleteModal = true;
    },

    deleteSolution() {
      this.openDeleteModal = false;
    },
  },
});
</script>

<style>
.result-group {
  margin: 5px;
  padding: 10px;
  border-bottom-style: solid;
  border-bottom-color: lightgray;
  border-bottom-width: 1px;
}

.result-group-title {
  vertical-align: middle;
}

.result-group-title .badge {
  display: inline;
  vertical-align: middle;
  padding: 0.45em 0.4em 0.3em 0.4em;
}

.result-group-body {
  padding: 4px 0;
}

.solution-button {
  cursor: pointer;
}
.solution-button:hover {
  opacity: 0.5;
}

.result-group-selected {
  padding: 9px;
  border-style: solid;
  border-color: #007bff;
  box-shadow: 0 0 10px #007bff;
  border-width: 1px;
  border-radius: 2px;
  padding-bottom: 10px;
}

.result-group:not(.result-group-selected):hover {
  padding: 9px;
  border-style: solid;
  border-color: lightgray;
  border-width: 1px;
  border-radius: 2px;
  padding-bottom: 10px;
}

.result-container {
  position: relative;
  box-shadow: none;
}

.result-container {
  box-shadow: none;
}

.result-container .facets-group,
.residual-container .facets-group {
  box-shadow: none;
}

.result-group,
.result-container .facets-group,
.result-container .facets-group .group-header,
.residual-container .facets-group,
.residual-container .facets-group .group-header {
  cursor: pointer !important;
}

.residual-container .facets-facet-horizontal .facet-histogram-bar-highlighted {
  fill: #e05353;
}

.residual-container
  .facets-facet-horizontal
  .facet-histogram-bar-highlighted:hover {
  fill: #662424;
}

.residual-container
  .facets-facet-horizontal
  .facet-histogram-bar-highlighted.select-highlight {
  fill: #007bff;
}

.residual-container .facets-facet-vertical .facet-bar-selected {
  box-shadow: inset 0 0 0 1000px #e0535e;
}

.residual-container .facets-facet-horizontal .facet-range-filter {
  box-shadow: inset 0 0 0 1000px rgba(225, 0, 11, 0.15);
}

.residual-group-container {
  position: relative;
}
</style>
