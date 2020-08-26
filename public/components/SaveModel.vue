<template>
  <div>
    <!-- Modal to save the model. -->
    <b-modal
      title="Save Model"
      id="save-model-modal"
      no-stacking
      @ok="handleSaveOk"
      @cancel="resetModal"
      @close="resetModal"
    >
      <!-- show form to save model if unsaved -->
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

    <!-- Modal to offer to apply the model once saved. -->
    <b-modal
      id="save-success-modal"
      :title-html="successTitle"
      header-class="success-modal-header"
    >
      <p>
        The model {{ saveName.toUpperCase() }} will now be available on the
        start page for re-use. To use it now on new data, click
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
import { actions as appActions } from "../store/app/module";
import { SEARCH_ROUTE } from "../store/route";
import { getters as routeGetters } from "../store/route/module";
import { createRouteEntry } from "../util/routes";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import router from "../router/router";

export default Vue.extend({
  name: "save-model",

  props: {
    solutionId: String as () => string,
    fittedSolutionId: String as () => string,
  },

  data() {
    return {
      saveName: "",
      saveNameState: null,
      saveDescription: "",
      saveDescriptionState: null,
    };
  },

  computed: {
    actionName(): string {
      return this.isTimeseries ? "Forecast" : "Apply Model";
    },

    successTitle(): string {
      return `<i class="fa fa-check-circle header-icon"></i> Model ${this.saveName.toUpperCase()} was successfully saved`;
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
  },

  methods: {
    // process or reject model save based on form state
    handleSaveOk(bvModalEvt) {
      // Prevent modal from closing
      bvModalEvt.preventDefault();

      // Trigger submit handler
      this.saveModel();
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

      appActions.logUserEvent(this.$store, {
        feature: Feature.EXPORT_MODEL,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_SAVE,
        details: {
          solution: this.solutionId,
          fittedSolution: this.fittedSolutionId,
        },
      });

      try {
        await appActions.saveModel(this.$store, {
          fittedSolutionId: this.fittedSolutionId,
          modelName: this.saveName,
          modelDescription: this.saveDescription,
        });
      } catch (err) {
        console.warn(err);
      }
      this.$bvModal.show("save-success-modal");
    },

    // ensure required fields are filled out
    validForm() {
      const valid = this.saveName.length > 0;
      this.saveNameState = valid;
      return valid;
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
});
</script>

<style>
.success-modal-header {
  background: #d5ecdb;
}
.header-icon {
  color: #35a54c;
  margin-right: 5px;
}
</style>
