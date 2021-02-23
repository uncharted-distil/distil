<template>
  <loading-spinner v-if="loading" :state="loadingState" />
  <div v-else class="row flex-1 pb-3 h-100">
    <div
      class="col-12 col-md-3 d-flex h-100 flex-column border-right border-color"
    >
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
        :summaries="featureSummaries"
        :pagination="
          featureSummaries && searchedActiveVariables.length > numRowsPerPage
        "
        :facet-count="featureSummaries && searchedActiveVariables.length"
        :rows-per-page="numRowsPerPage"
        :instance-name="instance"
      />
    </div>
    <div class="col-12 col-md-6 d-flex flex-column h-100">
      <div class="flex-1 d-flex flex-column pb-1 pt-2">
        <labeling-data-slot
          :summaries="summaries"
          :variables="variables"
          :has-confidence="hasConfidence"
          @DataChanged="onAnnotationChanged"
        />
        <create-labeling-form
          :is-loading="isLoadingData"
          :low-shot-summary="labelSummary"
          @export="onExport"
          @apply="onApply"
          @save="onSaveClick"
        />
      </div>
    </div>
    <div
      class="col-12 col-md-3 d-flex h-100 flex-column border-left border-color"
    >
      <h5 class="header-title">Scores</h5>
      <variable-facets
        enable-highlighting
        enable-type-filtering
        :summaries="scoreSummary"
        :instance-name="instance"
        class="h-18"
      />
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
    <save-dataset :dataset-name="dataset" @save="onSaveValid" />
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
import { CATEGORICAL_TYPE, DISTIL_ROLES } from "../util/types";
import {
  getVariableSummariesByState,
  searchVariables,
  NUM_PER_TARGET_PAGE,
  cloneDatasetUpdateRoute,
  LowShotLabels,
  LOW_SHOT_LABEL_COLUMN_NAME,
  LOW_SHOT_SCORE_COLUMN_NAME,
  minimumRouteKey,
  addOrderBy,
  downloadFile,
} from "../util/data";
import {
  Variable,
  VariableSummary,
  TableRow,
  TableColumn,
} from "../store/dataset/index";
import VariableFacets from "../components/facets/VariableFacets.vue";
import SaveDataset from "../components/labelingComponents/SaveDataset.vue";
import CreateLabelingForm from "../components/labelingComponents/CreateLabelingForm.vue";
import LabelingDataSlot from "../components/labelingComponents/LabelingDataSlot.vue";
import { EXCLUDE_FILTER, Filter, INCLUDE_FILTER } from "../util/filters";
import { Dictionary } from "vue-router/types/router";
import { updateHighlight, clearHighlight } from "../util/highlights";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { overlayRouteEntry } from "../util/routes";
import { actions as requestActions } from "../store/requests/module";
import { clearRowSelection } from "../util/row";
import LoadingSpinner from "../components/labelingComponents/LoadingSpinner.vue";

const LABEL_KEY = "label";

