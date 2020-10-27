<template>
  <b-modal
    :id="id"
    title="Add Dataset"
    ok-title="Add Selected Dataset"
    @ok="handleOK"
    @show="clearForm"
  >
    <b-button variant="primary" v-b-modal.upload-modal>
      <i class="fa fa-upload" /> Upload a file
    </b-button>
    <hr />
    <strong>Select an available dataset</strong>
    <i
      class="fa fa-question-circle-o"
      title="Lookup datasets already available in $D3MOUTPUTDIR/augmented."
    />
    <b-form-select
      v-model="availableDatasetSelected"
      :options="availableDatasets"
      :select-size="Math.min(availableDatasets.length, 10)"
    />
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import {
  getAvailableDatasets,
  removeExtension,
  generateUniqueDatasetName,
} from "../util/uploads";
import { actions as datasetActions } from "../store/dataset/module";

export default Vue.extend({
  name: "AddDataset",

  props: {
    id: { type: String as () => string, required: true },
  },

  data() {
    return {
      availableDatasets: [],
      availableDatasetSelected: null,
      file: null as File,
      importName: null as string,
      importNameState: null as boolean,
      isUpload: null as boolean, // flag to know if we upload a file or import an available dataset
    };
  },

  computed: {
    // Create a unique name for the dataset.
    deconflictedName(): string {
      return generateUniqueDatasetName(this.importName);
    },
  },

  watch: {
    // Watches for file name changes, setting a dataset import name value
    // if the user hasn't done so.
    file() {
      if (!this.importName && this.file?.name) {
        // use the filname without the extension
        this.importName = removeExtension(this.file.name);
        this.importNameState = true;
      }
    },

    // Watches the import data name to update the valid/invalid state.
    importDataName() {
      // allowed transitions are: null -> true, true -> false, false -> true
      if (this.importNameState === null && !!this.importName) {
        this.importNameState = true;
      } else if (this.importNameState !== null) {
        this.importNameState = !!this.importName;
      }
    },
  },

  beforeMount() {
    this.getAvailableDatasets();
  },

  methods: {
    // Make sure everything is neat and tidy on opening.
    clearForm() {
      const $refs = this.$refs as any;
      if ($refs && $refs.fileinput) $refs.fileinput.reset();
      if ($refs && $refs.importnameinput) $refs.fileinput.reset();
      this.file = null;
      this.importName = "";
      this.importNameState = null;
      this.getAvailableDatasets();
    },

    // Fetch the list of available dataset in the $D3MOUTPUTDIR/augmented folder.
    async getAvailableDatasets() {
      this.availableDatasets = [];
      try {
        const response = await getAvailableDatasets();

        // Format the response to fit the <b-form-select> options format {value, text}
        this.availableDatasets = response.map((dataset) => {
          return {
            value: dataset,
            text: dataset.name,
          };
        });
      } catch (error) {
        console.error("Error fetching available datasets.");
      }
    },

    async handleOK() {
      if (this.isUpload) {
        this.uploadFile();
      } else {
        this.importAvailableDataset();
      }
    },

    async importAvailableDataset() {
      // Make sure that the arguments are sound.
      const { name, path } = this.availableDatasetSelected;
      const location = path + "/" + name;
      if (!location) return;

      try {
        const response = await datasetActions.importAvailableDataset(
          this.$store,
          { datasetID: this.deconflictedName, path: location }
        );

        console.debug(response);
      } catch (error) {
        this.$emit("", error, null);
      }
    },

    async uploadFile() {
      // Notify external listeners that the file upload is starting
      this.$emit("uploadstart", {
        file: this.file,
        filename: this.file.name,
        datasetID: this.deconflictedName,
      });

      try {
        // Upload the file and notify when complete
        const response = await datasetActions.uploadDataFile(this.$store, {
          datasetID: this.deconflictedName,
          file: this.file,
        });
        this.$emit("uploadfinish", null, response);
      } catch (err) {
        this.$emit("uploadfinish", err, null);
      }
    },
  },
});
</script>
