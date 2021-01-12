<template>
  <div class="result-panel">
    <h6 class="sidebar-title">Ground Truth</h6>
    <result-target-variable class="result-target-variable" />

    <!-- Searchs results -->
    <h6 class="sidebar-title">Results</h6>
    <section class="result-options">
      <error-threshold-slider v-if="showResiduals && !isTimeseries" />
    </section>
    <section class="result-summaries">
      <result-facets
        :show-residuals="showResiduals"
        :single-solution="isSingleSolution"
      />
    </section>

    <!-- Action buttons -->
    <aside
      v-if="isActiveSolutionCompleted"
      class="d-flex flex-row flex-shrink-0 justify-content-end"
    >
      <!-- Modal boxes to apply new data to models. -->
      <forecast-horizon
        v-if="isTimeseries"
        :dataset="dataset"
        :fitted-solution-id="fittedSolutionId"
        :target="target"
        :target-type="targetType"
      />
      <predictions-data-uploader
        v-else
        :fitted-solution-id="fittedSolutionId"
        :target="target"
        :target-type="targetType"
      />
      <template v-if="isSingleSolution || isActiveSolutionSaved">
        <b-button
          v-if="isTimeseries"
          variant="primary"
          class="apply-button"
          @click="$bvModal.show('forecast-horizon-modal')"
        >
          Forecast
        </b-button>
        <b-button
          v-else
          variant="primary"
          class="apply-button"
          @click="$bvModal.show('predictions-data-upload-modal')"
        >
          Apply Model
        </b-button>
      </template>
      <template v-else>
        <save-model
          :solution-id="solutionId"
          :fitted-solution-id="fittedSolutionId"
        />
        <b-button
          variant="success"
          class="save-button"
          @click="$bvModal.show('save-model-modal')"
        >
          <i class="fa fa-floppy-o" />
          Save Model
        </b-button>
      </template>
    </aside>
  </div>
</template>

<script lang="ts">
import ResultFacets from "../components/ResultFacets.vue";
import PredictionsDataUploader from "../components/PredictionsDataUploader.vue";
import ForecastHorizon from "../components/ForecastHorizon.vue";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider.vue";
import SaveModel from "../components/SaveModel.vue";
import ResultTargetVariable from "../components/ResultTargetVariable.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { Variable, TaskTypes } from "../store/dataset/index";
import Vue from "vue";
import { Solution, SolutionStatus } from "../store/requests/index";
import { isFittedSolutionIdSavedAsModel } from "../util/models";

export default Vue.extend({
  name: "ResultSummaries",

  components: {
    ErrorThresholdSlider,
    ForecastHorizon,
    PredictionsDataUploader,
    ResultFacets,
    ResultTargetVariable,
    SaveModel,
  },

  data() {
    return {
      formatter(arg) {
        return arg ? arg.toFixed(2) : "";
      },
      symmetricSlider: true,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    targetType(): string {
      const targetName = this.target;
      const variables = this.variables;
      return variables.find((v) => v.key === targetName)?.colType;
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    taskArgs(): string[] {
      return routeGetters.getRouteTask(this.$store).split(",");
    },

    showResiduals(): boolean {
      return (
        this.taskArgs &&
        !!this.taskArgs.find(
          (t) => t === TaskTypes.REGRESSION || t === TaskTypes.FORECASTING
        )
      );
    },

    solutionId(): string {
      return requestGetters.getActiveSolution(this.$store)?.solutionId;
    },

    fittedSolutionId(): string {
      return requestGetters.getActiveSolution(this.$store)?.fittedSolutionId;
    },

    activeSolution(): Solution {
      return requestGetters.getActiveSolution(this.$store);
    },

    activeSolutionName(): string {
      return this.activeSolution ? this.activeSolution.feature : "";
    },

    instanceName(): string {
      return "groundTruth";
    },

    /**
     * Check that the active solution is completed.
     * This is used to display possible actions on the selected model.
     */
    isActiveSolutionCompleted(): boolean {
      return !!(
        this.activeSolution &&
        this.activeSolution.progress === SolutionStatus.SOLUTION_COMPLETED
      );
    },

    /**
     * Check that the active solution is saved as a model.
     * This is used to display possible actions on the selected model.
     */
    isActiveSolutionSaved(): boolean {
      return this.isFittedSolutionIdSavedAsModel(this.fittedSolutionId);
    },

    // Indicates whether or not the contained result facets should show "relevant"
    // results, which consist of those that match the target/dataset, or a single
    // result, which matches the route solutionID.  The latter case occurs when the
    // user selects a model directly from the search screen.
    isSingleSolution(): boolean {
      return routeGetters.isSingleSolution(this.$store);
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
  },

  methods: {
    isFittedSolutionIdSavedAsModel,
  },
});
</script>

<style>
.result-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.result-options {
  flex-shrink: 0;
}

.result-summaries {
  overflow-x: hidden;
  overflow-y: auto;
}

.result-summaries .facets-facet-base {
  overflow: visible;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
  box-shadow: inset 0 0 0 1000px var(--blue);
}

.check-message-container {
  display: flex;
  justify-content: flex-start;
  flex-direction: row;
  align-items: center;
}

.check-icon {
  display: flex;
  flex-shrink: 0;
  color: #00c851;
  padding-right: 15px;
}

.fail-icon {
  display: flex;
  flex-shrink: 0;
  color: #ee0701;
  padding-right: 15px;
}

.check-button {
  width: 60%;
  margin: 0 20%;
}

.save-button {
  flex-shrink: 0;
  flex-grow: 0;
  margin-top: 15px;
  margin-bottom: 0px;
  margin-right: 8px;
}

.apply-button {
  flex-shrink: 0;
  flex-grow: 0;
  margin-top: 15px;
  margin-bottom: 0px;
  margin-right: 8px;
}

.result-target-variable .variable-facets-item {
  margin-top: 0px;
  padding-top: 0px;
}
</style>
