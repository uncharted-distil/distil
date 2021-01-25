<template>
  <div class="result-facets">
    <div
      v-for="request in requestGroups"
      :key="request.requestId"
      class="request-group-container"
    >
      <header v-if="!singleSolution" class="sidebar-title">
        Search <sup>{{ request.requestIndex }}</sup>
      </header>

      <aside class="request-group-status">
        <template v-if="isErrored(request)">
          <b-badge variant="danger">ERROR</b-badge>
        </template>
        <template v-else-if="isStopRequested(request) && !isCompleted(request)">
          <b-badge variant="info">STOPPING</b-badge>
        </template>
        <template v-else-if="!isCompleted(request)">
          <b-badge variant="info">
            {{ request.progress }}
          </b-badge>
          <b-button
            variant="danger"
            size="sm"
            class="pull-right abort-search-button"
            :disabled="isStopRequested(request)"
            @click="stopRequest(request)"
          >
            Stop
          </b-button>
        </template>
      </aside>

      <result-group
        v-for="group in request.groups"
        :key="group.solutionId"
        class="result-group-container"
        :name="group.groupName"
        :timestamp="group.timestamp"
        :request-id="group.requestId"
        :solution-id="group.solutionId"
        :single-solution="singleSolution"
        :scores="group.scores"
        :target-summary="group.targetSummary"
        :predicted-summary="group.predictedSummary"
        :residuals-summary="group.residualsSummary"
        :correctness-summary="group.correctnessSummary"
        :confidence-summary="group.confidenceSummary"
      />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from "lodash";
import ResultGroup from "../components/ResultGroup.vue";
import { VariableSummary } from "../store/dataset/index";
import { SolutionRequestStatus, Score } from "../store/requests/index";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import {
  getters as requestGetters,
  actions as requestActions,
} from "../store/requests/module";
import {
  getSolutionRequestIndex,
  getSolutionById,
  getSolutionIndex,
} from "../util/solutions";
import {
  getSolutionResultSummary,
  getResidualSummary,
  getCorrectnessSummary,
  getConfidenceSummary,
} from "../util/summaries";

interface SummaryGroup {
  requestId: string;
  solutionId: string;
  groupName: string;
  predictedSummary: VariableSummary;
  residualsSummary: VariableSummary;
  correctnessSummary: VariableSummary;
  confidenceSummary: VariableSummary;
  targetSummary: VariableSummary;
  scores: Score[];
}

interface RequestGroup {
  requestId: string;
  progress: string;
  groups: SummaryGroup[];
}

export default Vue.extend({
  name: "ResultFacets",

  components: {
    ResultGroup,
  },

  props: {
    // display results in regression vs. classification mode
    showResiduals: { type: Boolean, default: false },
    singleSolution: { type: Boolean, default: false },
  },

  data() {
    return {
      stopRequested: new Set<string>(),
    };
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
      let solutionRequests = requestGetters.getRelevantSolutionRequests(
        this.$store
      );

      let solutions = [];
      if (this.singleSolution) {
        const solutionId = routeGetters.getRouteSolutionId(this.$store);
        const solution = getSolutionById(
          requestGetters.getSolutions(this.$store),
          solutionId
        );
        if (solution) {
          solutions = [solution];
          solutionRequests = [
            solutionRequests.find(
              (request) => request.requestId === solution.requestId
            ),
          ];
        }
      } else {
        // multiple solutions
        solutions = requestGetters.getRelevantSolutions(this.$store);
      }

      const requestsMap = _.keyBy(solutionRequests, (s) => s.requestId);

      // Create a summary group for each search result.
      const summaryGroups: SummaryGroup[] = solutions.map((solution) => {
        const solutionId = solution.solutionId;
        const requestId = solution.requestId;
        const predictedSummary = getSolutionResultSummary(solutionId);
        const residualSummary = this.showResiduals
          ? getResidualSummary(solutionId)
          : null;
        const correctnessSummary = !this.showResiduals
          ? getCorrectnessSummary(solutionId)
          : null;
        const confidenceSummary = !this.showResiduals
          ? getConfidenceSummary(solutionId)
          : null;
        const scores = solution.scores;

        return {
          requestId: requestId,
          solutionId: solutionId,
          groupName: solution.featureLabel,
          predictedSummary: predictedSummary,
          residualsSummary: residualSummary,
          correctnessSummary: correctnessSummary,
          confidenceSummary: confidenceSummary,
          targetSummary: this.resultTargetSummary,
          scores: scores,
        };
      });

      // Group the requests by their request ID.
      const summariesByRequestId = _.groupBy(summaryGroups, "requestId");

      // Map them as a RequestGroup array sorted by DESC requestIndex,
      // with their groups sorted by Scores DESC.
      return _.map(summariesByRequestId, (groups, requestId) => ({
        groups: groups.sort(this.sortByScoreDESC),
        progress: requestsMap[requestId]?.progress,
        requestId: requestId,
        requestIndex: this.getRequestIndex(requestId),
      })).sort(this.sortByRequestIndexDESC);
    },
  },

  methods: {
    isPending(requestGroup: RequestGroup): boolean {
      return (
        requestGroup.progress === SolutionRequestStatus.SOLUTION_REQUEST_PENDING
      );
    },

    isCompleted(requestGroup: RequestGroup): boolean {
      return (
        requestGroup.progress ===
        SolutionRequestStatus.SOLUTION_REQUEST_COMPLETED
      );
    },

    isErrored(requestGroup: RequestGroup): boolean {
      return (
        requestGroup.progress === SolutionRequestStatus.SOLUTION_REQUEST_ERRORED
      );
    },

    isStopRequested(requestGroup: RequestGroup): boolean {
      return this.stopRequested.has(requestGroup.requestId);
    },

    stopRequest(requestGroup: RequestGroup) {
      this.stopRequested.add(requestGroup.requestId);
      requestActions.stopSolutionRequest(this.$store, {
        requestId: requestGroup.requestId,
      });
    },

    getRequestIndex(requestId: string) {
      return getSolutionRequestIndex(requestId);
    },

    /* Sort SummaryGroup DESC by Scores, or by SolutionIndex. */
    sortByScoreDESC(a: SummaryGroup, b: SummaryGroup): number {
      if (_.isEmpty(a.scores) || _.isEmpty(b.scores)) {
        return 0;
      }
      const aScore = a.scores[0].value;
      const bScore = b.scores[0].value;
      if (aScore !== bScore) {
        return bScore - aScore;
      }

      return getSolutionIndex(b.solutionId) - getSolutionIndex(b.solutionId);
    },

    /* Sort RequestGroup DESC by RequestIndex. */
    sortByRequestIndexDESC(a, b) {
      return b.requestIndex - a.requestIndex;
    },
  },
});
</script>

<style scoped>
.request-group-status {
  align-items: center; /* Keep the button taller. */
  display: flex;
  margin-bottom: 0.25rem;
  margin-top: 0.25rem;
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
