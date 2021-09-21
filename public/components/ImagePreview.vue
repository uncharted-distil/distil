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
  <div
    v-if="!!imageUrl"
    v-observe-visibility="visibilityChanged"
    :class="{ 'is-hidden': !isVisible && !preventHiding }"
    :style="{
      width: `${width}px`, // + 2 for boarder
      height: `${height}px`, // boarder
      filter: `grayscale(${gray}%)`,
      '--confidence': confidenceColor,
    }"
  >
    <div class="image-container" :class="{ selected: isSelected && isLoaded }">
      <div v-if="!isLoaded" v-html="spinnerHTML" />
      <template v-else-if="!stopSpinner">
        <div
          class="image-elem clickable"
          @click.stop.exact="handleClick"
          @click.shift.exact.stop="handleShiftClick"
        >
          <img ref="imageElem" />
        </div>
        <div
          v-show="imageAttentionHasRendered"
          class="filter-elem clickable"
          @click.stop.exact="handleClick"
          @click.shift.exact.stop="handleShiftClick"
        >
          <img ref="imageAttentionElem" />
        </div>
        <i class="fa fa-search-plus zoom-icon" @click.stop="showZoomedImage" />
      </template>
      <img v-else alt="image unavailable" :class="imgClass" />
    </div>

    <image-drilldown
      :enable-cycling="enableCycling"
      :field-key="fieldKey"
      :info="imageDrilldown"
      :items="items"
      :type="type"
      :url="imageUrl"
      :visible="zoomImage"
      :dataset-name="dataset"
      :index="index"
      :label-feature-name="labelFeatureName"
      :date-column="dateColumn"
      @close="hideZoomImage"
      @cycle-images="onCycleImages"
    />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import ImageDrilldown, { DrillDownInfo } from "./ImageDrilldown.vue";
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { circleSpinnerHTML as spinnerHTML } from "../util/spinner";
import { VariableSummary } from "../store/dataset/index";
import {
  D3M_INDEX_FIELD,
  TableRow,
  RowSelection,
} from "../store/dataset/index";
import { isRowSelected } from "../util/row";
import { Dictionary } from "../util/dict";
import {
  MULTIBAND_IMAGE_TYPE,
  IMAGE_TYPE,
  DATE_TIME_TYPE,
} from "../util/types";
import { COLOR_SCALES, colorByFacet } from "../util/color";
import { EI, EventList } from "../util/events";

