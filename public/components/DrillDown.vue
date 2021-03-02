<template>
  <div class="drill-down-container">
    <div>
      <div class="toolbar">
        <div class="title">{{ title }}</div>
        <b-button class="exit-button" @click="onExitClicked"
          ><span aria-hidden="true">&times;</span></b-button
        >
      </div>
      <div class="grid-container" :style="gridColStyle">
        <template v-for="(r, i) in renderTiles.length">
          <template v-for="(c, j) in renderTiles[i].length">
            <div class="image-container">
              <image-label
                class="image-label"
                :dataFields="dataFields"
                includedActive
                shortenLabels
                alignHorizontal
                :item="renderTiles[i][j].selected.item"
              />
              <image-preview
                class="image-preview"
                :row="renderTiles[i][j].selected.item"
                :image-url="renderTiles[i][j].selected.imageUrl"
                :width="imageWidth"
                :height="imageHeight"
                :type="imageType"
                :gray="renderTiles[i][j].selected.gray"
                @click="onImageClick"
                :overlappedUrls="
                  renderTiles[i][j].overlapped.map((o) => o.imageUrl)
                "
              />
              <overlap-selection
                :items="renderTiles[i][j].overlapped"
                :indices="{ y: i, x: j }"
                :instanceName="`over-lap-${i}-${j}`"
                :width="imageWidth"
                :height="imageHeight"
                :imageType="imageType"
                @item-selected="onOverlapSelected"
              />
            </div>
          </template>
        </template>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import ImagePreview from "./ImagePreview.vue";
import ImageLabel from "./ImageLabel.vue";
import {
  TableRow,
  TableColumn,
  D3M_INDEX_FIELD,
  RowSelection,
} from "../store/dataset/index";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
} from "../util/row";
import { clearAreaOfInterest } from "../util/data";
import { Dictionary } from "../util/dict";
import { getters as routeGetters } from "../store/route/module";
import { LatLngBounds, LatLngBoundsLiteral } from "leaflet";
import OverlapSelection from "./OverlapSelection.vue";
import { CoordinateInfo } from "../util/rendering/coordinates";

interface Tile {
  imageUrl: string;
  item: TableRow;
  info: CoordinateInfo;
}
interface RenderTile {
  selected: Tile;
  overlapped: Tile[];
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
    OverlapSelection,
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
    instanceName: { type: String as () => string, default: "" },
  },
  data() {
    return {
      renderTiles: [] as RenderTile[][],
    };
  },
  mounted() {
    this.renderTiles = this.spatialSort();
  },
  watch: {
    tiles() {
      this.renderTiles = this.spatialSort();
    },
  },
  computed: {
    tileDims(): Dimensions {
      return {
        width:
          this.centerTile.info.coordinates[1][1] -
          this.centerTile.info.coordinates[0][1],
        height:
          this.centerTile.info.coordinates[1][0] -
          this.centerTile.info.coordinates[0][0],
      };
    },
    gridColStyle(): string {
      return `grid-template-columns: repeat(${this.cols}, ${this.imageWidth}px); grid-template-rows: repeat(${this.rows}, ${this.imageHeight}px);padding:5px;`;
    },
    title(): string {
      return `coordinates [${this.bounds[0][1].toFixed(
        2
      )}, ${this.bounds[0][0].toFixed(2)}] to [${this.bounds[1][1].toFixed(
        2
      )}, ${this.bounds[1][0].toFixed(2)}]`;
    },
    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
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
    spatialSort(): RenderTile[][] {
      const result = [];
      for (let i = 0; i < this.rows; ++i) {
        result.push([]);
        for (let j = 0; j < this.cols; ++j) {
          result[i].push({
            selected: {
              imageUrl: null,
              item: null,
              coordinates: null,
              gray: 0,
            },
            overlapped: [],
          });
        }
      }
      if (!this.tiles.length) {
        return result;
      }
      // loop through and build spatial array
      this.tiles.forEach((t) => {
        const center = new LatLngBounds(
          t.info.coordinates as LatLngBoundsLiteral
        ).getCenter();
        const indices = this.getIndex(center.lng, center.lat);
        if (
          indices.x < 0 ||
          indices.y < 0 ||
          indices.x >= this.cols ||
          indices.y >= this.rows
        ) {
          // tile outside defined area
          return;
        }
        const invertY = this.rows - 1 - indices.y;
        result[invertY][indices.x].selected = t;
        result[invertY][indices.x].overlapped.push(t);
      });
      // normalize coordinates
      return result;
    },
    onExitClicked() {
      this.$emit("close");
      clearAreaOfInterest();
    },
    onImageClick(event: any) {
      if (!isRowSelected(this.rowSelection, event.row[D3M_INDEX_FIELD])) {
        addRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          event.row[D3M_INDEX_FIELD]
        );
      } else {
        removeRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          event.row[D3M_INDEX_FIELD]
        );
      }
    },
    onOverlapSelected(info: { item: Tile; key: { x: number; y: number } }) {
      this.renderTiles[info.key.y][info.key.x].selected = info.item;
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
.stack-button {
  position: absolute;
  bottom: 0px;
  left: 0px;
}
</style>
