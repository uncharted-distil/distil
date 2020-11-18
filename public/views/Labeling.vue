<template>
  <div class="row flex-1 pb-3 h-100">
    <div class="col-12 col-md-3 d-flex h-100 flex-column">
      <h5 class="header-title">Labels</h5>
      <div class="mb-5">
        <facet-categorical :summary="labelSummary" />
      </div>
      <h5 class="header-title">Features</h5>
      <variable-facets
        enable-highlighting
        enable-type-filtering
        :summaries="summaries"
      />
    </div>
    <div class="col-12 col-md-6 d-flex flex-column h-100">
      <div class="h-80 flex-1 d-flex flex-column pb-1 pt-2">
        <labeling-data-slot
          :summaries="summaries"
          :variables="variables"
          @DataChanged="onDataChanged"
        />
        <create-labeling-form @export="onExport" />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { getters as routeGetters } from "../store/route/module";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../store/dataset/module";
import { actions as viewActions } from "../store/view/module";
import {
  getVariableSummariesByState,
  searchVariables,
  NUM_PER_TARGET_PAGE,
  cloneDatasetUpdateRoute,
  LowShotLabels,
  LOW_SHOT_LABEL_COLUMN_NAME,
} from "../util/data";
import { Variable, VariableSummary } from "../store/dataset/index";
import VariableFacets from "../components/facets/VariableFacets.vue";
import FacetCategorical from "../components/facets/FacetCategorical.vue";
import CreateLabelingForm from "../components/labelingComponents/CreateLabelingForm.vue";
import { MULTIBAND_IMAGE_TYPE } from "../util/types";
import LabelingDataSlot from "../components/labelingComponents/LabelingDataSlot.vue";
import { EXCLUDE_FILTER } from "../util/filters";
const LABEL_KEY = "label";

export default Vue.extend({
  name: "labeling-view",
  components: {
    VariableFacets,
    LabelingDataSlot,
    CreateLabelingForm,
    FacetCategorical,
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
    labelSummary(): VariableSummary {
      const lowShotLabel = this.summaries.filter((s) => {
        return s.key === LOW_SHOT_LABEL_COLUMN_NAME;
      });
      return lowShotLabel.length
        ? lowShotLabel[0]
        : this.getDefaultLabelFacet();
    },
    numOfMultiBands(): number {
      const multiBandSummary = this.summaries.filter((s) => {
        return s.varType === MULTIBAND_IMAGE_TYPE;
      });
      return multiBandSummary.length
        ? multiBandSummary[0].baseline.buckets.length
        : 0;
    },
    summaries(): VariableSummary[] {
      const pageIndex = routeGetters.getRouteAvailableTargetVarsPage(
        this.$store
      );

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
            { key: LowShotLabels.unlabeled, count: this.numOfMultiBands },
          ],
          extrema: { min: 0, max: this.numOfMultiBands },
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
      // TODO download csv from response
    },
  },
  watch: {
    highlight() {
      this.onDataChanged();
    },
  },
  async created() {
    const entry = await cloneDatasetUpdateRoute();
    if (entry === null) {
      return;
    }
    this.$router.push(entry).catch((err) => console.warn(err));
    // add new field
    await datasetActions.addField<string>(this.$store, {
      dataset: this.dataset,
      name: LOW_SHOT_LABEL_COLUMN_NAME,
      fieldType: typeof LowShotLabels.unlabeled,
      defaultValue: LowShotLabels.unlabeled,
    });
    await datasetActions.fetchVariables(this.$store, {
      dataset: this.dataset,
    });
    // pull the cloned data
    viewActions.updateLabelData(this.$store);
  },
});
</script>
