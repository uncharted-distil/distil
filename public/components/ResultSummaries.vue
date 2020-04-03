<template>
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

    <file-uploader
      class="result-button-alignment"
      @uploadstart="onUploadStart"
      @uploadfinish="onUploadFinish"
      :upload-type="uploadType"
      :fitted-solution-id="fittedSolutionId"
      :target="target"
      :target-type="targetType"
    ></file-uploader>
    <b-button
      block
      variant="primary"
      class="result-button-alignment"
      v-on:click="saveModel"
    >
      Save Model
    </b-button>
  </div>
</template>

<script lang="ts">
import ResultFacets from "../components/ResultFacets";
import FileUploader from "../components/FileUploader";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider";
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
import { Solution } from "../store/requests/index";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { createRouteEntry } from "../util/routes";
import { PREDICTION_UPLOAD } from "../util/uploads";

export default Vue.extend({
  name: "result-summaries",

  components: {
    ResultFacets,
    FileUploader,
    ErrorThresholdSlider,
    vueSlider
  },

  data() {
    return {
      formatter(arg) {
        return arg ? arg.toFixed(2) : "";
      },
      symmetricSlider: true,
      file: null,
      uploadData: {},
      uploadStatus: "",
      uploadType: PREDICTION_UPLOAD
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

    taskArgs(): string {
      return routeGetters.getRouteTask(this.$store);
    },

    showResiduals(): boolean {
      return this.taskArgs && this.taskArgs.includes(TaskTypes.REGRESSION);
    },

    solutionId(): string {
      return requestGetters.getActiveSolution(this.$store)?.solutionId;
    },

    fittedSolutionId(): string {
      return requestGetters.getActiveSolution(this.$store)?.fittedSolutionId;
    },

    activeSolution(): Solution {
      return requestGetters.getActiveSolution(this.$store)
    },

    activeSolutionName(): string {
      return this.activeSolution ? this.activeSolution.feature : "";
    },

    instanceName(): string {
      return "groundTruth";
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
        const routeArgs = {
          fittedSolutionId: this.fittedSolutionId,
          produceRequestId: response.produceRequestId,
          inferenceDataset: response.dataset,
          target: this.target
        };
        const entry = createRouteEntry(PREDICTION_ROUTE, routeArgs);
        this.$router.push(entry);
      }
    },
    saveModel () {
      appActions.logUserEvent(this.$store, {
        feature: Feature.EXPORT_MODEL,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_SAVE,
        details: {
          solution: this.activeSolution.solutionId,
          fittedSolution: this.fittedSolutionId
        }
      });
      appActions
        .saveModel(this.$store, {
          solutionId: this.activeSolution.solutionId,
          fittedSolutionId: this.fittedSolutionId
        })
        .then(err => {
          if (err) {
            console.log(err);
          }
        });
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

.result-button-alignment {
  padding: 0 0 15px;
}
</style>
