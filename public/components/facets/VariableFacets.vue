<template>
  <div class="d-flex flex-column align-items-stretch h-100 w-100">
    <div v-if="enableSearch" class="py-1">
      <b-form-input v-model="search" size="sm" placeholder="Search" />
    </div>
    <!-- TODO: this should be passed in as title HTML -->
    <div v-if="enableTitle" class="py-1">
      <p>
        <b>Select Feature to Predict</b> Select from potential features of
        interest below. Each feature tile shown summarizes count of records by
        value.
      </p>
    </div>
    <div>
      <!-- injectable slot -->
      <slot />
    </div>
    <div
      class="my-2 flex-fill w-100 variable-facets-wrapper flex-wrap justify-content-between"
    >
      <div class="variable-facets-container">
        <div
          v-for="summary in summaries"
          :key="summary.key"
          class="variable-facets-item flex-fill my-2 mx-1"
        >
          <template v-if="summary.pending">
            <facet-loading :summary="summary" />
          </template>
          <template v-else-if="summary.err">
            <facet-error
              :summary="summary"
              :enabled-type-changes="enabledTypeChanges"
            />
          </template>
          <template v-else-if="summary.varType === 'timeseries'">
            <facet-timeseries
              :style="facetColors"
              :summary="summary"
              :highlight="highlight"
              :row-selection="rowSelection"
              :html="html"
              :enabled-type-changes="enabledTypeChanges"
              :enable-highlighting="[enableHighlighting, enableHighlighting]"
              :ignore-highlights="[ignoreHighlights, ignoreHighlights]"
              :instance-name="instanceName"
              :expanded="expandGeoAndTimeseriesFacets"
              @numerical-click="onNumericalClick"
              @categorical-click="onCategoricalClick"
              @facet-click="onFacetClick"
              @range-change="onRangeChange"
              @histogram-numerical-click="onNumericalClick"
              @histogram-categorical-click="onCategoricalClick"
              @histogram-range-change="onRangeChange"
            />
          </template>
          <template v-else-if="isGeoLocated(summary.varType)">
            <geocoordinate-facet
              :summary="summary"
              :enable-highlighting="enableHighlighting"
              :ignore-highlights="ignoreHighlights"
              :is-available-features="isAvailableFeatures"
              :is-features-to-model="isFeaturesToModel"
              :log-activity="logActivity"
              :expanded="expandGeoAndTimeseriesFacets"
              @histogram-numerical-click="onNumericalClick"
              @histogram-range-change="onRangeChange"
            />
          </template>
          <template v-else-if="isImage(summary.varType)">
            <facet-image
              :style="facetColors"
              :summary="summary"
              :highlight="highlight"
              :row-selection="rowSelection"
              :ranking="ranking[summary.key]"
              :html="html"
              :enabled-type-changes="enabledTypeChanges"
              :enable-highlighting="enableHighlighting"
              :ignore-highlights="ignoreHighlights"
              :instance-name="instanceName"
              @facet-click="onFacetClick"
            />
          </template>
          <template v-else-if="summary.varType === 'dateTime'">
            <facet-date-time
              :style="facetColors"
              :summary="summary"
              :highlight="highlight"
              :row-selection="rowSelection"
              :importance="ranking[summary.key]"
              :ranking="ranking[summary.key]"
              :html="html"
              :enabled-type-changes="enabledTypeChanges"
              :enable-highlighting="enableHighlighting"
              :ignore-highlights="ignoreHighlights"
              :instance-name="instanceName"
              @facet-click="onFacetClick"
              @range-change="onRangeChange"
            />
          </template>
          <template v-else-if="summary.type === 'categorical'">
            <facet-categorical
              :style="facetColors"
              :summary="summary"
              :highlight="highlight"
              :row-selection="rowSelection"
              :importance="ranking[summary.key]"
              :html="html"
              :enabled-type-changes="enabledTypeChanges"
              :enable-highlighting="enableHighlighting"
              :ignore-highlights="ignoreHighlights"
              :instance-name="instanceName"
              @facet-click="onFacetClick"
            />
          </template>
          <template v-else-if="summary.type === 'numerical'">
            <facet-numerical
              :style="facetColors"
              :summary="summary"
              :highlight="highlight"
              :row-selection="rowSelection"
              :importance="ranking[summary.key]"
              :html="html"
              :enabled-type-changes="enabledTypeChanges"
              :enable-highlighting="enableHighlighting"
              :ignore-highlights="ignoreHighlights"
              :instance-name="instanceName"
              @numerical-click="onNumericalClick"
              @categorical-click="onCategoricalClick"
              @range-change="onRangeChange"
              @facet-click="onFacetClick"
            />
          </template>
        </div>
      </div>
    </div>
    <div v-if="pagination" class="p-1">
      <div class="flex-fill">
        <b-pagination
          v-if="pagination"
          v-model="currentPage"
          align="center"
          first-number
          last-number
          size="sm"
          :total-rows="facetCount"
          :per-page="rowsPerPage"
          class="mb-0"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import {
  getVariableRanking,
  getSolutionFeatureImportance,
  NUM_PER_PAGE,
} from "../../util/data";
import FacetImage from "./FacetImage.vue";
import FacetDateTime from "./FacetDateTime.vue";
import FacetTimeseries from "./FacetTimeseries.vue";
import FacetCategorical from "./FacetCategorical.vue";
import FacetNumerical from "./FacetNumerical.vue";
import FacetLoading from "./FacetLoading.vue";
import FacetError from "./FacetError.vue";
import GeocoordinateFacet from "./GeocoordinateFacet.vue";
import { overlayRouteEntry, getRouteFacetPage } from "../../util/routes";
import { Dictionary } from "../../util/dict";
import {
  applyColor,
  FACET_COLOR_SELECT,
  FACET_COLOR_EXCLUDE,
  FACET_COLOR_FILTERED,
} from "../../util/facets";
import {
  Highlight,
  RowSelection,
  TimeseriesGrouping,
  Variable,
  VariableSummary,
} from "../../store/dataset";
import {
  getters as datasetGetters,
  actions as datasetActions,
} from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import {
  ROUTE_PAGE_SUFFIX,
  ROUTE_SEARCH_SUFFIX,
} from "../../store/route/index";
import {
  isGeoLocatedType,
  isImageType,
  TIMESERIES_TYPE,
} from "../../util/types";
import { actions as appActions } from "../../store/app/module";
import { Feature, Activity, SubActivity } from "../../util/userEvents";
import { updateHighlight, clearHighlight } from "../../util/highlights";
import Vue from "vue";

