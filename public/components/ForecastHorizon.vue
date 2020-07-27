<template>
  <b-modal id="forecast-horizon-modal" title="Forecast Horizon" @ok="handleOk">
    <b-form-group label="Interval size">
      <b-form-spinbutton v-model="intervalLength" inline min="1" />
    </b-form-group>
    <b-form-group label="Number of intervals">
      <b-form-spinbutton v-model="intervalCount" inline min="1" />
    </b-form-group>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { getters as routeGetters } from "../store/route/module";
import {
  actions as requestActions,
  getters as requestGetters
} from "../store/requests/module";
import { getPredictionsById } from "../util/predictions";
import { varModesToString, createRouteEntry } from "../util/routes";
import { PREDICTION_ROUTE } from "../store/route";

/**
 * Modal to request a Forecast Horizon.
 *
 * TODO - Add test on the Interval Length and Count to offer the user with useful
 * limits (Yearly, Monthly, days, minutes, etc.) and limited counts.
 */
export default Vue.extend({
  name: "forecast-horizon",

  data() {
    return {
      intervalCount: 0,
      intervalLength: 0
    };
  },

  props: {
    dataset: String,
    fittedSolutionId: String,
    target: String,
    targetType: String
  },

  methods: {
    async handleOk() {
      try {
        const requestMsg = {
          datasetId: this.dataset,
          fittedSolutionId: this.fittedSolutionId,
          target: this.target,
          targetType: this.targetType,
          intervalCount: this.intervalCount, // in seconds
          intervalLength: this.intervalLength
        };

        const response = await requestActions.createPredictRequest(
          this.$store,
          requestMsg
        );

        this.predidctionFinish(response);
      } catch (err) {
        /**
         * TODO
         *
         * Display a visual error message on the message box.
         */
        console.error("Forecast Horizon", "Prediction could not finish");
      }
    },

    /* Once the prediction is finished, we send the user to the prediction page. */
    predidctionFinish(response: any) {
      const predictionDataset = getPredictionsById(
        requestGetters.getPredictions(this.$store),
        response.produceRequestId
      ).dataset;

      const varModes = varModesToString(
        routeGetters.getDecodedVarModes(this.$store)
      );

      const routeArgs = {
        applyModel: true.toString(),
        dataset: this.dataset,
        fittedSolutionId: this.fittedSolutionId,
        predictionDataset: predictionDataset,
        produceRequestId: response.produceRequestId,
        target: this.target,
        varModes: varModes
      };

      const entry = createRouteEntry(PREDICTION_ROUTE, routeArgs);

      this.$router.push(entry);
    }
  }
});
</script>
