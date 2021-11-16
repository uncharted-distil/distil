<template>
  <b-table td-class="min-height-40" sticky-header="100%" :items="items" />
</template>

<script lang="ts">
import Vue from "vue";
import { Dataset } from "../../store/dataset";
import { formatBytes } from "../../util/bytes";
import { filterVariablesByFeature } from "../../util/data";
export default Vue.extend({
  name: "DatasetPreviewTable",
  props: {
    datasets: {
      type: Array as () => Dataset[],
      default: () => [] as Dataset[],
    },
  },
  computed: {
    items() {
      return this.datasets.map((d) => {
        return {
          "Dataset Name": d.name,
          Features: filterVariablesByFeature(d.variables).length,
          Rows: d.numRows,
          Size: formatBytes(d.numBytes),
        };
      });
    },
  },
});
</script>
<style scoped>
.min-height-40 {
  min-height: 40px;
}
</style>
