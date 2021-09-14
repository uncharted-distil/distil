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
  <div class="view-container">
    <action-column
      ref="action-column"
      :actions="activeActions"
      :current-action="currentAction"
      @set-active-pane="onSetActive"
    />
    <left-side-panel v-if="currentAction !== ''" :panel-title="currentAction">
      <add-variable-pane
        v-if="activePane === 'add'"
        :enable-label="imageVarExists"
        @label="switchToLabelState"
      />
      <template v-else>
        <template v-if="hasNoVariables">
          <p v-if="activePane === 'selected'">Select a variable to explore.</p>
          <p v-else>All the variables of that type are selected.</p>
        </template>
        <facet-list-pane
          v-else
          :is-target-panel="activePane === 'target' && isSelectState"
          :variables="activeVariables"
          :enable-color-scales="geoVarExists"
          :include="include"
          :summaries="summaries"
          :enable-footer="isSelectState"
          @fetch-summaries="fetchSummaries"
          @type-change="fetchSummaries"
        />
      </template>
    </left-side-panel>
    <main class="content">
      <loading-spinner v-show="isBusy" :state="busyState" />
      <template v-show="!isBusy">
        <search-bar
          :variables="allVariables"
          :filters="filters"
          :highlights="routeHighlight"
          handle-updates
        />

        <!-- Tabs to switch views -->

        <div class="d-flex flex-row align-items-end mt-2">
          <div class="flex-grow-1 mr-2">
            <b-tabs v-model="activeView" class="tab-container">
              <b-tab
                v-for="(view, index) in activeViews"
                :key="index"
                :active="view === activeViews[activeView]"
                :title="capitalize(view)"
                @click="onTabClick(view)"
              >
                <template v-slot:title>
                  <b-spinner v-if="dataLoading" small />
                </template>
              </b-tab>
            </b-tabs>
          </div>
          <layer-selection
            v-if="isMultiBandImage"
            :has-image-attention="isResultState"
            class="align-self-center mr-2"
          />
          <b-button
            v-if="include && isSelectState"
            class="select-data-action-exclude align-self-center"
            variant="outline-secondary"
            :disabled="isExcludeDisabled"
            @click="onExcludeClick"
          >
            <i
              class="fa fa-minus-circle pr-1"
              :class="{
                'exclude-highlight': isFilteringHighlights,
                'exclude-selection': isFilteringSelection,
              }"
            />
            Exclude
          </b-button>
          <b-button
            v-if="!include && isSelectState"
            variant="outline-secondary"
            :disabled="!isFilteringSelection"
            @click="onReincludeClick"
          >
            <i
              class="fa fa-plus-circle pr-1"
              :class="{ 'include-selection': isFilteringSelection }"
            />
            Reinclude
          </b-button>
          <label-header-buttons
            v-if="isLabelState"
            class="height-36"
            @button-event="onAnnotationChanged"
            @select-all="onSelectAll"
          />
          <legend-weight
            v-if="hasWeight && isResultState"
            class="ml-5 mr-auto"
          />
        </div>
        <section class="data-container">
          <component
            :is="viewComponent"
            ref="dataView"
            :instance-name="instanceName"
            :included-active="include"
            :dataset="dataset"
            :data-fields="fields"
            :timeseries-info="timeseries"
            :data-items="items"
            :item-count="items.length"
            :baseline-items="baselineItems"
            :baseline-map="baselineMap"
            :summaries="summaries"
            :solution="solution"
            :residual-extrema="residualExtrema"
            :enable-selection-tool-event="isLabelState"
            :variables="allVariables"
            :label-feature-name="labelName"
            :label-score-name="labelName"
            :area-of-interest-items="{
              inner: drillDownBaseline,
              outer: drillDownFiltered,
            }"
            :get-timeseries="state.getTimeseries"
            @tile-clicked="onTileClick"
            @selection-tool-event="onToolSelection"
            @fetch-timeseries="fetchTimeseries"
            @finished-loading="onMapFinishedLoading"
          />
        </section>

        <footer
          class="d-flex align-items-end d-flex justify-content-between mt-1 mb-0"
        >
          <div v-if="!isGeoView" class="flex-grow-1">
            <data-size
              :current-size="numRows"
              :total="totalNumRows"
              @submit="onDataSizeSubmit"
            />
            <strong class="matching-color">matching</strong> samples of
            {{ totalNumRows }} to model<template v-if="selectionNumRows > 0">
              , {{ selectionNumRows }}
              <strong class="selected-color">selected</strong>
            </template>
          </div>
          <div v-else class="flex-grow-1">
            <p class="m-0">
              Selected Area Coverage:
              <strong class="matching-color">
                {{ areaCoverage }}km<sup>2</sup>
              </strong>
            </p>
          </div>
          <b-button-toolbar v-if="isSelectState">
            <b-button-group class="ml-2 mt-1">
              <b-button
                variant="primary"
                :disabled="include"
                @click="setIncludedActive"
              >
                Included
              </b-button>
              <b-button
                variant="secondary"
                :disabled="!include"
                @click="setExcludedActive"
              >
                Excluded
              </b-button>
            </b-button-group>
          </b-button-toolbar>
          <!-- RESULT AND PREDICTION VIEW COMPONENTS-->
          <create-solutions-form
            v-if="isSelectState"
            ref="model-creation-form"
            :aria-disabled="isCreateModelPossible"
            class="ml-2"
            @create-model="onModelCreation"
          />
          <predictions-data-uploader
            :fitted-solution-id="fittedSolutionId"
            :target="targetName"
            :target-type="targetType"
            @model-apply="onApplyModel"
          />
          <save-modal
            ref="saveModel"
            :solution-id="solutionId"
            :fitted-solution-id="fittedSolutionId"
            @save="onSaveModel"
          />
          <forecast-horizon
            v-if="isTimeseries"
            :dataset="dataset"
            :fitted-solution-id="fittedSolutionId"
            :target="targetName"
            :target-type="targetType"
            @model-apply="onApplyModel"
          />
          <template
            v-if="isResultState && (isSingleSolution || isActiveSolutionSaved)"
          >
            <b-button
              v-if="isTimeseries"
              variant="success"
              class="apply-button"
              @click="$bvModal.show('forecast-horizon-modal')"
            >
              Forecast
            </b-button>
            <b-button
              v-else
              variant="success"
              class="apply-button"
              @click="$bvModal.show('predictions-data-upload-modal')"
            >
              Apply Model
            </b-button>
          </template>
          <b-button
            v-else-if="isResultState"
            variant="success"
            class="save-button"
            @click="$bvModal.show('save-model-modal')"
          >
            <i class="fa fa-floppy-o" />
            Save Model
          </b-button>
          <b-button v-if="isPredictState" v-b-modal.save>
            Create Dataset
          </b-button>
          <b-button v-if="isPredictState" v-b-modal.export variant="primary">
            Export Predictions
          </b-button>
          <create-labeling-form
            v-if="isLabelState"
            class="d-flex justify-content-between h-100 align-items-center"
            :is-loading="isBusy"
            :low-shot-summary="labelSummary"
            :is-saving="isBusy"
            @export="onExport"
            @apply="onSearchSimilar"
            @save="onLabelSaveClick"
          />
        </footer>
      </template>
    </main>
    <left-side-panel
      v-if="isOutcomeToggled"
      panel-title="Outcome Variables"
      class="overflow-auto"
    >
      <div v-if="state.name === 'result'">
        <error-threshold-slider v-if="showResiduals && !isTimeseries" />
        <result-facets
          :single-solution="isSingleSolution"
          :show-residuals="showResiduals"
          @fetch-summary-solution="fetchSummarySolution"
        />
      </div>
      <facet-list-pane
        v-else-if="isLabelState"
        :variables="secondaryVariables"
        :enable-color-scales="geoVarExists"
        :include="include"
        :summaries="secondarySummaries"
        :enable-footer="isSelectState"
        @fetch-summaries="fetchSummaries"
      />
      <prediction-summaries
        v-else
        @fetch-summary-prediction="fetchSummaryPrediction"
      />
    </left-side-panel>
    <status-sidebar />
    <status-panel />
    <b-modal :id="labelModalId" @ok="onLabelSubmit">
      <template #modal-header>
        {{ labelModalTitle }}
      </template>
      <b-form-group
        v-if="!isClone"
        id="input-group-1"
        label="Label name:"
        label-for="label-input-field"
        description="Enter the name of label."
      >
        <b-form-input
          id="label-input-field"
          v-model="labelName"
          type="text"
          required
          :placeholder="labelName"
        />
      </b-form-group>
      <b-form-group
        v-else
        label="Label name:"
        label-for="label-select-field"
        description="Select the label field."
      >
        <b-form-select
          id="label-select-field"
          v-model="labelName"
          :options="options"
        />
      </b-form-group>
    </b-modal>
    <save-dataset
      modal-id="save-dataset-modal"
      :dataset-name="dataset"
      @save="onSaveDataset"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { capitalize } from "lodash";

