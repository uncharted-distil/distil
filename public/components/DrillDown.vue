<template>
  <div class="drill-down-container">
    <div>
      <div class="toolbar">
        <div class="title">{{ title }}</div>
        <b-button class="exit-button" @click="onExitClicked">x</b-button>
      </div>
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
    title(): string {
      return `coordinates [${this.bounds[0][1].toFixed(
        2
      )}, ${this.bounds[0][0].toFixed(2)}] to [${this.bounds[1][1].toFixed(
        2
      )}, ${this.bounds[1][0].toFixed(2)}]`;
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
    onExitClicked() {
      this.$emit("close");
    },
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
.exit-button {
  width: 30px;
  height: 30px;
  z-index: 999;
  text-align: center;
  float: right;
  background: none;
  border: none;
  color: #8b8b8b;
  border-top-right-radius: 0px;
  font-size: 1.407rem;
  font-weight: 600;
  line-height: 0;
}
.toolbar {
  background: rgba(255, 255, 255, 0.7);
  height: 30px;
  border-bottom: 1px solid #8b8b8b;
}
.title {
  background-color: #255dcc;
  display: inline;
  color: #fff;
  padding-left: 8px;
  border-radius: 4px;
  margin: 2px;
  height: 26px;
  position: absolute;
  padding-right: 8px;
}
</style>
