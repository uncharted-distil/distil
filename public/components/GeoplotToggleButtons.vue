<template>
  <div class="position-absolute top-5">
    <ul style="list-style-type: none; padding-left: 1rem">
      <li class="pt-1">
        <div
          class="toggle"
          :class="{ active: isSelectionMode }"
          title="Select area"
          aria-label="Select area"
          @click="selectionToolToggle"
        >
          <icon-base
            width="100%"
            height="100%"
            class="d-flex justify-content-center align-items-center"
          >
            <icon-crop-free />
          </icon-base>
        </div>
      </li>
      <li class="pt-1">
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
      <li class="pt-1">
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
      <li class="pt-1">
        <div
          class="baseline-toggle toggle"
          title="Hide gray nodes"
          :class="{ active: isHidingBaseline }"
          @click="baselineToggle"
        >
          <i class="fa fa-eye-slash icon" aria-hidden="true" />
        </div>
      </li>
    </ul>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import IconBase from "./icons/IconBase.vue";
import IconCropFree from "./icons/IconCropFree.vue";
import { EventList } from "../util/events";
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
    isHidingBaseline: { type: Boolean as () => boolean, default: false },
  },
  methods: {
    baselineToggle() {
      this.$emit(EventList.MAP.BASELINE_TOGGLE_EVENT);
    },
    mapToggle() {
      this.$emit(EventList.MAP.MAP_TOGGLE_EVENT);
    },
    clusteringToggle() {
      this.$emit(EventList.MAP.CLUSTERING_TOGGLE_EVENT);
    },
    selectionToolToggle() {
      this.$emit(EventList.MAP.SELECTION_TOOL_TOGGLE_EVENT);
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
.toggle:hover {
  background-color: #f4f4f4;
}
.toggle.active {
  color: #26b8d1;
}
</style>
