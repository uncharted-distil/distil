<template>
  <div
    v-observe-visibility="visibilityChanged"
    v-bind:class="{ 'is-hidden': !isVisible && !preventHiding }"
    v-bind:style="{ width: `${width}px`, height: `${height}px` }"
  >
    <div
      class="image-container"
      v-bind:class="{ selected: isSelected && isLoaded }"
    >
      <div
        class="image-elem"
        v-bind:class="{ clickable: hasClick }"
        ref="imageElem"
        @click.stop="handleClick"
        v-bind:style="{ 'max-width': `${width}px` }"
      >
        <div v-if="!isLoaded" v-html="spinnerHTML"></div>
      </div>
    </div>
    <b-modal
      id="image-zoom-modal"
      :title="imageUrl"
      @hide="hideModal"
      :visible="!!zoomImage"
      hide-footer
    >
      <div v-if="availableBands.length > 0">
        <b-dropdown :text="band" size="sm">
          <b-dropdown-item
            v-for="bandInfo in availableBands"
            :key="bandInfo.id"
            @click="setBandCombination(bandInfo.id)"
            >{{ bandInfo.displayName }}</b-dropdown-item
          >
        </b-dropdown>
      </div>
      <div class="image-elem-zoom" ref="imageElemZoom"></div>
    </b-modal>
  </div>
</template>

<script lang="ts">
import $ from "jquery";
import _ from "lodash";
import Vue from "vue";
import {
  getters as datasetGetters,
  actions as datasetActions
} from "../store/dataset/module";
import { getters as routeGetters } from "../store/route/module";
import { circleSpinnerHTML } from "../util/spinner";
import {
  D3M_INDEX_FIELD,
  TableRow,
  RowSelection,
  BandID,
  BandCombination,
  TaskTypes
} from "../store/dataset/index";
import { isRowSelected } from "../util/row";
import { Dictionary } from "../util/dict";
import { REMOTE_SENSING_TYPE, IMAGE_TYPE } from "../util/types";
import { createRouteEntry, overlayRouteEntry } from "../util/routes";

