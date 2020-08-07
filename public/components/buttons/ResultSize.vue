<template>
  <b-form-group label="Number of Results">
    <b-overlay
      :show="isUpdating"
      opacity="0.6"
      spinner-small
      spinner-variant="primary"
    >
      <b-input-group size="sm" :prepend="numDisplay">
        <b-form-input
          v-model="resultSize"
          number
          min="1"
          :max="total"
          step="1"
          type="range"
        />
        <b-input-group-append>
          <b-button
            :disabled="updateDisabled"
            variant="primary"
            @click="onUpdate"
          >
            Update
          </b-button>
        </b-input-group-append>
      </b-input-group>
    </b-overlay>
  </b-form-group>
</template>

<script lang="ts">
import Vue from "vue";
import { Highlight } from "../../store/dataset";
import {
  actions as resultsActions,
  getters as resultsGetters
} from "../../store/results/module";
import { getters as requestsGetters } from "../../store/requests/module";
import { getters as routeGetters } from "../../store/route/module";
import { overlayRouteEntry } from "../../util/routes";

/**
 * Button to change the size of a current result.
 * @param {Number} currentSize - the current number of results.
 * @param {Number} total - the total number of results.
 * @param {Boolean} excluded - display only excluded results.
 * @emits updated - a boolean to signal that the size has been updated.
 */
export default Vue.extend({
  name: "result-size",

  props: {
    currentSize: Number,
    total: Number,
    excluded: Boolean
  },

  data() {
    return {
      resultSize: this.currentSize,
      isUpdating: false
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    solutionId(): string {
      return requestsGetters.getActiveSolution(this.$store)?.solutionId;
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    /* Display the selected number of results displayed. */
    numDisplay(): string {
      return this.resultSize.toString();
    },

    /* Disable the Update button */
    updateDisabled(): boolean {
      return this.isUpdating || this.resultSize === this.currentSize;
    }
  },

  methods: {
    /* Set the resultSize in the URI, and reload the page */
    onUpdate() {
      this.isUpdating = true;
      const args = {
        dataset: this.dataset,
        solutionId: this.solutionId,
        highlight: this.highlight,
        size: this.resultSize
      };

      if (this.excluded) {
        resultsActions.fetchExcludedResultTableData(this.$store, args);
      } else {
        resultsActions.fetchIncludedResultTableData(this.$store, args);
      }
    }
  },

  watch: {
    currentSize(oldValue, newValue) {
      if (oldValue === newValue) return;
      this.resultSize = this.currentSize; // Set the input range to the appropriate value.
      this.isUpdating = false;
      this.$emit("updated");
    }
  }
});
</script>
