<template>
  <div class="result-panel">
    <div class="result-summaries">
      <p class="nav-link font-weight-bold">Results</p>
      <p></p>
      <div v-if="showResiduals" class="result-summaries-error">
        <error-threshold-slider></error-threshold-slider>
      </div>
      <p class="nav-link font-weight-bold">
        Predictions by Model
      </p>
      <result-facets :showResiduals="showResiduals" />
    </div>
    <template v-if="isActiveSolutionCompleted">
      <save-model
        :solutionId="solutionId"
        :fittedSolutionId="fittedSolutionId"
      ></save-model>
      <hr />
      <b-button
        variant="success"
        class="save-button"
        v-b-modal.save-model-modal
      >
        <i class="fa fa-floppy-o"></i>
        Save Model
      </b-button>
    </template>
  </div>
</template>

<script lang="ts">
import ResultFacets from "../components/ResultFacets";
import PredictionsDataUploader from "../components/PredictionsDataUploader";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider";
import SaveModel from "../components/SaveModel";
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
    SaveModel
  },

  data() {
    return {
      formatter(arg) {
        return arg ? arg.toFixed(2) : "";
      },
      symmetricSlider: true,
      file: null,
      uploadData: {},
      uploadStatus: ""
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
    }
  },

  methods: {
    onUploadStart(uploadData) {
      this.uploadData = uploadData;
      this.uploadStatus = "started";
      appActions.logUserEvent(this.$store, {
        feature: Feature.EXPORT_MODEL,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.IMPORT_INFERENCE,
        details: {
          activeSolution: this.activeSolution.solutionId
        }
      });
    },

    onUploadFinish(err: Error, response: any) {
      this.uploadStatus = err ? "error" : "success";

      if (this.uploadStatus !== "error" && !response.complete) {
        const predictionDataset = getPredictionsById(
          requestGetters.getPredictions(this.$store),
          response.produceRequestId
        ).dataset;

        const varModes = varModesToString(
          routeGetters.getDecodedVarModes(this.$store)
        );

        const routeArgs = {
          fittedSolutionId: this.fittedSolutionId,
          produceRequestId: response.produceRequestId,
          target: this.target,
          predictionDataset: predictionDataset,
          dataset: this.dataset,
          varModes: varModes
        };
        const entry = createRouteEntry(PREDICTION_ROUTE, routeArgs);
        this.$router.push(entry);
      }
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
  margin-left: auto;
  margin-right: 8px;
}

.result-panel {
  display: flex;
  flex-direction: column;
}
</style>
