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
  <b-modal
    id="forecast-horizon-modal"
    title="Forecast Horizon"
    @ok="handleOk"
    @cancel="showError = false"
  >
    <b-form-group label="Interval size" description="Size of the prediction">
      <b-form-spinbutton v-model="intervalLength" inline min="1" />
      <!-- Add a dateTime scale selection. -->
      <b-dropdown
        v-if="isDateTime"
        :text="intervalScaleTitle"
        variant="outline-secondary"
      >
        <b-dropdown-item
          v-for="(scale, index) in intervalScale"
          :key="index"
          @click="intervalScaleSelected = index"
        >
          {{ scale.caption }}
        </b-dropdown-item>
      </b-dropdown>
    </b-form-group>
    <b-form-group
      label="Number of intervals"
      description="How many interval should the prediction made."
    >
      <b-form-spinbutton v-model="intervalCount" inline min="1" />
    </b-form-group>

    <template v-slot:modal-footer="{ ok, cancel }">
      <b-button :disabled="isWaiting" @click="cancel()">Cancel</b-button>

      <b-overlay
        :show="isWaiting"
        rounded
        opacity="0.6"
        spinner-small
        spinner-variant="primary"
        class="d-inline-block"
      >
        <b-button variant="primary" :disabled="isWaiting" @click="ok()">
          Forecast
        </b-button>
      </b-overlay>
    </template>

    <b-alert v-model="showError" variant="danger" dismissible>
      The Forecast prediction could not be made.
    </b-alert>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { TaskTypes } from "../store/dataset/index";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import {
  actions as requestActions,
  getters as requestGetters,
} from "../store/requests/module";
import { getPredictionsById } from "../util/predictions";
import { varModesToString, createRouteEntry } from "../util/routes";
import { PREDICTION_ROUTE } from "../store/route";
import { EventList } from "../util/events";

/**
 * Modal to request a Forecast Horizon.
 */
export default Vue.extend({
  name: "ForecastHorizon",

  props: {
    dataset: { type: String, default: null },
    fittedSolutionId: { type: String, default: null },
    target: { type: String, default: null },
    targetType: { type: String, default: null },
    handleStateChange: { type: Boolean as () => boolean, default: false },
  },

  data() {
    return {
      intervalCount: 1,
      intervalLength: 1,
      intervalScale: [
        { caption: "Seconds", value: 1 },
        { caption: "Minutes", value: 60 },
        { caption: "Hours", value: 3600 },
        { caption: "Days", value: 86400 },
        { caption: "Weeks", value: 604800 },
        { caption: "Months", value: 2629800 },
        { caption: "Years", value: 31557600 },
        { caption: "Decades", value: 315576000 },
      ],
      intervalScaleSelected: 0,
      showError: false,
      isWaiting: false,
    };
  },

  computed: {
    /* Get the interval length in seconds. */
    intervalLengthFormatted(): number {
      if (this.isDateTime) {
        return (
          this.intervalLength *
          this.intervalScale[this.intervalScaleSelected].value
        );
      } else {
        return this.intervalLength;
      }
    },

    intervalScaleTitle(): string {
      return this.intervalScale[this.intervalScaleSelected].caption;
    },

    /* Get the current timeseries extremas. */
    /* TODO - to be used to calculate "safe" extremas for the interval values.
    timeseriesExtremas(): Extrema {
      const extremas = datasetGetters.getTimeseriesExtrema(this.$store);
      if (!extremas[this.dataset]) return { max: 1, min: 1 };
      return extremas[this.dataset].x;
    },
    */

    /* Test if all the current timeseries variables are DateTime. */
    isDateTime(): boolean {
      const timeseries = datasetGetters.getTimeseries(this.$store);
      if (!timeseries[this.dataset]) return false;

      const values = Object.values(timeseries[this.dataset].isDateTime);
      return values.every((value) => value);
    },
  },

  methods: {
    handleOk(bvModalEvt) {
      // Prevent modal from closing
      bvModalEvt.preventDefault();
      this.makePredictionRequest();
    },

    /* Send the prediction to the server. */
    async makePredictionRequest() {
      this.isWaiting = true;

      const requestMsg = {
        datasetId: this.dataset,
        fittedSolutionId: this.fittedSolutionId,
        target: this.target,
        targetType: this.targetType,
        task: TaskTypes.TIME_SERIES,
        intervalCount: this.intervalCount,
        intervalLength: this.intervalLengthFormatted,
        existingDataset: false,
      };

      try {
        const response = await requestActions.createPredictRequest(
          this.$store,
          requestMsg
        );

        this.redirectToPredictionPage(response);
      } catch (error) {
        this.showError = true;
        console.error("Forecast prediction could not be made", error);
      }

      this.isWaiting = false;
    },

    /* Once the prediction is requested, we send the user to the prediction page. */
    redirectToPredictionPage(response: any) {
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
        predictionsDataset: predictionDataset,
        produceRequestId: response.produceRequestId,
        target: this.target,
        task: TaskTypes.TIME_SERIES,
        varModes: varModes,
        solutionId: routeGetters.getRouteSolutionId(this.$store),
      };
      if (!this.handleStateChange) {
        this.$emit(EventList.MODEL.APPLY_EVENT, routeArgs);
        return;
      }
      const entry = createRouteEntry(PREDICTION_ROUTE, routeArgs);
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>