// Components
import ActionColumn from "../components/layout/ActionColumn.vue";
import AddVariablePane from "../components/panel/AddVariablePane.vue";
import CreateLabelingForm from "../components/labelingComponents/CreateLabelingForm.vue";
import CreateSolutionsForm from "../components/CreateSolutionsForm.vue";
import DataSize from "../components/buttons/DataSize.vue";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider.vue";
import FacetListPane from "../components/panel/FacetListPane.vue";
import ForecastHorizon from "../components/ForecastHorizon.vue";
import GeoPlot from "../components/GeoPlot.vue";
import ImageMosaic from "../components/ImageMosaic.vue";
import LabelHeaderButtons from "../components/labelingComponents/LabelHeaderButtons.vue";
import LayerSelection from "../components/LayerSelection.vue";
import LeftSidePanel from "../components/layout/LeftSidePanel.vue";
import LegendWeight from "../components/LegendWeight.vue";
import LoadingSpinner from "../components/LoadingSpinner.vue";
import PredictionsDataUploader from "../components/PredictionsDataUploader.vue";
import PredictionSummaries from "../components/PredictionSummaries.vue";
import ResultFacets from "../components/ResultFacets.vue";
import SaveDataset from "../components/labelingComponents/SaveDataset.vue";
import SaveModal from "../components/SaveModal.vue";
import SearchBar from "../components/layout/SearchBar.vue";
import SelectDataTable from "../components/SelectDataTable.vue";
import SelectGraphView from "../components/SelectGraphView.vue";
import SelectTimeseriesView from "../components/SelectTimeseriesView.vue";
import StatusPanel from "../components/StatusPanel.vue";
import StatusSidebar from "../components/StatusSidebar.vue";
// Store
import { viewActions, datasetActions } from "../store";
import { Variable } from "../store/dataset/index";
import { DATA_EXPLORER_VAR_INSTANCE } from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";

