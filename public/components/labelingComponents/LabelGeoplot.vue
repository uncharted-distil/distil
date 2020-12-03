<template>
  <geo-plot
    :instance-name="instanceName"
    :data-fields="fields"
    :data-items="items"
    :summaries="summaries"
    @tileClicked="onTileClick"
  >
  </geo-plot>
</template>

<script lang="ts">
import Vue from "vue";
import GeoPlot, { TileClickData } from "../GeoPlot.vue";
import { getters as datasetGetters } from "../../store/dataset/module";
import { Dictionary } from "../../util/dict";
import {
  TableColumn,
  TableRow,
  Variable,
  VariableSummary,
} from "../../store/dataset/index";
import { getters as routeGetters } from "../../store/route/module";
import { getVariableSummariesByState } from "../../util/data";
import { isGeoLocatedType } from "../../util/types";
import { actions as viewActions } from "../../store/view/module";
import { INCLUDE_FILTER, Filter } from "../../util/filters";
export default Vue.extend({
  name: "label-geo-plot",

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

    items(): TableRow[] {
      const tableData = this.includedActive
        ? datasetGetters.getHighlightedIncludeTableDataItems(this.$store)
        : datasetGetters.getHighlightedExcludeTableDataItems(this.$store);
      const highlighted = tableData
        ? tableData.map((h) => {
            return { ...h, isExcluded: true };
          })
        : [];
      return this.includedActive
        ? [
            ...highlighted,
            ...datasetGetters.getIncludedTableDataItems(this.$store),
          ]
        : [
            ...highlighted,
            ...datasetGetters.getExcludedTableDataItems(this.$store),
          ];
    },
    availableTargetVarsSearch(): string {
      return routeGetters.getRouteAvailableTargetVarsSearch(this.$store);
    },
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
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
      // get area of interests
      const inner = this.includedActive
        ? datasetGetters.getAreaOfInterestIncludeInnerItems(this.$store)
        : datasetGetters.getAreaOfInterestExcludeInnerItems(this.$store);
      const outer = this.includedActive
        ? datasetGetters.getAreaOfInterestIncludeOuterItems(this.$store)
        : datasetGetters.getAreaOfInterestExcludeOuterItems(this.$store);
      // send data back to geoplot
      data.callback(inner, outer);
    },
  },
});
</script>

<style></style>