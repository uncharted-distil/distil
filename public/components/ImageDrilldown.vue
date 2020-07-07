<template>
  <b-modal @hide="hide" hide-footer :title="visibleTitle" :visible="visible">
    <image-label
      v-if="item && dataFields"
      class="image-label"
      :dataFields="dataFields"
      includedActive
      :item="item"
    />
    <div class="image-container" ref="imageContainer"></div>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import ImageLabel from "./ImageLabel";
import {
  BandID,
  BandCombination,
  TableColumn,
  TableRow
} from "../store/dataset/index";
import {
  getters as datasetGetters,
  actions as datasetActions
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import { overlayRouteEntry } from "../util/routes";

const imageId = imageUrl => imageUrl?.split(/_B[0-9][0-9a-zA-Z][.]/)[0];

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
    ImageLabel
  },

  props: {
    dataFields: Object as () => Dictionary<TableColumn>,
    imageUrl: String,
    item: Object as () => TableRow,
    title: String,
    visible: Boolean
  },

  mounted() {
    this.injectImage();
  },

  updated() {
    this.$nextTick(this.injectImage);
  },

  computed: {
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

    visibleTitle(): string {
      return this.title ?? this.imageUrl ?? "Image Drilldown";
    }
  },

  methods: {
    hide() {
      this.$emit("hide");
    },

    injectImage() {
      const container = this.$refs.imageContainer as any;

      if (!!this.image && container) {
        container.innerHTML = "";
        container.appendChild(this.image.cloneNode() as HTMLImageElement);
      }
    }
  }
});
</script>

<style scoped>
.image-container {
  /* Keep the image under 25% of screen width. */
  max-height: 25vw;
  max-width: 25vw;

  text-align: center;
}

/* Keep the image in its container. */
.image-container img {
  max-height: 100%;
  max-width: 100%;
  position: relative;
}
</style>
