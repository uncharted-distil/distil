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
  <b-modal
    size="lg"
    hide-footer
    :title="visibleTitle"
    :visible="visible"
    dialog-class="fit-content"
    no-fade
    @hide="hide"
  >
    <main class="drill-down">
      <header class="d-flex flex-row pb-1 justify-content-between w-100">
        <ul v-if="hasDate" style="list-style-type: none" class="pr-3 pl-0">
          <li>Date: {{ date }}</li>
          <li>Time: {{ time }}</li>
        </ul>
        <color-scale-selection v-if="isMultiBandImage" class="pr-1 height-36" />
        <layer-selection
          v-if="isMultiBandImage"
          class="pr-1 height-36"
          :has-image-attention="hasImageAttention"
        />
        <label-header-buttons
          v-if="isLabelState && hasD3mIndex"
          disable-select-all
          class="height-36"
          :local-emit="onAnnotation"
        />
      </header>
      <div class="d-flex justify-content-center align-items-center">
        <b-button
          v-if="enableCycling"
          :disabled="carouselPosition === 0"
          @click="cycleImage(-1)"
        >
          <i class="fas fa-arrow-left" />
        </b-button>
        <div class="position-relative">
          <image-label
            class="image-label pt-2 pl-2"
            included-active
            shorten-labels
            align-horizontal
            :item="items[carouselPosition]"
            :label-feature-name="labelFeatureName"
          />
          <image-transformer
            ref="transformer"
            :selected="isSelected"
            :width="width"
            :height="height"
            :img-srcs="imageSources"
            :hidden="hidden"
            @row-selection="onRowSelection"
          />
        </div>
        <b-button
          v-if="enableCycling"
          :disabled="carouselPosition === items.length - 1"
          @click="cycleImage(1)"
        >
          <i class="fas fa-arrow-right" />
        </b-button>
      </div>
      <div
        v-show="isFilteredToggled"
        ref="imageAttentionElem"
        class="filter-elem"
      />
      <footer class="d-flex flex-row justify-content-between w-100">
        <div>
          <b-button
            v-if="shouldImagesScale"
            :disabled="disableUpscale || scale"
            class="height-36 mt-3"
            @click="upscaleFetch"
          >
            <b-spinner v-if="fetchingUpscale" small />
            Upscale Image
          </b-button>
          <b-button class="height-36 mt-3" @click="resetImage">
            Reset View
          </b-button>
        </div>
        <div v-if="isMultiBandImage" class="information-brightness">
          <b-input-group prepend="0" append="1.0" class="mt-3 mb-1" size="sm">
            <b-form-input
              type="range"
              name="brightness"
              :min="brightnessMin"
              :max="brightnessMax"
              step="1"
              class="brightness-slider"
              v-model="sliderPosition"
              @change="onBrightnessChanged"
            />
          </b-input-group>
          <label class="brightness-label">
            <i class="fa fa-adjust fa-rotate-180" aria-hidden="true" />
            {{ brightnessValue }}
          </label>
        </div>
      </footer>
    </main>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { RowSelection, TableRow } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { getters as appGetters } from "../store/app/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import { IMAGE_TYPE, MULTIBAND_IMAGE_TYPE } from "../util/types";
import ImageLabel from "./ImageLabel.vue";
import ImageTransformer from "./ImageTransformer.vue";
import LabelHeaderButtons from "./labelingComponents/LabelHeaderButtons.vue";
import ColorScaleSelection from "../components/ColorScaleSelection.vue";
import LayerSelection from "./LayerSelection.vue";
import { EventList, EI } from "../util/events";
import {
  addRowSelection,
  isRowSelected,
  removeRowSelection,
} from "../util/row";
import { ExplorerStateNames } from "../util/explorer";
const IMAGE_MAX_SIZE = 750; // Maximum size of an image in the drill-down in pixels.
const IMAGE_MAX_ZOOM = 2.5; // We don't want an image to be too magnified to avoid blurriness.

const imageId = (imageUrl) => imageUrl?.split(/_B[0-9][0-9a-zA-Z][.]/)[0];

export interface DrillDownInfo {
  band?: string;
  title?: string;
  confidence?: string;
}

/**
 * Display a modal with drill-downed information about an image.
 *
 * @param info {DrillDownInfo} List of information to be displayed.
 * @param url {String} URL of the image to be drill-down.
 * @param type {ImageType=} Type of the image, default to IMAGE_TYPE.
 * @param item {TableRow=} item being drill-down.
 * @param visible {Boolean} Display or hide the modal.
 */
