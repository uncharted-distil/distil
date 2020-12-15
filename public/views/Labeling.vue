<template>
  <div class="row flex-1 pb-3 h-100">
    <div class="col-12 col-md-3 d-flex h-100 flex-column">
      <h5 class="header-title">Labels</h5>
      <variable-facets
        enable-highlighting
        enable-type-filtering
        :summaries="[labelSummary]"
        :instance-name="instance"
        class="h-18"
      />
      <h5 class="header-title">Features</h5>
      <variable-facets
        enable-highlighting
        enable-type-filtering
        :summaries="summaries"
        :pagination="
          summaries && searchedActiveVariables.length > numRowsPerPage
        "
        :facet-count="summaries && searchedActiveVariables.length"
        :rows-per-page="numRowsPerPage"
        :instance-name="instance"
      />
    </div>
    <div class="col-12 col-md-6 d-flex flex-column h-100">
      <div class="flex-1 d-flex flex-column pb-1 pt-2">
        <labeling-data-slot
          :summaries="summaries"
          :variables="variables"
          :ranked-set="rankedSet"
          @DataChanged="onAnnotationChanged"
        />
        <create-labeling-form
          :is-loading="isLoadingData"
          @export="onExport"
          @apply="onApply"
        />
      </div>
    </div>
    <b-modal :id="modalId" title="Label Creation" @hide="onLabelSubmit">
      <b-form-group
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
    </b-modal>
    <label-score-pop-up
      :data="dataItems"
      :data-fields="dataFields"
      :summaries="summaries"
      :ranked-set="rankedSet"
      :is-loading="isLoadingData"
      :is-remote-sensing="isRemoteSensing"
      @button-event="onAnnotationChanged"
      @apply="onApply"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { getters as routeGetters } from "../store/route/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { LABEL_FEATURE_INSTANCE } from "../store/route/index";
import { actions as viewActions } from "../store/view/module";
import {
  getVariableSummariesByState,
  searchVariables,
  NUM_PER_TARGET_PAGE,
  cloneDatasetUpdateRoute,
  LowShotLabels,
  LOW_SHOT_LABEL_COLUMN_NAME,
  minimumRouteKey,
  parseBinaryScoreResponse,
  BinaryScoreResponse,
} from "../util/data";
import {
  Variable,
  VariableSummary,
  TableRow,
  TableColumn,
} from "../store/dataset/index";
import { CATEGORICAL_TYPE } from "../util/types";
import VariableFacets from "../components/facets/VariableFacets.vue";
import CreateLabelingForm from "../components/labelingComponents/CreateLabelingForm.vue";
import LabelingDataSlot from "../components/labelingComponents/LabelingDataSlot.vue";
import { EXCLUDE_FILTER, Filter } from "../util/filters";
import { Dictionary } from "vue-router/types/router";
import { updateHighlight, clearHighlight } from "../util/highlights";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { overlayRouteEntry } from "../util/routes";
import { actions as requestActions } from "../store/requests/module";
import LabelScorePopUp from "../components/labelingComponents/LabelScorePopUp.vue";
import { clearRowSelection } from "../util/row";

const LABEL_KEY = "label";

