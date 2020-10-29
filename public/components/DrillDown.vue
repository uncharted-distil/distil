<template>
  <div class="drill-down-container">
    <b-container v-if="rows * cols == tilesToRender.length">
      <b-row v-for="r in rows" :key="r">
        <b-col v-for="c in cols" :key="c">
          <image-label
            class="image-label"
            :dataFields="dataFields"
            includedActive
            shortenLabels
            alignHorizontal
            :item="tilesToRender[r * c + c].item"
          />
          <image-preview
            class="image-preview"
            :row="tilesToRender[r * c].item"
            :image-url="tilesToRender[r * c + c].imageUrl"
            :width="imageWidth"
            :height="imageHeight"
            :type="imageType"
          />
        </b-col>
      </b-row>
    </b-container>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import ImagePreview from "./ImagePreview.vue";
import ImageLabel from "./ImageLabel.vue";
import { TableRow, TableColumn } from "../store/dataset/index";
import { Dictionary } from "../util/dict";

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
    imageWidth: { type: Number, default: 128 },
    imageHeight: { type: Number, default: 128 },
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
    tilesToRender(): Tile[] {
      return this.spatialSort();
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
    spatialSort(): Tile[] {
      if (!this.tiles.length) {
        return [];
      }
      const result = Array(this.rows * this.cols).fill({
        imageUrl: "null",
        item: { isExcluded: true },
        coordinates: null,
      });
      // loop through and build spatial array
      this.tiles.forEach((t) => {
        const indices = this.getIndex(t.coordinates[0][1], t.coordinates[0][0]);
        result[indices.y * indices.x + indices.x] = t;
      });
      // normalize coordinates
      return result;
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
}
</style>
