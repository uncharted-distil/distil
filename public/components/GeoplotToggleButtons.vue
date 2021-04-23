<template>
  <ul
    class="position-absolute top-5"
    style="list-style-type: none; width: 40px"
  >
    <li>
      <div
        class="toggle"
        :class="{ active: isSelectionMode }"
        @click="selectionToolToggle"
      >
        <a
          class="selection-toggle-control"
          title="Select area"
          aria-label="Select area"
        >
          <icon-base width="100%" height="100%">
            <icon-crop-free />
          </icon-base>
        </a>
      </div>
    </li>
    <li>
      <div
        class="toggle"
        title="Cluster"
        aria-label="Cluster Tiles"
        :class="{ active: isClustering }"
        @click="clusteringToggle"
      >
        <i class="fa fa-object-group fa-lg icon" aria-hidden="true" />
      </div>
    </li>
    <li v-if="dataHasConfidence">
      <div
        class="toggle"
        :class="{ active: isColoringByConfidence }"
        @click="confidenceColoringToggle"
      >
        <a
          :class="confidenceClass"
          title="confidence"
          aria-label="Color by Confidence"
          :style="colorGradient"
        >
          C
        </a>
      </div>
    </li>
    <li>
      <div
        class="toggle"
        title="Change Map"
        aria-label="Change Map"
        :class="{ active: isSatelliteView }"
        @click="mapToggle"
      >
        <i class="fa fa-globe icon" aria-hidden="true" />
      </div>
    </li>
    <li>
      <div
        class="baseline-toggle toggle"
        title="Hide gray nodes"
        @click="baselineToggle"
      >
        <i class="fa fa-eye-slash icon" aria-hidden="true"></i>
      </div>
    </li>
  </ul>
</template>

<script lang="ts">
import Vue from "vue";
import IconBase from "./icons/IconBase.vue";
import IconCropFree from "./icons/IconCropFree.vue";
export default Vue.extend({
  name: "GeoplotToggleButtons",
  components: {
    IconBase,
    IconCropFree,
  },
  props: {
    isSatelliteView: { type: Boolean as () => boolean, default: false },
    isColoringByConfidence: { type: Boolean as () => boolean, default: false },
    isClustering: { type: Boolean as () => boolean, default: false },
    isSelectionMode: { type: Boolean as () => boolean, default: false },
    dataHasConfidence: { type: Boolean as () => boolean, default: false },
  },
  methods: {
    baselineToggle() {
      this.$emit("baseline-toggle");
    },
    mapToggle() {
      this.$emit("map-toggle");
    },
    confidenceColoringToggle() {
      this.$emit("confidence-toggle");
    },
    clusteringToggle() {
      this.$emit("clustering-toggle");
    },
    selectionToolToggle() {
      this.$emit("selection-tool-toggle");
    },
  },
});
</script>

<style scoped>
.top-5 {
  top: 5%;
}
.icon {
  height: 15px;
}
.toggle {
  position: relative;
  z-index: 999;
  width: 34px;
  height: 34px;
  background-color: #fff;
  border: 2px solid rgba(0, 0, 0, 0.2);
  background-clip: padding-box;
  text-align: center;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  justify-content: center;
  align-items: center;
}
.toggle.active {
  color: #26b8d1;
}
.confidence-toggle.active:hover::after {
  content: "----Less Confidence";
  position: absolute;
  white-space: nowrap;
  left: 30px;
  top: 15px; /*works out to 4 pixels from bottom (this is based off the font size)*/
  display: inline;
  position: absolute;
}
.confidence-toggle.active:hover::before {
  content: "----More Confidence";
  white-space: nowrap;
  left: 30px;
  top: -7px; /*works out to 4 pixels from top (this is based off the font size)*/
  display: inline;
  position: absolute;
}
</style>
