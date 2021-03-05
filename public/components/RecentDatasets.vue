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
  <b-card header="Recent Dataset">
    <div v-if="recentDatasets.length === 0">None</div>
    <b-list-group v-bind:key="dataset.id" v-for="dataset in recentDatasets">
      <dataset-preview :dataset="dataset"> </dataset-preview>
    </b-list-group>
  </b-card>
</template>

<script lang="ts">
import _ from "lodash";
import DatasetPreview from "../components/DatasetPreview";
import { getters as datasetGetters } from "../store/dataset/module";
import { Dataset } from "../store/dataset/index";
import Vue from "vue";
import localStorage from "store";

export default Vue.extend({
  name: "recent-datasets",

  components: {
    DatasetPreview,
  },
  props: {
    maxDatasets: {
      default: 5,
      type: Number as () => number,
    },
  },

  computed: {
    recentDatasets(): Dataset[] {
      const recent = localStorage.get("recent-datasets") || [];
      const datasets = recent.slice(0, this.maxDatasets);
      return this.filterDatasets(
        datasets,
        datasetGetters.getDatasets(this.$store)
      );
    },
  },

  methods: {
    filterDatasets(ids: string[], datasets: Dataset[]): Dataset[] {
      if (_.isUndefined(ids)) {
        return datasets;
      }
      const idSet = new Set(ids);
      return _.filter(datasets, (d) => idSet.has(d.id));
    },
  },
});
</script>

<style></style>
