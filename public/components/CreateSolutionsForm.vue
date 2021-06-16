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
  <div class="create-solutions-form d-flex justify-content-center mt-2">
    <error-modal
      title="Model Failed"
      :show="showCreateFailure"
      :error="createErrorMessage"
      @close="showCreateFailure = !showCreateFailure"
    />
    <settings-modal :time-range="dateTimeExtrema" />
    <b-overlay
      :show="isPending"
      rounded
      opacity="0.6"
      spinner-small
      spinner-variant="success"
      class="d-inline-block"
    >
      <b-button-group>
        <b-button
          :variant="createVariant"
          :disabled="disableCreate"
          @click="create"
        >
          Create Models
        </b-button>
        <b-button
          v-b-modal.settings
          :variant="createVariant"
          :disabled="disableCreate"
        >
          <i class="fa fa-cog" aria-hidden="true" />
        </b-button>
      </b-button-group>
    </b-overlay>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import { createRouteEntry, varModesToString } from "../util/routes";
import ErrorModal from "../components/ErrorModal.vue";
import SettingsModal from "../components/SettingsModal.vue";
import {
  actions as appActions,
  getters as appGetters,
} from "../store/app/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { RESULTS_ROUTE } from "../store/route/index";
import { actions as requestActions } from "../store/requests/module";
import { Solution } from "../store/requests/index";
import { SolutionRequestMsg } from "../store/requests/actions";
import { Variable, DataMode } from "../store/dataset/index";
import { DATE_TIME_TYPE } from "../util/types";
import {
  FilterParams,
  FilterSetsParams,
  groupFiltersBySet,
} from "../util/filters";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { EventList } from "../util/events";
import Vue from "vue";

export default Vue.extend({
  name: "CreateSolutionsForm",

  components: {
    ErrorModal,
    SettingsModal,
  },
  props: {
    handleInput: { type: Boolean as () => boolean, default: false },
  },
  data() {
    return {
      pending: false,
      showExport: false,
      showExportSuccess: false,
      showExportFailure: false,
      showCreateFailure: false,
      createErrorMessage: null,
      $bvModal: null,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    filterParams(): FilterParams {
      return routeGetters.getDecodedSolutionRequestFilterParams(this.$store);
    },
    metrics(): string[] {
      return routeGetters.getModelMetrics(this.$store);
    },
    trainingSelected(): boolean {
      return !_.isEmpty(this.training);
    },
    targetSelected(): boolean {
      return !_.isEmpty(this.target);
    },
    training(): string[] {
      return routeGetters.getDecodedTrainingVariableNames(this.$store);
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    targetVariable(): Variable {
      return _.find(this.variables, (v) => {
        return _.toLower(v.key) === _.toLower(this.target);
      });
    },
    dateTimeExtrema(): { min: number; max: number } {
      const dateTimeVar = this.variables.find(
        (variable) => variable.colType === DATE_TIME_TYPE
      );
      if (!dateTimeVar) return;
      return { min: dateTimeVar.min, max: dateTimeVar.max };
    },
    isPending(): boolean {
      return this.pending;
    },
    disableCreate(): boolean {
      return this.isPending || !this.targetSelected || !this.trainingSelected;
    },
    createVariant(): string {
      return !this.disableCreate ? "success" : "outline-success";
    },
    percentComplete(): number {
      return 100;
    },
  },

  methods: {
    // create button handler
    create() {
      appActions.logUserEvent(this.$store, {
        feature: Feature.CREATE_MODEL,
        activity: Activity.DATA_PREPARATION,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: {},
      });

      // flag as pending
      this.pending = true;
      // dispatch action that triggers request send to server
      const routeSplit = routeGetters.getRouteTrainTestSplit(this.$store);
      const defaultSplit = appGetters.getTrainTestSplit(this.$store);
      const timestampSplit = routeGetters.getRouteTimestampSplit(this.$store);
      const filterSet = {
        ...this.filterParams,
        filters: {
          list: groupFiltersBySet(this.filterParams.filters.list),
          invert: this.filterParams.filters.invert,
        },
      } as FilterSetsParams;
      const solutionRequestMsg = {
        dataset: this.dataset,
        filters: filterSet,
        target: routeGetters.getRouteTargetVariable(this.$store),
        metrics: this.metrics,
        maxSolutions: routeGetters.getModelLimit(this.$store),
        maxTime: routeGetters.getModelTimeLimit(this.$store),
        quality: routeGetters.getModelQuality(this.$store),
        trainTestSplit: !!routeSplit ? routeSplit : defaultSplit,
        timestampSplitValue: timestampSplit,
      } as SolutionRequestMsg;

      // Add optional values to the request
      const positiveLabel = routeGetters.getPositiveLabel(this.$store);
      if (positiveLabel) solutionRequestMsg.positiveLabel = positiveLabel;
      if (!this.handleInput) {
        this.$emit(EventList.MODEL.CREATE_EVENT, solutionRequestMsg);
        return;
      }
      requestActions
        .createSolutionRequest(this.$store, solutionRequestMsg)
        .then((res: Solution) => {
          this.pending = false;
          const dataMode = routeGetters.getDataMode(this.$store);
          const dataModeDefault = dataMode ? dataMode : DataMode.Default;

          // transition to result screen
          const entry = createRouteEntry(RESULTS_ROUTE, {
            dataset: routeGetters.getRouteDataset(this.$store),
            target: routeGetters.getRouteTargetVariable(this.$store),
            solutionId: res.solutionId,
            task: routeGetters.getRouteTask(this.$store),
            dataMode: dataModeDefault,
            varModes: varModesToString(
              routeGetters.getDecodedVarModes(this.$store)
            ),
            modelLimit: routeGetters.getModelLimit(this.$store),
            modelTimeLimit: routeGetters.getModelTimeLimit(this.$store),
            modelQuality: routeGetters.getModelQuality(this.$store),
          });
          this.$router.push(entry).catch((err) => console.warn(err));
        })
        .catch((err) => {
          // display error modal
          this.pending = false;
          this.createErrorMessage = err.message;
          this.showCreateFailure = true;
        });
    },
  },
});
</script>

<style>
.close-modal {
  width: 35% !important;
}

.solution-progress {
  margin: 6px 10%;
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

.radio-container {
  padding: 0 15px;
}
</style>
