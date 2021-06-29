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
  <div class="result-panel">
    <h6 class="sidebar-title">Ground Truth</h6>
    <result-target-variable class="result-target-variable" />

    <!-- Searchs results -->
    <h6 class="sidebar-title">Results</h6>
    <section class="result-options">
      <error-threshold-slider v-if="showResiduals && !isTimeseries" />
    </section>
    <section class="result-summaries">
      <result-facets
        :show-residuals="showResiduals"
        :single-solution="isSingleSolution"
      />
    </section>

    <!-- Action buttons -->
    <aside
      v-if="isActiveSolutionCompleted"
      class="d-flex flex-row flex-shrink-0 justify-content-end"
    >
      <!-- Modal boxes to apply new data to models. -->
      <forecast-horizon
        v-if="isTimeseries"
        :dataset="dataset"
        :fitted-solution-id="fittedSolutionId"
        :target="target"
        :target-type="targetType"
      />
      <predictions-data-uploader
        v-else
        :fitted-solution-id="fittedSolutionId"
        :target="target"
        :target-type="targetType"
        handle-model-creation
      />
      <template v-if="isSingleSolution || isActiveSolutionSaved">
        <b-button
          v-if="isTimeseries"
          variant="primary"
          class="apply-button"
          @click="$bvModal.show('forecast-horizon-modal')"
        >
          Forecast
        </b-button>
        <b-button
          v-else
          variant="primary"
          class="apply-button"
          @click="$bvModal.show('predictions-data-upload-modal')"
        >
          Apply Model
        </b-button>
      </template>
      <template v-else>
        <save-modal
          ref="saveModel"
          :solution-id="solutionId"
          :fitted-solution-id="fittedSolutionId"
          @save="onSaveModel"
        />
        <b-button
          variant="success"
          class="save-button"
          @click="$bvModal.show('save-model-modal')"
        >
          <i class="fa fa-floppy-o" />
          Save Model
        </b-button>
      </template>
    </aside>
  </div>
</template>

<script lang="ts">
import ResultFacets from "../components/ResultFacets.vue";
import PredictionsDataUploader from "../components/PredictionsDataUploader.vue";
import ForecastHorizon from "../components/ForecastHorizon.vue";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider.vue";
import SaveModal from "./SaveModal.vue";
import ResultTargetVariable from "../components/ResultTargetVariable.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { getters as requestGetters } from "../store/requests/module";
import { Variable, TaskTypes } from "../store/dataset/index";
import Vue from "vue";
import { Solution, SolutionStatus } from "../store/requests/index";
import { isFittedSolutionIdSavedAsModel } from "../util/models";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { actions as appActions } from "../store/app/module";
import { EI } from "../util/events";

export default Vue.extend({
  name: "ResultSummaries",

  components: {
    ErrorThresholdSlider,
    ForecastHorizon,
    PredictionsDataUploader,
    ResultFacets,
    ResultTargetVariable,
    SaveModal,
  },

  data() {
    return {
      formatter(arg) {
        return arg ? arg.toFixed(2) : "";
      },
      symmetricSlider: true,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    targetType(): string {
      const targetName = this.target;
      const variables = this.variables;
      return variables.find((v) => v.key === targetName)?.colType;
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    taskArgs(): string[] {
      return routeGetters.getRouteTask(this.$store).split(",");
    },

    showResiduals(): boolean {
      return (
        this.taskArgs &&
        !!this.taskArgs.find(
          (t) => t === TaskTypes.REGRESSION || t === TaskTypes.FORECASTING
        )
      );
    },

    solutionId(): string {
      return requestGetters.getActiveSolution(this.$store)?.solutionId;
    },

    fittedSolutionId(): string {
      return requestGetters.getActiveSolution(this.$store)?.fittedSolutionId;
    },

    activeSolution(): Solution {
      return requestGetters.getActiveSolution(this.$store);
    },

    activeSolutionName(): string {
      return this.activeSolution ? this.activeSolution.feature : "";
    },

    instanceName(): string {
      return "groundTruth";
    },

    /**
     * Check that the active solution is completed.
     * This is used to display possible actions on the selected model.
     */
    isActiveSolutionCompleted(): boolean {
      return !!(
        this.activeSolution &&
        this.activeSolution.progress === SolutionStatus.SOLUTION_COMPLETED
      );
    },

    /**
     * Check that the active solution is saved as a model.
     * This is used to display possible actions on the selected model.
     */
    isActiveSolutionSaved(): boolean {
      return this.isFittedSolutionIdSavedAsModel(this.fittedSolutionId);
    },

    // Indicates whether or not the contained result facets should show "relevant"
    // results, which consist of those that match the target/dataset, or a single
    // result, which matches the route solutionID.  The latter case occurs when the
    // user selects a model directly from the search screen.
    isSingleSolution(): boolean {
      return routeGetters.isSingleSolution(this.$store);
    },

    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
  },

  methods: {
    isFittedSolutionIdSavedAsModel,
    async onSaveModel(args: EI.RESULT.SaveInfo) {
      appActions.logUserEvent(this.$store, {
        feature: Feature.EXPORT_MODEL,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_SAVE,
        details: {
          solution: args.solutionId,
          fittedSolution: args.fittedSolution,
        },
      });

      try {
        const err = await appActions.saveModel(this.$store, {
          fittedSolutionId: this.fittedSolutionId,
          modelName: args.name,
          modelDescription: args.description,
        });
        // should probably change UI based on error
        if (!err) {
          const modal = this.$refs.saveModel as InstanceType<typeof SaveModal>;

          modal.showSuccessModel();
        }
      } catch (err) {
        console.warn(err);
      }
    },
  },
});
</script>

<style>
.result-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.result-options {
  flex-shrink: 0;
}

.result-summaries {
  overflow-x: hidden;
  overflow-y: auto;
}

.result-summaries .facets-facet-base {
  overflow: visible;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
  box-shadow: inset 0 0 0 1000px var(--blue);
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

.check-button {
  width: 60%;
  margin: 0 20%;
}

.save-button {
  flex-shrink: 0;
  flex-grow: 0;
  margin-top: 15px;
  margin-bottom: 0px;
  margin-right: 8px;
}

.apply-button {
  flex-shrink: 0;
  flex-grow: 0;
  margin-top: 15px;
  margin-bottom: 0px;
  margin-right: 8px;
}

.result-target-variable .variable-facets-item {
  margin-top: 0px;
  padding-top: 0px;
}
</style>
