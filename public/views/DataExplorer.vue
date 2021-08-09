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

    <left-side-panel :panel-title="currentAction">
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
        />
      </template>
    </left-side-panel>

    <main class="content">
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
            />
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
          @button-event="onLabelAnnotationClicked"
          @select-all="onLabelSelectAll"
        />
        <legend-weight v-if="hasWeight && isResultState" class="ml-5 mr-auto" />
      </div>
      <!-- <layer-selection v-if="isMultiBandImage" class="layer-select-dropdown" /> -->
      <section class="data-container">
        <div v-if="!hasData" v-html="spinnerHTML" />
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
          :confidence-access-func="confidenceGetter"
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
        />
      </section>

      <footer
        class="d-flex align-items-end d-flex justify-content-between mt-1 mb-0"
      >
        <div class="flex-grow-1">
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
          @export="onLabelExport"
          @apply="onLabelApply"
          @save="onLabelSaveClick"
        />
      </footer>
    </main>
    <left-side-panel
      v-if="isOutcomeToggled"
      panel-title="Outcome Variables"
      class="overflow-auto"
    >
      <template v-if="hasNoVariables">
        <p>No Outcome Variables available.</p>
      </template>
      <div v-else-if="state.name === 'result'">
        <error-threshold-slider v-if="showResiduals && !isTimeseries" />
        <result-facets
          :single-solution="isSingleSolution"
          :show-residuals="showResiduals"
        />
      </div>
      <facet-list-pane
        v-else-if="state.name === 'label'"
        :variables="secondaryVariables"
        :enable-color-scales="geoVarExists"
        :include="include"
        :summaries="secondarySummaries"
        :enable-footer="isSelectState"
        @fetch-summaries="fetchSummaries"
      />
      <prediction-summaries v-else />
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
          :options="labelOptions"
        />
      </b-form-group>
    </b-modal>
    <save-dataset
      modal-id="save-dataset-modal"
      :dataset-name="dataset"
      @save="onSaveValid"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { capitalize, isEmpty } from "lodash";

// Components
import ActionColumn from "../components/layout/ActionColumn.vue";
import AddVariablePane from "../components/panel/AddVariablePane.vue";
import CreateSolutionsForm from "../components/CreateSolutionsForm.vue";
import DataSize from "../components/buttons/DataSize.vue";
import ErrorThresholdSlider from "../components/ErrorThresholdSlider.vue";
import FacetListPane from "../components/panel/FacetListPane.vue";
import LeftSidePanel from "../components/layout/LeftSidePanel.vue";
import LayerSelection from "../components/LayerSelection.vue";
import ImageMosaic from "../components/ImageMosaic.vue";
import SearchBar from "../components/layout/SearchBar.vue";
import SelectDataTable from "../components/SelectDataTable.vue";
import GeoPlot from "../components/GeoPlot.vue";
import SelectGraphView from "../components/SelectGraphView.vue";
import SaveDataset from "../components/labelingComponents/SaveDataset.vue";
import SelectTimeseriesView from "../components/SelectTimeseriesView.vue";
import StatusPanel from "../components/StatusPanel.vue";
import StatusSidebar from "../components/StatusSidebar.vue";
import ResultFacets from "../components/ResultFacets.vue";
import LegendWeight from "../components/LegendWeight.vue";
import SaveModal from "../components/SaveModal.vue";
import PredictionsDataUploader from "../components/PredictionsDataUploader.vue";
import PredictionSummaries from "../components/PredictionSummaries.vue";
import CreateLabelingForm from "../components/labelingComponents/CreateLabelingForm.vue";
import LabelHeaderButtons from "../components/labelingComponents/LabelHeaderButtons.vue";
import ForecastHorizon from "../components/ForecastHorizon.vue";
// Store
import {
  viewActions,
  datasetGetters,
  requestGetters,
  resultGetters,
  datasetActions,
} from "../store";
import {
  Extrema,
  Highlight,
  RowSelection,
  TableColumn,
  TableRow,
  TaskTypes,
  TimeSeries,
  Variable,
  VariableSummary,
} from "../store/dataset/index";
import {
  DATA_EXPLORER_VAR_INSTANCE,
  ROUTE_PAGE_SUFFIX,
} from "../store/route/index";
import { getters as routeGetters } from "../store/route/module";

