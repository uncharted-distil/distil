<template>
  <b-modal
    size="lg"
    @hide="hide"
    hide-footer
    :title="visibleTitle"
    :visible="visible"
  >
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
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import ImageLabel from "./ImageLabel.vue";
import { TableColumn, TableRow } from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions,
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
    };
  },

  mounted() {
    this.injectImage();
  },

  updated() {
    this.requestImage();
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
  },

  methods: {
    hide() {
      this.$emit("hide");
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

    async requestImage() {
      if (this.isMultiBandImage) {
        await datasetActions.fetchMultiBandImage(this.$store, {
          dataset: this.dataset,
          imageId: imageId(this.imageUrl),
          bandCombination: this.band,
          isThumbnail: false,
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
</style>
