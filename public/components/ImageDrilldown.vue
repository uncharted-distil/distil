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
    @hide="hide"
  >
    <main class="drill-down">
      <b-button
        v-if="isCarousel"
        class="position-absolute left"
        :disabled="carouselPosition === 0"
        @click="rotateSelection(-1)"
      >
        <b>&lt;</b>
      </b-button>
      <b-button
        v-if="isCarousel"
        class="position-absolute right"
        :disabled="carouselPosition === imageUrls.length - 1"
        @click="rotateSelection(1)"
      >
        <b>&gt;</b>
      </b-button>
      <image-transformer
        ref="transformer"
        :width="width"
        :height="height"
        :img-srcs="imageSources"
        :hidden="hidden"
      />
      <div
        v-show="isFilteredToggled"
        ref="imageAttentionElem"
        class="filter-elem"
      />
      <div class="row">
        <div class="col-md-6">
          <ul class="information">
            <li v-if="bandName"><label>Image Layer:</label> {{ bandName }}</li>
            <li v-if="latLongValue">
              <label>Lat/Long:</label> {{ latLongValue }}
            </li>
            <li v-if="isResultScreen" class="d-flex justify-content-between">
              <label> {{ toggleStateString }} image explanation: </label>
              <div>
                <input
                  id="drill-down-filter-toggle"
                  v-model="isFilteredToggled"
                  class="form-check-input"
                  type="checkbox"
                  value=""
                />
              </div>
            </li>
          </ul>
        </div>
        <div class="col-md-6">
          <ul>
            <li v-if="isMultiBandImage" class="information-brightness">
              <b-input-group
                prepend="0"
                append="1.0"
                class="mt-3 mb-1"
                size="sm"
              >
                <b-form-input
                  type="range"
                  name="brightness"
                  :min="brightnessMin"
                  :max="brightnessMax"
                  step="1"
                  class="brightness-slider"
                  @change="onBrightnessChanged"
                />
              </b-input-group>
              <label class="brightness-label">
                <i class="fa fa-adjust fa-rotate-180" aria-hidden="true" />
                {{ brightnessValue }}
              </label>
              <b-button
                v-if="shouldImagesScale"
                :disabled="disableUpscale"
                @click="upscaleFetch"
              >
                <b-spinner v-if="fetchingUpscale" small />
                Upscale Image
              </b-button>
            </li>
          </ul>
        </div>
      </div>
    </main>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { TableRow } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { getters as appGetters } from "../store/app/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import { IMAGE_TYPE, MULTIBAND_IMAGE_TYPE } from "../util/types";
import ImageTransformer from "./ImageTransformer.vue";

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
    ImageTransformer,
  },
  props: {
    info: Object as () => DrillDownInfo,
    type: { type: String, default: IMAGE_TYPE },
    url: { type: String, default: null },
    visible: Boolean,
    imageUrls: { type: Array as () => string[], default: () => [] as string[] },
    items: { type: Array as () => TableRow[], default: () => [] as TableRow[] },
    initialPosition: { type: Number as () => number, default: 0 },
    datasetName: { type: String as () => string, default: null },
  },

  data() {
    return {
      IMAGE_MAX_SIZE: IMAGE_MAX_SIZE,
      currentVal: 0.5,
      carouselPosition: this.initialPosition,
      currentBrightness: 0.5,
      brightnessMin: 0,
      brightnessMax: 100,
      isFilteredToggled: true,
      hidden: false,
      disableUpscale: false,
      fetchingUpscale: false,
      scale: 0,
    };
  },

  computed: {
    shouldImagesScale(): boolean {
      return appGetters.getShouldScaleImages(this.$store);
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
    isCarousel(): boolean {
      return this.imageUrls.length > 0;
    },
    selectedImageUrl(): string {
      return this.imageUrls.length
        ? this.imageUrls[this.carouselPosition]
        : this.url;
    },
    image(): HTMLImageElement {
      return (
        this.files[this.selectedImageUrl] ??
        this.files[imageId(this.selectedImageUrl)] ??
        null
      );
    },
    imageSources(): string[] {
      const sources = [];
      if (!!this.image) {
        sources.push(this.image.src);
      }
      if (!!this.imageAttention && this.isFilteredToggled) {
        sources.push(this.imageAttention.src);
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
      return routeGetters.getRouteSolutionId(this.$store) != null;
    },
    isMultiBandImage(): boolean {
      return this.type === MULTIBAND_IMAGE_TYPE;
    },
    hasImageAttention(): boolean {
      return routeGetters.getImageAttention(this.$store);
    },
    visibleTitle(): string {
      return this.info.title ?? this.selectedImageUrl ?? "Image Drilldown";
    },
    sliderVal(): string {
      return this.currentVal.toFixed(2);
    },
    solutionId(): string {
      return routeGetters.getRouteSolutionId(this.$store);
    },
    ratio(): number {
      return Math.min(
        IMAGE_MAX_SIZE / Math.max(this.image?.height, this.image?.width),
        IMAGE_MAX_ZOOM
      );
    },
    height(): number {
      return this.image?.height * this.ratio;
    },
    width(): number {
      return this.image?.width * this.ratio;
    },
  },

  watch: {
    visible() {
      if (this.visible) {
        this.disableUpscale = false;
        this.hidden = false;
        this.isFilteredToggled = this.hasImageAttention;
        this.carouselPosition = this.initialPosition;
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
  },

  methods: {
    async upscaleFetch() {
      const MAX_GAINL = 2.0;
      this.disableUpscale = true;
      this.fetchingUpscale = true;
      this.scale += 1;
      await this.requestImage({
        gainL: this.currentBrightness * MAX_GAINL,
        gamma: 2.2,
        gain: 2.5,
        scale: this.scale,
      });
      this.fetchingUpscale = false;
      this.disableUpscale = false;
    },
    rotateSelection(direction: number) {
      this.scale = 0; // reset scale when fetching new images
      this.carouselPosition = Math.min(
        Math.max(0, this.carouselPosition + direction),
        this.imageUrls.length - 1
      );
      this.requestImage();
      if (this.hasImageAttention) {
        this.requestFilter();
      }
    },
    hide() {
      this.hidden = true;
      this.$emit("hide");
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
        this.scale = 0;
        datasetMutations.removeFile(
          this.$store,
          imageId(this.selectedImageUrl)
        );
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
      scale: number;
    }) {
      if (this.isMultiBandImage) {
        await datasetActions.fetchMultiBandImage(this.$store, {
          dataset: this.dataset,
          imageId: imageId(this.selectedImageUrl),
          bandCombination: this.info.band,
          isThumbnail: false,
          options: imageOptions,
        });
      } else {
        await datasetActions.fetchImage(this.$store, {
          dataset: this.dataset,
          url: this.selectedImageUrl,
        });
      }
    },
  },
});
</script>

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
.left {
  left: 0px;
  top: 50%;
}
.right {
  right: 0px;
  top: 50%;
}
</style>
