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
      v-b-modal.save-model-modal
    >
      Save Model
    </b-button>
    <b-modal
      title="Save Model"
      id="save-model-modal"
      @ok="handleOk"
      @cancel="resetModal"
      @close="resetModal"
    >
      <form ref="saveModelForm" @submit.stop.prevent="saveModel">
        <b-form-group
          label="Model Name"
          label-for="model-name-input"
          invalid-feedback="Model Name is Required"
          :state="saveNameState"
        >
          <b-form-input
            id="model-name-input"
            v-model="saveName"
            :state="saveNameState"
            required
          ></b-form-input>
        </b-form-group>
        <b-form-group
          label="Model Description"
          label-for="model-desc-input"
          :state="saveDescriptionState"
        >
          <b-form-input
            id="model-desc-input"
            v-model="saveDescription"
            :state="saveDescriptionState"
          ></b-form-input>
        </b-form-group>
      </form>
    </b-modal>
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
import { createRouteEntry, varModesToString } from "../util/routes";
import { PREDICTION_UPLOAD } from "../util/uploads";
import { getPredictionsById } from "../util/predictions";

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
      uploadType: PREDICTION_UPLOAD,
      saveName: "",
      saveNameState: null,
      saveDescription: "",
      saveDescriptionState: null
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
      return this.taskArgs &&
        !!this.taskArgs.find
          (t => t ===  TaskTypes.REGRESSION || t === TaskTypes.FORECASTING);
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
    validForm() {
      const valid = this.saveName.length > 0
      this.saveNameState = valid;
      return valid
    },
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

        const varModes = varModesToString(routeGetters.getDecodedVarModes(this.$store));

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
    },
    handleOk(bvModalEvt) {
      // Prevent modal from closing
      bvModalEvt.preventDefault();
      // Trigger submit handler
      this.saveModel();
    },
    resetModal() {
      this.saveName = '';
      this.saveNameState = null;
      this.saveDescription = '';
      this.saveDescriptionState = null;
    },
    saveModel () {
      if (!this.validForm()) {
        return;
      }

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
          fittedSolutionId: this.fittedSolutionId,
          modelName: this.saveName,
          modelDescription: this.saveDescription
        })
        .then(err => {
          if (err) {
            console.warn(err);
          }
        });

      this.$nextTick(() => {
        this.$bvModal.hide('save-model-modal');
        this.resetModal();
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
