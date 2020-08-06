<template>
  <b-form-group label="Number of Results">
    <b-overlay
      :show="isUpdating"
      opacity="0.6"
      spinner-small
      spinner-variant="primary"
    >
      <b-input-group size="sm" :prepend-html="numDisplay">
        <b-form-input
          v-model="resultSize"
          number
          min="1"
          :max="numRows"
          step="1"
          type="range"
        />
        <b-input-group-append>
          <b-button
            :disabled="updateDisabled"
            variant="outline-secondary"
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
 */
export default Vue.extend({
  name: "result-size",

  data() {
    return {
      resultSize: 1,
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
      const num = this.resultSize.toString();
      const length = this.numRows.toString().length;
      return num.padStart(length, "‎‎-").replace(/-/gi, "&nbsp;");
    },

    /* Number of result exluded by the filter */
    numExcludedResults(): number {
      return resultsGetters.getExcludedResultTableDataCount(this.$store);
    },

    /* Number of result inluded by the filter */
    numIncludedResults(): number {
      return resultsGetters.getIncludedResultTableDataCount(this.$store);
    },

    /* Get the number of results items returned by the back-end */
    numResults(): number {
      // If they are identical, this mean that no filter has been applied
      if (this.numIncludedResults === this.numExcludedResults) {
        return this.numExcludedResults;
      }

      // To know the true number of results, we add those included and excluded.
      return this.numIncludedResults + this.numExcludedResults;
    },

    /* Get the total number of items available */
    numRows(): number {
      return resultsGetters.getResultDataNumRows(this.$store);
    },

    /* Disable the Update button */
    updateDisabled(): boolean {
      return this.isUpdating || this.resultSize === this.numResults;
    }
  },

  methods: {
    /* Set the resultSize in the URI, and reload the page */
    onUpdate() {
      this.isUpdating = true;
      resultsActions.fetchResultTableData(this.$store, {
        dataset: this.dataset,
        solutionId: this.solutionId,
        highlight: this.highlight,
        size: this.resultSize
      });
    }
  },

  watch: {
    numResults() {
      this.resultSize = this.numResults; // Set the input range to the appropriate value.
      this.isUpdating = false;
    }
  }
});
</script>
