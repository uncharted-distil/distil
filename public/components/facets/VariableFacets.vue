<template>
  <div class="variable-facets row">
    <div class="col-12 flex-column d-flex h-100">
      <div v-if="enableSearch" class="row align-items-center facet-filters">
        <div class="col-12 flex-column d-flex">
          <b-form-input size="sm" v-model="filter" placeholder="Search" />
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
            v-for="summary in paginatedSummaries"
            :key="summary.key"
          >
            <template v-if="summary.pending">
              <facet-loading :summary="summary"> </facet-loading>
            </template>
            <template v-else-if="summary.err">
              <facet-error
                :summary="summary"
                :enabled-type-changes="enabledTypeChanges"
              ></facet-error>
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
                @numerical-click="onNumericalClick"
                @categorical-click="onCategoricalClick"
                @facet-click="onFacetClick"
                @range-change="onRangeChange"
                @histogram-numerical-click="onNumericalClick"
                @histogram-categorical-click="onCategoricalClick"
                @histogram-range-change="onRangeChange"
              >
              </facet-timeseries>
            </template>
            <template v-else-if="isGeoLocated(summary.varType)">
              <geocoordinate-facet
                :summary="summary"
                :enable-highlighting="enableHighlighting"
                :ignore-highlights="ignoreHighlights"
                :isAvailableFeatures="isAvailableFeatures"
                :isFeaturesToModel="isFeaturesToModel"
                :log-activity="logActivity"
                @histogram-numerical-click="onNumericalClick"
                @histogram-range-change="onRangeChange"
              >
              </geocoordinate-facet>
            </template>
            <template
              v-else-if="
                summary.varType === 'image' ||
                  summary.varType === 'multiband_image'
              "
            >
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
              >
              </facet-image>
            </template>
            <template v-else-if="summary.varType === 'dateTime'">
              <facet-date-time
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
              >
              </facet-date-time>
            </template>
            <template v-else-if="summary.type === 'categorical'">
              <facet-categorical
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
              >
              </facet-categorical>
            </template>
            <template v-else-if="summary.type === 'numerical'">
              <facet-numerical
                :summary="summary"
                :highlight="highlight"
                :row-selection="rowSelection"
                :ranking="ranking[summary.key]"
                :html="html"
                :enabled-type-changes="enabledTypeChanges"
                :enable-highlighting="enableHighlighting"
                :ignore-highlights="ignoreHighlights"
                :instanceName="instanceName"
                @numerical-click="onNumericalClick"
                @categorical-click="onCategoricalClick"
                @range-change="onRangeChange"
                @facet-click="onFacetClick"
              >
              </facet-numerical>
            </template>
          </div>
        </div>
      </div>
      <div
        v-if="numSummaries > rowsPerPage"
        class="row align-items-center variable-page-nav"
      >
        <div class="col-12 flex-column">
          <b-pagination
            size="sm"
            align="center"
            :total-rows="numSummaries"
            :per-page="rowsPerPage"
            v-model="currentPage"
            class="mb-0"
          />
        </div>
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
  sortSummariesByImportance,
  filterVariablesByPage,
  getVariableRanking,
  getVariableImportance
} from "../../util/data";
import {
  Highlight,
  RowSelection,
  Variable,
  VariableSummary
} from "../../store/dataset";
import {
  getters as datasetGetters,
  actions as datasetActions
} from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import { ROUTE_PAGE_SUFFIX } from "../../store/route/index";
import { Group } from "../../util/facets";
import {
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  isLocationType,
  isGeoLocatedType
} from "../../util/types";
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
    FacetError
  },

  props: {
    enableSearch: Boolean as () => boolean,
    enableTitle: Boolean as () => boolean,
    enableTypeChange: Boolean as () => boolean,
    enableHighlighting: Boolean as () => boolean,
    enableTypefiltering: Boolean as () => boolean,
    isAvailableFeatures: Boolean as () => boolean,
    isFeaturesToModel: Boolean as () => boolean,
    ignoreHighlights: Boolean as () => boolean,
    summaries: Array as () => VariableSummary[],
    subtitle: String as () => string,
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function
    ],
    instanceName: { type: String as () => string, default: "variableFacets" },
    rowsPerPage: { type: Number as () => number, default: 10 },
    logActivity: {
      type: String as () => Activity,
      default: Activity.DATA_PREPARATION
    }
  },

  data() {
    return {
      filter: ""
    };
  },

  computed: {
    currentPage: {
      set(page: number) {
        const entry = overlayRouteEntry(this.$route, {
          [this.routePageKey()]: page
        });
        this.$router.push(entry);
      },
      get(): number {
        return getRouteFacetPage(this.routePageKey(), this.$route);
      }
    },

    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },

    filteredSummaries(): VariableSummary[] {
      return this.summaries.filter(summary => {
        return (
          this.filter === "" ||
          summary.key.toLowerCase().includes(this.filter.toLowerCase())
        );
      });
    },

    sortedFilteredSummaries(): VariableSummary[] {
      return sortSummariesByImportance(this.filteredSummaries, this.variables);
    },

    paginatedSummaries(): VariableSummary[] {
      const filteredVariables = filterVariablesByPage(
        this.currentPage,
        this.rowsPerPage,
        this.sortedFilteredSummaries
      );
      return filteredVariables;
    },

    numSummaries(): number {
      return this.filteredSummaries.length;
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    ranking(): Dictionary<number> {
      const ranking: Dictionary<number> = {};
      this.variables.forEach(variable => {
        ranking[variable.colName] = getVariableRanking(variable);
      });
      return ranking;
    },

    enabledTypeChanges(): string[] {
      const typeChangeStatus: string[] = [];
      this.variables.forEach(variable => {
        if (this.enableTypeChange && !this.isSeriesID(variable.colName)) {
          const datasetName = routeGetters.getRouteDataset(this.$store);
          typeChangeStatus.push(`${variable.datasetName}:${variable.colName}`);
        }
      });
      return typeChangeStatus;
    }
  },

  methods: {
    // creates a facet key for the route from the instance-name component arg
    // or uses a default if unset
    routePageKey(): string {
      return `${this.instanceName}${ROUTE_PAGE_SUFFIX}`;
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
        value: value
      });
      this.$emit("range-change", key, value);
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: this.logActivity,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: { key: key, value: value }
      });
    },

    onFacetClick(context: string, key: string, value: string, dataset: string) {
      if (this.enableHighlighting) {
        if (key && value) {
          updateHighlight(this.$router, {
            context: context,
            dataset: dataset,
            key: key,
            value: value
          });
        } else {
          clearHighlight(this.$router);
        }
        appActions.logUserEvent(this.$store, {
          feature: Feature.CHANGE_HIGHLIGHT,
          activity: this.logActivity,
          subActivity: SubActivity.DATA_TRANSFORMATION,
          details: { key: key, value: value }
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
            value: value
          });
        }
      }
      this.$emit("numerical-click", key);
    },

    availableVariables(): string[] {
      // NOTE: used externally, not internally by the component
      // filter by search
      const searchFiltered = this.summaries.filter(summary => {
        return (
          this.filter === "" ||
          summary.key.toLowerCase().includes(this.filter.toLowerCase())
        );
      });
      return searchFiltered.map(v => v.key);
    },

    isSeriesID(colName: string): boolean {
      // Check to see if this facet is being used as a series ID
      const targetVar = routeGetters.getTargetVariable(this.$store);
      if (targetVar && targetVar.grouping) {
        if (targetVar.grouping.subIds.length > 0) {
          return !!targetVar.grouping.subIds.find(v => v === colName);
        }
      }
      return false;
    },

    isGeoLocated(location: string): boolean {
      return isGeoLocatedType(location);
    }
  }
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
  color: #868e96;
}

.page-item.active .page-link {
  z-index: 2;
  color: #fff;
  background-color: #868e96;
  border-color: #868e96;
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
  background: white;
  font-size: 0.867rem;
  color: rgba(0, 0, 0, 0.87);
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
</style>