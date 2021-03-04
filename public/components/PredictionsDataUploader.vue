<template>
  <b-modal
    id="predictions-data-upload-modal"
    title="Select Input Data"
    @ok="handleOk"
    @show="clearForm"
    @hide="hide"
  >
    <b-form-group label="Import dataset or choose existing dataset">
      <b-form-select v-model="inputAvenue" :options="options" />
    </b-form-group>
    <b-form-group
      v-if="isActive(options[0])"
      class="p-2 mt-4"
      label="Select a Source File (csv, zip) or dataset"
    >
      <b-form-file
        ref="fileinput"
        v-model="file"
        :state="Boolean(file)"
        accept=".csv, .zip"
        plain
      />
    </b-form-group>
    <b-form-group
      v-if="isActive(options[1])"
      class="mt-4"
      label="Choose an existing dataset"
    >
      <b-form-select v-model="selectedDataset" name="model-scoring" size="sm">
        <b-form-select-option
          v-for="dataset in datasets"
          :key="dataset.id"
          :value="dataset.id"
        >
          {{ dataset.name }}
        </b-form-select-option>
      </b-form-select>
    </b-form-group>
    <!-- <div class="mt-3">Selected file: {{ file ? file.name : "" }}</div> -->

    <template v-slot:modal-footer="{ ok, cancel }">
      <b-button :disabled="isWaiting" @click="cancel()">Cancel</b-button>

      <b-overlay
        :show="isWaiting"
        rounded
        opacity="0.6"
        spinner-small
        spinner-variant="primary"
        class="d-inline-block"
      >
        <b-button variant="primary" :disabled="!canApply" @click="ok()">
          Apply Model to Input Data
        </b-button>
      </b-overlay>
    </template>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import {
  actions as requestActions,
  getters as requestGetters,
} from "../store/requests/module";
import { actions as appActions } from "../store/app/module";
import { getters as routeGetters } from "../store/route/module";
import { Dataset } from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import { generateUniqueDatasetName, removeExtension } from "../util/uploads";
import { getPredictionsById } from "../util/predictions";
import { varModesToString, createRouteEntry } from "../util/routes";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { PREDICTION_ROUTE } from "../store/route";

export default Vue.extend({
  name: "PredictionsUploader",

  props: {
    fittedSolutionId: String as () => string,
    target: String as () => string,
    targetType: String as () => string,
  },

  data() {
    return {
      file: null as File,
      uploadData: {},
      uploadStatus: "",
      isWaiting: false,
      selectedDataset: "",
      inputAvenue: null,
      options: ["Import Dataset", "Select Existing Dataset"],
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    datasets(): Dataset[] {
      return datasetGetters.getDatasets(this.$store);
    },
    canApply(): boolean {
      return (
        !this.isWaiting &&
        ((Boolean(this.file) && this.isActive(this.options[0])) ||
          (this.selectedDataset !== "" && this.isActive(this.options[1])))
      );
    },
  },

  methods: {
    hide() {
      this.selectedDataset = "";
      this.inputAvenue = null;
      this.file = null as File;
    },
    isActive(option: string): boolean {
      return this.inputAvenue === option;
    },
    clearForm() {
      this.file = null;
      const $refs = this.$refs as any;
      if ($refs && $refs.fileinput) $refs.fileinput.reset();
    },

    handleOk(bvModalEvt) {
      // Prevent modal from closing
      bvModalEvt.preventDefault();

      if (!this.canApply) {
        return;
      }

      this.isWaiting = true;
      if (this.selectedDataset !== "" && this.isActive(this.options[1])) {
        this.makePrediction(true, this.selectedDataset, "");
      } else {
        this.makeRequest();
      }
    },

    async makeRequest() {
      var deconflictedName = generateUniqueDatasetName(
        removeExtension(this.file.name)
      );

      this.uploadStart({
        file: this.file,
        filename: this.file.name,
        datasetID: deconflictedName,
      });

      // Apply model to a new prediction set.  The selected file's contents will be uploaded and
      // ingeested.  The request then applies to the uploaded file.
      try {
        // upload and ingest the dataset
        const uploadResponse = await datasetActions.uploadDataFile(
          this.$store,
          {
            file: this.file,
            datasetID: deconflictedName,
          }
        );

        this.makePrediction(
          false,
          deconflictedName,
          uploadResponse.data.location
        );
      } catch (err) {
        this.predictionFinish(err, null);
      }
    },

    async makePrediction(existing, dataset, datasetPath) {
      // Apply model to a prediction dataset.
      try {
        const requestMsg = {
          datasetId: dataset,
          fittedSolutionId: this.fittedSolutionId,
          target: this.target,
          targetType: this.targetType,
          datasetPath: datasetPath,
          existingDataset: existing,
        };
        const predictResponse = await requestActions.createPredictRequest(
          this.$store,
          requestMsg
        );

        this.predictionFinish(null, predictResponse);
      } catch (err) {
        this.predictionFinish(err, null);
      }
    },

    uploadStart(uploadData) {
      this.uploadData = uploadData;
      this.uploadStatus = "started";
      appActions.logUserEvent(this.$store, {
        feature: Feature.EXPORT_MODEL,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.IMPORT_INFERENCE,
        details: {
          activeSolution: this.fittedSolutionId,
        },
      });
    },

    predictionFinish(err: Error, response: any) {
      this.isWaiting = false;
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
          predictionsDataset: predictionDataset,
          dataset: this.dataset,
          varModes: varModes,
          applyModel: true.toString(),
          solutionId: routeGetters.getRouteSolutionId(this.$store),
        };
        const entry = createRouteEntry(PREDICTION_ROUTE, routeArgs);
        this.$router.push(entry).catch((err) => console.warn(err));
      }
    },
  },
});
</script>
