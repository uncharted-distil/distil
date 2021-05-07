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
  <b-dropdown variant="outline-secondary p-0 pl-1 pr-1" size="dropdown">
    <template v-slot:button-content>
      <div class="d-inline-flex align-items-center justify-content-center">
        <i class="fas fa-palette fa-sm"></i>
        <div
          v-if="!isFacetScale || isSelected"
          class="selected-bar d-inline-flex ml-1"
          :style="selectedColorScale.gradient"
        />
      </div>
    </template>
    <b-dropdown-item
      v-for="item in colorScales"
      :key="item.name"
      @click.stop="onScaleClick(item.name)"
    >
      {{ item.name[0].toUpperCase() + item.name.slice(1) }}
      <div class="w-100 bar" :style="item.gradient" />
    </b-dropdown-item>
  </b-dropdown>
</template>
<script lang="ts">
import Vue from "vue";
import { VariableSummary } from "../store/dataset/index";
import {
  ColorScaleNames,
  COLOR_SCALES,
  getGradientScales,
  getDiscreteScales,
  DISCRETE_COLOR_MAPS,
} from "../util/color";
import { getters as routeGetters } from "../store/route/module";
import { overlayRouteEntry, RouteArgs } from "../util/routes";
import { isCategoricalType } from "../util/types";

interface ColorScaleItem {
  name: string; // name of color scale
  gradient: string; // css linear-gradient string
}

export default Vue.extend({
  name: "color-scale-drop-down",
  props: {
    isFacetScale: { type: Boolean as () => boolean, default: false },
    variableSummary: Object as () => VariableSummary,
  },
  computed: {
    isCategorical(): boolean {
      return isCategoricalType(this.variableSummary.type);
    },
    colorScales(): ColorScaleItem[] {
      return this.isCategorical ? this.discreteScales : this.gradientScales;
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
    discreteScales(): ColorScaleItem[] {
      const result = [];
      for (const key of getDiscreteScales()) {
        const name = key;
        const gradient = this.getDiscrete(key);
        result.push({ name, gradient });
      }
      return result;
    },
    selectedColorScale(): ColorScaleItem {
      const selected = routeGetters.getColorScale(this.$store);
      return {
        name: selected,
        gradient: this.isCategorical
          ? this.getDiscrete(selected)
          : this.getGradient(selected),
      };
    },
    selectedFacet(): string {
      return routeGetters.getColorScaleVariable(this.$store);
    },
    isSelected(): boolean {
      return this.variableSummary?.key === this.selectedFacet;
    },
    dropDownClass(): string {
      return this.isSelected ? "selected-dropdown" : "not-selected-dropdown";
    },
    min(): number {
      return this.variableSummary?.baseline.extrema.min ?? 0;
    },
    max(): number {
      return this.variableSummary?.baseline.extrema.max ?? 0;
    },
  },
  methods: {
    getDiscrete(colorScaleName: ColorScaleNames): string {
      const colors = DISCRETE_COLOR_MAPS.get(colorScaleName);
      const stepLength = 100 / colors.length;
      let currentStep = 0;
      let linearGradient = "";
      for (let i = 0; i < colors.length; i++) {
        linearGradient += `${colors[i]} ${currentStep}%, ${colors[i]} ${
          currentStep + stepLength
        }%,`;
        currentStep += stepLength;
      }
      return `background: linear-gradient(to right, ${linearGradient.slice(
        0,
        linearGradient.length - 1
      )});`;
    },
    getGradient(colorScaleName: ColorScaleNames): string {
      const vals = [0.0, 0.25, 0.5, 0.75, 1.0]; // array to get the values to generate linear gradient
      const colors = vals.map(COLOR_SCALES.get(colorScaleName));
      return `background: linear-gradient(to right, ${colors.join(", ")});`;
    },
    onScaleClick(colorScaleName: ColorScaleNames) {
      const route = routeGetters.getRoute(this.$store);
      const routeArgs = { colorScale: colorScaleName } as RouteArgs;
      if (this.isFacetScale && !!this.variableSummary) {
        routeArgs.colorScaleVariable = this.variableSummary.key;
      }
      const entry = overlayRouteEntry(route, routeArgs);
      this.$router.push(entry).catch((err) => console.warn(err));
    },
  },
});
</script>

<style scoped>
.selected-dropdown {
  position: absolute;
  right: 120px;
}
.not-selected-dropdown {
  position: absolute;
  right: 80px;
}
.bar {
  height: 30px;
}
.dropdown {
  height: 22px;
}
.selected-bar {
  width: 100px;
  height: 13px;
  display: inline-block;
}
</style>
