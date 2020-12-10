<template>
  <div class="error-threshold-slider">
    <div>
      <div class="error-header">
        Error:
        <div
          class="asym-button float-right"
          v-bind:class="{ active: !symmetricSlider }"
          @click="enableAsymmetric"
        >
          <div class="button-line"></div>
          <div class="button-center"></div>
          <div class="button-left-circle"></div>
        </div>
        <div
          class="sym-button float-right"
          v-bind:class="{ active: symmetricSlider }"
          @click="enableSymmetric"
        >
          <div class="button-line"></div>
          <div class="button-center"></div>
          <div class="button-left-circle"></div>
          <div class="button-right-circle"></div>
        </div>
      </div>
    </div>

    <div class="error-slider">
      <div class="error-center-line"></div>
      <div class="error-center-label">0</div>
      <vue-slider
        ref="slider"
        :min="sliderMin"
        :max="sliderMax"
        :interval="sliderInterval"
        :value="sliderValue"
        :formatter="formatter"
        :lazy="true"
        width="100%"
        tooltip-dir="bottom"
        @callback="onSlide"
      />
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import { overlayRouteEntry } from "../util/routes";
import { Extrema } from "../store/dataset/index";
import { getters as resultsGetters } from "../store/results/module";
import { getters as routeGetters } from "../store/route/module";
import { actions as appActions } from "../store/app/module";
import { Feature, Activity, SubActivity } from "../util/userEvents";
import vueSlider from "vue-slider-component";
import Vue from "vue";

const DEFAULT_PERCENTILE = 0.25;
const NUM_STEPS = 100;
const ERROR_DECIMALS = 2;

export default Vue.extend({
  name: "error-threshold-slider",

  components: {
    vueSlider,
  },

  data() {
    return {
      symmetricSlider: true,
      min: null as number,
      max: null as number,
      hasModified: false,
      lastExtrema: { min: -1.0, max: 1.0 } as Extrema,
    };
  },

  computed: {
    sliderValue(): number[] {
      return [
        Math.ceil(this.normalize(this.thresholdMin)),
        Math.floor(this.normalize(this.thresholdMax)),
      ];
    },

    sliderMin(): number {
      return 0;
    },

    sliderMax(): number {
      return NUM_STEPS;
    },

    sliderInterval(): number {
      return 1;
    },

    residualExtrema(): Extrema {
      const extrema = resultsGetters.getResidualsExtrema(this.$store);
      // cache the last non-null max min we found - we display that until we get an update from the server
      if (extrema.min === null || extrema.max === null) {
        return this.lastExtrema;
      } else {
        this.lastExtrema = {
          min: extrema.min,
          max: extrema.max,
        };
      }
      return {
        min: extrema.min,
        max: extrema.max,
      };
    },

    thresholdMin(): number {
      const min = routeGetters.getRouteResidualThresholdMin(this.$store);
      return min !== undefined ? _.toNumber(min) : null;
    },

    thresholdMax(): number {
      const max = routeGetters.getRouteResidualThresholdMax(this.$store);
      return max !== undefined ? _.toNumber(max) : null;
    },

    hasThreshold(): boolean {
      return this.thresholdMin !== null && this.thresholdMax !== null;
    },
  },

  methods: {
    formatter(arg) {
      return arg ? this.denormalize(arg).toFixed(2) : "";
    },

    normalize(value: number): number {
      const range = this.residualExtrema.max - this.residualExtrema.min;
      return ((value - this.residualExtrema.min) / range) * NUM_STEPS;
    },

    denormalize(value: number): number {
      const range = this.residualExtrema.max - this.residualExtrema.min;
      return this.residualExtrema.min + (value / NUM_STEPS) * range;
    },

    enableAsymmetric() {
      this.symmetricSlider = false;
    },

    enableSymmetric() {
      this.symmetricSlider = true;
      const newVal = Math.min(Math.abs(this.min), this.max);
      this.forceSymmetric([-newVal, newVal]);
    },

    forceSymmetric(value: number[]): number[] {
      const newValues = [value[0], value[1]];
      if (this.symmetricSlider) {
        const origin = NUM_STEPS / 2;
        if (value[0] !== this.min) {
          // min changed, force max
          const dist = Math.abs(origin - value[0]);
          newValues[0] = origin - dist;
          newValues[1] = origin + dist;
        }
        if (value[1] !== this.max) {
          // max changed, force min
          const dist = Math.abs(origin - value[1]);
          newValues[0] = origin - dist;
          newValues[1] = origin + dist;
        }
        const $slider = <any>this.$refs.slider;
        $slider.setValue(newValues, true);
      }
      return newValues;
    },

    updateThreshold(min: number, max: number) {
      this.min = min;
      this.max = max;

      const entry = overlayRouteEntry(this.$route, {
        residualThresholdMin: `${this.denormalize(min)}`,
        residualThresholdMax: `${this.denormalize(max)}`,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },

    onSlide(value: number[]) {
      this.hasModified = true;
      const newValues = this.forceSymmetric(value);

      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_ERROR_THRESHOLD,
        activity: Activity.MODEL_SELECTION,
        subActivity: SubActivity.MODEL_EXPLANATION,
        details: { min: this.min, max: this.max },
      });

      this.updateThreshold(newValues[0], newValues[1]);
    },
  },

  watch: {
    residualExtrema(newR, oldR) {
      if (!newR || newR === oldR) return;

      // update threshold if there isnt one, or if the user hasn't touched
      // the slider yet.
      if (!this.hasThreshold || !this.hasModified) {
        // set the route
        const ORIGIN = NUM_STEPS / 2;
        const DEVIATION = NUM_STEPS * DEFAULT_PERCENTILE;
        const defaultMin = Math.ceil(ORIGIN - DEVIATION);
        const defaultMax = Math.floor(ORIGIN + DEVIATION);
        this.updateThreshold(defaultMin, defaultMax);
      }
    },
  },
});
</script>

