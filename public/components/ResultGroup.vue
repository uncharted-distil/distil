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
          @click.stop="onCollapseClick"
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
          v-if="!isRoc"
        >
          <component
            :is="getFacetByType(summary.type)"
            enable-highlighting
            :geo-enabled="hasGeoData && isActiveSolution"
            :summary="summary"
            :enable-importance="false"
            :highlights="highlights"
            :enabled-type-changes="[]"
            :row-selection="rowSelection"
            :instance-name="predictedInstanceName"
            :style="facetColors"
            @numerical-click="onResultNumericalClick"
            @range-change="onResultRangeChange"
            @facet-click="onResultCategoricalClick"
          />
        </div>

        <div class="residual-group-container" v-if="!isRoc">
          <component
            :is="getFacetByType(summary.type)"
            v-for="summary in residualSummaries"
            :key="summary.key"
            class="residual-container"
            show-origin
            enable-highlighting
            :summary="summary"
            :enable-importance="false"
            :highlights="highlights"
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
          v-if="!isRoc"
          v-for="summary in correctnessSummaries"
          :key="summary.key"
          :enable-importance="false"
          enable-highlighting
          color-scale-toggle
          :geo-enabled="hasGeoData && isActiveSolution"
          :summary="summary"
          :highlights="highlights"
          :enabled-type-changes="[]"
          :row-selection="rowSelection"
          :instance-name="correctnessInstanceName"
          @facet-click="onCorrectnessCategoricalClick"
        />

        <component
          :is="getFacetByType(summary.type)"
          v-for="summary in confidenceSummaries"
          :key="summary.key"
          enable-highlighting
          :enable-importance="false"
          :geo-enabled="hasGeoData && isActiveSolution"
          :summary="summary"
          :highlights="highlights"
          :enabled-type-changes="[]"
          :row-selection="rowSelection"
          :instance-name="confidenceInstanceName"
          :style="facetColors"
          @range-change="onConfidenceRangeChange"
          @facet-click="onConfidenceClick"
        />
        <component
          :is="getFacetByType(summary.type)"
          v-for="summary in rankingSummaries"
          :key="summary.key"
          enable-highlighting
          :enable-importance="false"
          :geo-enabled="hasGeoData && isActiveSolution"
          :summary="summary"
          :highlight="highlights"
          :enabled-type-changes="[]"
          :row-selection="rowSelection"
          :instance-name="rankingInstanceName"
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
import { SolutionStatus } from "../store/requests/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
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
  isTopSolutionByTime,
  SOLUTION_PROGRESS,
  SOLUTION_LABELS,
  reviseOpenSolutions,
} from "../util/solutions";
import { getModelNameByFittedSolutionId } from "../util/models";
import { overlayRouteEntry } from "../util/routes";
import {
  updateHighlight,
  clearHighlight,
  UPDATE_FOR_KEY,
} from "../util/highlights";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import _ from "lodash";
import store from "../store/store";
import { EventList } from "../util/events";