export default Vue.extend({
  name: "image-preview",

  props: {
    row: Object as () => TableRow,
    imageUrl: String as () => string,
    type: String as () => string,
    width: {
      default: 64,
      type: Number as () => number
    },
    height: {
      default: 64,
      type: Number as () => number
    },
    preventHiding: {
      default: false,
      type: Boolean as () => boolean
    },
    onClick: Function
  },

  watch: {
    imageUrl(newUrl: string, oldUrl: string) {
      if (newUrl !== oldUrl) {
        this.hasRendered = false;
        this.hasRequested = false;
        if (!this.image) {
          this.clearImage();
          this.requestImage();
        } else {
          this.injectImage();
        }
      }
    },
    // Refresh image on band change
    band(newBand: string, oldBand: string) {
      if (newBand !== oldBand) {
        this.clearImage();
        this.requestImage();
      }
    }
  },

  data() {
    return {
      zoomImage: false,
      entry: null,
      zoomedWidth: 400,
      zoomedHeight: 400,
      isVisible: false,
      hasRendered: false,
      hasRequested: false
    };
  },

  updated() {
    this.$nextTick(this.injectZoomedImage);
  },

  computed: {
    imageId(): string {
      return this.imageUrl.split(/_B[0-9][0-9a-zA-Z][.]/)[0];
    },
    files(): Dictionary<any> {
      return datasetGetters.getFiles(this.$store);
    },
    isLoaded(): boolean {
      return !!this.files[this.imageUrl] && !!this.files[this.imageId];
    },
    image(): HTMLImageElement {
      return this.files[this.imageUrl]
        ? this.files[this.imageUrl]
        : this.files[this.imageId]
        ? this.files[this.imageId]
        : null;
    },
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
    dataset(): string {
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
    },
    band(): string {
      const bandID = routeGetters.getBandCombinationId(this.$store);
      return this.availableBands.find(b => b.id === bandID)?.displayName;
    },
    availableBands(): BandCombination[] {
      // Show available bands for remote sensing only
      if (this.type === REMOTE_SENSING_TYPE) {
        return datasetGetters.getMultiBandCombinations(this.$store);
      }
      return [];
    }
  },

  methods: {
    visibilityChanged(isVisible: boolean) {
      this.isVisible = isVisible;
      if (this.isVisible && !this.hasRequested) {
        this.requestImage();
        return;
      }
      if (this.isVisible && this.hasRequested && !this.hasRendered) {
        this.injectImage();
      }
    },

    handleClick() {
      if (this.onClick) {
        this.onClick({
          row: this.row,
          imageUrl: this.imageUrl,
          image: this.image
        });
      }
    },

    // Writes a new band combination into the route
    setBandCombination(bandID: BandID) {
      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        bandCombinationId: bandID
      });
      this.$router.push(entry);
    },

    injectZoomedImage() {
      const $elem = this.$refs.imageElemZoom as any;
      if (this.image && $elem) {
        $elem.innerHTML = "";
        $elem.appendChild(
          this.clonedImageElement(this.zoomedWidth, this.zoomedHeight)
        );
      }
    },

    showZoomedImage() {
      this.zoomImage = true;
    },

    hideModal() {
      this.zoomImage = false;
    },

    clonedImageElement(width: number, height: number): HTMLImageElement {
      const img = this.image.cloneNode();
      $(img).css("max-width", `${width}px`);
      $(img).css("max-height", `${height}px`);
      return img as HTMLImageElement;
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
        elem.appendChild(this.clonedImageElement(this.width, this.height));
        const icon = document.createElement("i");
        icon.className += "fa fa-search-plus zoom-icon";
        $(icon).click(event => {
          this.showZoomedImage();
          event.stopPropagation();
        });
        elem.appendChild(icon);
        this.hasRendered = true;
      }
    },

    async requestImage() {
      this.hasRequested = true;
      if (this.type === IMAGE_TYPE) {
        await datasetActions.fetchImage(this.$store, {
          dataset: this.dataset,
          url: this.imageUrl
        });
        if (this.isVisible) {
          this.injectImage();
        }
      } else if (this.type === REMOTE_SENSING_TYPE) {
        await datasetActions.fetchMultiBandImage(this.$store, {
          dataset: this.dataset,
          imageId: this.imageId,
          bandCombination: routeGetters.getBandCombinationId(this.$store)
        });
        if (this.isVisible) {
          this.injectImage();
        }
      } else {
        console.warn(`Image Data Type ${this.type} is not supported`);
      }
    }
  },
  async beforeMount() {
    // lazy fetch available band types
    if (
      this.type === REMOTE_SENSING_TYPE &&
      _.isEmpty(datasetGetters.getMultiBandCombinations(this.$store))
    ) {
      await datasetActions.fetchMultiBandCombinations(this.$store, {
        dataset: this.dataset
      });
    }
  }
});
</script>

<style>
.image-container {
  border: 2px solid rgba(0, 0, 0, 0);
}
.image-container.selected {
  border: 2px solid #ff0067;
}

.image-elem {
  position: relative;
}
.image-elem:hover {
  background-color: #000;
}
.image-elem img {
  position: relative;
}
.image-elem.clickable {
  cursor: pointer;
}
.image-elem.clickable img:hover {
  opacity: 0.7;
}

.image-elem.clickable zoom-icon:hover {
  opacity: 0.7;
}

.image-elem-zoom {
  position: relative;
  text-align: center;
}
.image-elem-zoom img {
  position: relative;
  padding: 8px 16px;
  max-width: 100%;
  border-radius: 4px;
}
.image-elem .zoom-icon {
  position: absolute;
  right: 0;
  top: 0;
  padding: 4px;
  color: #fff;
  visibility: hidden;
}
.image-elem:hover .zoom-icon {
  visibility: visible;
}

.zoom-icon {
  cursor: pointer;
  background-color: #424242;
}

.is-hidden {
  visibility: hidden;
}
</style>
