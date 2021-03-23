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
  <aside
    v-if="isRanked"
    class="importance d-inline-flex align-items-baseline ml-1"
    :class="{ 'not-available': !importance }"
    :title="importanceTitle"
  >
    <div
      v-for="index in numBars"
      :key="index"
      :class="{ active: index <= numActive }"
      :style="{ '--index': index }"
    />
  </aside>
</template>

<script lang="ts">
import Vue from "vue";
import { getters as routeGetters } from "../store/route/module";

// Labels associated with confidences for tooltips
const TOOLTIP_LABELS = ["LOW", "MEDIUM", "HIGH"];

// Bias exponent to apply to importance values.
const IMPORTANCE_EXPONENT = 0.3;

export default Vue.extend({
  name: "ImportanceBars",

  props: {
    // Feature importance value, assumed to be [0,1]
    importance: {
      type: Number as () => number,
      default: null,
    },

    // Number of bars in the display
    numBars: {
      type: Number as () => number,
      default: 5,
    },
  },

  computed: {
    // biased bar
    biasedImportance(): number {
      if (!this.importance) return;
      return Math.pow(this.importance, IMPORTANCE_EXPONENT);
    },

    // Threshold to display a bar active
    numActive(): number {
      if (!this.importance) return -1;
      return Math.round(this.biasedImportance * this.numBars);
    },

    // Generate the title tooltip
    importanceTitle(): string {
      if (!this.importance) return "Importance not available";

      const label =
        TOOLTIP_LABELS[
          Math.min(
            Math.round(this.biasedImportance * (TOOLTIP_LABELS.length - 1)),
            TOOLTIP_LABELS.length - 1
          )
        ];
      return `${label} estimated importance`;
    },

    // Check that the variables have been ranked
    isRanked(): boolean {
      return routeGetters.getRouteIsTrainingVariablesRanked(this.$store);
    },
  },
});
</script>

<style scoped>
.importance {
  position: relative;
}
.importance div {
  background-color: lightgray;
  border-radius: 2px;
  height: calc(3px * var(--index));
  margin-left: 1px;
  width: 3px;
}
.importance div.active {
  background-color: black;
}
.importance.not-available::after {
  background-color: lightgray;
  border-top: 2px solid white;
  content: "\0A";
  height: 4px;
  position: absolute;
  transform: rotate(45deg) translate(10%, 100%);
  width: 100%;
}
</style>
