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
  <b-dropdown v-if="bandRamp" variant="outline-secondary">
    <template v-slot:button-content>
      <i class="fas fa-palette"></i> Color Scale:
      <div class="selected-bar d-inline-flex mx-1" :style="selectedGradient" />
    </template>
    <b-dropdown-item
      v-for="gradientScale in gradientScales"
      :key="gradientScale.name"
      @click="setGradient(gradientScale.name)"
      class=""
    >
      {{ formatGradientName(gradientScale.name) }}
      <div class="w-100 bar" :style="gradientScale.gradient" />
    </b-dropdown-item>
  </b-dropdown>
</template>

<script lang="ts">
import Vue from "vue";
import {
  ColorScaleNames,
  COLOR_SCALES,
  getGradientScales,
} from "../util/color";
import { getters as routeGetters } from "../store/route/module";
import { getters as datasetGetters } from "../store/dataset/module";
import { overlayRouteEntry } from "../util/routes";

interface ColorScaleItem {
  name: ColorScaleNames; // name of color scale
  gradient: string; // css linear-gradient string
}

export default Vue.extend({
  name: "color-scale-selection",

  data() {
    return {
      imageAttentionEnabled: false,
    };
  },

  computed: {
    bandId(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },
    bandRamp(): boolean {
      return datasetGetters
        .getMultiBandCombinations(this.$store)
        .find((multiBandCombo) => multiBandCombo.id === this.bandId)?.ramp;
    },
    gradientScales(): ColorScaleItem[] {
      const result = [];
      for (const key of getGradientScales()) {
        const name = key;
        const gradient = this.getGradient(key);
        result.push({ name, gradient });
      }
      return result;
    },
    selectedGradient(): string {
      return this.getGradient(
        routeGetters.getImageLayerScale(this.$store) as ColorScaleNames
      );
    },
  },

  methods: {
    formatGradientName(str: string): string {
      // Capitalize the first letter of every word
      return str.replace(/(\b[a-z](?!\s))/g, (s) => s.toUpperCase());
    },

    getGradient(colorScaleName: ColorScaleNames): string {
      const vals = [0.0, 0.25, 0.5, 0.75, 1.0]; // array to get the values to generate linear gradient
      const colors = vals.map(COLOR_SCALES.get(colorScaleName));
      return `background: linear-gradient(to right, ${colors.join(", ")});`;
    },

    setGradient(colorScale: ColorScaleNames) {
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        imageLayerScale: colorScale,
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>

<style scoped>
.bar {
  height: 30px;
}
.selected-bar {
  width: 100px;
  height: 13px;
}
</style>
