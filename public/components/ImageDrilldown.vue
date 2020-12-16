<template>
  <b-modal
    size="lg"
    hide-footer
    :title="visibleTitle"
    :visible="visible"
    @hide="hide"
  >
    <main class="drill-down">
      <image-label
        v-if="item && dataFields"
        class="image-label"
        :data-fields="dataFields"
        included-active
        :item="item"
      />

      <section
        ref="imageContainer"
        class="image-container"
        :style="{ '--IMAGE_MAX_SIZE': IMAGE_MAX_SIZE + 'px' }"
      />

      <ul class="information">
        <li v-if="bandName"><label>Band:</label> {{ bandName }}</li>
        <li v-if="latLongValue"><label>Lat/Long:</label> {{ latLongValue }}</li>
        <li v-if="isMultiBandImageType" class="information-brightness">
          <b-input-group prepend="0" append="1.0" class="mt-3 mb-1" size="sm">
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
        </li>
      </ul>
    </main>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import ImageLabel from "./ImageLabel.vue";
import { TableColumn, TableRow } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import { MULTIBAND_IMAGE_TYPE, IMAGE_TYPE } from "../util/types";

const IMAGE_MAX_SIZE = 750; // Maximum size of an image in the drill-down in pixels.
const IMAGE_MAX_ZOOM = 3; // We don't want an image to be too magnified to avoid blurriness.

const imageId = (imageUrl) => imageUrl?.split(/_B[0-9][0-9a-zA-Z][.]/)[0];

export interface DrillDownInfo {
  band?: string;
  title?: string;
}

/**
 * Display a modal with drill-downed information about an image.
 *
 * @param dataFields {Array<TableColumn>}
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
  },

  props: {
    dataFields: Object as () => Dictionary<TableColumn>,
    url: String,
    type: { type: String, default: IMAGE_TYPE },
    info: Object as () => DrillDownInfo,
    item: Object as () => TableRow,
    visible: Boolean,
  },

  data() {
    return {
      IMAGE_MAX_SIZE: IMAGE_MAX_SIZE,
      currentBrightness: 0.5,
      brightnessMin: 0,
      brightnessMax: 100,
    };
  },

  computed: {
    bandName(): string {
      return datasetGetters
        .getMultiBandCombinations(this.$store)
        .find((band) => band.id === this.info.band)?.displayName;
    },

    brightnessValue(): string {
      return this.currentBrightness.toFixed(2);
    },

    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    files(): Dictionary<any> {
      return datasetGetters.getFiles(this.$store);
    },

    image(): HTMLImageElement {
      return this.files[this.url] ?? this.files[imageId(this.url)] ?? null;
    },

    isMultiBandImageType(): boolean {
      return this.type === MULTIBAND_IMAGE_TYPE;
    },

    latLongValue(): string {
      if (!this.item?.coordinates) return;
      const coordinates = this.item.coordinates.value.Elements;
      if (coordinates.some((x) => x === undefined)) return;

      /*
        Item store the coordinates as a list of 8 values being four pairs of 
        [Long, Lat], one for each corner of the isMultiBandImage-sensing image.

        [0,1]     [2,3]
          A-------B
          |       |
          |       |
          D-------C
        [6,7]     [4,5]
      */

      // Corner A as [Lat, Long]
      const cornerA = `[${coordinates[1].Float}, ${coordinates[0].Float}]`;
      // Corner C as [Lat, Long]
      const cornerC = `[${coordinates[5].Float}, ${coordinates[4].Float}]`;

      return `From ${cornerA} to ${cornerC}`;
    },

    visibleTitle(): string {
      return this.info.title ?? this.url ?? "Image Drilldown";
    },
  },

  watch: {
    visible() {
      if (this.visible) {
        this.requestImage();
      }
    },
  },

  methods: {
    hide() {
      this.$emit("hide");
    },

    onBrightnessChanged(e) {
      const MAX_GAINL = 2.0;
      const val = Number(e) / this.brightnessMax;
      const gainL = val * MAX_GAINL;
      this.currentBrightness = val;
      this.requestImage({ gainL, gamma: 2.2, gain: 2.5 }); // gamma, gain, are default. They are here if we need to edit them later down the road
    },

    cleanUp() {
      if (!this.isMultiBandImageType) return;
      datasetMutations.removeFile(this.$store, imageId(this.url));
    },

    injectImage() {
      const container = this.$refs.imageContainer as any;

      if (!!this.image && container) {
        const image = this.image.cloneNode() as HTMLImageElement;

        // Calculate how much bigger we can make the image to fit in the modal box.
        const ratio = Math.min(
          IMAGE_MAX_SIZE / Math.max(this.image.height, this.image.width),
          IMAGE_MAX_ZOOM
        );

        // Update the image to be bigger, but not bigger than the modal box.
        image.height = this.image.height * ratio;
        image.width = this.image.width * ratio;

        // Add the image to the container.
        container.innerHTML = "";
        container.appendChild(image);
      }
    },

    async requestImage(imageOptions?: {
      gamma: number;
      gain: number;
      gainL: number;
    }) {
      if (this.isMultiBandImageType) {
        await datasetActions.fetchMultiBandImage(this.$store, {
          dataset: this.dataset,
          imageId: imageId(this.url),
          bandCombination: this.info.band,
          isThumbnail: false,
          options: imageOptions,
        });
      } else {
        await datasetActions.fetchImage(this.$store, {
          dataset: this.dataset,
          url: this.url,
        });
      }

      this.injectImage();
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

.brightness-slider {
  width: 70%;
}
</style>
