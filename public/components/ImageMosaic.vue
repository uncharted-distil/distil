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
          <div class="image-tile" :key="idx">
            <template v-for="(fieldInfo, fieldKey) in fields">
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
                @click="onImageClick"
                @shift-click="onImageShiftClick"
              />
            </template>
            <image-label
              class="image-label"
              :dataFields="dataFields"
              includedActive
              shortenLabels
              alignHorizontal
              :item="item"
              :is-result="isResult"
              :label-feature-name="labelFeatureName"
            />
          </div>
        </template>
      </template>
    </div>
    <b-pagination
      v-if="items && items.length > perPage"
      align="center"
      first-number
      last-number
      size="sm"
      v-model="currentPage"
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
} from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
  updateTableRowSelection,
  bulkRowSelectionUpdate,
} from "../util/row";
import { MULTIBAND_IMAGE_TYPE } from "../util/types";
import { getImageFields } from "../util/data";

export default Vue.extend({
  name: "image-mosaic",

  components: {
    ImageLabel,
    ImagePreview,
  },

  props: {
    instanceName: String as () => string,
    dataItems: Array as () => any[],
    dataFields: Object as () => Dictionary<TableColumn>,
    isResult: { type: Boolean as () => boolean, default: false },
    labelFeatureName: { type: String, default: "" },
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

  watch: {
    band() {
      this.debounceImageFetch();
    },
    paginatedItems(prev, cur) {
      // check if all the indices are in the same order and prev == cur
      if (this.sameData(prev, cur)) {
        return;
      }
      this.debounceImageFetch();
    },
  },

  destroyed() {
    window.removeEventListener("keyup", this.shiftRelease);
  },

  mounted() {
    this.removeImages();
    this.fetchImagePack(this.paginatedItems);
    window.addEventListener("keyup", this.shiftRelease);
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    items(): TableRow[] {
      if (this.dataItems) {
        return this.dataItems;
      }
      const items = this.includedActive
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);

      return updateTableRowSelection(
        items,
        this.rowSelection,
        this.instanceName
      );
    },

    paginatedItems(): TableRow[] {
      const page = this.currentPage - 1; // currentPage starts at 1
      const start = page * this.perPage;
      const end = start + this.perPage;

      return this.items.slice(start, end);
    },

    itemCount(): number {
      return this.items.length;
    },

    fields(): Dictionary<TableColumn> {
      const currentFields = this.dataFields
        ? this.dataFields
        : this.includedActive
        ? datasetGetters.getIncludedTableDataFields(this.$store)
        : datasetGetters.getExcludedTableDataFields(this.$store);
      return currentFields;
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    imageFields(): { key: string; type: string }[] {
      return getImageFields(this.fields);
    },

    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },

    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },
  },

  methods: {
    debounceImageFetch() {
      clearTimeout(this.debounceKey);
      this.debounceKey = setTimeout(() => {
        this.removeImages();
        this.fetchImagePack(this.paginatedItems);
      }, 1000);
    },
    sameData(old: [], cur: []): boolean {
      if (old === null || cur === null) {
        return false;
      }
      if (old.length !== cur.length) {
        return false;
      }
      for (let i = 0; i < old.length; ++i) {
        if (old[i][D3M_INDEX_FIELD] !== cur[i][D3M_INDEX_FIELD]) {
          return false;
        }
      }
      return true;
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
      if (!this.imageFields.length) {
        return;
      }
      const imageKey = this.imageFields[0].key;
      datasetMutations.bulkRemoveFiles(this.$store, {
        urls: this.paginatedItems.map((item) => {
          return `${item[imageKey].value}/${this.uniqueTrail}`;
        }),
      });
    },
    fetchImagePack(items) {
      if (!this.imageFields.length) {
        return;
      }
      const key = this.imageFields[0].key;
      const type = this.imageFields[0].type;
      // if band is "" the route assumes it is an image not a multi-band image
      datasetActions.fetchImagePack(this.$store, {
        multiBandImagePackRequest: {
          imageIds: items.map((item) => {
            return item[key].value;
          }),
          dataset: this.dataset,
          band: type === MULTIBAND_IMAGE_TYPE ? this.band : "",
        },
        uniqueTrail: this.uniqueTrail,
      });
    },
    onImageClick(event: any) {
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
        this.shiftClickInfo.second = this.items.findIndex(
          (x) => x.d3mIndex === data.d3mIndex
        );
        this.onShiftSelect();
        return;
      }
      this.shiftClickInfo.first = this.items.findIndex(
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
      const subSet = this.items.slice(start, end).map((item) => item.d3mIndex);
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
