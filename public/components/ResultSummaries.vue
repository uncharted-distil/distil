<template>
  <div class="result-panel">
    <p class="nav-link font-weight-bold flex-shrink-0">Ground Truth</p>
    <result-target-variable
      class="result-target-variable"
    ></result-target-variable>

    <p class="nav-link font-weight-bold flex-shrink-0">Results</p>
    <div class="result-summaries">
      <div v-if="showResiduals" class="result-summaries-error">
        <error-threshold-slider></error-threshold-slider>
      </div>
      <result-facets
        :showResiduals="showResiduals"
        :single-solution="isSingleSolution"
      />
    </div>
    <template v-if="isActiveSolutionCompleted">
      <div class="d-flex flex-row flex-shrink-0 justify-content-end">
        <predictions-data-uploader
          class="result-button-alignment"
          :fitted-solution-id="fittedSolutionId"
          :target="target"
          :target-type="targetType"
        ></predictions-data-uploader>
        <b-button
          v-if="isSingleSolution"
          variant="primary"
          class="apply-button"
          v-b-modal.predictions-data-upload-modal
          >Apply Model
        </b-button>
        <save-model
          :solutionId="solutionId"
          :fittedSolutionId="fittedSolutionId"
        ></save-model>
        <b-button
          v-if="!isSingleSolution"
          variant="success"
          class="save-button"
          v-b-modal.save-model-modal
        >
          <i class="fa fa-floppy-o"></i>
          Save Model
        </b-button>
      </div>
    </template>
  </div>
</template>

<script lang="ts">
import ResultFacets from "../components/ResultFacets";
import PredictionsDataUploader from "../components/PredictionsDataUploader";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider";
import SaveModel from "../components/SaveModel";
import ResultTargetVariable from "../components/ResultTargetVariable";
import { getSolutionById } from "../util/solutions";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as resultGetters } from "../store/results/module";
import {
  actions as appActions,
  getters as appGetters
} from "../store/app/module";
import store from "../store/store";
import {
  EXPORT_SUCCESS_ROUTE,
  ROOT_ROUTE,
  PREDICTION_ROUTE
} from "../store/route/index";
import { Variable, TaskTypes } from "../store/dataset/index";
import vueSlider from "vue-slider-component";
import Vue from "vue";
import { Solution, SOLUTION_COMPLETED } from "../store/requests/index";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { createRouteEntry, varModesToString } from "../util/routes";
import { getPredictionsById } from "../util/predictions";

export default Vue.extend({
  name: "result-summaries",

  components: {
    ResultFacets,
    PredictionsDataUploader,
    ErrorThresholdSlider,
    vueSlider,
    SaveModel,
    ResultTargetVariable
  },

  data() {
    return {
      formatter(arg) {
        return arg ? arg.toFixed(2) : "";
      },
      symmetricSlider: true
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
      return variables.find(v => v.colName === targetName)?.colType;
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
          t => t === TaskTypes.REGRESSION || t === TaskTypes.FORECASTING
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
     * @returns {Boolean}
     */
    isActiveSolutionCompleted(): boolean {
      return !!(
        this.activeSolution &&
        this.activeSolution.progress === SOLUTION_COMPLETED
      );
    },

    // Indicates whether or not the contained result facets should show "relevant"
    // results, which consist of those that match the target/dataset, or a single
    // result, which matches the route solutionID.  The latter case occurs when the
    // user selects a model directly from the search screen.
    isSingleSolution(): boolean {
      return routeGetters.isSingleSolution(this.$store);
    }
  }
});
</script>

<style>
.result-summaries {
  overflow-x: hidden;
  overflow-y: auto;
}

.result-summaries .facets-facet-base {
  overflow: visible;
}

.result-summaries-error {
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  margin-bottom: 30px;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
  box-shadow: inset 0 0 0 1000px #007bff;
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

.result-panel {
  display: flex;
  flex-direction: column;
}

.result-target-variable .variable-facets-item {
  margin-top: 0px;
  padding-top: 0px;
}
</style>
