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
  <b-table
    bordered
    hover
    small
    :items="items"
    :fields="tableFields"
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
      />
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
      />
    </template>
  </b-table>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import SparklinePreview from "./SparklinePreview.vue";
import ImagePreview from "./ImagePreview.vue";
import { Dictionary } from "../util/dict";
import {
  TableColumn,
  TableRow,
  Variable,
  TimeseriesGrouping,
} from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import {
  getTimeseriesGroupingsFromFields,
  formatFieldsAsArray,
  getImageFields,
} from "../util/data";
import { getters as routeGetters } from "../store/route/module";

export default Vue.extend({
  name: "JoinDataPreviewTable",

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
      return getImageFields(this.fields, this.topDataset);
    },

    tableFields(): TableColumn[] {
      return formatFieldsAsArray(this.fields);
    },
    topDataset(): string {
      const joinDatasets = routeGetters.getRouteJoinDatasets(this.$store);
      return joinDatasets.length >= 1 ? joinDatasets[0] : null;
    },
    timeseriesGroupings(): TimeseriesGrouping[] {
      return getTimeseriesGroupingsFromFields(this.variables, this.fields);
    },
  },
});
</script>

<style></style>
