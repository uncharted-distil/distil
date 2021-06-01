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
  <div class="type-change-dropdown-wrapper">
    <d-drop-down
      id="type-change-dropdown"
      :value="label"
      label="label"
      class="btn-secondary"
      fontColor="#fff"
      :disabled="isDisabled"
      :options="getSuggestedList()"
      @input="onTypeChange"
    >
      <template v-slot:option="option">
        <div class="option-slot">
          <i v-if="option.isSelected" class="fa fa-check" aria-hidden="true" />
          {{ option.label }}
          <icon-base
            v-if="option.isRecommended"
            icon-name="bookmark"
            class="recommended-icon"
          >
            <icon-bookmark />
          </icon-base>
        </div>
      </template>
    </d-drop-down>
    <i v-if="isUnsure" class="unsure-type-icon fa fa-circle" />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import IconBase from "./icons/IconBase.vue";
import IconBookmark from "./icons/IconBookmark.vue";
import { SuggestedType, Variable } from "../store/dataset/index";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import {
  addTypeSuggestions,
  getLabelFromType,
  TIMESERIES_TYPE,
  isClusterType,
  isEquivalentType,
  normalizedEquivalentType,
  GEOCOORDINATE_TYPE,
  hasComputedVarPrefix,
  COLLAPSE_ACTION_TYPE,
  EXPAND_ACTION_TYPE,
  EXPLODE_ACTION_TYPE,
  isTimeSeriesType,
  isGeoLocatedType,
} from "../util/types";
import { hasFilterInRoute } from "../util/filters";
import { createRouteEntry } from "../util/routes";
import {
  GROUPING_ROUTE,
  PREDICTION_ROUTE,
  RESULTS_ROUTE,
} from "../store/route";
import DDropDown from "./DDropDown.vue";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import { hasHighlightInRoute } from "../util/highlights";