export default Vue.extend({
  name: "LabelingView",
  components: {
    VariableFacets,
    LabelingDataSlot,
    CreateLabelingForm,
    LabelScorePopUp,
  },
  props: {
    logActivity: {
      type: String as () => Activity,
      default: Activity.DATA_PREPARATION,
    },
  },
  data() {
    return {
      labelName: LOW_SHOT_LABEL_COLUMN_NAME,
      modalId: "label-input-form",
      rankedSet: null,
      isLoadingData: false,
    };
  },
  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    availableTargetVarsSearch(): string {
      return routeGetters.getRouteAvailableTargetVarsSearch(this.$store);
    },
    isRemoteSensing(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },
    searchedActiveVariables(): Variable[] {
      // remove variables used in groupedFeature;
      const activeVariables = this.variables.filter(
        (v) => !this.groupedFeatures.includes(v.colName)
      );

      return searchVariables(activeVariables, this.availableTargetVarsSearch);
    },
    groupedFeatures(): string[] {
      // Fetch the grouped features.
      const groupedFeatures = datasetGetters
        .getGroupings(this.$store)
        .filter((group) => Array.isArray(group.grouping.subIds))
        .map((group) => group.grouping.subIds)
        .flat();
      return groupedFeatures;
    },
    numRowsPerPage(): number {
      return NUM_PER_TARGET_PAGE;
    },
    lowShotSummary(): Dictionary<VariableSummary> {
      const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
        this.$store
      );
      return summaryDictionary
        ? summaryDictionary[LOW_SHOT_LABEL_COLUMN_NAME]
        : null;
    },
    dataItems(): TableRow[] {
      return datasetGetters.getIncludedTableDataItems(this.$store);
    },
    dataFields(): Dictionary<TableColumn> {
      return datasetGetters.getIncludedTableDataFields(this.$store);
    },
    labelSummary(): VariableSummary {
      if (!this.lowShotSummary) {
        return this.getDefaultLabelFacet();
      }
      const routeKey = minimumRouteKey();
      const lowShotLabel = this.lowShotSummary[routeKey];
      return !!lowShotLabel ? lowShotLabel : this.getDefaultLabelFacet();
    },
    numData(): number {
      const tableData = datasetGetters.getIncludedTableDataItems(this.$store);
      return tableData ? tableData.length : 0;
    },
    summaries(): VariableSummary[] {
      const pageIndex = routeGetters.getLabelFeaturesVarsPage(this.$store);

      const summaryDictionary = datasetGetters.getVariableSummariesDictionary(
        this.$store
      );

      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        this.numRowsPerPage,
        this.searchedActiveVariables,
        summaryDictionary
      );

      return currentSummaries;
    },
    highlight(): string {
      return routeGetters.getRouteHighlight(this.$store);
    },
    filters(): Filter[] {
      return routeGetters.getDecodedFilters(this.$store);
    },
    instance(): string {
      return LABEL_FEATURE_INSTANCE;
    },
    isClone(): boolean {
      return this.variables.some((v) => {
        return v.colName === LOW_SHOT_LABEL_COLUMN_NAME;
      });
    },
  },
  watch: {
    highlight() {
      this.onDataChanged();
    },
  },
  async mounted() {
    await this.fetchData();
    if (this.isClone) {
      // dataset is already a clone don't clone again. (used for testing. might add button for cloning later.)
      this.updateRoute();
      return;
    }
    this.$bvModal.show(this.modalId);
    const entry = await cloneDatasetUpdateRoute();
    if (entry === null) {
      return;
    }
    this.$router.push(entry).catch((err) => console.warn(err));
  },
  methods: {
    // used for generating default labels in the instance where labels do not exist in the dataset
    getDefaultLabelFacet(): VariableSummary {
      return {
        label: LABEL_KEY,
        key: LABEL_KEY,
        dataset: this.dataset,
        description: "Generated labels.",
        baseline: {
          buckets: [
            { key: LowShotLabels.positive, count: 0 },
            { key: LowShotLabels.negative, count: 0 },
            { key: LowShotLabels.unlabeled, count: this.numData },
          ],
          extrema: { min: 0, max: this.numData },
        },
      };
    },
    async onDataChanged() {
      await this.fetchData();
      if (this.isRemoteSensing) {
        await viewActions.updateHighlight(this.$store);
      }
    },
    async onApply() {
      this.isLoadingData = true;
      const res = await requestActions.createQueryRequest(this.$store, {
        datasetId: this.dataset,
        target: LOW_SHOT_LABEL_COLUMN_NAME,
        filters: null,
      });
      const rankedSet = parseBinaryScoreResponse(res as BinaryScoreResponse);
      this.isLoadingData = false;
      if (!rankedSet) {
        console.error("Error parsing binary score response");
        return;
      }
      this.rankedSet = rankedSet;
    },
    onExport() {
      const highlight = {
        context: this.instance,
        dataset: this.dataset,
        key: LOW_SHOT_LABEL_COLUMN_NAME,
        value: LowShotLabels.unlabeled,
      }; // exclude unlabeled from data export
      const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
        this.$store
      );
      const dataMode = routeGetters.getDataMode(this.$store);
      datasetActions.extractDataset(this.$store, {
        dataset: this.dataset,
        filterParams,
        highlight,
        include: true,
        mode: EXCLUDE_FILTER,
        dataMode,
      });
    },
    onFacetClick(context: string, key: string, value: string, dataset: string) {
      if (key && value) {
        updateHighlight(this.$router, {
          context: this.instance,
          dataset: dataset,
          key: key,
          value: value,
        });
      } else {
        clearHighlight(this.$router);
      }
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: this.logActivity,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: { key: key, value: value },
      });
    },
    async onLabelSubmit() {
      // add new field
      await datasetActions.addField<string>(this.$store, {
        dataset: this.dataset,
        name: LOW_SHOT_LABEL_COLUMN_NAME,
        fieldType: CATEGORICAL_TYPE,
        defaultValue: LowShotLabels.unlabeled,
        displayName: this.labelName,
      });
      // fetch new dataset with the newly added field
      await this.fetchData();
      // update task based on the current training data
      this.updateRoute();
    },
    async fetchData() {
      await datasetActions.fetchVariables(this.$store, {
        dataset: this.dataset,
      });
      const training = routeGetters.getDecodedTrainingVariableNames(
        this.$store
      );

      this.variables.forEach((variable) => {
        training.push(variable.colName);
      });

      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        training: training.join(","),
      });

      this.$router.push(entry).catch((err) => console.warn(err));
      await viewActions.updateLabelData(this.$store);
    },
    onAnnotationChanged(label: LowShotLabels) {
      const rowSelection = routeGetters.getDecodedRowSelection(this.$store);
      const innerData = new Map<number, unknown>();
      const updateData = rowSelection.d3mIndices.map((i) => {
        innerData.set(i, { LowShotLabel: label });
        return {
          index: i.toString(),
          name: LOW_SHOT_LABEL_COLUMN_NAME,
          value: label,
        };
      });
      datasetMutations.updateAreaOfInterestIncludeInner(this.$store, innerData);
      datasetActions.updateDataset(this.$store, {
        dataset: this.dataset,
        updateData,
      });
      clearRowSelection(this.$router);
      this.onDataChanged();
    },
    async updateRoute() {
      const taskResponse = await datasetActions.fetchTask(this.$store, {
        dataset: this.dataset,
        targetName: LOW_SHOT_LABEL_COLUMN_NAME,
        variableNames: this.variables.map((v) => v.colName),
      });
      // update route with training data
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        task: taskResponse.data.task.join(","),
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>
<style scoped>
.h-18 {
  height: 18% !important;
}
</style>
