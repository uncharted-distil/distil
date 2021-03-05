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
  <div>
    <b-button block variant="primary" v-b-modal.upload-modal>
      <i class="fa fa-plus-circle"></i> Import File
    </b-button>
    <b-modal
      id="upload-modal"
      title="Import Local File"
      :ok-disabled="!Boolean(file) || this.importDataName.length <= 0"
      ok-title="Upload Local File"
      @ok="handleOk()"
      @show="clearForm()"
    >
      <b-form-group label="Source File (csv, zip)">
        <b-form-file
          ref="fileinput"
          v-model="file"
          :state="Boolean(file)"
          accept=".csv, .zip"
          plain
        />
      </b-form-group>

      <b-form-group
        label="Dataset Name"
        label-for="import-name-input"
        invalid-feedback="Dataset name required"
        :state="importDataNameState"
      >
        <b-form-input
          ref="importnameinput"
          id="import-name-input"
          v-model="importDataName"
          :state="importDataNameState"
          required
        />
      </b-form-group>
    </b-modal>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { actions as datasetActions } from "../store/dataset/module";
import { removeExtension, generateUniqueDatasetName } from "../util/uploads";

export default Vue.extend({
  name: "file-uploader",

  props: {
    target: String as () => string,
    targetType: String as () => string,
  },

  data() {
    return {
      file: null as File,
      importDataName: "",
      importDataNameState: null as boolean,
    };
  },

  watch: {
    // Watches for file name changes, setting a dataset import name value
    // if the user hasn't done so.
    file() {
      if (!this.importDataName && this.file?.name) {
        // use the filname without the extension
        this.importDataName = removeExtension(this.file.name);
        this.importDataNameState = true;
      }
    },

    // Watches the import data name to update the valid/invalid state.
    importDataName() {
      // allowed transitions are: null -> true, true -> false, false -> true
      if (this.importDataNameState === null && !!this.importDataName) {
        this.importDataNameState = true;
      } else if (this.importDataNameState !== null) {
        this.importDataNameState = !!this.importDataName;
      }
    },
  },

  methods: {
    clearForm() {
      const $refs = this.$refs as any;
      if ($refs && $refs.fileinput) $refs.fileinput.reset();
      if ($refs && $refs.importnameinput) $refs.fileinput.reset();
      this.file = null;
      this.importDataName = "";
      this.importDataNameState = null;
    },

    async handleOk() {
      const deconflictedName = generateUniqueDatasetName(this.importDataName);

      // Notify external listeners that the file upload is starting
      this.$emit("uploadstart", {
        file: this.file,
        filename: this.file.name,
        datasetID: deconflictedName,
      });

      try {
        // Upload and import the data file and notify when complete
        const response = await datasetActions.importDataFile(this.$store, {
          datasetID: deconflictedName,
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
``
