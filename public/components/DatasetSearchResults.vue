<template>
  <div class="search-results">
    <div class="row justify-content-center" v-if="isPending">
      <div v-html="spinnerHTML"></div>
    </div>
    <div class="search-results-container" ref="datasetResults">
      <div class="mb-3" :key="dataset.id" v-for="dataset in filteredDatasets">
        <dataset-preview :dataset="dataset" allow-join allow-import>
        </dataset-preview>
      </div>
      <div
        class="row justify-content-center"
        v-if="
          !isPending && (!filteredDatasets || filteredDatasets.length === 0)
        "
      >
        <h5>No datasets found for search</h5>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import $ from "jquery";
import DatasetPreview from "../components/DatasetPreview";
import Vue from "vue";
import { spinnerHTML } from "../util/spinner";
import { getters as datasetGetters } from "../store/dataset/module";
import { Dataset } from "../store/dataset/index";

export default Vue.extend({
  name: "dataset-search-results",

  components: {
    DatasetPreview
  },

  props: {
    isPending: Boolean as () => boolean
  },

  computed: {
    filteredDatasets(): Dataset[] {
      return datasetGetters.getFilteredDatasets(this.$store);
    },
    spinnerHTML(): string {
      return spinnerHTML();
    }
  },

  watch: {
    filteredDatasets() {
      // reset back to top on dataset change
      const $results = this.$refs.datasetResults as Element;
      $results.scrollTop = 0;
    }
  }
});
</script>

<style>
.search-results {
  width: 100%;
  overflow-x: hidden;
  overflow-y: auto;
}
.search-results-container {
  width: 100%;
}
</style>