// Util
import { overlayRouteEntry } from "../util/routes";
import { META_TYPES } from "../util/types";
import { SelectViewState } from "../util/state/AppStateWrapper";
import {
  bindMethods,
  SelectViewConfig,
  genericMethods,
  genericComputes,
  labelMethods,
  labelComputes,
  resultMethods,
  resultComputes,
  selectComputes,
  selectMethods,
  predictionMethods,
  predictionComputes,
} from "../util/explorer";
import _ from "lodash";
import { DataExplorerRef } from "../util/componentTypes";

const DataExplorer = Vue.extend({
  name: "DataExplorer",

  components: {
    ActionColumn,
    AddVariablePane,
    CreateLabelingForm,
    CreateSolutionsForm,
    DataSize,
    ErrorThresholdSlider,
    FacetListPane,
    ForecastHorizon,
    GeoPlot,
    ImageMosaic,
    LabelHeaderButtons,
    LayerSelection,
    LeftSidePanel,
    LegendWeight,
    LoadingSpinner,
    PredictionsDataUploader,
    PredictionSummaries,
    ResultFacets,
    SaveDataset,
    SaveModal,
    SearchBar,
    SelectDataTable,
    SelectGraphView,
    SelectTimeseriesView,
    StatusPanel,
    StatusSidebar,
  },

  data() {
    return {
      activeView: 0, // TABLE_VIEW
      busyState: "Busy", // contains the info to display to the user when the UI is busy
      config: new SelectViewConfig(), // this config controls what is displayed in the action bar
      dataLoading: false, // this controls the spinners for the data view tabs (table, mosaic, geoplot)
      include: true, // this controls the include exclude view for the select state
      instanceName: DATA_EXPLORER_VAR_INSTANCE, // component instance name
      isBusy: false, // controls spinners in label state when search similar or save is used
      labelModalId: "label-input-form", // modal id
      labelName: "", // labelName of the variable being annotated in the label view
      metaTypes: Object.keys(META_TYPES), // all of the meta types categories
      state: new SelectViewState(), // this state controls data flow
    };
  },

  // Update either the summaries or explore data on user interaction.
  watch: {
    solutionId() {
      this.dataLoading = true;
      this.state.fetchData();
      this.dataLoading = false;
    },
    produceRequestId() {
      this.dataLoading = true;
      this.state.fetchData();
      this.dataLoading = false;
    },
    activeVariables(n, o) {
      if (_.isEqual(n, o)) return;
      this.state.fetchVariableSummaries();
    },

    async filters(n, o) {
      if (n === o) return;
      this.dataLoading = true;
      await this.state.fetchData();
      this.dataLoading = false;
    },

    async highlights(n, o) {
      if (_.isEqual(n, o)) return;
      this.dataLoading = true;
      await this.state.fetchData();
      this.dataLoading = false;
    },

    async explore(n, o) {
      if (_.isEqual(n, o)) return;
      this.dataLoading = true;
      await viewActions.updateDataExplorerData(this.$store);
      this.dataLoading = false;
    },
    async geoVarExists() {
      const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
      if (
        (!self.geoVarExists && self.summaries.some((s) => s.pending)) ||
        self.geoVarExists === routeGetters.hasGeoData(this.$store)
      ) {
        return;
      }
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, { hasGeoData: self.geoVarExists });
      this.$router.push(entry).catch((err) => console.warn(err));
      this.dataLoading = true;
      await this.state.fetchMapBaseline();
      await this.state.fetchData();
      this.dataLoading = false;
    },
    targetName() {
      const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
      datasetActions.fetchOutliers(this.$store, self.dataset);
      datasetActions.fetchModelingMetrics(this.$store, self.task);
    },
  },
  beforeCreate() {
    const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
    // computes / methods need to be binded to the instance
    this.$options.computed = {
      ...this.$options.computed, // any computes defined in the component
      ...bindMethods(genericComputes, self), // generic computes used across all states
      ...bindMethods(resultComputes, self), // computes used in result state
      ...bindMethods(selectComputes, self), // computes used in select state
      ...bindMethods(predictionComputes, self), // computes used in prediction state
      ...bindMethods(labelComputes, self), // computes used in the label state
    };
    this.$options.methods = {
      ...this.$options.methods, // any methods defined in the component
      ...bindMethods(genericMethods, self), // generic computes used across all states
      ...bindMethods(selectMethods, self), // computes used in result state
      ...bindMethods(labelMethods, self), // computes used in select state
      ...bindMethods(resultMethods, self), // computes used in prediction state
      ...bindMethods(predictionMethods, self), // computes used in the label state
    };
  },
  async beforeMount() {
    const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
    if (self.isSelectState) {
      // First get the dataset informations
      await viewActions.fetchDataExplorerData(this.$store, [] as Variable[]);
      // Pre-select the top 5 variables by importance
      self.preSelectTopVariables();
      // Update the explore data
      await viewActions.updateDataExplorerData(this.$store);
    }
  },

  mounted() {
    const self = (this as unknown) as DataExplorerRef; // because the computes/methods are added in beforeCreate typescript does not work so we cast it to a type here
    self.changeStatesByName(self.explorerRouteState);
    self.labelName = routeGetters.getRouteLabel(this.$store);
  },
  methods: {
    capitalize,
  },
});
export default DataExplorer;
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  flex-grow: 1;
  height: var(--content-full-height);
  margin-top: var(--navbar-outer-height);
  overflow: hidden;
}

