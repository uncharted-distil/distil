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
  <b-dropdown variant="outline-secondary" no-flip>
    <template v-slot:button-content>
      <div class="selected-bar" :style="selectedColorScale.gradient" />
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
import { ColorScaleNames, COLOR_SCALES } from "../util/data";
import { getters as routeGetters } from "../store/route/module";
import { overlayRouteEntry } from "../util/routes";

interface ColorScaleItem {
  name: string; // name of color scale
  gradient: string; // css linear-gradient string
}

export default Vue.extend({
  name: "color-scale-drop-down",
  computed: {
    colorScales(): ColorScaleItem[] {
      const result = [];
      for (const [key] of COLOR_SCALES) {
        const name = key;
        const gradient = this.getGradient(key);
        result.push({ name, gradient });
      }
      return result;
    },
    selectedColorScale(): ColorScaleItem {
      const selected = routeGetters.getColorScale(this.$store);
      return { name: selected, gradient: this.getGradient(selected) };
    },
  },
  methods: {
    getGradient(colorScaleName: ColorScaleNames): string {
      const vals = [0.0, 0.25, 0.5, 0.75, 1.0]; // array to get the values to generate linear gradient
      const colors = vals.map(COLOR_SCALES.get(colorScaleName));
      return `background: linear-gradient(to right, ${colors.join(", ")});`;
    },
    onScaleClick(colorScaleName: ColorScaleNames) {
      const route = routeGetters.getRoute(this.$store);
      const entry = overlayRouteEntry(route, { colorScale: colorScaleName });
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
  height: 18px;
  display: inline-block;
}
</style>
