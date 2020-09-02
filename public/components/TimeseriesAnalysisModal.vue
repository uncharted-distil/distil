<template>
  <div>
    <b-modal title="Timeseries" v-model="show" cancel-disabled hide-footer>
      <div class="row justify-content-center">
        This dataset has a time related variable, perform timeseries analysis?
      </div>

      <div class="row justify-content-center mt-1 mb-1">
        <div class="col-6 text-center">
          <b>Time Column:</b>
        </div>
        <div class="col-6">
          <b-form-select v-model="timeCol" :options="variableOptions" />
        </div>
      </div>
      <div class="row justify-content-center">
        <b-btn
          class="mt-3 timeseries-modal-button"
          variant="outline-success"
          block
          @click="onClose(true)"
          >Yes</b-btn
        >
        <b-btn
          class="mt-3 timeseries-modal-button"
          variant="outline-secondary"
          block
          @click="onClose(false)"
          >No</b-btn
        >
      </div>
    </b-modal>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import { Variable } from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { isTimeType } from "../util/types";

export default Vue.extend({
  name: "timeseries-analysis-model",

  props: {
    show: Boolean as () => boolean,
  },

  data() {
    return {
      timeCol: null,
    };
  },

  computed: {
    variables(): Variable[] {
      return datasetGetters.getVariables(this.$store);
    },
    variableOptions(): Object[] {
      const def = [{ value: null, text: "Choose column", disabled: true }];
      const suggestions = this.variables
        .filter((v) => isTimeType(v.colType))
        .map((v) => {
          return { value: v.colName, text: v.colDisplayName };
        });

      if (suggestions.length === 1) {
        this.timeCol = suggestions[0].value;
        return suggestions;
      }

      return [].concat(def, suggestions);
    },
  },

  methods: {
    onClose(arg: boolean) {
      this.$emit("close", arg ? { col: this.timeCol } : null);
    },
  },
});
</script>

<style>
.timeseries-modal-button {
  margin: 0 8px;
  width: 25% !important;
  line-height: 32px !important;
}
</style>
