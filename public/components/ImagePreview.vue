<template>
  <div
    v-observe-visibility="visibilityChanged"
    :class="{ 'is-hidden': !isVisible && !preventHiding }"
    :style="{
      width: `${width}px`, // + 2 for boarder
      height: `${height}px`, // boarder
      filter: `grayscale(${gray}%)`,
    }"
  >
    <div class="image-container" :class="{ selected: isSelected && isLoaded }">
      <template v-if="!isLoaded">
        <div v-html="spinnerHTML"></div>
      </template>
      <template v-else-if="!stopSpinner">
        <div
          ref="imageElem"
          class="image-elem"
          :class="{ clickable: hasClick }"
          @click.stop="handleClick"
        />
        <div
          ref="imageAttentionElem"
          class="filter-elem"
          :class="{ clickable: hasClick }"
          @click.stop="handleClick"
        />
        <i class="fa fa-search-plus zoom-icon" @click.stop="showZoomedImage" />
      </template>
    </div>
    <image-drilldown
      :imageUrl="imageUrl"
      :title="imageUrl"
      :visible="!!zoomImage"
      @hide="hideZoomImage"
    />
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import ImageDrilldown from "./ImageDrilldown.vue";
import {
  getters as datasetGetters,
  actions as datasetActions,
  mutations as datasetMutations,
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { circleSpinnerHTML } from "../util/spinner";
import {
  D3M_INDEX_FIELD,
  TableRow,
  RowSelection,
} from "../store/dataset/index";
import { isRowSelected } from "../util/row";
import { Dictionary } from "../util/dict";
import { MULTIBAND_IMAGE_TYPE, IMAGE_TYPE } from "../util/types";
import { ColorScaleNames } from "../util/data";

export default Vue.extend({
  name: "image-preview",

  components: {
    ImageDrilldown,
  },

  props: {
    row: Object as () => TableRow,
    imageUrl: String as () => string,
    uniqueTrail: { type: String as () => string, default: "" },
    type: String as () => string,
    width: {
      default: 64,
      type: Number as () => number,
    },
    height: {
      default: 64,
      type: Number as () => number,
    },
    preventHiding: {
      default: false,
      type: Boolean as () => boolean,
    },
    onClick: Function,
    gray: { type: Number, default: 0 }, // support for graying images.
    debounce: { type: Boolean as () => boolean, default: false },
    debounceWaitTime: { type: Number as () => number, default: 500 },
  },

  watch: {
    isLoaded() {
      this.$nextTick(async () => {
        if (!this.isLoaded) {
          return;
        }
        await this.handleImageAttention();
      });
    },
    imageUrl(newUrl: string, oldUrl: string) {
      if (newUrl === null) {
        return;
      }
      if (newUrl !== oldUrl) {
        this.cleanUp();
        this.hasRendered = false;
        this.hasRequested = false;
        this.clearImage();
        this.getImage();
      }
    },
    async hasImageAttention() {
      await this.handleImageAttention();
    },
    async row(newRow: TableRow, oldRow: TableRow) {
      await this.handleImageAttention();
    },
    async colorScale() {
      if (!!this.solutionId && !!this.row) {
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
          this.clearImage();
          this.getImage();
        }
      }
    },
    debounce(newVal: boolean) {
      if (newVal) {
        this.getImage = this.debouncedRequestImage;
        return;
      }
      this.getImage = this.requestImage;
    },
  },

  data() {
    return {
      zoomImage: false,
      entry: null,
      zoomedWidth: 400,
      zoomedHeight: 400,
      isVisible: false,
      hasRendered: false,
      hasRequested: false,
      stopSpinner: false,
      debouncedRequestImage: null,
      getImage: null,
      imageAttentionHasRendered: false,
    };
  },

  computed: {
    colorScale(): ColorScaleNames {
      return routeGetters.getColorScale(this.$store);
    },
    imageId(): string {
      return this.imageUrl?.split(/_B[0-9][0-9a-zA-Z][.]/)[0];
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
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
    dataset(): string {
      const dataset = routeGetters.getRoutePredictionsDataset(this.$store);
      if (dataset) {
        return dataset;
      }
      return routeGetters.getRouteDataset(this.$store);
    },
    hasClick(): boolean {
      return !!this.onClick;
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
  },

  methods: {
    async visibilityChanged(isVisible: boolean) {
      this.isVisible = isVisible;
      if (this.isVisible && !this.hasRequested) {
        this.getImage();
        await this.handleImageAttention();
        return;
      }
      if (this.isVisible && this.hasRequested && !this.hasRendered) {
        this.injectImage();
      }
    },
    async handleImageAttention() {
      this.hasRendered = false;
      this.clearImage(this.$refs.imageAttentionElem as any);
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
        this.clearImage(this.$refs.imageAttentionElem as any);
        this.imageAttentionHasRendered = false;
      }
      if (!this.hasImageAttention && this.imageAttentionHasRendered) {
        this.clearImage(this.$refs.imageAttentionElem as any);
        this.imageAttentionHasRendered = false;
      }
    },
    handleClick() {
      if (this.onClick) {
        this.onClick({
          row: this.row,
          imageUrl: this.imageUrl,
          image: this.image,
        });
      }
    },

    showZoomedImage() {
      this.zoomImage = true;
    },
    hideZoomImage() {
      this.zoomImage = false;
    },

    clearImage(elem?: any) {
      const $elem = elem || (this.$refs.imageElem as any);
      if ($elem) {
        $elem.innerHTML = "";
      }
    },

    injectImage() {
      if (!this.image) {
        return;
      }

      const elem = this.$refs.imageElem as any;
      if (elem) {
        this.clearImage(elem);
        const image = this.image.cloneNode() as HTMLImageElement;
        elem.appendChild(image);

        // fit image preview to available area with no overflows
        if (
          this.width === this.height &&
          elem.children[0].height > elem.children[0].width
        ) {
          elem.children[0].style.height = elem.children[0].width + "px";
        }
        this.hasRendered = true;
      }
    },
    injectFilter() {
      if (!this.imageAttention) {
        return;
      }

      const elem = this.$refs.imageAttentionElem as any;
      if (elem) {
        this.clearImage(elem);
        const image = this.imageAttention.cloneNode() as HTMLImageElement;
        elem.appendChild(image);

        // fit image preview to available area with no overflows
        if (
          this.width === this.height &&
          elem.children[0].height > elem.children[0].width
        ) {
          elem.children[0].style.height = elem.children[0].width + "px";
        }
        this.imageAttentionHasRendered = true;
      }
    },

    async requestImage() {
      if (this.imageUrl === null) {
        this.stopSpinner = true; // imageUrl is null stop spinner
        return;
      }
      this.hasRequested = true;
      if (this.type === IMAGE_TYPE) {
        await datasetActions.fetchImage(this.$store, {
          dataset: this.dataset,
          url: this.imageUrl,
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
      const empty = "";
      if (this.uniqueTrail !== empty) {
        datasetMutations.removeFile(this.$store, this.imageParamId);
        if (!!this.imageAttention) {
          datasetMutations.removeFile(this.$store, this.imageAttentionId);
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
  created() {
    this.debouncedRequestImage = _.debounce(
      this.requestImage.bind(this),
      this.debounceWaitTime
    );
    if (this.debounce) {
      this.getImage = this.debouncedRequestImage;
    } else {
      this.getImage = this.requestImage;
    }
  },
  destroyed() {
    this.cleanUp();
    if (this.debounce) {
      this.getImage.cancel();
    }
  },
});
</script>

<style>
.is-hidden {
  visibility: hidden;
}

.image-container {
  position: relative;
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
