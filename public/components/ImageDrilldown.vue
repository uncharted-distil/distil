<template>
  <b-modal @hide="hide" hide-footer :title="visibleTitle" :visible="visible">
    <div v-if="isRemoteSensing && availableBands.length > 0">
      <b-dropdown :text="band" size="sm">
        <b-dropdown-item
          v-for="bandInfo in availableBands"
          :key="bandInfo.id"
          @click="setBandCombination(bandInfo.id)"
          >{{ bandInfo.displayName }}
        </b-dropdown-item>
      </b-dropdown>
    </div>
    <div class="image-container" ref="imageContainer"></div>
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
import { BandID, BandCombination } from "../store/dataset/index";
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
 * @param visible  {Boolean} Display or hide the modal.
 * @param imageUrl {String}  URL of the image to be drilldown.
 * @param title    {String=} Title of the modal.
 */
export default Vue.extend({
  name: "image-drilldown",

  props: {
    visible: Boolean,
    imageUrl: String,
    title: String
  },

  mounted() {
    this.injectImage();
  },

  updated() {
    this.$nextTick(this.injectImage);
  },

  computed: {
    availableBands(): BandCombination[] {
      return datasetGetters.getMultiBandCombinations(this.$store);
    },

    band(): string {
      const bandID = routeGetters.getBandCombinationId(this.$store);
      return this.availableBands.find(b => b.id === bandID)?.displayName;
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

    isRemoteSensing(): boolean {
      return routeGetters.isRemoteSensing(this.$store);
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
    },

    async requestMultiBandImage() {
      await datasetActions.fetchMultiBandImage(this.$store, {
        dataset: this.dataset,
        imageId: imageId(this.imageUrl),
        bandCombination: routeGetters.getBandCombinationId(this.$store)
      });

      // Display the image for that band.
      this.injectImage();
    },

    setBandCombination(bandID: BandID) {
      // Writes a new band combination into the route
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        bandCombinationId: bandID
      });
      this.$router.push(entry);

      // Request the new band of the image.
      this.requestMultiBandImage();
    }
  }
});
</script>

<style>
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
