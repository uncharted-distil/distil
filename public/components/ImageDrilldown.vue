<template>
  <b-modal
    size="lg"
    @hide="hide"
    hide-footer
    :title="visibleTitle"
    :visible="visible"
  >
    <div class="drill-down">
      <image-label
        v-if="item && dataFields"
        class="image-label"
        :dataFields="dataFields"
        includedActive
        :item="item"
      />
      <div
        class="image-container"
        ref="imageContainer"
        :style="{ '--IMAGE_MAX_SIZE': IMAGE_MAX_SIZE + 'px' }"
      ></div>
      <div class="slider-container">
        <label class="slider-label">0.0 </label>
        <b-form-input
          v-if="isMultiBandImage"
          type="range"
          name="brightness"
          :min="min"
          :max="max"
          step="1"
          class="slider"
          @change="onSliderChanged"
        />
        <label class="slider-label">1.0</label>
      </div>
      <div>
        <i class="fa fa-adjust" aria-hidden="true" />
        <label class="slider-label">{{ sliderVal }}</label>
      </div>
    </div>
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

const IMAGE_MAX_SIZE = 750; // Maximum size of an image in the drilldown in pixels.
const IMAGE_MAX_ZOOM = 3; // We don't want an image to be too magnified to avoid blurriness.

const imageId = (imageUrl) => imageUrl?.split(/_B[0-9][0-9a-zA-Z][.]/)[0];

/**
 * Display a modal with drilldowned information about an image.
 *
 * @param visible    {Boolean} Display or hide the modal.
 * @param imageUrl   {String}  URL of the image to be drilldown.
 * @param title      {String=} Title of the modal.
 * @param dataFields {Array<TableColumn>}
 * @param item       {TableRow} item being drilldown.
 */
export default Vue.extend({
  name: "image-drilldown",

  components: {
    ImageLabel,
  },

  props: {
    dataFields: Object as () => Dictionary<TableColumn>,
    imageUrl: String,
    item: Object as () => TableRow,
    title: String,
    visible: Boolean,
  },

  data() {
    return {
      IMAGE_MAX_SIZE: IMAGE_MAX_SIZE,
      currentVal: 0.5,
    };
  },
  computed: {
    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },

    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    files(): Dictionary<any> {
      return datasetGetters.getFiles(this.$store);
    },

    image(): HTMLImageElement {
      return (
        this.files[this.imageUrl] ?? this.files[imageId(this.imageUrl)] ?? null
      );
    },

    isMultiBandImage(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },

    visibleTitle(): string {
      return this.title ?? this.imageUrl ?? "Image Drilldown";
    },
    sliderVal(): string {
      return this.currentVal.toFixed(2);
    },
    max(): number {
      return 100;
    },
    min(): number {
      return 0;
    },
  },

  methods: {
    hide() {
      this.$emit("hide");
    },
    onSliderChanged(e) {
      const MAX_GAINL = 2.0;
      const val = Number(e) / this.max;
      const gainL = val * MAX_GAINL;
      this.currentVal = val;
      this.requestImage({ gainL, gamma: 2.2, gain: 2.5 }); // gamma, gain, are default. They are here if we need to edit them later down the road
    },
    cleanUp() {
      if (this.isMultiBandImage) {
        datasetMutations.removeFile(this.$store, imageId(this.imageUrl));
        return;
      }
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
      if (this.isMultiBandImage) {
        await datasetActions.fetchMultiBandImage(this.$store, {
          dataset: this.dataset,
          imageId: imageId(this.imageUrl),
          bandCombination: this.band,
          isThumbnail: false,
          options: imageOptions,
        });
      } else {
        await datasetActions.fetchImage(this.$store, {
          dataset: this.dataset,
          url: this.imageUrl,
        });
      }
      this.injectImage();
    },
  },
  watch: {
    visible() {
      if (this.visible) {
        this.requestImage();
      }
    },
  },
});
</script>

<style scoped>
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
.drill-down {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.slider-container {
  display: flex;
  align-items: center;
}
.slider-label {
  margin-bottom: 0px;
  padding-left: 5px;
  padding-right: 5px;
  display: inline-block;
}
.slider {
  width: 70%;
}
</style>
