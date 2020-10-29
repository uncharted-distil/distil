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
import GeoPlot from "./GeoPlot.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { Dictionary } from "../util/dict";
import {
  TableColumn,
  TableRow,
  Variable,
  VariableSummary,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { getVariableSummariesByState, searchVariables } from "../util/data";
import { isGeoLocatedType } from "../util/types";
import { actions as viewActions } from "../store/view/module";
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
        ? datasetGetters
            .getIncludedTableDataItems(this.$store)
            .concat(highlighted)
        : datasetGetters
            .getExcludedTableDataItems(this.$store)
            .concat(highlighted);
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
    onTileClick(args: { bounds: number[][]; callback: () => void }) {
      const filter = {
        displayName: "coordinates",
        key: "coordinates_group",
        maxX: args.bounds[1][1],
        maxY: args.bounds[0][0],
        minX: args.bounds[0][1],
        minY: args.bounds[1][0],
        mode: "exclude",
        type: "geobounds",
      };
      viewActions.updateAreaOfInterest(this.$store, filter);
      args.callback();
    },
  },
});
</script>

<style></style>
