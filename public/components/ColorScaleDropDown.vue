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
  <b-button
    v-if="isToggle"
    variant="outline-secondary"
    class="min-width-40"
    :class="toggleStyle"
    @click="onToggleClick"
  >
    <i class="fas fa-palette fa-sm" />
  </b-button>
  <div
    v-else
    class="d-flex justify-content-space align-items-center btn-secondary rounded"
  >
    <d-drop-down
      ref="drop-down"
      class="p-0 pl-1 pr-1 shadow-none cursor-pointer"
      :options="colorScales"
      :value="selectedColorScale"
      label="name"
      @input="onColorChange"
    >
      <template v-slot:selected-option-container>
        <div class="vs__selected">
          <div class="d-inline-flex align-items-center justify-content-center">
            <div
              v-if="isSelected"
              class="selected-bar d-inline-flex mr-1"
              :style="selectedColorScale.gradient"
            />
            <i
              class="fas fa-times fa-sm white-color"
              @mousedown.stop="onDisableScale"
            />
          </div>
        </div>
      </template>
      <template v-slot:option="option">
        {{ formatGradientName(option.name) }}
        <div class="w-100 bar" :style="option.gradient" />
      </template>
      <template v-slot:dropdown-caret-sibling-icon>
        <i
          v-if="!isSelected"
          class="dropdown-caret-sibling-icon fas fa-palette fa-sm"
        />
      </template>
    </d-drop-down>
  </div>
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
import DDropDown from "./DDropDown.vue";

interface ColorScaleItem {
  name: ColorScaleNames; // name of color scale
  gradient: string; // css linear-gradient string
}

export default Vue.extend({
  name: "color-scale-drop-down",
  components: {
    DDropDown,
  },
  props: {
    isFacetScale: { type: Boolean as () => boolean, default: false },
    variableSummary: Object as () => VariableSummary,
    isToggle: { type: Boolean as () => boolean, default: false },
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
      if (!this.isSelected) {
        return null;
      }
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
    toggleStyle(): string {
      return this.selectedFacet === this.variableSummary.key.split(":")[0]
        ? "selected-toggle d-flex align-items-center shadow-none"
        : "toggle d-flex align-items-center shadow-none";
    },
  },
  mounted() {
    if (this.isToggle && this.selectedFacet.length === 0) {
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, {
        colorScaleVariable: this.variableSummary.key.split(":")[0],
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    }
  },
  methods: {
    onColorChange(data: ColorScaleItem) {
      this.onScaleClick(data.name);
    },
    onToggleClick() {
      const splitKey = this.variableSummary.key.split(":")[0];
      const route = routeGetters.getRoute(this.$store);
      // we just pass in the resultUUID instead of the variable name for the green and red error colors
      const entry = overlayRouteEntry(route, {
        colorScaleVariable: this.selectedFacet !== splitKey ? splitKey : "",
      });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
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
    formatGradientName(str: string): string {
      // Capitalize the first letter of every word
      return str.replace(/(\b[a-z](?!\s))/g, (s) => s.toUpperCase());
    },
    getGradient(colorScaleName: ColorScaleNames): string {
      const vals = [0.0, 0.25, 0.5, 0.75, 1.0]; // array to get the values to generate linear gradient
      const colors = vals.map(COLOR_SCALES.get(colorScaleName));
      return `background: linear-gradient(to right, ${colors.join(", ")});`;
    },
    onDisableScale() {
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, { colorScaleVariable: "" });
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    onScaleClick(colorScaleName: ColorScaleNames) {
      this.$bvModal.hide(this.variableSummary.key);
      const route = routeGetters.getRoute(this.$store);
      const routeArgs = { colorScale: colorScaleName } as RouteArgs;
      if (this.isFacetScale && !!this.variableSummary) {
        routeArgs.colorScaleVariable = this.variableSummary.key;
      }
      const entry = overlayRouteEntry(route, routeArgs);
      this.$router.push(entry).catch((err) => console.warn(err));
    },
    toggleModal() {
      this.$bvModal.show(this.variableSummary.key);
    },
  },
});
</script>

<style scoped>
.list-item {
  color: #424242 !important;
}
.list-item:hover {
  cursor: pointer;
}
.white-color:hover {
  color: white;
}
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
.dropdown,
.toggle {
  height: 22px;
  height: 1.375rem;
}
.selected-toggle {
  height: 22px;
  height: 1.375rem;
  color: #424242;
  background-color: #9e9e9e;
}
.selected-bar {
  width: 70px;
  height: 13px;
  display: inline-block;
}
.min-width-40 {
  min-width: 40px;
}
.dropdown-caret-sibling-icon {
  margin-right: 5px;
}
</style>
