<template>
  <div class="variable-facets row h-100">
    <div :class="variableFacetListClass + ' col-12 flex-column d-flex'">
      <div v-if="enableSearch" class="row align-items-center facet-filters">
        <div class="col-12 flex-column d-flex">
          <b-form-input size="sm" v-model="search" placeholder="Search" />
        </div>
      </div>
      <!-- TODO: this should be passed in as title HTML -->
      <div v-if="enableTitle" class="row align-items-center">
        <div class="col-12 flex-column d-flex">
          <p>
            <b>Select Feature to Predict</b> Select from potential features of
            interest below. Each feature tile shown summarizes count of records
            by value.
          </p>
        </div>
      </div>
      <div class="pl-1 pr-1">
        <!-- injectable slot -->
        <slot></slot>
      </div>
      <div class="row flex-1 variable-facets-wrapper">
        <div class="col-12 flex-column variable-facets-container">
          <div
            class="variable-facets-item"
            v-for="summary in summaries"
            :key="summary.key"
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
                :summary="summary"
                :highlight="highlight"
                :row-selection="rowSelection"
                :html="html"
                :enabled-type-changes="enabledTypeChanges"
                :enable-highlighting="[enableHighlighting, enableHighlighting]"
                :ignore-highlights="[ignoreHighlights, ignoreHighlights]"
                :instanceName="instanceName"
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
                :isAvailableFeatures="isAvailableFeatures"
                :isFeaturesToModel="isFeaturesToModel"
                :log-activity="logActivity"
                :expanded="expandGeoAndTimeseriesFacets"
                @histogram-numerical-click="onNumericalClick"
                @histogram-range-change="onRangeChange"
              />
            </template>
            <template v-else-if="isImage(summary.varType)">
              <facet-image
                :summary="summary"
                :highlight="highlight"
                :row-selection="rowSelection"
                :ranking="ranking[summary.key]"
                :html="html"
                :enabled-type-changes="enabledTypeChanges"
                :enable-highlighting="enableHighlighting"
                :ignore-highlights="ignoreHighlights"
                :instanceName="instanceName"
                @facet-click="onFacetClick"
              />
            </template>
            <template v-else-if="summary.varType === 'dateTime'">
              <facet-date-time
                :summary="summary"
                :highlight="highlight"
                :row-selection="rowSelection"
                :importance="ranking[summary.key]"
                :ranking="ranking[summary.key]"
                :html="html"
                :enabled-type-changes="enabledTypeChanges"
                :enable-highlighting="enableHighlighting"
                :ignore-highlights="ignoreHighlights"
                :instanceName="instanceName"
                @facet-click="onFacetClick"
              />
            </template>
            <template v-else-if="summary.type === 'categorical'">
              <facet-categorical
                :summary="summary"
                :highlight="highlight"
                :row-selection="rowSelection"
                :importance="ranking[summary.key]"
                :html="html"
                :enabled-type-changes="enabledTypeChanges"
                :enable-highlighting="enableHighlighting"
                :ignore-highlights="ignoreHighlights"
                :instanceName="instanceName"
                @facet-click="onFacetClick"
              />
            </template>
            <template v-else-if="summary.type === 'numerical'">
              <facet-numerical
                :summary="summary"
                :highlight="highlight"
                :row-selection="rowSelection"
                :importance="ranking[summary.key]"
                :html="html"
                :enabled-type-changes="enabledTypeChanges"
                :enable-highlighting="enableHighlighting"
                :ignore-highlights="ignoreHighlights"
                :instanceName="instanceName"
                @numerical-click="onNumericalClick"
                @categorical-click="onCategoricalClick"
                @range-change="onRangeChange"
                @facet-click="onFacetClick"
              />
            </template>
          </div>
        </div>
      </div>
    </div>
    <div
      v-if="pagination"
      class="col-12 row align-items-center variable-page-nav"
    >
      <div class="col-12 flex-column">
        <b-pagination
          v-if="pagination"
          size="sm"
          align="center"
          :total-rows="facetCount"
          :per-page="rowsPerPage"
          v-model="currentPage"
          class="mb-0"
        />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
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
  getVariableRanking,
  getSolutionFeatureImportance,
  NUM_PER_PAGE,
} from "../../util/data";
import {
  Highlight,
  RowSelection,
  Variable,
  VariableSummary,
} from "../../store/dataset";
import { getters as datasetGetters } from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import {
  ROUTE_PAGE_SUFFIX,
  ROUTE_SEARCH_SUFFIX,
} from "../../store/route/index";
import { isGeoLocatedType, isImageType } from "../../util/types";
import { actions as appActions } from "../../store/app/module";
import { Feature, Activity, SubActivity } from "../../util/userEvents";
import { updateHighlight, clearHighlight } from "../../util/highlights";
import Vue from "vue";

