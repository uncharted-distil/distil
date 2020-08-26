<template>
  <b-modal
    id="predictions-data-upload-modal"
    title="Select input data"
    @ok="handleOk"
    @show="clearForm"
  >
    <p>Select a csv or zip file to import</p>
    <b-form-file
      ref="fileinput"
      v-model="file"
      :state="Boolean(file)"
      accept=".csv, .zip"
      plain
    />
    <div class="mt-3">Selected file: {{ file ? file.name : "" }}</div>

    <template v-slot:modal-footer="{ ok, cancel }">
      <b-button @click="cancel()" :disabled="isWaiting">Cancel</b-button>

      <b-overlay
        :show="isWaiting"
        rounded
        opacity="0.6"
        spinner-small
        spinner-variant="primary"
        class="d-inline-block"
      >
        <b-button
          variant="primary"
          @click="ok()"
          :disabled="isWaiting || !Boolean(file)"
        >
          Apply Model
        </b-button>
      </b-overlay>
    </template>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { actions as datasetActions } from "../store/dataset/module";
import {
  actions as requestActions,
  getters as requestGetters,
} from "../store/requests/module";
import { actions as appActions } from "../store/app/module";
import { getters as routeGetters } from "../store/route/module";
import { filterSummariesByDataset } from "../util/data";
import {
  getBase64,
  generateUniqueDatasetName,
  removeExtension,
} from "../util/uploads";
import moment from "moment";
import { getPredictionsById } from "../util/predictions";
import { varModesToString, createRouteEntry } from "../util/routes";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { PREDICTION_ROUTE } from "../store/route";

export default Vue.extend({
  name: "predictions-uploader",

  data() {
    return {
      file: null as File,
      importDataName: "",
      uploadData: {},
      uploadStatus: "",
      isWaiting: false,
    };
  },

  props: {
    fittedSolutionId: String as () => string,
    target: String as () => string,
    targetType: String as () => string,
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
  },

  methods: {
    clearForm() {
      this.file = null;
      const $refs = this.$refs as any;
      if ($refs && $refs.fileinput) $refs.fileinput.reset();
    },

    handleOk(bvModalEvt) {
      // Prevent modal from closing
      bvModalEvt.preventDefault();

      if (!this.file) {
        return;
      }

      this.isWaiting = true;
      this.makeRequest();
    },

    async makeRequest() {
      const deconflictedName = generateUniqueDatasetName(
        removeExtension(this.file.name)
      );

      this.uploadStart({
        file: this.file,
        filename: this.file.name,
        datasetID: deconflictedName,
      });

      // Apply model to a new prediction set.  The selected file's contents will be uploaded and
      // fed into a fitted solution.  The prediction request goes through a websocket similar to
      try {
        const dataset = await getBase64(this.file);
        const requestMsg = {
          datasetId: deconflictedName,
          dataset: dataset,
          fittedSolutionId: this.fittedSolutionId,
          target: this.target,
          targetType: this.targetType,
        };
        const response = await requestActions.createPredictRequest(
          this.$store,
          requestMsg
        );
        this.uploadFinish(null, response);
      } catch (err) {
        this.uploadFinish(err, null);
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

    uploadFinish(err: Error, response: any) {
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
          predictionDataset: predictionDataset,
          dataset: this.dataset,
          varModes: varModes,
          applyModel: true.toString(),
        };
        const entry = createRouteEntry(PREDICTION_ROUTE, routeArgs);
        this.$router.push(entry);
      }
    },
  },
});
</script>
