<template>
  <div class="drill-down-container">
    <div class="grid-container" :style="gridColStyle">
      <template v-for="(r, i) in rows">
        <template v-for="(c, j) in cols">
          <div class="image-container">
            <image-label
              class="image-label"
              :dataFields="dataFields"
              includedActive
              shortenLabels
              alignHorizontal
              :item="tilesToRender[i][j].item"
            />
            <image-preview
              class="image-preview"
              :row="tilesToRender[i][j].item"
              :image-url="tilesToRender[i][j].imageUrl"
              :width="imageWidth"
              :height="imageHeight"
              :type="imageType"
              :gray="tilesToRender[i][j].gray"
            />
          </div>
        </template>
      </template>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import ImagePreview from "./ImagePreview.vue";
import ImageLabel from "./ImageLabel.vue";
import { TableRow, TableColumn } from "../store/dataset/index";
import { Dictionary } from "../util/dict";
import { LatLngBounds, LatLngBoundsLiteral } from "leaflet";

interface Tile {
  imageUrl: string;
  item: TableRow;
  coordinates: number[][];
}
interface SpatialIndex {
  x: number;
  y: number;
}
interface Dimensions {
  width: number;
  height: number;
}
export default Vue.extend({
  name: "drill-down",

  components: {
    ImagePreview,
    ImageLabel,
  },

  props: {
    tiles: { type: Array as () => Tile[], default: [] },
    rows: { type: Number, default: 5 },
    cols: { type: Number, default: 7 },
    imageWidth: { type: Number, default: 124 },
    imageHeight: { type: Number, default: 124 },
    imageType: { type: String },
    dataFields: Object as () => Dictionary<TableColumn>,
    bounds: { type: Array as () => number[][] },
    centerTile: {
      type: Object as () => Tile,
      default: { imageUrl: "", item: null, coordinates: null },
    },
  },

  computed: {
    tileDims(): Dimensions {
      return {
        width:
          this.centerTile.coordinates[1][1] - this.centerTile.coordinates[0][1],
        height:
          this.centerTile.coordinates[1][0] - this.centerTile.coordinates[0][0],
      };
    },
    tilesToRender(): Tile[][] {
      return this.spatialSort();
    },
    gridColStyle(): string {
      return `grid-template-columns: repeat(${this.cols}, ${this.imageWidth}px); grid-template-rows: repeat(${this.rows}, ${this.imageHeight}px);`;
    },
  },
  methods: {
    getIndex(x: number, y: number): SpatialIndex {
      const minX = this.bounds[0][1];
      const minY = this.bounds[1][0];
      return {
        x: Math.floor((x - minX) / this.tileDims.width),
        y: Math.floor((y - minY) / this.tileDims.height),
      };
    },
    spatialSort(): Tile[][] {
      if (!this.tiles.length) {
        return [];
      }
      const result = Array.from({ length: this.rows }, (e) =>
        Array(this.cols).fill({
          imageUrl: null,
          item: null,
          coordinates: null,
          gray: 0,
        })
      );
      // loop through and build spatial array
      this.tiles.forEach((t) => {
        const center = new LatLngBounds(
          t.coordinates as LatLngBoundsLiteral
        ).getCenter();
        const indices = this.getIndex(center.lng, center.lat);
        const invertY = this.rows - 1 - indices.y;
        result[invertY][indices.x] = t;
      });
      // normalize coordinates
      return result;
    },
    emitCloseEvent(event) {
      if (!this.$el.contains(event.target)) {
        window.removeEventListener("mousedown", this.emitCloseEvent);
        this.$emit("close");
      }
    },
  },
  mounted() {
    window.addEventListener("mousedown", this.emitCloseEvent);
  },
});
</script>

<style scoped>
.drill-down-container {
  position: absolute;
  display: flex;
  height: 100%;
  width: 100%;
  top: 0;
  right: 0;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-box-direction: normal;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background: rgba(0, 0, 0, 0.54);
}
.image-container {
  position: relative;
  z-index: 0;
  width: 100%;
  height: 100%;
}
.grid-container {
  background: rgba(255, 255, 255, 0.7);
  display: grid;
  grid-gap: 2px;
  padding-bottom: 4px;
}
</style>
