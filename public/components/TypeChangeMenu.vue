<template>
  <div class="type-change-menu">
    <div class="type-change-dropdown-wrapper">
      <b-dropdown
        variant="secondary"
        class="var-type-button"
        id="type-change-dropdown"
        :text="label"
        :disabled="isDisabled"
      >
        <template v-if="!isComputedFeature">
          <template v-if="!isGroupedCluster">
            <b-dropdown-item
              v-for="suggested in getSuggestedList()"
              v-bind:class="{
                selected: suggested.isSelected,
                recommended: suggested.isRecommended
              }"
              @click.stop="onTypeChange(suggested.type)"
              :key="suggested.type"
            >
              <i
                v-if="suggested.isSelected"
                class="fa fa-check"
                aria-hidden="true"
              ></i>
              {{ suggested.label }}
              <icon-base
                v-if="suggested.isRecommended"
                icon-name="bookmark"
                class="recommended-icon"
                ><icon-bookmark
              /></icon-base>
            </b-dropdown-item>
          </template>
          <template v-if="!isGroupedCluster">
            <b-dropdown-divider></b-dropdown-divider>
          </template>
          <template>
            <b-dropdown-item
              v-for="grouping in groupingOptions()"
              @click.stop="onGroupingSelect(grouping.type)"
              :key="grouping.type"
            >
              {{ grouping.label }}
            </b-dropdown-item>
          </template>
        </template>
      </b-dropdown>
      <i v-if="isUnsure" class="unsure-type-icon fa fa-circle"></i>
    </div>
    <b-tooltip
      :delay="delay"
      :disabled="!isDisabled"
      target="type-change-dropdown"
    >
      Cannot change type when actively filtering or viewing models or
      predictions
    </b-tooltip>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import IconBase from "./icons/IconBase";
import IconBookmark from "./icons/IconBookmark";
import {
  SuggestedType,
  Variable,
  Highlight,
  RemoteSensingGrouping
} from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import {
  addTypeSuggestions,
  getLabelFromType,
  TIMESERIES_TYPE,
  isClusterType,
  getTypeFromLabel,
  isEquivalentType,
  isLocationType,
  normalizedEquivalentType,
  BASIC_SUGGESTIONS,
  GEOCOORDINATE_TYPE,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  REMOTE_SENSING_TYPE,
  hasComputedVarPrefix,
  COLLAPSE_ACTION_TYPE,
  EXPAND_ACTION_TYPE,
  EXPLODE_ACTION_TYPE
} from "../util/types";
import { hasFilterInRoute } from "../util/filters";
import { createRouteEntry } from "../util/routes";
import {
  GROUPING_ROUTE,
  PREDICTION_ROUTE,
  RESULTS_ROUTE
} from "../store/route";
import { getComposedVariableKey } from "../util/data";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";

const PROBABILITY_THRESHOLD = 0.8;

