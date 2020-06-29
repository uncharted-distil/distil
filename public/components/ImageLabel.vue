<template>
  <ol class="labels">
    <li
      v-for="(label, index) in labels"
      :key="index"
      class="label"
      :class="label.status"
    >
      {{ label.value }}
    </li>
  </ol>
</template>

<script lang="ts">
import Vue from "vue";
import { Dictionary } from "../util/dict";
import { TableRow, TableColumn } from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as routeGetters } from "../store/route/module";

interface Label {
  status: string;
  value: string;
}

/**
 * Display the prediction result as a label.
 */
export default Vue.extend({
  name: "image-label",

  props: {
    dataFields: Object as () => Dictionary<TableColumn>,
    includedActive: Boolean as () => boolean,
    item: Object as () => TableRow
  },

  computed: {
    fields(): Dictionary<TableColumn> {
      return this.dataFields
        ? this.dataFields
        : this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
    },

    labels(): Label[] {
      const labels = [];
      let status;

      for (const key in this.fields) {
        status = null;

        // Define status of label (correct or incorrect) if we want to show the predicted error.
        if (key === this.predictedField && this.showError) {
          status = this.correct() ? "correct" : "incorrect";
        }

        // Display the label
        if (key === this.targetField || key === this.predictedField) {
          labels.push({ status, value: this.item[key].value } as Label);
        }
      }

      return labels;
    },

    predictedField(): string {
      const predictions = requestGetters.getActivePredictions(this.$store);
      if (predictions) {
        return predictions.predictedKey;
      }

      const solution = requestGetters.getActiveSolution(this.$store);
      return solution ? `${solution.predictedKey}` : "";
    },

    showError(): boolean {
      return (
        this.predictedField && !requestGetters.getActivePredictions(this.$store)
      );
    },

    targetField(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    }
  },

  methods: {
    correct(): boolean {
      return (
        this.item[this.targetField].value ===
        this.item[this.predictedField].value
      );
    }
  }
});
</script>

<style scoped>
.labels {
  color: white;
  list-style: none;
  list-style-position: outside;
  max-width: 100%;
}

.label {
  background-color: #424242;
  text-overflow: ellipsis;
  overflow: hidden;
}
.label.correct {
  background-color: #03c003;
}
.label.incorrect {
  background-color: #be0000;
}
</style>
