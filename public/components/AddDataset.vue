<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

<template>
  <b-modal
    :id="id"
    title="Add Dataset"
    ok-title="Add Selected Dataset"
    :ok-disabled="disabledOK"
    @ok="handleOK"
    @show="clearForm"
    @hidden="clearForm"
  >
    <!-- Upload a file -->
    <b-form-group label="Upload a Source File (csv, zip)">
      <b-form-file
        ref="fileinput"
        v-model="file"
        :state="Boolean(file)"
        accept=".csv, .zip"
        plain
      />
    </b-form-group>

    <template v-if="isAvailableDatasets">
      <strong>OR</strong>

      <!-- Available Datasets -->
      <b-form-group label="Select an available dataset">
        <b-form-select
          v-model="availableDatasetSelected"
          :options="availableDatasets"
          :select-size="Math.min(availableDatasets.length, 10)"
        />
        <i
          slot="description"
          class="fa fa-question-circle-o"
          title="Lookup datasets available in $D3MOUTPUTDIR/PUBLIC_SUBFOLDER."
        />
      </b-form-group>

      <hr />
    </template>

    <!-- Dataset Name -->
    <b-form-group
      label="Dataset Name"
      label-for="name-input"
      invalid-feedback="Dataset name required"
      :state="nameState"
    >
      <b-form-input
        id="name-input"
        ref="nameinput"
        v-model="name"
        :state="nameState"
        required
      />
    </b-form-group>
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
import { isEmpty } from "lodash";
import { EventList } from "../util/events";

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
      name: "" as string,
      nameState: null as boolean,
      isUpload: null as boolean, // flag to know if we upload a file or import an available dataset
    };
  },

  computed: {
    // Create a unique name for the dataset.
    deconflictedName(): string {
      return generateUniqueDatasetName(this.name);
    },

    // Boolean to disable the submit button.
    disabledOK(): boolean {
      const noDataset = this.isUpload
        ? !Boolean(this.file)
        : !this.availableDatasetSelected;
      return noDataset || this.name?.length <= 0;
    },

    // Check if we have any available datasets in the public folder
    isAvailableDatasets(): boolean {
      return !isEmpty(this.availableDatasets);
    },
  },

  watch: {
    // Watches for file name changes, setting a dataset import name value
    // if the user hasn't done so.
    file() {
      if (!!this.file?.name) {
        // use the filename without the extension
        this.name = removeExtension(this.file.name);
        this.isUpload = true;
      }
    },

    availableDatasetSelected() {
      const name = this.availableDatasetSelected?.name;
      if (!name) return;
      this.name = removeExtension(name);
      this.isUpload = false;
    },

    // Watches the import data name to update the valid/invalid state.
    name() {
      // allowed transitions are: null -> true, true -> false, false -> true
      if (this.nameState === null && !!this.name) {
        this.nameState = true;
      } else if (this.nameState !== null) {
        this.nameState = !!this.name;
      }
    },
  },

  methods: {
    // Make sure everything is neat and tidy on opening.
    clearForm() {
      const $refs = this.$refs as any;
      if ($refs && $refs.fileinput) $refs.fileinput.reset();
      if ($refs && $refs.nameinput) $refs.fileinput.reset();
      this.file = null;
      this.name = "";
      this.nameState = null;
      this.availableDatasetSelected = null;
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
      const path =
        this.availableDatasetSelected.path +
        "/" +
        this.availableDatasetSelected.name;
      if (!path) return;

      // Notify external listeners that the file upload is starting
      this.$emit(EventList.UPLOAD.START_EVENT, {
        name: this.availableDatasetSelected.name,
        datasetID: this.deconflictedName,
      });

      try {
        const response = await datasetActions.importAvailableDataset(
          this.$store,
          { datasetID: this.deconflictedName, path }
        );

        this.$emit(EventList.UPLOAD.FINISHED_EVENT, null, response);
      } catch (error) {
        this.$emit(EventList.UPLOAD.FINISHED_EVENT, error, null);
      }
    },

    async uploadFile() {
      // Notify external listeners that the file upload is starting
      this.$emit(EventList.UPLOAD.START_EVENT, {
        file: this.file,
        name: this.file?.name ?? this.name,
        datasetID: this.deconflictedName,
      });

      try {
        // Upload the file and notify when complete
        const response = await datasetActions.importDataFile(this.$store, {
          datasetID: this.deconflictedName,
          file: this.file,
        });
        this.$emit(EventList.UPLOAD.FINISHED_EVENT, null, response);
      } catch (err) {
        this.$emit(EventList.UPLOAD.FINISHED_EVENT, err, null);
      }
    },
  },
});
</script>
