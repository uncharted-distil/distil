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
    :summaries="summaries"
    :area-of-interest-items="{ inner: inner, outer: outer }"
    :confidence-access-func="confidenceGetter"
    enable-selection-tool-event
    @tileClicked="onTileClick"
    @selection-tool-event="onToolSelection"
  />
</template>

<script lang="ts">
import Vue from "vue";
import GeoPlot, { TileClickData, SelectionHighlight } from "../GeoPlot.vue";
import { getters as datasetGetters } from "../../store/dataset/module";
import { Dictionary } from "../../util/dict";
import {
  D3M_INDEX_FIELD,
  TableColumn,
  TableRow,
  Variable,
  VariableSummary,
} from "../../store/dataset/index";
import { getters as routeGetters } from "../../store/route/module";
import {
  getVariableSummariesByState,
  getAllDataItems,
  LOW_SHOT_LABEL_COLUMN_NAME,
  LowShotLabels,
  LOW_SHOT_SCORE_COLUMN_NAME,
} from "../../util/data";
import { isGeoLocatedType } from "../../util/types";
import { actions as viewActions } from "../../store/view/module";
import { INCLUDE_FILTER, Filter } from "../../util/filters";
import { actions as datasetActions } from "../../store/dataset/module";
import { bulkRowSelectionUpdate } from "../../util/row";

export default Vue.extend({
  name: "label-geo-plot",

  components: {
    GeoPlot,
  },

  props: {
    instanceName: String as () => string,
    includedActive: Boolean as () => boolean,
    dataItems: { type: Array as () => TableRow[], default: null },
    hasConfidence: { type: Boolean as () => boolean, default: false },
  },

  computed: {
    fields(): Dictionary<TableColumn> {
      return this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
    },

    items(): TableRow[] {
      if (this.dataItems) {
        return this.dataItems;
      }
      return getAllDataItems(this.includedActive);
    },
    availableTargetVarsSearch(): string {
      return routeGetters.getRouteAvailableTargetVarsSearch(this.$store);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
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
    summaries(): VariableSummary[] {
      const pageIndex = routeGetters.getRouteTrainingVarsPage(this.$store);
      const include = routeGetters.getRouteInclude(this.$store);
      const summaryDictionary = include
        ? datasetGetters.getIncludedVariableSummariesDictionary(this.$store)
        : datasetGetters.getExcludedVariableSummariesDictionary(this.$store);

      const currentSummaries = getVariableSummariesByState(
        pageIndex,
        this.variables.length,
        this.variables,
        summaryDictionary
      ) as VariableSummary[];

      return currentSummaries.filter((cs) => {
        return isGeoLocatedType(cs.varType);
      });
    },
    confidenceGetter(): Function {
      if (!this.hasConfidence) {
        return (item: TableRow, idx: number) => {
          return undefined;
        };
      }
      return this.getConfidenceRank;
    },
  },
  methods: {
    getConfidenceRank(item: TableRow, idx: number): number {
      if (item[LOW_SHOT_LABEL_COLUMN_NAME] === LowShotLabels.positive) {
        return 1.0;
      }
      if (item[LOW_SHOT_LABEL_COLUMN_NAME] === LowShotLabels.negative) {
        return 0;
      }
      // comes back order by confidence so the rank is already engrained in the array
      if (item[LOW_SHOT_SCORE_COLUMN_NAME]) {
        return 1.0 - idx / this.dataItems.length;
      }
      return undefined;
    },
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
    async onToolSelection(selection: SelectionHighlight) {
      const filterParams = routeGetters.getDecodedSolutionRequestFilterParams(
        this.$store
      );
      filterParams.size = datasetGetters.getIncludedTableDataNumRows(
        this.$store
      );
      // fetch data selected by map tool
      const resp = await datasetActions.fetchTableData(this.$store, {
        dataset: selection.dataset,
        highlights: [selection],
        filterParams: filterParams,
        dataMode: null,
        include: true,
      });
      // find d3mIndex
      const labelIndex = resp.columns.findIndex((c) => {
        return c.key === D3M_INDEX_FIELD;
      });
      // if -1 then something failed
      if (labelIndex === -1) {
        return;
      }
      // map the values
      const indices = resp.values.map((v) => {
        return v[labelIndex].value.toString();
      });
      // update row selection
      const rowSelection = routeGetters.getDecodedRowSelection(this.$store);
      bulkRowSelectionUpdate(
        this.$router,
        selection.context,
        rowSelection,
        indices
      );
    },
  },
});
</script>

<style></style>