export default Vue.extend({
  name: "ImagePreview",

  components: {
    ImageDrilldown,
  },

  props: {
    datasetName: { type: String as () => string, default: null },
    dateColumn: { type: String as () => string, default: "" },
    enableCycling: { type: Boolean as () => boolean, default: false },
    fieldKey: { type: String as () => string, default: "" },
    gray: { type: Number, default: 0 }, // support for graying images.
    height: { type: Number as () => number, default: 64 },
    imageUrl: { type: String as () => string, default: null },
    index: { type: Number as () => number },
    items: { type: Array as () => TableRow[] },
    labelFeatureName: { type: String as () => string, default: "" },
    preventHiding: { type: Boolean as () => boolean, default: false },
    row: { type: Object as () => TableRow, default: null as TableRow },
    shouldCleanUp: { type: Boolean as () => boolean, default: true },
    shouldFetchImage: { type: Boolean as () => boolean, default: true },
    summaries: {
      type: Array as () => VariableSummary[],
      default: () => [] as VariableSummary[],
    },
    type: { type: String as () => string, default: IMAGE_TYPE },
    uniqueTrail: { type: String as () => string, default: "" },
    width: { type: Number as () => number, default: 64 },
  },

  data() {
    return {
      hasRendered: false,
      hasRequested: false,
      imageAttentionHasRendered: false,
      isVisible: false,
      stopSpinner: false,
      zoomedHeight: 400,
      zoomImage: false,
      zoomedWidth: 400,
    };
  },

  computed: {
    confidenceColor(): string {
      if (!this.summaries.length || this.colorScaleByVar === "") {
        return;
      }
      // index is not needed
      return this.colorScale(this.colorByVariable(this.row, 0));
    },
    colorScaleByVar(): string {
      return routeGetters.getColorScaleVariable(this.$store);
    },
    colorByVariable(): (item: TableRow, idx: number) => number {
      const findKey = (v) => {
        return v.key === this.colorScaleByVar;
      };
      if (!this.summaries.some(findKey)) {
        return (item: TableRow, idx: number) => {
          return 0;
        };
      }
      return colorByFacet(this.summaries.find(findKey));
    },
    colorScale(): (t: number) => string {
      const colorScale = routeGetters.getColorScale(this.$store);
      return COLOR_SCALES.get(colorScale);
    },
    imageId(): string {
      return this.imageUrl?.split(/_B[0-9][0-9a-zA-Z][.]/)[0];
    },
    imgClass(): string {
      return this.imageUrl == null ? "d-none" : "";
    },
    imageDrilldown(): DrillDownInfo {
      return {
        band: this.band,
        title: this.imageUrl,
      };
    },
    files(): Dictionary<any> {
      return datasetGetters.getFiles(this.$store);
    },

    imageParamUrl(): string {
      return this.uniqueTrail.length
        ? `${this.imageUrl}/${this.uniqueTrail}`
        : this.imageUrl;
    },

    imageParamId(): string {
      return this.uniqueTrail.length
        ? `${this.imageId}/${this.uniqueTrail}`
        : this.imageId;
    },

    isLoaded(): boolean {
      if (this.type === IMAGE_TYPE) {
        return !!this.files[this.imageParamUrl] || this.stopSpinner;
      }

      return (
        (!!this.files[this.imageParamUrl] && !!this.files[this.imageParamId]) ||
        this.stopSpinner
      );
    },

    imageAttentionIsLoaded(): boolean {
      return (
        !!this.solutionId &&
        !!this.row &&
        !!this.files[this.solutionId + this.row.d3mIndex]
      );
    },

    image(): HTMLImageElement {
      if (this.type === IMAGE_TYPE) {
        return this.files[this.imageParamUrl] ?? null;
      }
      return (
        this.files[this.imageParamUrl] ?? this.files[this.imageParamId] ?? null
      );
    },

    imageAttentionId(): string {
      return this.solutionId + this.row.d3mIndex;
    },

    imageAttention(): HTMLImageElement {
      return this.files[this.solutionId + this.row?.d3mIndex] ?? null;
    },

    dataset(): string {
      return this.datasetName ?? routeGetters.getRouteDataset(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    isSelected(): boolean {
      if (this.row) {
        return isRowSelected(this.rowSelection, this.row[D3M_INDEX_FIELD]);
      }
      return false;
    },

    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },

    hasImageAttention(): boolean {
      return routeGetters.getImageAttention(this.$store);
    },

    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },

    spinnerHTML,
  },

  watch: {
    isLoaded() {
      this.$nextTick(async () => {
        if (!this.isLoaded) {
          return;
        }
        this.injectImage();
        await this.handleImageAttention();
      });
    },
    imageUrl(newUrl: string, oldUrl: string) {
      if (newUrl === null) return;
      if (newUrl !== oldUrl) {
        this.cleanUp();
        if (this.image) {
          this.injectImage();
          return;
        }
        this.hasRendered = false;
        this.hasRequested = false;
        this.requestImage();
      }
    },
    async hasImageAttention() {
      await this.handleImageAttention();
    },
    async row() {
      await this.handleImageAttention();
    },
    async colorScale() {
      if (!!this.solutionId && !!this.row && this.hasImageAttention) {
        await datasetActions.fetchImageAttention(this.$store, {
          dataset: this.dataset,
          resultId: this.solutionId,
          d3mIndex: this.row.d3mIndex,
        });
        this.injectFilter();
      }
    },
    // Refresh image on band change
    band(newBand: string, oldBand: string) {
      if (newBand !== oldBand) {
        this.cleanUp();
        this.hasRendered = false;
        this.hasRequested = false;
        if (this.isVisible) {
          this.requestImage();
        }
      }
    },
  },

  async beforeMount() {
    // lazy fetch available band types
    if (
      this.type === MULTIBAND_IMAGE_TYPE &&
      _.isEmpty(datasetGetters.getMultiBandCombinations(this.$store))
    ) {
      await datasetActions.fetchMultiBandCombinations(this.$store, {
        dataset: this.dataset,
      });
    }
  },

  destroyed() {
    this.cleanUp();
  },

  methods: {
    onCycleImages(cycleInfo: EI.IMAGES.CycleImage) {
      this.hideZoomImage();
      this.$emit(EventList.IMAGES.CYCLE_IMAGES, cycleInfo);
    },
    handleShiftClick() {
      this.$emit(EventList.BASIC.SHIFT_CLICK_EVENT, this.row);
    },

    async visibilityChanged(isVisible: boolean) {
      this.isVisible = isVisible;
      if (this.isVisible && !this.hasRequested) {
        this.requestImage();
        await this.handleImageAttention();
        return;
      }
      if (this.isVisible && this.hasRequested && !this.hasRendered) {
        this.injectImage();
      }
    },

    async handleImageAttention() {
      this.hasRendered = false;
      if (
        this.hasImageAttention &&
        !this.imageAttentionIsLoaded &&
        !!this.row
      ) {
        await datasetActions.fetchImageAttention(this.$store, {
          dataset: this.dataset,
          resultId: this.solutionId,
          d3mIndex: this.row.d3mIndex,
        });
      }
      if (this.hasImageAttention && this.imageAttentionIsLoaded) {
        this.injectFilter();
      }
      if (!this.hasImageAttention && this.imageAttentionHasRendered) {
        this.imageAttentionHasRendered = false;
      }
    },

    handleClick() {
      this.$emit(EventList.BASIC.CLICK_EVENT, {
        row: this.row,
        imageUrl: this.imageUrl,
        image: this.image,
      });
    },

    showZoomedImage() {
      this.zoomImage = true;
    },

    hideZoomImage() {
      this.zoomImage = false;
    },

    injectImage() {
      if (!this.image) return;
      const elem = this.$refs.imageElem as HTMLImageElement;
      if (elem) {
        // this.image can be an HTMLImageElement or Binary data
        if (typeof this.image === "object") {
          elem.src = this.image.src;
        } else {
          // in this instance this.image is binary data
          elem.src = ("data:image/jpeg;base64," + this.image) as string;
        }
        this.hasRendered = true;
      }
    },

    injectFilter() {
      if (!this.imageAttention) {
        return;
      }

      const elem = this.$refs.imageAttentionElem as HTMLImageElement;
      if (elem) {
        elem.src = this.imageAttention.src;
        this.imageAttentionHasRendered = true;
      }
    },

    async requestImage() {
      if (!this.shouldFetchImage) {
        return;
      }
      if (this.imageUrl === null) {
        this.stopSpinner = true; // imageUrl is null stop spinner
        return;
      }
      this.hasRequested = true;
      if (this.type === IMAGE_TYPE) {
        await datasetActions.fetchImage(this.$store, {
          dataset: this.dataset,
          url: this.imageUrl,
          isThumbnail: true,
        });
        if (this.isVisible) {
          this.injectImage();
        }
      } else if (this.type === MULTIBAND_IMAGE_TYPE) {
        await datasetActions.fetchMultiBandImage(this.$store, {
          dataset: this.dataset,
          imageId: this.imageId,
          bandCombination: routeGetters.getBandCombinationId(this.$store),
          isThumbnail: true,
          uniqueTrail: this.uniqueTrail,
        });
        if (this.isVisible) {
          this.injectImage();
        }
      } else {
        console.warn(`Image Data Type ${this.type} is not supported`);
      }
    },

    cleanUp() {
      if (!this.shouldCleanUp) {
        return;
      }
      const empty = "";
      if (this.uniqueTrail !== empty) {
        datasetMutations.removeFile(this.$store, this.imageParamId);
        if (!!this.imageAttention) {
          datasetMutations.removeFile(this.$store, this.imageAttentionId);
        }
      }
    },
  },
});
</script>

<style>
.is-hidden {
  visibility: hidden;
}

.image-container {
  position: relative;
  outline: solid 2px var(--confidence);
}

.image-container.selected {
  border: 2px solid #ff0067;
}

.image-elem:hover {
  background-color: #000;
}

.image-elem.clickable {
  cursor: pointer;
}

.image-elem.clickable img:hover {
  opacity: 0.7;
}

/* Keep the image in its container. */
.image-elem img {
  max-height: 100%;
  max-width: 100%;
  height: 100%;
  width: 100%;
  position: relative;
}

.filter-elem img {
  position: absolute;
  top: 0px;
  max-height: 100%;
  max-width: 100%;
}

/* Zoom icon */

.image-container .zoom-icon {
  background-color: #424242;
  color: #fff;
  cursor: pointer;
  opacity: 0.7;
  padding: 4px;
  position: absolute;
  right: 0;
  top: 0;
  visibility: hidden;
}

/* Make the icon visible on image hover. */
.image-container:hover .zoom-icon {
  visibility: visible;
}

/* Make the icon dimmer on hover. */
.image-container .zoom-icon:hover {
  opacity: 1;
}
</style>
