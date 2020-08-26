<template>
  <div
    class="d-inline-flex flex-row align-items-baseline"
    :title="`${importanceLabel} estimated importance`"
  >
    <div
      v-for="bar of bars"
      class="importance-bar"
      :class="bar.colorClass"
      :key="bar.height"
      :style="{ height: bar.height + 'px', background: bar.colorClass }"
    ></div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import * as d3 from "d3";

// Rendering description for bar
interface Bar {
  colorClass: string;
  height: number;
}

// Labels associated with confidences for tooltips
const TOOLTIP_LABELS = ["LOW", "MEDIUM", "HIGH"];

// Bias exponent to apply to importance values.
const IMPORTANCE_EXPONENT = 0.3;

export default Vue.extend({
  name: "importance-bars",

  props: {
    // Feature importance value, assumed to be [0,1]
    importance: {
      type: Number as () => number,
      required: true,
    },
    // Number of bars in the display
    numBars: {
      type: Number as () => number,
      default: 5,
    },
    // Width of bars in pixels
    barWidth: {
      type: Number as () => number,
      default: 3,
    },
    // Bar height step in pixels
    barHeightIncrement: {
      type: Number as () => number,
      default: 3,
    },
  },

  computed: {
    // biased bar
    biasedImportance(): number {
      return Math.pow(this.importance, IMPORTANCE_EXPONENT);
    },

    // Render descriptions of bars
    bars(): Bar[] {
      const entries: Bar[] = [];
      const numActive = Math.round(this.biasedImportance * this.numBars);
      for (let i = 0; i < this.numBars; i++) {
        const entry = {
          height: i * this.barHeightIncrement,
          colorClass:
            i <= numActive ? "importance-active" : "importance-inactive",
        };
        entries.push(entry);
      }
      return entries;
    },

    // Generate the title tooltip
    importanceLabel(): string {
      const label =
        TOOLTIP_LABELS[
          Math.min(
            Math.round(this.biasedImportance * (TOOLTIP_LABELS.length - 1)),
            TOOLTIP_LABELS.length - 1,
          )
        ];
      return label;
    },
  },
});
</script>

<style scoped>
.importance-bar {
  width: 3px;
  margin-left: 1px;
  border-radius: 2px;
}
.importance-active {
  background: #000000;
}
.importance-inactive {
  background: lightgray;
}
</style>