export default Vue.extend({
  name: "ImageDrilldown",
  components: {
    ImageLabel,
    ImageTransformer,
    ColorScaleSelection,
    LayerSelection,
    LabelHeaderButtons,
  },
  props: {
    datasetName: { type: String as () => string, default: null },
    dateColumn: { type: String as () => string, default: "" },
    enableCycling: { type: Boolean as () => boolean, default: false },
    url: { type: String as () => string, default: null },
    fieldKey: { type: String as () => string, default: "" },
    index: { type: Number as () => number },
    info: Object as () => DrillDownInfo,
    items: { type: Array as () => TableRow[], default: () => [] as TableRow[] },
    labelFeatureName: { type: String as () => string, default: "" },
    type: { type: String, default: IMAGE_TYPE },
    visible: Boolean,
  },

  data() {
    return {
      IMAGE_MAX_SIZE: IMAGE_MAX_SIZE,
      carouselPosition: 0,
      currentBrightness: 0.5,
      brightnessMin: 0,
      brightnessMax: 100,
      sliderPosition: 50,
      isFilteredToggled: true,
      hidden: false,
      disableUpscale: false,
      fetchingUpscale: false,
      scale: false,
      imgHeight: 0,
      imgWidth: 0,
      uniqueTrail: "image-drilldown",
    };
  },

  computed: {
    hasDate(): boolean {
      return this.dateColumn !== "";
    },
    isSelected(): boolean {
      const d3mIndex = this.items[this.carouselPosition]?.d3mIndex;
      return d3mIndex ? isRowSelected(this.rowSelection, d3mIndex) : false;
    },
    shouldImagesScale(): boolean {
      return appGetters.getShouldScaleImages(this.$store);
    },
    date(): string {
      return new Date(
        this.items[this.carouselPosition][this.dateColumn]?.value
      ).toDateString();
    },
    time(): string {
      return new Date(this.items[this.carouselPosition][this.dateColumn]?.value)
        .toTimeString()
        .split("(")[0];
    },
    bandName(): string {
      return datasetGetters
        .getMultiBandCombinations(this.$store)
        .find((band) => band.id === this.info.band)?.displayName;
    },
    latLongValue(): string | null {
      if (!this.items.length) {
        return null;
      }

      return this.items[
        this.carouselPosition
      ]?.coordinates?.value.Elements?.slice(0, 2).map((x) => x.Float);
    },
    brightnessValue(): string {
      return this.currentBrightness.toFixed(2);
    },
    dataset(): string {
      const predDataset = routeGetters.getRoutePredictionsDataset(this.$store);
      const result = !!predDataset
        ? predDataset
        : routeGetters.getRouteDataset(this.$store);
      return this.datasetName ?? result;
    },
    files(): Dictionary<any> {
      return datasetGetters.getFiles(this.$store);
    },
    selectedImageUrl(): string {
      return this.items.length
        ? this.items[this.carouselPosition][this.fieldKey]?.value
        : this.url;
    },
    image(): HTMLImageElement {
      return (
        this.files[this.uniqueId] ??
        this.files[imageId(this.selectedImageUrl) + "/" + this.uniqueTrail] ??
        null
      );
    },
    imageSources(): HTMLImageElement[] {
      const sources = [];
      if (!!this.image) {
        sources.push(this.image);
      }
      if (!!this.imageAttention && this.isFilteredToggled) {
        sources.push(this.imageAttention);
      }
      return sources;
    },
    imageAttention(): HTMLImageElement {
      return (
        this.files[
          this.solutionId + this.items[this.carouselPosition]?.d3mIndex
        ] ?? null
      );
    },
    toggleStateString(): string {
      return this.isFilteredToggled ? "Disable" : "Enable";
    },
    isResultScreen(): boolean {
      return (
        routeGetters.getDataExplorerState(this.$store) ===
        ExplorerStateNames.RESULT_VIEW
      );
    },
    isMultiBandImage(): boolean {
      return this.type === MULTIBAND_IMAGE_TYPE;
    },
    hasImageAttention(): boolean {
      return routeGetters.getImageAttention(this.$store);
    },
    d3mIndex(): number {
      return this.items[this.carouselPosition]?.d3mIndex;
    },
    hasD3mIndex(): boolean {
      return !!this.d3mIndex;
    },
    visibleTitle(): string {
      return this.selectedImageUrl ?? "Image Drilldown";
    },
    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },
    ratio(): number {
      return Math.min(
        IMAGE_MAX_SIZE / Math.max(this.imgHeight, this.imgWidth),
        IMAGE_MAX_ZOOM
      );
    },
    height(): number {
      return this.imgHeight * this.ratio;
    },
    width(): number {
      return this.imgWidth * this.ratio;
    },
    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },
    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },
    isLabelState(): boolean {
      return (
        routeGetters.getDataExplorerState(this.$store) ===
        ExplorerStateNames.LABEL_VIEW
      );
    },
    uniqueId(): string {
      return this.selectedImageUrl + "-" + this.uniqueTrail;
    },
    imageLayerScale(): string {
      return routeGetters.getImageLayerScale(this.$store);
    },
  },

  watch: {
    visible() {
      if (this.visible) {
        this.disableUpscale = false;
        this.hidden = false;
        this.isFilteredToggled = this.hasImageAttention;
        this.carouselPosition = this.index;
        this.requestImage({
          gainL: 1.0,
          gamma: 2.2,
          gain: 2.5,
          scale: this.scale,
        });
        if (this.hasImageAttention) {
          this.requestFilter();
        }
      }
    },
    band() {
      // only fetch new image if image drill down is visible
      if (!this.visible) {
        return;
      }
      this.requestImage({
        gainL: 1.0,
        gamma: 2.2,
        gain: 2.5,
        scale: this.scale,
      });
    },
    imageLayerScale() {
      // only fetch new image if image drill down is visible
      if (!this.visible) {
        return;
      }
      this.requestImage({
        gainL: 1.0,
        gamma: 2.2,
        gain: 2.5,
        scale: this.scale,
      });
    },
  },

  methods: {
    onAnnotation(event: string) {
      // add to selection if not added already
      if (!isRowSelected(this.rowSelection, this.d3mIndex)) {
        addRowSelection(this.$router, "", this.rowSelection, this.d3mIndex);
      }
      // emit a global event of the annotation
      this.$eventBus.$emit(EventList.LABEL.ANNOTATION_EVENT, event);
    },
    resetImage() {
      this.$eventBus.$emit(EventList.IMAGE_DRILL_DOWN.RESET_IMAGE_EVENT);
    },
    onRowSelection() {
      const d3mIndex = this.items[this.carouselPosition]?.d3mIndex;
      if (d3mIndex) {
        if (!isRowSelected(this.rowSelection, d3mIndex)) {
          addRowSelection(this.$router, "", this.rowSelection, d3mIndex);
          return;
        }
        removeRowSelection(this.$router, "", this.rowSelection, d3mIndex);
      }
    },
    resetDrillDownData() {
      this.scale = false; // reset scale
      this.currentBrightness = 0.5; // reset brightness
      this.sliderPosition = 50; // reset slider position
    },
    cycleImage(sideToCycleTo: EI.IMAGES.Side) {
      this.carouselPosition = Math.max(
        0,
        Math.min(this.carouselPosition + sideToCycleTo, this.items.length - 1)
      );
      this.resetDrillDownData();
      this.requestImage({
        gainL: 1.0,
        gamma: 2.2,
        gain: 2.5,
        scale: this.scale,
      });
      if (this.hasImageAttention) {
        this.requestFilter();
      }
    },
    async upscaleFetch() {
      const MAX_GAINL = 2.0;
      this.disableUpscale = true;
      this.fetchingUpscale = true;
      this.scale = true;
      await this.requestImage({
        gainL: this.currentBrightness * MAX_GAINL,
        gamma: 2.2,
        gain: 2.5,
        scale: this.scale,
      });
      this.fetchingUpscale = false;
      this.disableUpscale = false;
    },
    hide() {
      this.hidden = true;
      this.$emit(EventList.BASIC.CLOSE_EVENT);
    },

    onBrightnessChanged(e) {
      const MAX_GAINL = 2.0;
      const val = Number(e) / this.brightnessMax;
      const gainL = val * MAX_GAINL;
      this.currentBrightness = val;
      this.requestImage({ gainL, gamma: 2.2, gain: 2.5, scale: this.scale }); // gamma, gain, are default. They are here if we need to edit them later down the road
    },

    cleanUp() {
      if (this.isMultiBandImage) {
        this.scale = false;
        datasetMutations.removeFile(this.$store, imageId(this.uniqueId));
      }
    },
    async requestFilter() {
      await datasetActions.fetchImageAttention(this.$store, {
        dataset: this.dataset,
        resultId: this.solutionId,
        d3mIndex: this.items[this.carouselPosition].d3mIndex,
      });
    },

    async requestImage(imageOptions?: {
      gamma: number;
      gain: number;
      gainL: number;
      scale: boolean;
    }) {
      if (this.isMultiBandImage) {
        await datasetActions.fetchMultiBandImage(this.$store, {
          dataset: this.dataset,
          imageId: imageId(this.selectedImageUrl),
          bandCombination: this.band,
          isThumbnail: false,
          colorScale: this.imageLayerScale,
          options: imageOptions,
          uniqueTrail: this.uniqueTrail,
        });
      } else {
        await datasetActions.fetchImage(this.$store, {
          dataset: this.dataset,
          url: this.selectedImageUrl,
          scale: imageOptions.scale,
          uniqueTrail: this.uniqueTrail,
        });
      }
      // save as data so when we fetch next image we retain the UI size
      this.imgHeight = this.image?.height;
      this.imgWidth = this.image?.width;
    },
  },
});
</script>
<style>
.fit-content {
  max-width: fit-content !important;
}
</style>
<style scoped>
.drill-down {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.image-container {
  max-height: 100%;
  max-width: 100%;
  overflow: auto;
  text-align: center;
}

/* Keep the image in its container. */
.image-container /deep/ img {
  max-height: var(--IMAGE_MAX_SIZE);
  max-width: var(--IMAGE_MAX_SIZE);
  position: relative;
}

.information {
  list-style: none;
  margin: 2rem 0 0;
  padding: 0;
}

.information li label {
  font-weight: bold;
}

.information-brightness {
  align-items: center;
  display: flex;
  flex-direction: column;
}

.brightness-label {
  margin-bottom: 0px;
  padding-left: 5px;
  padding-right: 5px;
  display: inline-block;
}
.filter-elem {
  position: absolute;
  top: 0px;
  left: 0px;
}
.brightness-slider {
  width: 70%;
}
.height-36 {
  height: 36px;
}
</style>