export default Vue.extend({
  name: "type-change-menu",

  components: {
    IconBase,
    IconBookmark
  },
  props: {
    dataset: String as () => string,
    field: String as () => string,
    expandCollapse: Function as () => Function
  },
  computed: {
    isPredictionOrResultsView(): boolean {
      const routePath = routeGetters.getRoutePath(this.$store);
      return (
        routePath &&
        (routePath === PREDICTION_ROUTE || routePath === RESULTS_ROUTE)
      );
    },
    isGroupedCluster(): boolean {
      return this.isCluster && this.isGrouping;
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    variable(): Variable {
      if (!this.variables) {
        return null;
      }

      const selectedVariable = this.variables.find(v => {
        if (this.field === null) {
          return;
        }
        return (
          v.colName.toLowerCase() === this.field.toLowerCase() &&
          v.datasetName === this.dataset
        );
      });

      const geocoordVariable = this.variables.find(v => {
        return v.colOriginalType === "real" && v.datasetName === this.dataset;
      });
      return selectedVariable ? selectedVariable : geocoordVariable;
    },
    isGrouping(): boolean {
      if (!this.variable) {
        return false;
      }
      return !!this.variable.grouping;
    },
    type(): string {
      return this.variable ? this.variable.colType : "";
    },
    isColTypeReviewed(): boolean {
      return this.variable ? this.variable.isColTypeReviewed : false;
    },
    originalType(): string {
      return this.variable ? this.variable.colOriginalType : "";
    },
    label(): string {
      return this.type !== "" ? getLabelFromType(this.type) : "";
    },
    suggestedTypes(): SuggestedType[] {
      const suggestedType = this.variable ? this.variable.suggestedTypes : [];
      return _.orderBy(suggestedType, "probability", "desc");
    },
    suggestedNonSchemaTypes(): SuggestedType[] {
      const nonSchemaTypes = _.filter(this.suggestedTypes, t => {
        return t.provenance !== "schema";
      });
      return nonSchemaTypes;
    },
    topNonSchemaType(): SuggestedType {
      return this.suggestedNonSchemaTypes.length > 0
        ? this.suggestedNonSchemaTypes[0]
        : undefined;
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },
    isCluster(): boolean {
      return isClusterType(normalizedEquivalentType(this.type));
    },
    isDisabled(): boolean {
      return (
        hasFilterInRoute(this.field) ||
        (this.highlight && this.highlight.key === this.field) ||
        this.isComputedFeature
      );
    },
    isComputedFeature(): boolean {
      return this.variable && hasComputedVarPrefix(this.variable.colName);
    },
    hasSchemaType(): boolean {
      return !!this.schemaType;
    },
    hasNonSchemaTypes(): boolean {
      return (
        _.find(this.suggestedTypes, t => {
          return t.provenance !== "schema";
        }) !== undefined
      );
    },
    schemaType(): SuggestedType {
      return _.find(this.suggestedTypes, t => {
        return t.provenance === "schema";
      });
    },
    isUnsure(): boolean {
      return (
        this.type === this.originalType && // we haven't changed the type (check from server)
        !this.isColTypeReviewed && // check if user ever reviewed the col type (client)
        this.hasSchemaType &&
        this.hasNonSchemaTypes &&
        this.topNonSchemaType.probability >= PROBABILITY_THRESHOLD && // it has both schema and ML types
        !isEquivalentType(this.schemaType.type, this.topNonSchemaType.type)
      ); // they don't agree
    },
    delay(): any {
      return {
        show: 10,
        hide: 10
      };
    }
  },

  methods: {
    groupingOptions() {
      const options = [];
      if (this.isGrouping) {
        options.push(
          {
            type: COLLAPSE_ACTION_TYPE,
            label: "Collapse"
          },
          {
            type: EXPAND_ACTION_TYPE,
            label: "Expand"
          }
        );
        if (!this.isPredictionOrResultsView) {
          options.push({
            type: EXPLODE_ACTION_TYPE,
            label: "Explode"
          });
        }
      } else {
        options.push(
          {
            type: TIMESERIES_TYPE,
            label: "Timeseries..."
          },
          {
            type: GEOCOORDINATE_TYPE,
            label: "Geocoordinate..."
          },
          {
            type: REMOTE_SENSING_TYPE,
            label: "Satellite Image..."
          }
        );
      }
      return options;
    },

    onGroupingSelect(type) {
      if (type === TIMESERIES_TYPE || type === GEOCOORDINATE_TYPE) {
        const entry = createRouteEntry(GROUPING_ROUTE, {
          dataset: routeGetters.getRouteDataset(this.$store),
          groupingType: type
        });
        this.$router.push(entry);
      } else if (type === REMOTE_SENSING_TYPE) {
        // CDB: Temporary for dev/debug.  Needs to be removed.
        datasetActions.setGrouping(this.$store, {
          dataset: this.dataset,
          grouping: {
            dataset: this.dataset,
            idCol: "group_id",
            type: REMOTE_SENSING_TYPE,
            imageCol: "image_file",
            bandCol: "band",
            coordinateCol: "coordinates",
            subIds: [],
            hidden: ["image_file", "band", "coordinates", "group_id"]
          } as RemoteSensingGrouping
        });
      } else if (
        this.expandCollapse &&
        (type === COLLAPSE_ACTION_TYPE || type === EXPAND_ACTION_TYPE)
      ) {
        this.expandCollapse(type);
      } else {
        datasetActions.removeGrouping(this.$store, {
          dataset: this.dataset,
          variable: this.variable.colName
        });
      }
    },

    addMissingSuggestions() {
      const flatSuggestedTypes = this.suggestedTypes.map(st => st.type);
      const missingSuggestions = addTypeSuggestions(flatSuggestedTypes);
      const nonSchemaSuggestions = this.suggestedNonSchemaTypes.map(suggested =>
        normalizedEquivalentType(suggested.type)
      );
      const menuSuggestions = _.uniq([
        ...nonSchemaSuggestions,
        ...missingSuggestions
      ]);
      return menuSuggestions;
    },
    getSuggestedList() {
      const currentNormalizedType = normalizedEquivalentType(this.type);
      const combinedSuggestions = this.addMissingSuggestions().map(type => {
        const normalizedType = normalizedEquivalentType(type);
        return {
          type: normalizedType,
          label: getLabelFromType(normalizedType),
          isRecommended:
            this.topNonSchemaType &&
            this.topNonSchemaType.type.toLowerCase() === type.toLowerCase(),
          isSelected: currentNormalizedType === normalizedType
        };
      });
      return combinedSuggestions;
    },
    onTypeChange(suggestedType) {
      const type = suggestedType;
      const field = this.field;
      const dataset = this.dataset;

      appActions.logUserEvent(this.$store, {
        feature: Feature.RETYPE_FEATURE,
        activity: Activity.PROBLEM_DEFINITION,
        subActivity: SubActivity.PROBLEM_SPECIFICATION,
        details: { from: this.type, to: type }
      });
      datasetActions
        .setVariableType(this.$store, {
          dataset: dataset,
          field: field,
          type: type
        })
        .then(() => {
          if (isLocationType(type)) {
            return datasetActions.geocodeVariable(this.$store, {
              dataset: dataset,
              field: field
            });
          } else if (type === "image") {
            return datasetActions.fetchClusters(this.$store, {
              dataset: this.dataset
            });
          }
          return null;
        })
        .then(() => {
          if (this.target && !this.isPredictionOrResultsView) {
            return datasetActions.fetchVariableRankings(this.$store, {
              dataset: dataset,
              target: this.target
            });
          }

          return null;
        });
    }
  }
});
</script>

<style>
.var-type-button button {
  border: none;
  border-radius: 0;
  padding: 2px 4px;
  width: 100%;
  text-align: left;
  outline: none;
  font-size: 0.75rem;
  color: white;
}
.var-type-button button:hover,
.var-type-button button:active,
.var-type-button button:focus,
.var-type-button.show > .dropdown-toggle {
  border: none;
  border-radius: 0;
  padding: 2px 4px;
  color: white;
  background-color: #424242;
  border-color: #424242;
  box-shadow: none;
}
.type-change-menu .dropdown-item {
  font-size: 0.867rem;
  text-transform: none;
  position: relative;
}
.type-change-menu .dropdown-item.selected {
  font-size: 0.867rem;
  text-transform: none;
  padding-left: 0;
}
.recommended-icon {
  position: absolute;
  right: 10px;
  bottom: 5px;
}
.unsure-type-icon {
  position: absolute;
  color: #dc3545;
  top: -5px;
  right: -5px;
  z-index: 2;
}
.type-change-dropdown-wrapper {
  position: relative;
}
</style>