interface Score {
  label: string;
  metric: string;
  solutionId: string;
  sortMultiplier: number;
  value: number;
}
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
    scores: Array as () => Score[],
    targetSummary: Object as () => VariableSummary,
    predictedSummary: Object as () => VariableSummary,
    residualsSummary: Object as () => VariableSummary,
    correctnessSummary: Object as () => VariableSummary,
    confidenceSummary: Object as () => VariableSummary,
    rankingSummary: Object as () => VariableSummary,
  },

  data() {
    return {
      openDeleteModal: false,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    isActiveSolution(): boolean {
      return (
        requestGetters.getActiveSolution(this.$store).solutionId ===
        this.solutionId
      );
    },
    errorColor(): string {
      return applyColor([FACET_COLOR_ERROR, null, null, FACET_COLOR_FILTERED]);
    },
    facetColors(): string {
      return applyColor([
        null,
        !!this.rowSelection ? FACET_COLOR_SELECT : null,
        null,
        FACET_COLOR_FILTERED,
      ]);
    },
    hasGeoData(): boolean {
      return routeGetters.hasGeoData(this.$store);
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
    rankingInstanceName(): string {
      return `ranking-result-facet-${this.solutionId}`;
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
    rankingSummaries(): VariableSummary[] {
      return this.rankingSummary ? [this.rankingSummary] : [];
    },
    residualSummaries(): VariableSummary[] {
      return this.residualsSummary ? [this.residualsSummary] : [];
    },

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },

    residualThreshold(): Extrema {
      return {
        min: _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store)),
        max: _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store)),
      };
    },

    isPending(): boolean {
      return (
        this.solutionStatus !== SolutionStatus.SOLUTION_COMPLETED &&
        this.solutionStatus !== SolutionStatus.SOLUTION_ERRORED &&
        this.solutionStatus !== SolutionStatus.SOLUTION_CANCELLED
      );
    },

    isCompleted(): boolean {
      return this.solutionStatus === SolutionStatus.SOLUTION_COMPLETED;
    },

    isErrored(): boolean {
      return this.solutionStatus === SolutionStatus.SOLUTION_ERRORED;
    },

    isCancelled(): boolean {
      return this.solutionStatus === SolutionStatus.SOLUTION_CANCELLED;
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
    isOpenInRoute(): boolean {
      return this.openSolutions.some((s) => {
        return s === this.requestId;
      });
    },
    isMaximized(): boolean {
      return (
        (this.routeSolutionId === this.solutionId || !this.isErrored) &&
        this.isOpenInRoute
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
      return isTopSolutionByTime(
        store.state.requestsModule.solutions,
        this.solutionId,
        3
      );
    },

    hasExplanations(): boolean {
      // waiting for explanation enum to be added to results summaries
      return !!this.predictedSummary?.weighted;
    },
    openSolutions(): string[] {
      return routeGetters.getRouteOpenSolutions(this.$store);
    },
    isRoc(): boolean {
      return "ROC AUC" === this.scores[0]?.label;
    },
  },
  mounted() {
    if (
      (this.routeSolutionId === this.solutionId &&
        !this.isErrored &&
        !this.isOpenInRoute) ||
      (this.isTopN &&
        this.openSolutions.length < 3 &&
        !this.isErrored &&
        !this.isOpenInRoute)
    ) {
      reviseOpenSolutions(this.requestId, this.$route, this.$router);
    }
  },
  watch: {
    isActiveSolution() {
      if (!this.isOpenInRoute) {
        reviseOpenSolutions(this.requestId, this.$route, this.$router);
      }
    },
  },
  methods: {
    onCollapseClick() {
      reviseOpenSolutions(this.requestId, this.$route, this.$router);
    },
    getFacetByType: getFacetByType,
    onResultCategoricalClick(
      context: string,
      key: string,
      value: string[],
      dataset: string
    ) {
      let highlight = this.highlights.find((h) => {
        return h.key === key;
      });
      if (key && value && Array.isArray(value) && value.length > 0) {
        highlight = highlight ?? {
          context: context,
          dataset: dataset,
          key: key,
          value: [],
        };
        highlight.value = value;
        updateHighlight(this.$router, highlight, UPDATE_FOR_KEY);
      } else {
        clearHighlight(this.$router, highlight.key);
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
      value: string[],
      dataset: string
    ) {
      let highlight = this.highlights.find((h) => {
        return h.key === key;
      });
      if (key && value && Array.isArray(value) && value.length > 0) {
        highlight = highlight ?? {
          context: context,
          dataset: dataset,
          key: key,
          value: [],
        };
        highlight.value = value;
        updateHighlight(this.$router, highlight, UPDATE_FOR_KEY);
      } else {
        clearHighlight(this.$router, highlight.key);
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
      const uniqueHighlight = this.highlights.reduce(
        (acc, highlight) => highlight.key !== key || acc,
        false
      );
      if (uniqueHighlight) {
        if (key && value) {
          updateHighlight(this.$router, {
            context: context,
            dataset: dataset,
            key: key,
            value: value,
          });
        } else {
          clearHighlight(this.$router, key);
        }
      }
    },

    onResultRangeChange(
      context: string,
      key: string,
      value: { from: { label: string[] }; to: { label: string[] } },
      dataset: string
    ) {
      if (key && value) {
        updateHighlight(
          this.$router,
          {
            context: context,
            dataset: dataset,
            key: key,
            value: value,
          },
          UPDATE_FOR_KEY
        );
      } else {
        clearHighlight(this.$router, key);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value },
      });
      this.$emit(EventList.FACETS.RANGE_CHANGE_EVENT, key, value);
    },

    onResidualNumericalClick(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      const uniqueHighlight = this.highlights.reduce(
        (acc, highlight) => highlight.key !== key || acc,
        false
      );
      if (uniqueHighlight) {
        if (key && value) {
          updateHighlight(this.$router, {
            context: context,
            dataset: dataset,
            key: key,
            value: value,
          });
        } else {
          clearHighlight(this.$router, key);
        }
      }
    },

    onResidualRangeChange(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      if (key && value) {
        updateHighlight(
          this.$router,
          {
            context: context,
            dataset: dataset,
            key: key,
            value: value,
          },
          UPDATE_FOR_KEY
        );
      } else {
        clearHighlight(this.$router, key);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value },
      });
      this.$emit(EventList.FACETS.RANGE_CHANGE_EVENT, key, value);
    },

    onConfidenceClick(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      const uniqueHighlight = this.highlights.reduce(
        (acc, highlight) => highlight.key !== key || acc,
        false
      );
      if (uniqueHighlight) {
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
      if (key && value) {
        updateHighlight(
          this.$router,
          {
            context: context,
            dataset: dataset,
            key: key,
            value: value,
          },
          UPDATE_FOR_KEY
        );
      } else {
        clearHighlight(this.$router, key);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value },
      });
      this.$emit(EventList.FACETS.RANGE_CHANGE_EVENT, key, value);
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
          colorScaleVariable: "",
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
