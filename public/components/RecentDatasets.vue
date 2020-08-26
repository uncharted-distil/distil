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
        datasetGetters.getDatasets(this.$store),
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
