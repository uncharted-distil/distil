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
    <b-modal :id="modalId" :title="title" @close="resetModal">
      <!-- show form to save model if unsaved -->
      <form ref="saveModelForm" @submit.stop.prevent="saveModel">
        <b-form-group
          :label="subjectNameLabel"
          label-for="model-name-input"
          :invalid-feedback="invalidFeedback"
          :state="saveNameState"
        >
          <b-form-input
            id="model-name-input"
            v-model="saveName"
            :state="saveNameState"
            required
          />
        </b-form-group>
        <b-form-group
          :label="subjectDescriptionLabel"
          label-for="model-desc-input"
          :state="saveDescriptionState"
        >
          <b-form-input
            id="model-desc-input"
            v-model="saveDescription"
            :state="saveDescriptionState"
          />
        </b-form-group>
      </form>
      <template v-slot:modal-footer>
        <b-button variant="secondary" @click="resetModal"> Cancel </b-button>
        <b-button variant="primary" @click="handleSaveOk" :disabled="isSaving">
          <b-spinner v-if="isSaving" small />
          <span v-else>OK</span>
        </b-button>
      </template>
    </b-modal>

    <!-- Modal to offer to apply the model once saved. -->
    <b-modal
      id="save-success-modal"
      :title-html="successTitle"
      header-class="success-modal-header"
    >
      <p>
        The {{ subject }} {{ saveName.toUpperCase() }} will now be available on
        the start page for re-use. To use it now on new data, click
        <b>{{ actionName }}</b> or <b>Go Back to Start Page</b> to work on
        something else.
      </p>

      <template v-slot:modal-footer>
        <b-button variant="secondary" @click="back()">
          Go Back to Start Page
        </b-button>
        <b-button variant="primary" @click="apply()">{{ actionName }}</b-button>
      </template>
    </b-modal>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { SEARCH_ROUTE } from "../store/route";
import { getters as routeGetters } from "../store/route/module";
import { createRouteEntry } from "../util/routes";
import router from "../router/router";
import { EventList, EI } from "../util/events";

export default Vue.extend({
  name: "SaveModal",

  props: {
    solutionId: { type: String as () => string, default: "" },
    fittedSolutionId: { type: String as () => string, default: "" },
    subject: { type: String as () => string, default: "Model" },
    modalId: { type: String as () => string, default: "save-model-modal" },
  },

  data() {
    return {
      saveName: "",
      saveNameState: null,
      saveDescription: "",
      saveDescriptionState: null,
      isSaving: false,
    };
  },

  computed: {
    actionName(): string {
      return this.isTimeseries ? "Forecast" : "Apply Model";
    },

    successTitle(): string {
      return `<i class="fa fa-check-circle header-icon"/> ${
        this.subject
      } ${this.saveName.toUpperCase()} was successfully saved`;
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
    title(): string {
      return `Save ${this.subject}`;
    },
    subjectDescriptionLabel(): string {
      return `${this.subject} Description`;
    },
    subjectNameLabel(): string {
      return `${this.subject} Name`;
    },
    invalidFeedback(): string {
      return `${this.subject} Name is Required`;
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
    // process or reject model save based on form state
    handleSaveOk(bvModalEvt) {
      // Prevent modal from closing
      bvModalEvt.preventDefault();

      // Trigger submit handler
      this.saveModel();
      this.isSaving = true;
    },

    // CDB: Currently will open up the file upload dialog. Should transition to the
    // apply model workflow.
    apply() {
      this.resetModal();

      if (this.isTimeseries) {
        this.$bvModal.show("forecast-horizon-modal");
      } else {
        this.$bvModal.show("predictions-data-upload-modal");
      }
    },

    // Return to the search screen.
    back() {
      this.resetModal();
      const routeEntry = createRouteEntry(SEARCH_ROUTE);
      router.push(routeEntry);
    },

    // clear dialog state
    resetModal() {
      this.saveName = "";
      this.saveNameState = null;
      this.saveDescription = "";
      this.saveDescriptionState = null;
    },

    // save model
    async saveModel() {
      if (!this.validForm()) {
        return;
      }
      this.$emit(EventList.MODEL.SAVE_EVENT, {
        solutionId: this.solutionId,
        fittedSolution: this.fittedSolutionId,
        name: this.saveName,
        description: this.saveDescription,
      } as EI.RESULT.SaveInfo);
    },
    showSuccessModal() {
      this.isSaving = false;
      this.$bvModal.show("save-success-modal");
    },
    hideSuccessModal() {
      this.$bvModal.hide("save-success-modal");
    },
    hideSaveForm() {
      this.$bvModal.hide(this.modalId);
    },
    // ensure required fields are filled out
    validForm() {
      const valid = this.saveName.length > 0;
      this.saveNameState = valid;
      return valid;
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
