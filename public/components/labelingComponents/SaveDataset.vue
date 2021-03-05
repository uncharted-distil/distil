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
    <!-- Modal to save the model. -->
    <b-modal
      id="save-model-modal"
      title="Save Dataset"
      no-stacking
      @ok="handleSaveOk"
      @cancel="resetModal"
      @close="resetModal"
      @show="onShow"
    >
      <!-- show form to save model if unsaved -->
      <form ref="saveDatasetForm" @submit.stop.prevent="saveModel">
        <b-form-group
          label="Dataset Name"
          label-for="dataset-name-input"
          invalid-feedback="Dataset Name is Required"
          :state="saveNameState"
        >
          <b-form-input
            id="dataset-name-input"
            v-model="saveName"
            :state="saveNameState"
            :value="datasetName"
            required
          />
        </b-form-group>
      </form>
    </b-modal>

    <!-- Modal to offer to apply the model once saved. -->
    <b-modal
      id="save-success-dataset"
      :title-html="successTitle"
      header-class="success-modal-header"
    >
      <p>
        The dataset {{ saveName.toUpperCase() }} will now be available on the
        start page for re-use. Click <b>Go Back to Select Target Page</b> or
        <b>Go Back to Start Page</b> to work on something else.
      </p>

      <template v-slot:modal-footer>
        <b-button variant="secondary" @click="startPage()">
          Go Back to Start Page
        </b-button>
        <b-button variant="primary" @click="selectPage()">
          Go Back to Select Target Page
        </b-button>
      </template>
    </b-modal>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { SEARCH_ROUTE, SELECT_TARGET_ROUTE } from "../../store/route";
import { createRouteEntry } from "../../util/routes";
import router from "../../router/router";

export default Vue.extend({
  name: "save-dataset",

  props: {
    datasetName: String as () => string,
  },

  data() {
    return {
      saveName: "",
      saveNameState: null,
    };
  },

  computed: {
    successTitle(): string {
      return `<i class="fa fa-check-circle header-icon"/> Dataset ${this.saveName.toUpperCase()} was successfully saved`;
    },
  },

  watch: {
    // Watches the dataset save name so that the valid/invalid state can
    // be updated in response to user action.
    saveName() {
      // allowed transitions are null -> true, true -> false, false -> true
      if (this.saveNameState === null && !!this.saveName) {
        this.saveNameState = true;
      } else if (this.saveNameState !== null) {
        this.saveNameState = !!this.saveName;
      }
    },
  },

  methods: {
    // process or reject dataset save based on form state
    handleSaveOk(bvModalEvt) {
      // Prevent modal from closing
      bvModalEvt.preventDefault();
      if (!this.validForm()) {
        return;
      }
      // Trigger submit handler
      this.saveDataset();
    },

    selectPage() {
      this.resetModal();
      const routeEntry = createRouteEntry(SELECT_TARGET_ROUTE, {
        dataset: this.datasetName,
      });
      router.push(routeEntry);
    },

    // Return to the search screen.
    startPage() {
      this.resetModal();
      const routeEntry = createRouteEntry(SEARCH_ROUTE);
      router.push(routeEntry);
    },

    // clear dialog state
    resetModal() {
      this.saveName = "";
      this.saveNameState = null;
    },

    // save model
    async saveDataset() {
      this.$emit("save", this.saveName);
    },

    // ensure required fields are filled out
    validForm() {
      const valid = this.saveName.length > 0;
      this.saveNameState = valid;
      return valid;
    },
    onShow() {
      this.saveName = this.datasetName;
    },
  },
});
</script>

<style scoped>
.success-modal-header {
  background: #d5ecdb;
}

.header-icon {
  color: #35a54c;
  margin-right: 5px;
}
</style>