export default Vue.extend({
  name: "LabelingView",
  components: {
    VariableFacets,
    LabelingDataSlot,
    CreateLabelingForm,
    SaveDataset,
    LoadingSpinner,
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
      isLoadingData: false,
      scorePopUpId: "modal-score-pop-up",
      loading: true,
      loadingState: "",
      hasConfidence: false,
    };
  },
  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store).filter((v) => {
        return (
          v.distilRole !== DISTIL_ROLES.SystemData ||
          v.key !== LOW_SHOT_SCORE_COLUMN_NAME
        );
      });
    },
    scores(): Variable {
      return datasetGetters.getVariables(this.$store).find((v) => {
        return v.key === LOW_SHOT_SCORE_COLUMN_NAME;
      });
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
        (v) => !this.groupedFeatures.includes(v.key)
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
    // filters out the low shot labels
    featureSummaries(): VariableSummary[] {
      return this.summaries.filter((s) => {
        return (
          s.key !== LOW_SHOT_LABEL_COLUMN_NAME &&
          s.key !== LOW_SHOT_SCORE_COLUMN_NAME
        );
      });
    },
    scoreSummary(): VariableSummary[] {
      const score = this.summaries.find((s) => {
        return s.key === LOW_SHOT_SCORE_COLUMN_NAME;
      });
      return !score ? [] : [score];
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
    training(): string[] {
      return routeGetters.getDecodedTrainingVariableNames(this.$store);
    },
    isClone(): boolean | null {
      const datasets = datasetGetters.getDatasets(this.$store);
      const dataset = datasets.find((d) => d.id === this.dataset);
      if (!dataset) {
        return null;
      }
      return dataset.clone === undefined ? false : dataset.clone;
    },
  },
  watch: {
    highlight() {
      this.onDataChanged();
    },
    training(prev: string[], cur: string[]) {
      if (prev.length !== cur.length) {
        this.fetchData();
      }
    },
  },
  async mounted() {
    this.loadingState = "Fetching Data";
    await datasetActions.fetchDataset(this.$store, { dataset: this.dataset });
    await this.fetchData();
    this.loadingState = "Checking Clone";
    this.checkClone();
  },
  methods: {
    async checkClone() {
      if (this.isClone) {
        // dataset is already a clone don't clone again. (used for testing. might add button for cloning later.)
        this.updateRoute();
        this.loading = false;
        return;
      }
      const entry = await cloneDatasetUpdateRoute();
      if (entry === null) {
        return;
      }
      await this.$router.push(entry).catch((err) => console.warn(err));
      this.loading = false;
      this.$nextTick(() => {
        this.$bvModal.show(this.modalId);
      });
    },
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
      const res = (await requestActions.createQueryRequest(this.$store, {
        datasetId: this.dataset,
        target: LOW_SHOT_LABEL_COLUMN_NAME,
        filters: null,
      })) as { success: boolean; error: string };
      if (!res.success) {
        this.$bvToast.toast(res.error, {
          title: "Error",
          autoHideDelay: 5000,
          appendToast: true,
          variant: "danger",
          toaster: "b-toaster-bottom-right",
        });
      }
      addOrderBy(LOW_SHOT_SCORE_COLUMN_NAME);
      this.isLoadingData = false;
      await this.fetchData();
      this.hasConfidence = true;
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        annotationHasChanged: false,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    async onExport() {
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
      const file = await datasetActions.extractDataset(this.$store, {
        dataset: this.dataset,
        filterParams,
        highlight,
        include: true,
        mode: EXCLUDE_FILTER,
        dataMode,
      });
      downloadFile(file, this.dataset, ".csv");
    },
    onSaveClick() {
      this.$bvModal.show("save-model-modal");
    },
    async onSaveValid(saveName: string) {
      const highlight = {
        context: this.instance,
        dataset: this.dataset,
        key: LOW_SHOT_LABEL_COLUMN_NAME,
        value: LowShotLabels.unlabeled,
      }; // exclude unlabeled from data export
      const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
        this.$store
      );
      if (
        this.variables.some((v) => {
          return v.key === LOW_SHOT_SCORE_COLUMN_NAME;
        })
      ) {
        // delete confidence variable when saving
        await datasetActions.deleteVariable(this.$store, {
          dataset: this.dataset,
          key: LOW_SHOT_SCORE_COLUMN_NAME,
        });
      }
      const dataMode = routeGetters.getDataMode(this.$store);
      await datasetActions.saveDataset(this.$store, {
        dataset: this.dataset,
        datasetNewName: saveName,
        filterParams,
        highlight,
        include: true,
        mode: INCLUDE_FILTER,
        dataMode,
      });
      this.$bvModal.show("save-success-dataset");
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
      if (!updateData.length) {
        return;
      }
      datasetMutations.updateAreaOfInterestIncludeInner(this.$store, innerData);
      datasetActions.updateDataset(this.$store, {
        dataset: this.dataset,
        updateData,
      });
      clearRowSelection(this.$router);
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        annotationHasChanged: true,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
      this.onDataChanged();
    },
    async updateRoute() {
      const taskResponse = await datasetActions.fetchTask(this.$store, {
        dataset: this.dataset,
        targetName: LOW_SHOT_LABEL_COLUMN_NAME,
        variableNames: this.variables.map((v) => v.key),
      });
      const training = routeGetters.getDecodedTrainingVariableNames(
        this.$store
      );
      const check = training.length;
      const trainingMap = new Map(
        training.map((t) => {
          return [t, true];
        })
      );
      this.variables.forEach((variable) => {
        if (!trainingMap.has(variable.key)) {
          training.push(variable.key);
        }
      });
      if (check === training.length) {
        return;
      }
      // update route with training data
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        task: taskResponse.data.task.join(","),
        training: training.join(","),
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
.border-color {
  border-color: var(--gray-500) !important;
}
</style>