// Util
import { Filter, INCLUDE_FILTER } from "../util/filters";
import { clearHighlight } from "../util/highlights";
import { overlayRouteEntry, RouteArgs } from "../util/routes";
import { clearRowSelection, getNumIncludedRows } from "../util/row";
import { spinnerHTML } from "../util/spinner";
import {
  DISTIL_ROLES,
  isGeoLocatedType,
  isImageType,
  isMultibandImageType,
  META_TYPES,
} from "../util/types";
import {
  GEO_VIEW,
  GRAPH_VIEW,
  IMAGE_VIEW,
  TABLE_VIEW,
  TIMESERIES_VIEW,
  filterViews,
} from "../util/view";
import { Dictionary } from "vue-router/types/router";
import { BaseState, SelectViewState } from "../util/state/AppStateWrapper";
import { SolutionRequestMsg } from "../store/requests/actions";
import { Solution } from "../store/requests";
import { EI } from "../util/events";
import ExplorerConfig, {
  Action,
  ActionNames,
  ACTION_MAP,
  ExplorerStateNames,
  getConfigFromName,
  getStateFromName,
  SelectViewConfig,
  LABEL_COMPUTES,
  SELECT_COMPUTES,
  RESULT_METHODS,
  RESULT_COMPUTES,
  SELECT_METHODS,
  LABEL_METHODS,
  GENERIC_METHODS,
} from "../util/dataExplorer";
import {
  LowShotLabels,
  LOW_SHOT_SCORE_COLUMN_PREFIX,
  sortVariablesByImportance,
} from "../util/data";
import _ from "lodash";

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
    LabelHeaderButtons,
    LayerSelection,
    LeftSidePanel,
    LegendWeight,
    ImageMosaic,
    PredictionsDataUploader,
    PredictionSummaries,
    SearchBar,
    SelectDataTable,
    GeoPlot,
    ResultFacets,
    SaveModal,
    SaveDataset,
    SelectGraphView,
    SelectTimeseriesView,
    StatusPanel,
    StatusSidebar,
  },

  data() {
    return {
      activeView: 0, // TABLE_VIEW
      instanceName: DATA_EXPLORER_VAR_INSTANCE,
      metaTypes: Object.keys(META_TYPES),
      include: true,
      state: new SelectViewState(),
      config: new SelectViewConfig(),
      labelName: "",
      labelModalId: "label-input-form",
      isBusy: false,
    };
  },
  computed: {
    /* Actions displayed on the Action column */
    activeActions(): Action[] {
      return this.availableActions.map((action) => {
        const count = this.variablesPerActions[action.paneId]?.length;
        return count ? { ...action, count } : action;
      });
    },

    /* Variables displayed on the Facet Panel */
    activeVariables(): Variable[] {
      return this.variablesPerActions[this.config.currentPane] ?? [];
    },

    activeViews(): string[] {
      return filterViews(this.variables);
    },

    /* All variables, only used for lex as we need to parse the hidden variables from groupings */
    allVariables(): Variable[] {
      const variables = [...this.state.getLexBarVariables()];
      return sortVariablesByImportance(variables);
    },

    /* Actions available based on the variables meta types */
    availableActions(): Action[] {
      // Remove the inactive MetaTypes
      return this.config.actionList.filter(
        (action) => !this.inactiveMetaTypes.includes(action.paneId)
      );
    },
    showResiduals(): boolean {
      const tasks = routeGetters.getRouteTask(this.$store).split(",");
      return (
        tasks &&
        !!tasks.find(
          (t) => t === TaskTypes.REGRESSION || t === TaskTypes.FORECASTING
        )
      );
    },
    targetName(): string {
      return this.target?.key;
    },
    isMultiBandImage(): boolean {
      return this.allVariables.some((v) => {
        return isMultibandImageType(v.colType);
      });
    },
    targetType(): string {
      const target = this.target;
      if (!target) {
        return null;
      }
      const variables = this.variables;
      return variables.find((v) => v.key === target.key)?.colType;
    },
    currentAction(): string {
      return (
        this.config.currentPane &&
        this.config.actionList.find((a) => a.paneId === this.config.currentPane)
          .name
      );
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    explore(): string[] {
      return routeGetters.getExploreVariables(this.$store);
    },

    filters(): string {
      return routeGetters.getRouteFilters(this.$store);
    },

    hasData(): boolean {
      return this.state.hasData();
    },
    activePane(): string {
      return this.config.currentPane;
    },
    hasNoVariables(): boolean {
      return isEmpty(this.activeVariables);
    },
    isTimeseries(): boolean {
      return routeGetters.isTimeseries(this.$store);
    },
    highlights(): Highlight[] {
      return _.cloneDeep(routeGetters.getDecodedHighlights(this.$store));
    },
    solution(): Solution {
      return requestGetters.getActiveSolution(this.$store);
    },
    solutionId(): string {
      return this.solution?.solutionId;
    },
    drillDownBaseline(): TableRow[] {
      return this.state.getMapDrillDownBaseline(this.include);
    },
    drillDownFiltered(): TableRow[] {
      return this.state.getMapDrillDownFiltered(this.include);
    },
    fittedSolutionId(): string {
      return this.solution?.fittedSolutionId;
    },
    residualExtrema(): Extrema {
      return resultGetters.getResidualsExtrema(this.$store);
    },
    isSingleSolution(): boolean {
      return routeGetters.isSingleSolution(this.$store);
    },
    timeseries(): TimeSeries {
      return this.state.getTimeseries();
    },
    routeHighlight(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },

    inactiveMetaTypes(): string[] {
      // Go trough each meta type
      return this.metaTypes.map((metaType) => {
        // test if some variables types...
        const typeNotInMetaTypes = !this.variablesTypes.some((t) =>
          // ...is in that meta type
          META_TYPES[metaType].includes(t)
        );
        if (typeNotInMetaTypes) return metaType;
      });
    },
    fields(): Dictionary<TableColumn> {
      return this.state.getFields(this.include);
    },

    isFilteringHighlights(): boolean {
      return this.highlights && this.highlights.length > 0;
    },

    isFilteringSelection(): boolean {
      return !!this.rowSelection;
    },

    numRows(): number {
      return this.hasData
        ? datasetGetters.getIncludedTableDataLength(this.$store)
        : 0;
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    selectionNumRows(): number {
      return getNumIncludedRows(this.rowSelection);
    },

    spinnerHTML,

    target(): Variable {
      return this.state.getTargetVariable();
    },

    totalNumRows(): number {
      return this.hasData
        ? datasetGetters.getIncludedTableDataNumRows(this.$store)
        : 0;
    },

    variables(): Variable[] {
      const variables = this.state
        .getVariables()
        .filter((v) => v.distilRole !== DISTIL_ROLES.Meta);
      return sortVariablesByImportance(variables);
    },

    variablesPerActions() {
      const variables = {};
      this.availableActions.forEach((action) => {
        variables[action.paneId] = action.variables(this);
      });
      return variables;
    },

    variablesTypes(): string[] {
      return [...new Set(this.variables.map((v) => v.colType))];
    },
    // enables or disables coloring by facets
    geoVarExists(): boolean {
      const varSums = this.summaries;
      return varSums.some((v) => {
        return isGeoLocatedType(v.type);
      });
    },
    imageVarExists(): boolean {
      const varSums = this.allVariables;
      return varSums.some((v) => {
        return isImageType(v.colType);
      });
    },
    viewComponent() {
      const viewType = this.activeViews[this.activeView] as string;
      if (viewType === GEO_VIEW) return "GeoPlot";
      if (viewType === GRAPH_VIEW) return "SelectGraphView";
      if (viewType === IMAGE_VIEW) return "ImageMosaic";
      if (viewType === TABLE_VIEW) return "SelectDataTable";
      if (viewType === TIMESERIES_VIEW) return "SelectTimeseriesView";

      // Default is TABLE_VIEW
      return "SelectDataTable";
    },
    // used to enable certain UI components
    isResultState(): boolean {
      return ExplorerStateNames.RESULT_VIEW === this.explorerRouteState;
    },
    // used to enable certain UI components
    isSelectState(): boolean {
      return ExplorerStateNames.SELECT_VIEW === this.explorerRouteState;
    },
    // used to enable certain UI components
    isPredictState(): boolean {
      return ExplorerStateNames.PREDICTION_VIEW === this.explorerRouteState;
    },
    isLabelState(): boolean {
      return ExplorerStateNames.LABEL_VIEW === this.explorerRouteState;
    },
    // basic table data used by all view components
    items(): TableRow[] {
      return this.state.getData(this.include);
    },
    // baselineMap is used to maintain index order for faster buffer changes
    baselineMap(): Dictionary<number> {
      const result = {};
      const base = this.baselineItems ?? [];
      base.forEach((item, i) => {
        result[item.d3mIndex] = i;
      });
      return result;
    },
    // used for map is the baseline
    baselineItems(): TableRow[] {
      return this.state.getMapBaseline();
    },
    // returns all summaries
    summaries(): VariableSummary[] {
      return this.state.getAllVariableSummaries(this.include);
    },
    // available summaries, result summaries, prediction summaries
    secondarySummaries(): VariableSummary[] {
      return this.state.getSecondaryVariableSummaries(this.include);
    },
    // available variables, result variables, prediction variables
    secondaryVariables(): Variable[] {
      return this.state.getSecondaryVariables();
    },
    explorerRouteState(): ExplorerStateNames {
      return routeGetters.getDataExplorerState(this.$store);
    },
    // toggles right side variable pane
    isOutcomeToggled(): boolean {
      const outcome = ACTION_MAP.get(ActionNames.OUTCOME_VARIABLES).paneId;
      return routeGetters
        .getToggledActions(this.$store)
        .some((a) => a === outcome);
    },
    isRemoteSensing(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },
    labelScoreName(): string {
      return LOW_SHOT_SCORE_COLUMN_PREFIX + this.labelName;
    },
    training(): string[] {
      return SELECT_COMPUTES.training(this);
    },
    isCreateModelPossible(): boolean {
      return SELECT_COMPUTES.isCreateModelPossible(this);
    },
    /* Disable the Exclude filter button. */
    isExcludeDisabled(): boolean {
      return SELECT_COMPUTES.isExcludedDisabled(this);
    },
    isClone(): boolean {
      return LABEL_COMPUTES.isClone(this);
    },
    labelOptions(): { value: string; text: string }[] {
      return LABEL_COMPUTES.options(this);
    },
    labelSummary(): VariableSummary {
      return LABEL_COMPUTES.labelSummary(this);
    },
    labelModalTitle(): string {
      return LABEL_COMPUTES.labelModalTitle(this);
    },
    confidenceGetter(): Function {
      if (!this.items?.length || this.labelName in this.items[0]) {
        return () => {
          return undefined;
        };
      }
      return LABEL_METHODS.confidenceGetter(this);
    },
    isActiveSolutionSaved(): boolean {
      return RESULT_COMPUTES.isActiveSolutionSaved(this);
    },
    hasWeight(): boolean {
      return RESULT_COMPUTES.hasWeight(this);
    },
  },

  // Update either the summaries or explore data on user interaction.
  watch: {
    activeVariables(n, o) {
      if (_.isEqual(n, o)) return;
      viewActions.fetchDataExplorerData(this.$store, this.activeVariables);
    },

    filters(n, o) {
      if (n === o) return;
      viewActions.updateDataExplorerData(this.$store);
    },

    highlights(n, o) {
      if (_.isEqual(n, o)) return;
      this.state.fetchData();
    },

    explore(n, o) {
      if (_.isEqual(n, o)) return;
      viewActions.updateDataExplorerData(this.$store);
    },
    geoVarExists() {
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, { hasGeoData: this.geoVarExists });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    targetName() {
      datasetActions.fetchOutliers(this.$store, this.dataset);
    },
  },

  async beforeMount() {
    // First get the dataset informations
    await viewActions.fetchDataExplorerData(this.$store, [] as Variable[]);
    // Pre-select the top 5 variables by importance
    this.preSelectTopVariables();
    // Update the explore data
    viewActions.updateDataExplorerData(this.$store);
  },
  mounted() {
    this.changeStatesByName(this.explorerRouteState);
    this.labelName = routeGetters.getRouteLabel(this.$store);
  },

  methods: {
    capitalize,
    async changeStatesByName(state: ExplorerStateNames) {
      // reset state
      this.state.resetState();
      // reset config state
      this.config.resetConfig(this);
      // get the new state object
      this.setState(getStateFromName(state));
      // set the config used for action bar, could be used for other configs
      this.setConfig(getConfigFromName(state));
      // init this is the basic fetches needed to get the information for the state
      await this.state.init();
    },
    /* When the user request to fetch a different size of data. */
    onDataSizeSubmit(dataSize: number) {
      this.updateRoute({ dataSize });
      viewActions.updateDataExplorerData(this.$store);
    },
    onSetActive(actionName: string): void {
      if (actionName === this.config.currentPane) return;

      let activePane = "available"; // default
      if (actionName !== "") {
        activePane = this.config.actionList.find((a) => a.name === actionName)
          .paneId;
      }
      this.config.currentPane = activePane;

      // update the selected pane, and reset the page var to 1
      this.updateRoute({
        pane: activePane,
        [`${DATA_EXPLORER_VAR_INSTANCE}${ROUTE_PAGE_SUFFIX}`]: 1,
      });
    },
    async onTileClick(data: EI.MAP.TileClickData) {
      // filter for area of interests
      const filter: Filter = {
        displayName: data.displayName,
        key: data.key,
        maxX: data.bounds[1][1],
        maxY: data.bounds[0][0],
        minX: data.bounds[0][1],
        minY: data.bounds[1][0],
        mode: INCLUDE_FILTER,
        type: data.type,
        set: "",
      };
      // fetch area of interests
      this.state.fetchMapDrillDown(filter);
    },
    setState(state: BaseState) {
      this.state = state;
      if (this.explorerRouteState !== state.name) {
        this.updateRoute({
          dataExplorerState: state.name,
          toggledActions: "[]",
        } as RouteArgs);
      }
    },
    setConfig(config: ExplorerConfig) {
      this.config = config;
      const toggledMap = new Map(
        routeGetters.getToggledActions(this.$store).map((t) => {
          return [t, true];
        })
      );
      // the switch to the new config will trigger a render of new elements
      // if the defaultActions is one of the new elements it will not exist in the dom yet
      // so we toggle the default actions after the next DOM cycle
      this.$nextTick(() => {
        this.config.defaultAction.forEach((actionName) => {
          const action = ACTION_MAP.get(actionName);
          if (!toggledMap.has(action.paneId)) {
            this.toggleAction(actionName);
          }
        });
      });
    },
    setIncludedActive() {
      this.include = true;
    },

    setExcludedActive() {
      this.include = false;
    },

    updateRoute(args) {
      const entry = overlayRouteEntry(this.$route, args);
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    resetHighlightsOrRow() {
      if (this.isFilteringHighlights) {
        clearHighlight(this.$router);
      } else {
        clearRowSelection(this.$router);
      }
    },

    preSelectTopVariables(number = 5): void {
      // if explore is already filled let's skip
      if (!isEmpty(this.explore)) return;

      // get the top 5 variables
      const top5Variables = [...this.variables]
        .slice(0, number)
        .map((variable) => variable.key)
        .join(",");

      // Update the route with the top 5 variable as training
      this.updateRoute({ explore: top5Variables });
    },
    fetchTimeseries(args: EI.TIMESERIES.FetchTimeseriesEvent) {
      this.state.fetchTimeseries(args);
    },
    fetchSummaries() {
      this.state.fetchVariableSummaries();
    },
    toggleAction(actionName: ActionNames) {
      GENERIC_METHODS.toggleAction(this)(actionName);
    },
    onModelCreation(solutionRequestMsg: SolutionRequestMsg) {
      SELECT_METHODS.onModelCreation(this)(solutionRequestMsg);
    },

    onExcludeClick() {
      SELECT_METHODS.onExcludeClick(this)();
    },

    onReincludeClick() {
      SELECT_METHODS.onReincludeClick(this)();
    },
    async updateTask() {
      LABEL_METHODS.updateTask(this)();
    },
    switchToLabelState() {
      this.$bvModal.show(this.labelModalId);
    },
    onLabelSubmit() {
      LABEL_METHODS.onLabelSubmit(this)();
    },
    onLabelAnnotationClicked(label: LowShotLabels) {
      LABEL_METHODS.onAnnotationChanged(this)(label);
    },
    onLabelSelectAll() {
      LABEL_METHODS.onSelectAll(this)();
    },
    onLabelExport() {
      LABEL_METHODS.onExport(this)();
    },
    onLabelApply() {
      LABEL_METHODS.onSearchSimilar(this)();
    },
    onLabelSaveClick() {
      this.$bvModal.show("save-dataset-modal");
    },
    onSaveValid(saveName: string, retainUnlabeled: boolean) {
      LABEL_METHODS.onSaveDataset(this)(saveName, retainUnlabeled);
    },
    onToolSelection(selection: EI.MAP.SelectionHighlight) {
      LABEL_METHODS.onToolSelection(this)(selection);
    },
    async onSaveModel(args: EI.RESULT.SaveInfo) {
      RESULT_METHODS.onSaveModel(this)(args);
    },
    isFittedSolutionIdSavedAsModel(id: string): boolean {
      return RESULT_METHODS.isFittedSolutionIdSavedAsModel(this)(id);
    },
    async onApplyModel(args: RouteArgs) {
      RESULT_METHODS.onApplyModel(this)(args);
    },
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
