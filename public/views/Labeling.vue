<template>
  <div class="row flex-1 pb-3 h-100">
    <div class="col-12 col-md-3 d-flex h-100 flex-column">
      <h5 class="header-title">Labels</h5>
      <div class="mb-5">
        <facet-categorical
          enable-highlighting
          enable-type-filtering
          :summary="labelSummary"
          @facet-click="onFacetClick"
        />
      </div>
      <h5 class="header-title">Features</h5>
      <variable-facets
        enable-highlighting
        enable-type-filtering
        :summaries="summaries"
        :pagination="
          summaries && searchedActiveVariables.length > numRowsPerPage
        "
        :facetCount="summaries && searchedActiveVariables.length"
        :rows-per-page="numRowsPerPage"
        :instanceName="instance"
      />
    </div>
    <div class="col-12 col-md-6 d-flex flex-column h-100">
      <div class="flex-1 d-flex flex-column pb-1 pt-2">
        <labeling-data-slot
          :summaries="summaries"
          :variables="variables"
          @DataChanged="onDataChanged"
        />
        <create-labeling-form @export="onExport" />
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
          required
          :placeholder="labelName"
        />
      </b-form-group>
    </b-modal>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { getters as routeGetters } from "../store/route/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
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
} from "../util/data";
import { Variable, VariableSummary } from "../store/dataset/index";
import { CATEGORICAL_TYPE } from "../util/types";
import VariableFacets from "../components/facets/VariableFacets.vue";
import FacetCategorical from "../components/facets/FacetCategorical.vue";
import CreateLabelingForm from "../components/labelingComponents/CreateLabelingForm.vue";
import LabelingDataSlot from "../components/labelingComponents/LabelingDataSlot.vue";
import { EXCLUDE_FILTER } from "../util/filters";
import { Dictionary } from "vue-router/types/router";
import { updateHighlight, clearHighlight } from "../util/highlights";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { overlayRouteEntry } from "../util/routes";
const LABEL_KEY = "label";

export default Vue.extend({
  name: "labeling-view",
  components: {
    VariableFacets,
    LabelingDataSlot,
    CreateLabelingForm,
    FacetCategorical,
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
    instance(): string {
      return LABEL_FEATURE_INSTANCE;
    },
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
      await datasetActions.fetchVariables(this.$store, {
        dataset: this.dataset,
      });
      await viewActions.updateLabelData(this.$store);
    },
    async onExport() {
      const highlight = {
        context: "VariableFacets",
        dataset: this.dataset,
        key: LOW_SHOT_LABEL_COLUMN_NAME,
        value: LowShotLabels.unlabeled,
      }; // exclude unlabled from data export
      const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
        this.$store
      );
      const dataMode = routeGetters.getDataMode(this.$store);
      const response = await datasetActions.extractDataset(this.$store, {
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
          context: context,
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
      await datasetActions.fetchVariables(this.$store, {
        dataset: this.dataset,
      });
      // pull the cloned data
      viewActions.updateLabelData(this.$store);
      // update task based on the current training data
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
  watch: {
    highlight() {
      this.onDataChanged();
    },
  },
  async mounted() {
    this.$bvModal.show(this.modalId);
    const entry = await cloneDatasetUpdateRoute();
    if (entry === null) {
      return;
    }
    this.$router.push(entry).catch((err) => console.warn(err));
  },
});
</script>
