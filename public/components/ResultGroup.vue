<template>
  <div v-bind:class="currentClass" @click="onClick()">
    <div class="result-group-title">
      <b
        >{{ name }} <sup>{{ solutionIndex }}</sup></b
      >
      <template v-if="!isErrored">
        <div
          class="pull-right pl-2 solution-button"
          @click.stop="minimized = !minimized"
        >
          <i
            class="fa"
            v-bind:class="{
              'fa-angle-down': !isMaximized,
              'fa-angle-up': isMaximized
            }"
          ></i>
        </div>
        <!--
				<div class="pull-right">|</div>
				-->
      </template>
      <!--
			<div class="pull-right pr-2 solution-button" @click.stop="onDelete"><i class="fa fa-trash"></i></div>
			-->
      <template v-if="isPending">
        <b-badge variant="info">{{ progressLabel }}</b-badge>
        <b-progress
          :value="percentComplete"
          variant="outline-secondary"
          striped
          :animated="true"
        ></b-progress>
      </template>
      <template v-if="isCompleted">
        <b-badge
          variant="info"
          v-bind:key="`${score.metric}-${solutionId}`"
          v-for="score in scores"
        >
          {{ score.label }}: {{ score.value.toFixed(2) }}
        </b-badge>
      </template>
      <template v-if="isErrored">
        <b-badge variant="danger">
          ERROR
        </b-badge>
      </template>
    </div>
    <div class="result-group-body" v-if="isMaximized">
      <template v-if="isCompleted">
        <div v-for="summary in predictedSummaries" :key="summary.key">
          <template v-if="summary.varType === 'timeseries'">
            <facet-timeseries
              :summary="summary"
              :highlight="highlight"
              :row-selection="rowSelection"
              :instanceName="predictedInstanceName"
              :enabled-type-changes="[]"
              :enable-highlighting="[true, true]"
              @numerical-click="onResultNumericalClick"
              @range-change="onResultRangeChange"
              @histogram-numerical-click="onResultNumericalClick"
              @histogram-range-change="onResultRangeChange"
            >
            </facet-timeseries>
          </template>
          <template v-else>
            <facet-entry
              enable-highlighting
              :summary="summary"
              :highlight="highlight"
              :enabled-type-changes="[]"
              :row-selection="rowSelection"
              :instanceName="predictedInstanceName"
              @numerical-click="onResultNumericalClick"
              @range-change="onResultRangeChange"
              @facet-click="onResultCategoricalClick"
            >
            </facet-entry>
          </template>
        </div>

        <div class="residual-group-container">
          <facet-entry
            v-for="summary in residualSummaries"
            :key="summary.key"
            class="residual-container"
            show-origin
            enable-highlighting
            :summary="summary"
            :highlight="highlight"
            :enabled-type-changes="[]"
            :row-selection="rowSelection"
            :instanceName="residualInstanceName"
            :deemphasis="residualThreshold"
            @numerical-click="onResidualNumericalClick"
            @range-change="onResidualRangeChange"
            @facet-click="onResultCategoricalClick"
          >
          </facet-entry>
        </div>

        <facet-entry
          v-for="summary in correctnessSummaries"
          :key="summary.key"
          enable-highlighting
          :summary="summary"
          :highlight="highlight"
          :enabled-type-changes="[]"
          :row-selection="rowSelection"
          :instanceName="correctnessInstanceName"
          @facet-click="onCorrectnessCategoricalClick"
        >
        </facet-entry>
      </template>
    </div>
    <b-modal v-model="openDeleteModal" hide-footer hide-header>
      <h6 class="my-4 text-center">
        Are you sure you would like to delete this solution?
      </h6>
      <footer class="modal-footer">
        <b-btn class="mt-3" variant="danger" @click="deleteSolution"
          >Delete</b-btn
        >
        <b-btn class="mt-3" variant="secondary" @click="openDeleteModal = false"
          >Cancel</b-btn
        >
      </footer>
    </b-modal>
  </div>
</template>

<script lang="ts">
// Component that contains a histogram of regression predictions, a histogram of the
// of prediction-truth residuals, and scoring information.