/* Make some elements of a container unsquishable. */
.view-container > *:not(.content),
.content > *:not(.data-container) {
  flex-shrink: 0;
}

.content {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  padding-bottom: 1rem;
  padding-top: 1rem;
}

/* Add padding to all elements but the tabs and data */
.content > *:not(.data-container),
.content > *:not(.tab-container) {
  padding-left: 1rem;
  padding-right: 1rem;
}

.tab-container,
.data-container {
  border-bottom: 1px solid var(--border-color);
}

.data-container {
  background-color: var(--white);
  display: flex;
  flex-flow: wrap;
  height: 100%;
  padding: 1rem;
  position: relative;
  width: 100%;
}
</style>
<style>
.view-container .tab-container ul.nav-tabs {
  border: none;
  margin-bottom: -1px;
}

.view-container .tab-container a.nav-link {
  border: 1px solid transparent;
  border-bottom-color: var(--border-color);
  border-top-width: 3px;
  color: var(--color-text-second);
  margin-bottom: 0;
}

.view-container .tab-container a.nav-link.active {
  background-color: var(--white);
  border-color: var(--border-color);
  border-top-color: var(--primary);
  border-bottom-width: 0;
  border-top-left-radius: 0.25rem;
  border-top-right-radius: 0.25rem;
  color: var(--primary);
  margin-bottom: -1px;
}

.select-data-action-exclude:not([disabled]) .include-highlight,
.select-data-action-exclude:not([disabled]) .exclude-highlight {
  color: var(--blue); /* #255dcc; */
}

.select-data-action-exclude:not([disabled]) .include-selection,
.select-data-action-exclude:not([disabled]) .exclude-selection {
  color: var(--red); /* #ff0067; */
}
</style>