<style>
.error-threshold-slider {
  padding-bottom: 2rem;
  position: relative;
  width: 100%;
}

.error-header {
  position: relative;
  margin: 0 15px;
}

.error-slider {
  position: relative;
  margin: 8px 10%;
}

.error-threshold-slider .vue-slider-component .vue-slider-process {
  background-color: #9e9e9e;
}

.error-threshold-slider .vue-slider-component .vue-slider-tooltip {
  border: 1px solid #9e9e9e;
  background-color: #9e9e9e;
}

.error-threshold-slider .vue-slider-component .vue-slider-piecewise {
  background-color: #e05353;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
  box-shadow: inset 0 0 0 1000px #007bff;
}

.error-center-line {
  position: absolute;
  left: 50%;
  height: 22px;
  width: 1px;
  background-color: #666;
}

.error-center-label {
  position: absolute;
  top: 22px;
  width: 100%;
  color: #666;
  text-align: center;
}

.sym-button,
.asym-button {
  position: relative;
  width: 36px;
  height: 23px;
  border-radius: 4px;
  margin: 2px;
  cursor: pointer;
}

.button-line {
  position: absolute;
  width: 26px;
  height: 2px;
  left: 4px;
  top: 10px;
}

.button-center {
  position: absolute;
  width: 1px;
  height: 14px;
  left: 17px;
  top: 4px;
}

.button-left-circle {
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 8px;
  top: 7px;
  left: 3px;
}

.button-right-circle {
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 8px;
  top: 7px;
  right: 3px;
}

.sym-button,
.asym-button {
  background-color: #eee;
  border: 1px solid #9e9e9e;
}

.sym-button:hover,
.asym-button:hover {
  opacity: 0.75;
}

.sym-button.active,
.asym-button.active {
  background-color: #9e9e9e;
  border: 1px solid #eee;
}

.button-line,
.button-center,
.button-left-circle,
.button-right-circle {
  background-color: #9e9e9e;
}

.active .button-line,
.active .button-center,
.active .button-left-circle,
.active .button-right-circle {
  background-color: #fff;
}
</style>
