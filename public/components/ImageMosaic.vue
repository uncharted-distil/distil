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
  <div class="mosaic-container" @keyup="shiftRelease">
    <div class="image-mosaic">
      <template v-for="imageField in imageFields">
        <template v-for="(item, idx) in paginatedItems">
          <div :key="idx" class="image-tile">
            <template v-for="(fieldInfo, fieldKey) in dataFields">
              <image-preview
                v-if="fieldKey === imageField.key"
                :key="fieldKey"
                class="image-preview"
                :row="item"
                :image-url="item[fieldKey].value"
                :width="imageWidth"
                :height="imageHeight"
                :type="imageField.type"
                :unique-trail="uniqueTrail"
                :should-clean-up="false"
                :should-fetch-image="false"
                :summaries="summaries"
                @click="onImageClick"
                @shift-click="onImageShiftClick"
              />
            </template>
            <image-label
              class="image-label"
              :data-fields="dataFields"
              included-active
              align-horizontal
              :item="item"
              :label-feature-name="labelFeatureName"
            />
          </div>
        </template>
      </template>
    </div>
    <b-pagination
      v-if="dataItems.length > perPage"
      v-model="currentPage"
      align="center"
      first-number
      last-number
      size="sm"
      :per-page="perPage"
      :total-rows="itemCount"
    />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import ImageLabel from "./ImageLabel.vue";
import ImagePreview from "./ImagePreview.vue";
import {
  RowSelection,
  TableColumn,
  TableRow,
  D3M_INDEX_FIELD,
  VariableSummary,
} from "../store/dataset/index";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  bulkRowSelectionUpdate,
} from "../util/row";
import { Field } from "../util/types";
import {
  getImageFields,
  sameData,
  bulkRemoveImages,
  debounceFetchImagePack,
} from "../util/data";

export default Vue.extend({
  name: "ImageMosaic",

  components: {
    ImageLabel,
    ImagePreview,
  },

  props: {
    instanceName: { type: String as () => string, default: "" },
    dataItems: {
      type: Array as () => TableRow[],
      default: () => {
        return [];
      },
    },
    dataFields: {
      type: Object as () => Dictionary<TableColumn>,
      default: () => {
        return {};
      },
    },
    labelFeatureName: { type: String as () => string, default: "" },
    dataset: { type: String as () => string, default: "" },
    summaries: {
      type: Array as () => VariableSummary[],
      default: Array as () => VariableSummary[],
    },
  },

  data() {
    return {
      imageWidth: 128,
      imageHeight: 128,
      currentPage: 1,
      perPage: 100,
      shiftClickInfo: { first: null, second: null },
      uniqueTrail: "image-mosiac",
      debounceKey: null,
    };
  },

  computed: {
    paginatedItems(): TableRow[] {
      const page = this.currentPage - 1; // currentPage starts at 1
      const start = page * this.perPage;
      const end = start + this.perPage;

      return this.dataItems.slice(start, end);
    },

    itemCount(): number {
      return this.dataItems.length;
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    imageFields(): Field[] {
      return getImageFields(this.dataFields);
    },

    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },

    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },
  },

  watch: {
    band() {
      this.debounceImageFetch();
    },

    paginatedItems(cur: TableRow[], prev: TableRow[]) {
      // check if all the indices are in the same order and prev == cur
      // and that images have been previously fetched
      if (sameData(prev, cur)) {
        return;
      }
      this.debounceImageFetch();
    },

    imageFields(cur: Field[], prev: Field[]) {
      if (prev.length == 0 && cur.length > 0) {
        this.debounceImageFetch();
      }
    },
  },

  destroyed() {
    window.removeEventListener("keyup", this.shiftRelease);
  },

  mounted() {
    this.debounceImageFetch();
    window.addEventListener("keyup", this.shiftRelease);
  },

  methods: {
    debounceImageFetch() {
      debounceFetchImagePack({
        items: this.paginatedItems,
        imageFields: this.imageFields,
        dataset: this.dataset,
        uniqueTrail: this.uniqueTrail,
        debounceKey: this.debounceKey,
      });
    },

    selectAll() {
      bulkRowSelectionUpdate(
        this.$router,
        this.instanceName,
        this.rowSelection,
        this.paginatedItems.map((pi) => pi.d3mIndex)
      );
    },

    removeImages() {
      bulkRemoveImages({
        items: this.paginatedItems,
        imageFields: this.imageFields,
        uniqueTrail: this.uniqueTrail,
      });
    },

    onImageClick(event) {
      if (!isRowSelected(this.rowSelection, event.row[D3M_INDEX_FIELD])) {
        addRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          event.row[D3M_INDEX_FIELD]
        );
      } else {
        removeRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          event.row[D3M_INDEX_FIELD]
        );
      }
    },

    onImageShiftClick(data: TableRow) {
      if (this.shiftClickInfo.first !== null) {
        this.shiftClickInfo.second = this.dataItems.findIndex(
          (x) => x.d3mIndex === data.d3mIndex
        );
        this.onShiftSelect();
        return;
      }
      this.shiftClickInfo.first = this.dataItems.findIndex(
        (x) => x.d3mIndex === data.d3mIndex
      );
    },

    onShiftSelect() {
      const start = Math.min(
        this.shiftClickInfo.second,
        this.shiftClickInfo.first
      );
      const end =
        Math.max(this.shiftClickInfo.second, this.shiftClickInfo.first) + 1; // +1 deals with slicing being exclusive
      const subSet = this.dataItems
        .slice(start, end)
        .map((item) => item.d3mIndex);
      this.resetShiftClickInfo();
      bulkRowSelectionUpdate(
        this.$router,
        this.instanceName,
        this.rowSelection,
        subSet
      );
    },

    shiftRelease(event) {
      if (event.key === "Shift") {
        this.resetShiftClickInfo();
      }
    },

    resetShiftClickInfo() {
      this.shiftClickInfo.first = null;
      this.shiftClickInfo.second = null;
    },
  },
});
</script>

<style scoped>
.image-mosaic {
  display: block;
  overflow: auto;
  padding-bottom: 0.5rem; /* To add some spacing on overflow. */
  height: 100%;
  width: 100%;
  margin-bottom: 0.25rem;
}
.mosaic-container {
  display: flex;
  height: 100%;
  width: 100%;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-box-direction: normal;
  flex-direction: column;
}
.image-tile {
  display: inline-block;
  position: relative;
  vertical-align: bottom;
  margin: 2px;
}

.image-preview {
  position: relative;
}

.image-label {
  position: absolute;
  top: 0px;
  z-index: 1;
}
</style>
