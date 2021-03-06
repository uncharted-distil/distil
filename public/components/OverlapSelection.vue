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
  <i
    v-if="shouldDisplay"
    :id="instanceName"
    class="stack-button"
    aria-hidden="true"
    >{{ items.length }}
    <b-popover :target="instanceName" triggers="hover" placement="left">
      <template #title>Overlapped Tiles</template>
      <div class="overflow-auto image-list">
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
              @click="onClick"
            />
          </div>
          <label v-if="hasTimeStamp">{{ items[i].item.timestamp.value }}</label>
        </template>
      </div>
    </b-popover>
  </i>
</template>

<script lang="ts">
import Vue from "vue";
import ImagePreview from "./ImagePreview.vue";
import ImageLabel from "./ImageLabel.vue";
import { Dictionary } from "../util/dict";
import { TableColumn } from "../store/dataset/index";
export default Vue.extend({
  name: "overlap-selection",
  components: {
    ImagePreview,
    ImageLabel,
  },
  props: {
    items: { type: Array, default: () => [] },
    indices: { type: Object as () => { x: number; y: number } },
    dataFields: Object as () => Dictionary<TableColumn>,
    instanceName: { type: String as () => string, default: "" },
    imageWidth: { type: Number, default: 124 },
    imageHeight: { type: Number, default: 124 },
    imageType: { type: String },
  },
  data() {
    return { eventName: "item-selected" };
  },
  computed: {
    shouldDisplay(): boolean {
      return this.items.length > 1;
    },
    hasTimeStamp(): boolean {
      return this.items[0]?.item?.timestamp !== null;
    },
  },
  methods: {
    onClick(item) {
      const res = this.items.find((i) => i.imageUrl === item.imageUrl);
      this.$emit(this.eventName, {
        item: res,
        key: this.indices,
      });
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
  max-height: 375px;
}
.image-container {
  position: relative;
  z-index: 0;
  width: 100%;
  height: 100%;
  display: block;
  margin: 2px;
}
</style>
