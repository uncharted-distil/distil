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
  <geo-plot
    :instance-name="instanceName"
    :data-fields="fields"
    :data-items="items"
    :baseline-items="baselineItems"
    :summaries="summaries"
    :areaOfInterestItems="{ inner: inner, outer: outer }"
    :is-exclude="!includedActive"
    @tileClicked="onTileClick"
  >
  </geo-plot>
</template>

<script lang="ts">
import Vue from "vue";
import GeoPlot, { TileClickData } from "./GeoPlot.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { Dictionary } from "../util/dict";
import {
  D3M_INDEX_FIELD,
  TableColumn,
  TableRow,
  Variable,
  VariableSummary,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getVariableSummariesByState, searchVariables } from "../util/data";
import { isGeoLocatedType } from "../util/types";
import { actions as viewActions } from "../store/view/module";
import { INCLUDE_FILTER, Filter } from "../util/filters";
export default Vue.extend({
  name: "select-geo-plot",

  components: {
    GeoPlot,
  },

  props: {
    instanceName: String as () => string,
    includedActive: Boolean as () => boolean,
  },

  computed: {
    fields(): Dictionary<TableColumn> {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
    },
    baselineItems(): TableRow[] {
      const bItems =
        datasetGetters.getBaselineIncludeTableDataItems(this.$store) ?? [];
      return bItems.sort((a, b) => {
        return a[D3M_INDEX_FIELD] - b[D3M_INDEX_FIELD];
      });
    },
    items(): TableRow[] {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);
    },
    inner(): TableRow[] {
      return this.includedActive
        ? datasetGetters.getAreaOfInterestIncludeInnerItems(this.$store)
        : datasetGetters.getAreaOfInterestExcludeInnerItems(this.$store);
    },
    outer(): TableRow[] {
      return this.includedActive
        ? datasetGetters.getAreaOfInterestIncludeOuterItems(this.$store)
        : datasetGetters.getAreaOfInterestExcludeOuterItems(this.$store);
    },
    trainingVarsSearch(): string {
      return routeGetters.getRouteTrainingVarsSearch(this.$store);
    },
    trainingVariables(): Variable[] {
      return searchVariables(
        routeGetters.getTrainingVariables(this.$store),
        this.trainingVarsSearch
      );
    },
    summaries(): VariableSummary[] {
      const pageIndex = routeGetters.getRouteTrainingVarsPage(this.$store);
      const include = routeGetters.getRouteInclude(this.$store);
      const summaryDictionary = include
        ? datasetGetters.getIncludedVariableSummariesDictionary(this.$store)
        : datasetGetters.getExcludedVariableSummariesDictionary(this.$store);

      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        this.trainingVariables.length,
        this.trainingVariables,
        summaryDictionary
      ) as VariableSummary[];

      return currentSummaries.filter((cs) => {
        return isGeoLocatedType(cs.varType);
      });
    },
  },
  methods: {
    async onTileClick(data: TileClickData) {
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
      };
      // fetch area of interests
      await viewActions.updateAreaOfInterest(this.$store, filter);
    },
  },
});
</script>

<style></style>
