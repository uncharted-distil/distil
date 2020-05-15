<template>
  <div>
    <b-button block variant="primary" v-b-modal.predictions-data-upload-modal
      >Apply Model
    </b-button>

    <!-- Modal Component -->
    <b-modal
      id="predictions-data-upload-modal"
      title="Select input data"
      @ok="handleOk()"
      @show="clearForm()"
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
    </b-modal>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { actions as datasetActions } from "../store/dataset/module";
import { actions as requestActions } from "../store/requests/module";
import { filterSummariesByDataset } from "../util/data";
import {
  getBase64,
  generateUniqueDatasetName,
  removeExtension
} from "../util/uploads";
import moment from "moment";

export default Vue.extend({
  name: "predictions-uploader",

  data() {
    return {
      file: null as File,
      importDataName: ""
    };
  },

  props: {
    fittedSolutionId: String as () => string,
    target: String as () => string,
    targetType: String as () => string
  },

  methods: {
    clearForm() {
      this.file = null;
      const $refs = this.$refs as any;
      if ($refs && $refs.fileinput) $refs.fileinput.reset();
    },
    async handleOk() {
      if (!this.file) {
        return;
      }

      const deconflictedName = generateUniqueDatasetName(
        removeExtension(this.file.name)
      );

      this.$emit("uploadstart", {
        file: this.file,
        filename: this.file.name,
        datasetID: deconflictedName
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
          targetType: this.targetType
        };
        const response = await requestActions.createPredictRequest(
          this.$store,
          requestMsg
        );
        this.$emit("uploadfinish", null, response);
      } catch (err) {
        this.$emit("uploadfinish", err, null);
      }
    }
  }
});
</script>

<style></style>