export default Vue.extend({
  name: "variable-facets",

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
    enableHighlighting: Boolean as () => boolean,
    enableSearch: Boolean as () => boolean,
    enableTitle: Boolean as () => boolean,
    enableTypeChange: Boolean as () => boolean,
    enableTypeFiltering: Boolean as () => boolean,
    facetCount: Number as () => number,
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function,
    ],
    instanceName: { type: String as () => string, default: "variableFacets" },
    isAvailableFeatures: Boolean as () => boolean,
    isFeaturesToModel: Boolean as () => boolean,
    isResultFeatures: Boolean as () => boolean,
    ignoreHighlights: Boolean as () => boolean,
    logActivity: {
      type: String as () => Activity,
      default: Activity.DATA_PREPARATION,
    },
    pagination: Boolean as () => boolean,
    summaries: Array as () => VariableSummary[],
    subtitle: String as () => string,
    rowsPerPage: { type: Number as () => number, default: NUM_PER_PAGE },
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
        this.$router.push(entry).catch((err) => console.warn(err));
      },
      get(): number {
        return getRouteFacetPage(this.routePageKey(), this.$route);
      },
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
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
        ranking[variable.colName] = this.isResultFeatures
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
        if (this.enableTypeChange && !this.isSeriesID(variable.colName)) {
          typeChangeStatus.push(`${variable.datasetName}:${variable.colName}`);
        }
      });
      return typeChangeStatus;
    },

    variableFacetListClass(): string {
      return this.pagination
        ? "variable-facets-list-with-footer"
        : "variable-facets-list";
    },

    expandGeoAndTimeseriesFacets(): Boolean {
      // The Geocoordinates and Timeseries Facets are expanded on SELECT_TARGET_ROUTE
      return routeGetters.isPageSelectTarget(this.$store);
    },
  },

  methods: {
    // creates a facet key for the route from the instance-name component arg
    // or uses a default if unset
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

    isSeriesID(colName: string): boolean {
      // Check to see if this facet is being used as a series ID
      const targetVar = routeGetters.getTargetVariable(this.$store);
      if (targetVar && targetVar.grouping) {
        if (targetVar.grouping.subIds.length > 0) {
          return !!targetVar.grouping.subIds.find((v) => v === colName);
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
  beforeMount() {
    this.search = routeGetters.getAllSearchesByQueryString(this.$store)[
      this.routeSearchKey()
    ];
  },
  watch: {
    search() {
      const entry = overlayRouteEntry(this.$route, {
        [this.routeSearchKey()]: this.search,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>

<style>
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

.variable-facets-container .variable-facets-item {
  margin: 0.5rem 0;
  vertical-align: bottom;
}

.variable-facets-container .facets-root-container .facets-group-container {
  background-color: inherit;
}

.variable-facets-container
  .facets-root-container
  .facets-group-container
  .facets-group {
  background: var(--white);
  font-size: 0.867rem;
  color: var(--color-text-base);
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.1);
  transition: box-shadow 0.3s ease-in-out;
}

.variable-facets-container
  .facets-root-container
  .facets-group-container
  .facets-group
  .group-header {
  margin: 0px !important;
  padding: 0 0 4px !important;
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  justify-content: flex-start;
  align-content: center;
  align-items: stretch;
}

.variable-facets-container
  .facets-root-container
  .facets-group-container
  .facets-group
  .group-header
  .header-text {
  order: 0;
  flex: 1 1 auto;
  align-self: auto;
  overflow: hidden;
  text-overflow: ellipsis;
  margin: 0 0 0 0.5rem;
  height: 20px;
  white-space: nowrap;
}

.variable-facets-container
  .facets-root-container
  .facets-group-container
  .facets-group
  .group-header
  .fa-info {
  margin: 5px 10px 5px 5px;
  order: 1;
  flex: 30 1 auto;
  align-self: auto;
  text-align: left;
}

.variable-facets-container
  .facets-root-container
  .facets-group-container
  .facets-group
  .group-header
  .type-change-menu {
  order: 2;
  flex: none;
  align-self: auto;
  text-align: right;
}

.variable-facets-container .dropdown-menu {
  max-height: 200px;
  overflow-y: auto;
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

.variable-facets-container .facet-header-container {
  overflow-y: scroll !important;
}

.variable-facets-container .facet-header-container .dropdown-menu {
  max-height: 200px;
  overflow-y: auto;
}

.variable-facets-list-with-footer {
  max-height: calc(100% - 45px);
}

.variable-facets-list {
  max-height: 100%;
}
</style>