import Vue from "vue";
import FacetEntry from "../components/FacetEntry";
import FacetTimeseries from "../components/FacetTimeseries";
import {
  Extrema,
  VariableSummary,
  RowSelection,
  Highlight
} from "../store/dataset/index";
import { SOLUTION_COMPLETED, SOLUTION_ERRORED } from "../store/solutions/index";
import { getters as routeGetters } from "../store/route/module";
import { getters as solutionGetters } from "../store/solutions/module";
import {
  getSolutionIndex,
  getSolutionById,
  isTopSolutionByScore,
  SOLUTION_PROGRESS,
  SOLUTION_LABELS
} from "../util/solutions";
import { overlayRouteEntry } from "../util/routes";
import { updateHighlight, clearHighlight } from "../util/highlights";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import _ from "lodash";
import { descending } from "d3";

export default Vue.extend({
  name: "result-group",

  components: {
    FacetEntry,
    FacetTimeseries
  },

  props: {
    name: String as () => string,
    timestamp: String as () => string,
    requestId: String as () => string,
    solutionId: String as () => string,
    scores: Array as () => number[],
    targetSummary: Object as () => VariableSummary,
    predictedSummary: Object as () => VariableSummary,
    residualsSummary: Object as () => VariableSummary,
    correctnessSummary: Object as () => VariableSummary
  },

  data() {
    return {
      minimized: null,
      openDeleteModal: false
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
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

    routeSolutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },

    solutionStatus(): string {
      const solution = getSolutionById(
        this.$store.state.solutionModule,
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

    residualSummaries(): VariableSummary[] {
      return this.residualsSummary ? [this.residualsSummary] : [];
      // groups.forEach(group => {
      // 	group.facets.forEach((facet: any) => {
      // 		if (facet.histogram) {
      // 			facet.histogram.showOrigin = true;
      // 		}
      // 	});
      // });
      // return groups;
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    currentClass(): string {
      return this.predictedSummary && this.solutionId === this.routeSolutionId
        ? "result-group-selected result-group"
        : "result-group";
    },

    residualThreshold(): Extrema {
      return {
        min: _.toNumber(routeGetters.getRouteResidualThresholdMin(this.$store)),
        max: _.toNumber(routeGetters.getRouteResidualThresholdMax(this.$store))
      };
    },

    isPending(): boolean {
      return (
        this.solutionStatus !== SOLUTION_COMPLETED &&
        this.solutionStatus !== SOLUTION_ERRORED
      );
    },

    isCompleted(): boolean {
      return this.solutionStatus === SOLUTION_COMPLETED;
    },

    isErrored(): boolean {
      return this.solutionStatus === SOLUTION_ERRORED || this.isBad;
    },

    isBad(): boolean {
      const solution = getSolutionById(
        this.$store.state.solutionModule,
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

    isTopN(): boolean {
      return isTopSolutionByScore(
        this.$store.state.solutionModule,
        this.requestId,
        this.solutionId,
        3
      );
    }
  },

  methods: {
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
          value: value
        });
      } else {
        clearHighlight(this.$router);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value }
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
          value: value
        });
      } else {
        clearHighlight(this.$router);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value }
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
          value: value
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
        value: value
      });
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value }
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
          value: value
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
        value: value
      });
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { key: key, value: value }
      });
      this.$emit("range-change", key, value);
    },

    onClick() {
      if (this.predictedSummary && this.routeSolutionId !== this.solutionId) {
        appActions.logUserEvent(this.$store, {
          feature: Feature.SELECT_MODEL,
          activity: Activity.MODEL_SELECTION,
          subActivity: SubActivity.MODEL_EXPLANATION,
          details: { solutionId: this.solutionId }
        });
        const routeEntry = overlayRouteEntry(this.$route, {
          solutionId: this.solutionId,
          highlights: null
        });
        this.$router.push(routeEntry);
      }
    },

    onDelete() {
      this.openDeleteModal = true;
    },

    deleteSolution() {
      this.openDeleteModal = false;
    }
  }
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
