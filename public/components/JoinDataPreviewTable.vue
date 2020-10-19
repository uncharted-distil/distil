<template>
  <b-table
    bordered
    hover
    small
    :items="items"
    :fields="fields"
    sticky-header="100%"
    class="distil-table"
  >
    <template v-slot:cell()="data">
      <span :title="data.value.value">{{ data.value.value }}</span>
    </template>

    <template
      v-for="imageField in imageFields"
      :slot="imageField"
      slot-scope="data"
    >
      <image-preview
        :key="imageField.key"
        :image-url="data.item[imageField.key]"
        :type="imageField.type"
      ></image-preview>
    </template>

    <template
      v-for="timeseriesGrouping in timeseriesGroupings"
      :slot="timeseriesGrouping.idCol"
      slot-scope="data"
    >
      <sparkline-preview
        :key="timeseriesGrouping.idCol"
        :dataset="dataset"
        :x-col="timeseriesGrouping.xCol"
        :y-col="timeseriesGrouping.yCol"
        :timeseries-col="timeseriesGrouping.idCol"
        :timeseries-id="data.item[timeseriesGrouping.idCol]"
      >
      </sparkline-preview>
    </template>
  </b-table>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import SparklinePreview from "./SparklinePreview";
import ImagePreview from "./ImagePreview";
import { Dictionary } from "../util/dict";
import {
  TableColumn,
  TableRow,
  D3M_INDEX_FIELD,
  Grouping,
  Variable,
  TimeseriesGrouping,
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import {
  IMAGE_TYPE,
  TIMESERIES_TYPE,
  MULTIBAND_IMAGE_TYPE,
} from "../util/types";
import { getTimeseriesGroupingsFromFields, getImageFields } from "../util/data";

export default Vue.extend({
  name: "join-data-preview-table",

  components: {
    ImagePreview,
    SparklinePreview,
  },

  props: {
    items: Array as () => TableRow[],
    fields: Object as () => Dictionary<TableColumn>,
    instanceName: String as () => string,
  },

  computed: {
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    imageFields(): { key: string; type: string }[] {
      return getImageFields(this.fields);
    },

    timeseriesGroupings(): TimeseriesGrouping[] {
      return getTimeseriesGroupingsFromFields(this.variables, this.fields);
    },
  },
});
</script>

<style></style>