const PROBABILITY_THRESHOLD = 0.8;
interface SuggestedInfo {
  type: string;
  label: string;
  isRecommended: boolean;
  isSelected: boolean;
}
export default Vue.extend({
  name: "TypeChangeMenu",

  components: {
    IconBase,
    IconBookmark,
    DDropDown,
  },

  props: {
    dataset: { type: String, default: null },
    field: { type: String, default: null },
    expandCollapse: Function as () => Function,
    expand: { type: Boolean, default: false },
  },

  data() {
    return {
      delay: {
        show: 10,
        hide: 10,
      },
      boundary: "scrollParent" as string | HTMLElement,
    };
  },

  computed: {
    isPredictionOrResultsView(): boolean {
      const routePath = routeGetters.getRoutePath(this.$store);
      return (
        routePath &&
        (routePath === PREDICTION_ROUTE || routePath === RESULTS_ROUTE)
      );
    },
    isPageSelectTraining(): boolean {
      return routeGetters.isPageSelectTraining(this.$store);
    },
    isGroupedCluster(): boolean {
      return this.isCluster && this.isGrouping;
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    variable(): Variable {
      if (!this.variables) return;

      const selectedVariable = this.variables.find((v) => {
        if (this.field === null) return;

        return (
          v.key.toLowerCase() === this.field.toLowerCase() &&
          v.datasetName === this.dataset
        );
      });

      const geocoordVariable = this.variables.find((v) => {
        return v.colOriginalType === "real" && v.datasetName === this.dataset;
      });

      return selectedVariable ?? geocoordVariable;
    },
    isGrouping(): boolean {
      if (!this.variable) {
        return false;
      }
      return !!this.variable.grouping;
    },
    type(): string {
      return this.variable?.colType ?? "";
    },
    isColTypeReviewed(): boolean {
      return this.variable?.isColTypeReviewed ?? false;
    },
    originalType(): string {
      return this.variable?.colOriginalType ?? "";
    },
    label(): string {
      return this.type !== "" ? getLabelFromType(this.type) : "";
    },
    suggestedTypes(): SuggestedType[] {
      const suggestedType = this.variable ? this.variable.suggestedTypes : [];
      return _.orderBy(suggestedType, "probability", "desc");
    },
    suggestedNonSchemaTypes(): SuggestedType[] {
      const nonSchemaTypes = _.filter(this.suggestedTypes, (t) => {
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
    isCluster(): boolean {
      return isClusterType(normalizedEquivalentType(this.type));
    },
    isDisabled(): boolean {
      return hasFilterInRoute(this.field) || hasHighlightInRoute(this.field);
    },
    isComputedFeature(): boolean {
      return this.variable && hasComputedVarPrefix(this.variable.key);
    },
    hasSchemaType(): boolean {
      return !!this.schemaType;
    },
    hasNonSchemaTypes(): boolean {
      return (
        _.find(this.suggestedTypes, (t) => {
          return t.provenance !== "schema";
        }) !== undefined
      );
    },
    schemaType(): SuggestedType {
      return _.find(this.suggestedTypes, (t) => {
        return t.provenance === "schema";
      });
    },
    isUnsure(): boolean {
      return (
        this.type === this.originalType && // we haven't changed the type (check from server)
        !this.isColTypeReviewed && // check if user ever reviewed the col type (client)
        this.hasSchemaType &&
        this.schemaType.type !== "unknown" && // don't flag for check when the schema type was unknown (which is the base type)
        this.hasNonSchemaTypes &&
        this.topNonSchemaType.probability >= PROBABILITY_THRESHOLD && // it has both schema and ML types
        !isEquivalentType(this.schemaType.type, this.topNonSchemaType.type)
      ); // they don't agree
    },
  },

  methods: {
    groupingOptions() {
      const options = [];
      if (this.isGrouping) {
        if (this.expand) {
          options.push({
            type: COLLAPSE_ACTION_TYPE,
            label: "Collapse",
          });
        } else {
          options.push({
            type: EXPAND_ACTION_TYPE,
            label: "Expand",
          });
        }
        if (!this.isPredictionOrResultsView && !this.isPageSelectTraining) {
          options.push({
            type: EXPLODE_ACTION_TYPE,
            label: "Explode",
          });
        }
      } else {
        options.push(
          {
            type: TIMESERIES_TYPE,
            label: "Timeseries...",
          },
          {
            type: GEOCOORDINATE_TYPE,
            label: "Geocoordinate...",
          }
        );
      }
      return options;
    },
    async onGroupingSelect(type) {
      if (type === TIMESERIES_TYPE || type === GEOCOORDINATE_TYPE) {
        const entry = createRouteEntry(GROUPING_ROUTE, {
          dataset: routeGetters.getRouteDataset(this.$store),
          groupingType: type,
        });
        this.$router.push(entry).catch((err) => console.warn(err));
      } else if (
        this.expandCollapse &&
        (type === COLLAPSE_ACTION_TYPE || type === EXPAND_ACTION_TYPE)
      ) {
        this.expandCollapse(type);
      } else if (type === EXPLODE_ACTION_TYPE) {
        // For timeseries, exploding one variable explodes them all
        const toRemove = datasetGetters
          .getGroupings(this.$store)
          .filter((g) => {
            return (
              (isTimeSeriesType(g.colType) || isGeoLocatedType(g.colType)) &&
              g.datasetName === this.dataset
            );
          });

        for (const g of toRemove) {
          // CDB: This needs to be converted into an API call that can handle removal of
          // multiple groups because the UI goes spastic updating after each invidiual operation.
          await datasetActions.removeGrouping(this.$store, {
            dataset: this.dataset,
            variable: g.key,
          });
        }
      } else {
        console.error(`Unhandled grouping action ${type}`);
      }
    },
    addMissingSuggestions() {
      const flatSuggestedTypes = this.suggestedTypes.map((st) => st.type);
      const missingSuggestions = addTypeSuggestions(flatSuggestedTypes);
      const nonSchemaSuggestions = this.suggestedNonSchemaTypes.map(
        (suggested) => normalizedEquivalentType(suggested.type)
      );
      const menuSuggestions = _.uniq([
        ...nonSchemaSuggestions,
        ...missingSuggestions,
      ]);
      return menuSuggestions;
    },
    getSuggestedList(): SuggestedInfo[] {
      const currentNormalizedType = normalizedEquivalentType(this.type);
      const combinedSuggestions = this.addMissingSuggestions().map((type) => {
        const normalizedType = normalizedEquivalentType(type);
        return {
          type: normalizedType,
          label: getLabelFromType(normalizedType),
          isRecommended:
            this.topNonSchemaType &&
            this.topNonSchemaType.type.toLowerCase() === type.toLowerCase(),
          isSelected: currentNormalizedType === normalizedType,
        };
      });
      return combinedSuggestions;
    },
    onTypeChange(suggestedType: SuggestedInfo) {
      const type = suggestedType.type;
      const field = this.field;
      const dataset = this.dataset;

      appActions.logUserEvent(this.$store, {
        feature: Feature.RETYPE_FEATURE,
        activity: Activity.PROBLEM_DEFINITION,
        subActivity: SubActivity.PROBLEM_SPECIFICATION,
        details: { from: this.type, to: type },
      });
      datasetActions
        .setVariableType(this.$store, {
          dataset: dataset,
          field: field,
          type: type,
        })
        .then(() => {
          /* TODO
           * Disabled because the current solution is not responsive enough:
           * https://github.com/uncharted-distil/distil/issues/1815
          if (isLocationType(type)) {
            return datasetActions.geocodeVariable(this.$store, {
              dataset: dataset,
              field: field
            });
          } else if (type === "image") {
          */
          if (type === "image") {
            return datasetActions.fetchClusters(this.$store, {
              dataset: this.dataset,
            });
          }
          return null;
        })
        .then(() => {
          if (this.target && !this.isPredictionOrResultsView) {
            return datasetActions.fetchVariableRankings(this.$store, {
              dataset: dataset,
              target: this.target,
            });
          }

          return null;
        })
        .then(() => {
          return datasetActions.fetchOutliers(this.$store, dataset);
        });
    },
  },
});
</script>

<style>
.option-slot {
  display: flex;
  align-items: center;
  justify-content: space-between;
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
  margin: auto;
}
.unsure-type-icon {
  position: absolute;
  color: var(--red);
  top: -5px;
  right: -5px;
  z-index: 2;
}
.type-change-dropdown-wrapper {
  position: relative;
}
</style>
