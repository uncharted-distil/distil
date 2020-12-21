<template>
  <div v-if="shouldDisplay">
    <i :id="instanceName" class="stack-button" aria-hidden="true">{{
      items.length
    }}</i>
    <b-popover :target="instanceName" triggers="hover" placement="top">
      <template #title>Overlapped Tiles</template>
      <div class="image-list">
        <template v-for="(r, i) in items.length">
          <div class="image-container">
            <image-label
              class="image-label"
              :dataFields="dataFields"
              includedActive
              shortenLabels
              alignHorizontal
              :item="items[i].item"
            />
            <image-preview
              :key="items[i].imageUrl"
              class="image-preview"
              :row="items[i].item"
              :image-url="items[i].imageUrl"
              :width="imageWidth"
              :height="imageHeight"
              :type="imageType"
              :gray="items[i].gray"
              :onClick="onClick"
            />
          </div>
        </template>
      </div>
    </b-popover>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import ImagePreview from "../ImagePreview.vue";
import ImageLabel from "../ImageLabel.vue";
import { Dictionary } from "../../util/dict";
import { TableColumn } from "../../store/dataset/index";
export default Vue.extend({
  name: "overlap-selection",
  components: {
    ImagePreview,
    ImageLabel,
  },
  props: {
    items: { type: Array, default: () => [] },
    dataFields: Object as () => Dictionary<TableColumn>,
    instanceName: { type: String as () => string, default: "" },
    imageWidth: { type: Number, default: 124 },
    imageHeight: { type: Number, default: 124 },
    imageType: { type: String },
  },
  computed: {
    shouldDisplay(): boolean {
      return this.items.length > 1;
    },
  },
  methods: {
    onClick(item) {
      console.log(item);
    },
  },
});
</script>

<style scoped>
.stack-button {
  background-color: #424242;
  color: #fff;
  cursor: pointer;
  padding: 4px;
  position: absolute;
  left: 0;
  bottom: 0;
  display: block;
}
.image-list {
  display: flex;
  flex-direction: row;
  overflow-x: auto;
}
.image-container {
  position: relative;
  z-index: 0;
  width: 100%;
  height: 100%;
}
</style>
