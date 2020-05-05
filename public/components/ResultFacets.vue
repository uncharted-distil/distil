<template>
  <div class="result-facets">
    <div
      class="request-group-container"
      :key="request.requestId"
      v-for="request in requestGroups"
    >
      <header class="request-group-header">
        Search <sup>{{ request.requestIndex }}</sup>
      </header>

      <aside class="request-group-status">
        <template v-if="isPending(request.progress)">
          <b-badge variant="info">{{ request.progress }}</b-badge>
          <b-button
            variant="danger"
            size="sm"
            class="pull-right abort-search-button"
            @click="stopRequest(request.requestId)"
            >Stop</b-button
          >
        </template>

        <template v-if="isErrored(request.progress)">
          <b-badge variant="danger">ERROR</b-badge>
        </template>
      </aside>

      <result-group
        class="result-group-container"
        :key="group.solutionId"
        v-for="group in request.groups"
        :name="group.groupName"
        :timestamp="group.timestamp"
        :request-id="group.requestId"
        :solution-id="group.solutionId"
        :scores="group.scores"
        :target-summary="group.targetSummary"
        :predicted-summary="group.predictedSummary"
        :residuals-summary="group.residualsSummary"
        :correctness-summary="group.correctnessSummary"
      >
      </result-group>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import moment from "moment";
import ResultGroup from "../components/ResultGroup";
import { VariableSummary } from "../store/dataset/index";
import {
  SOLUTION_REQUEST_COMPLETED,
  SOLUTION_REQUEST_ERRORED,
  Solution,
  Score
} from "../store/requests/index";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as predictionsGetters } from "../store/predictions/module";
import {
  getters as requestGetters,
  actions as requestActions
} from "../store/requests/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getSolutionRequestIndex } from "../util/solutions";
import {
  getSolutionResultSummary,
  getResidualSummary,
  getCorrectnessSummary
} from "../util/summaries";

interface SummaryGroup {
  requestId: string;
  solutionId: string;
  groupName: string;
  predictedSummary: VariableSummary;
  residualsSummary: VariableSummary;
  correctnessSummary: VariableSummary;
  targetSummary: VariableSummary;
  scores: Score[];
}

interface RequestGroup {
  requestId: string;
  progress: string;
  groups: SummaryGroup[];
}

export default Vue.extend({
  name: "result-facets",

  components: {
    ResultGroup
  },

  props: {
    // display results in regression vs. classification mode
    showResiduals: {
      type: Boolean as () => boolean,
      default: () => false
    }
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    resultTargetSummary(): VariableSummary {
      return resultsGetters.getTargetSummary(this.$store);
    },

    requestGroups(): RequestGroup[] {
      const solutionRequests = requestGetters.getRelevantSolutionRequests(
        this.$store
      );
      const requestsMap = _.keyBy(solutionRequests, s => s.requestId);
      const solutions = requestGetters.getRelevantSolutions(this.$store);

      // create a summary group for each search result
      const summaryGroups: SummaryGroup[] = solutions.map(solution => {
        const solutionId = solution.solutionId;
        const requestId = solution.requestId;
        const predictedSummary = getSolutionResultSummary(solutionId);
        const residualSummary = this.showResiduals
          ? getResidualSummary(solutionId)
          : null;
        const correctnessSummary = !this.showResiduals
          ? getCorrectnessSummary(solutionId)
          : null;
        const scores = solution.scores;

        return {
          requestId: requestId,
          solutionId: solutionId,
          groupName: solution.feature,
          predictedSummary: predictedSummary,
          residualsSummary: residualSummary,
          correctnessSummary: correctnessSummary,
          targetSummary: this.resultTargetSummary,
          scores: scores
        };
      });

      // group the requests by their request ID and return them as a RequestGroup array
      const summariesByRequestId = _.groupBy(summaryGroups, s => s.requestId);
      return _.map(summariesByRequestId, (groups, requestId) => ({
        requestId: requestId,
        requestIndex: this.getRequestIndex(requestId),
        progress: requestsMap[requestId].progress,
        groups: groups
      }));
    }
  },

  methods: {
    isPending(status: string): boolean {
      return (
        status !== SOLUTION_REQUEST_COMPLETED &&
        status !== SOLUTION_REQUEST_ERRORED
      );
    },

    isCompleted(status: string): boolean {
      return status === SOLUTION_REQUEST_COMPLETED;
    },

    isErrored(status: string): boolean {
      return status === SOLUTION_REQUEST_ERRORED;
    },

    stopRequest(requestId: string) {
      requestActions.stopSolutionRequest(this.$store, {
        requestId: requestId
      });
    },

    getRequestIndex(requestId: string) {
      return getSolutionRequestIndex(requestId);
    }
  }
});
</script>

<style>
.request-group-header {
  border-bottom: 1px solid #e0e0e0;
  color: rgba(0, 0, 0, 0.87);
  font-weight: 600;
  padding: 1rem 0 0.25rem;
}

.request-group-status {
  align-items: center; /* Keep the button taller. */
  display: flex;
  margin-bottom: 0.5rem;
  margin-top: 0.5rem;
}

.request-group-status button {
  cursor: pointer;
  margin-left: auto; /* Display on the right. */
}

.result-group-container {
  overflow-x: hidden;
  overflow-y: hidden;
}
</style>