export default Vue.extend({
  name: "VariableFacets",

  components: {
    FacetImage,
    FacetDateTime,
    FacetTimeseries,
    GeocoordinateFacet,
    FacetCategorical,
    FacetNumerical,
    FacetLoading,
    FacetError,
  },

  props: {
    enableHighlighting: Boolean,
    enableSearch: Boolean,
    enableTitle: Boolean,
    enableTypeChange: Boolean,
    enableTypeFiltering: Boolean,
    facetCount: { type: Number, default: 0 },
    html: { type: [String, Object, Function], default: null },
    instanceName: { type: String, default: "variableFacets" },
    isAvailableFeatures: Boolean,
    isFeaturesToModel: Boolean,
    isResultFeatures: Boolean,
    ignoreHighlights: Boolean,
    logActivity: {
      type: String as () => Activity,
      default: Activity.DATA_PREPARATION,
    },
    summaries: { type: Array as () => VariableSummary[], default: [] },
    subtitle: { type: String, default: null },
    rowsPerPage: { type: Number, default: 0 },
  },

  data() {
    return {
      search: "",
    };
  },

  computed: {
    currentPage: {
      set(page: number) {
        const entry = overlayRouteEntry(this.$route, {
          [this.routePageKey()]: page,
        });
        this.$router.push(entry).catch((err) => console.debug(err));
        this.$emit("page", page);
      },
      get(): number {
        return getRouteFacetPage(this.routePageKey(), this.$route);
      },
    },
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    timeseriesSummaries(): VariableSummary[] {
      return this.summaries.filter((s) => {
        return s.varType === TIMESERIES_TYPE;
      });
    },
    timeseriesVars(): Variable[] | null {
      const checkMap = new Map(
        this.timeseriesSummaries.map((ts) => {
          return [ts.key, true];
        })
      );
      return this.variables.filter((v) => {
        return checkMap.has(v.key);
      });
    },
    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    include(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },

    ranking(): Dictionary<number> {
      // Only show ranks for available feature, model features and result features
      if (
        !this.isAvailableFeatures &&
        !this.isFeaturesToModel &&
        !this.isResultFeatures
      ) {
        return {};
      }
      const ranking: Dictionary<number> = {};
      this.variables.forEach((variable) => {
        ranking[variable.key] = this.isResultFeatures
          ? getSolutionFeatureImportance(
              variable,
              routeGetters.getRouteSolutionId(this.$store)
            )
          : getVariableRanking(variable);
      });
      return ranking;
    },

    enabledTypeChanges(): string[] {
      const typeChangeStatus: string[] = [];
      this.variables.forEach((variable) => {
        if (this.enableTypeChange && !this.isSeriesID(variable.key)) {
          typeChangeStatus.push(`${variable.datasetName}:${variable.key}`);
        }
      });
      return typeChangeStatus;
    },

    expandGeoAndTimeseriesFacets(): boolean {
      // The Geocoordinate and Timeseries Facets are expanded on SELECT_TARGET_ROUTE
      return !!routeGetters.isPageSelectTarget(this.$store);
    },

    facetColors(): string {
      return applyColor([
        null,
        !!this.rowSelection ? FACET_COLOR_SELECT : null,
        !this.include ? FACET_COLOR_EXCLUDE : null,
        FACET_COLOR_FILTERED,
      ]);
    },
    numFacetPerPage(): number {
      return !this.rowsPerPage ? NUM_PER_PAGE : this.rowsPerPage;
    },
    pagination(): boolean {
      return this.facetCount > this.numFacetPerPage;
    },
  },

  watch: {
    async timeseriesSummaries() {
      if (this.timeseriesSummaries.length) {
        this.timeseriesSummaries.forEach(async (ts) => {
          const ids = ts.baseline.exemplars;
          const timeseriesVar = this.timeseriesVars.find((tsv) => {
            return tsv.key === ts.key;
          });
          const grouping = timeseriesVar.grouping as TimeseriesGrouping;
          await datasetActions.fetchTimeseries(this.$store, {
            dataset: this.dataset,
            variableKey: timeseriesVar.key,
            xColName: grouping.xCol,
            yColName: grouping.yCol,
            timeseriesIds: ids,
          });
        });
      }
    },
    search(newTerm, oldTerm) {
      if (!newTerm || newTerm === oldTerm) return;

      const entry = overlayRouteEntry(this.$route, {
        [this.routeSearchKey()]: this.search,
      });
      this.$router.push(entry).catch((err) => console.warn(err));

      // If the term searched has been updated, we emit an event.
      this.$emit("search", this.search);
    },
  },

  beforeMount() {
    this.search = routeGetters.getAllSearchesByQueryString(this.$store)[
      this.routeSearchKey()
    ];
  },

  methods: {
    // creates a facet key for the route from the instance-name component arg
    routePageKey(): string {
      return `${this.instanceName}${ROUTE_PAGE_SUFFIX}`;
    },
    routeSearchKey(): string {
      return `${this.instanceName}${ROUTE_SEARCH_SUFFIX}`;
    },

    onRangeChange(
      context: string,
      key: string,
      value: { from: number; to: number },
      dataset: string
    ) {
      updateHighlight(this.$router, {
        context: context,
        dataset: dataset,
        key: key,
        value: value,
      });
      this.$emit("range-change", key, value);
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: this.logActivity,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: { key: key, value: value },
      });
    },

    onFacetClick(context: string, key: string, value: string, dataset: string) {
      if (this.enableHighlighting) {
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
      }
      this.$emit("facet-click", context, key, value);
    },

    onCategoricalClick(context: string, key: string) {
      this.$emit("categorical-click", key);
    },

    onNumericalClick(
      context: string,
      key: string,
      value: { from: number; to: number; type: string },
      dataset: string
    ) {
      if (this.enableHighlighting) {
        if (!this.highlight || this.highlight.key !== key) {
          updateHighlight(this.$router, {
            context: this.instanceName,
            dataset: dataset,
            key: key,
            value: value,
          });
        }
      }
      this.$emit("numerical-click", key);
    },

    isSeriesID(key: string): boolean {
      // Check to see if this facet is being used as a series ID
      const targetVar = routeGetters.getTargetVariable(this.$store);
      if (targetVar && targetVar.grouping) {
        if (targetVar.grouping.subIds.length > 0) {
          return !!targetVar.grouping.subIds.find((v) => v === key);
        }
      }
      return false;
    },

    isGeoLocated(location: string): boolean {
      return isGeoLocatedType(location);
    },

    isImage(type: string): boolean {
      return isImageType(type);
    },
  },
});
</script>

<style scoped>
button {
  cursor: pointer;
}

.facet-terms-container {
  max-height: 200px !important;
  overflow-y: auto;
}

.page-link {
  color: var(--gray-600);
}

.page-item.active .page-link {
  z-index: 2;
  color: var(--white);
  background-color: var(--gray-700);
  border-color: var(--gray-700);
}

/* To display scrollbars on the list of variables facets. */
.variable-facets-wrapper {
  overflow-x: hidden;
  overflow-y: auto;
}

.variable-facets-wrapper .variable-facets-item {
  vertical-align: bottom;
}
.facet-filters {
  margin: 0 -10px 4px -10px;
}

.facet-filters span {
  font-size: 0.9rem;
}

.variable-page-nav {
  padding-top: 10px;
}

.geocoordinate {
  max-width: 500px;
  height: 300px;
}
</style>
