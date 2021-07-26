<!--

    Copyright © 2021 Uncharted Software Inc.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
-->

<template>
  <div
    id="geo-test"
    ref="geoPlotContainer"
    class="geo-plot-container"
    :class="{ 'selection-mode': isSelectionMode }"
  >
    <div
      :id="mapID"
      ref="geoPlot"
      class="geo-plot"
      tabindex="0"
      @keydown.esc="onEsc"
    />

    <drill-down
      v-if="showDrillDown"
      :data-fields="dataFields"
      :image-type="imageType"
      :tiles="drillDownState.tiles"
      :center-tile="drillDownState.centerTile"
      :bounds="drillDownState.bounds"
      :summaries="summaries"
      @close="onFocusOut"
    />
    <geoplot-toggle-buttons
      :is-clustering="isClustering"
      :is-satellite-view="isSatelliteView"
      :is-selection-mode="isSelectionMode"
      :is-hiding-baseline="isHidingBaseline"
      @map-toggle="mapToggle"
      @clustering-toggle="toggleClustering"
      @selection-tool-toggle="toggleSelectionTool"
      @baseline-toggle="baselineToggle"
    />
    <b-toast
      :id="toastId"
      :title="toastTitle"
      style="position: absolute; top: 0px; right: 0px"
      static
      no-auto-hide
    >
      <div class="geo-plot">
        <image-label
          class="image-label"
          :data-fields="dataFields"
          included-active
          shorten-labels
          align-horizontal
          :item="hoverItem"
        />
        <image-preview
          class="image-preview"
          :row="hoverItem"
          :image-url="hoverUrl"
          :width="imageWidth"
          :height="imageHeight"
          :type="imageType"
        />
      </div>
    </b-toast>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import ImageLabel from "./ImageLabel.vue";
import GeoplotToggleButtons from "./GeoplotToggleButtons.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import lumo from "lumo";
import BatchQuadOverlay from "../util/rendering/BatchQuadOverlay";
import {
  BatchQuadOverlayRenderer,
  EVENT_TYPES,
  DRAW_MODES,
} from "../util/rendering/BatchQuadOverlayRenderer";
import {
  TableColumn,
  TableRow,
  Highlight,
  RowSelection,
  GeoCoordinateGrouping,
  VariableSummary,
  GeoBoundsGrouping,
  Variable,
} from "../store/dataset/index";
import { updateHighlight, highlightsExist } from "../util/highlights";
import ImagePreview from "../components/ImagePreview.vue";
import {
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  REAL_VECTOR_TYPE,
  GEOCOORDINATE_TYPE,
  MULTIBAND_IMAGE_TYPE,
  GEOBOUNDS_TYPE,
} from "../util/types";
import { scaleThreshold } from "d3";
import Color from "color";
import "leaflet/dist/leaflet.css";
import "leaflet/dist/images/marker-icon.png";
import "leaflet/dist/images/marker-icon-2x.png";
import "leaflet/dist/images/marker-shadow.png";
import {
  BLUE_PALETTE,
  GRAY,
  BLUE,
  SELECTION_RED,
  BLACK,
  RESULT_RED,
  RESULT_GREEN,
  colorByFacet,
  COLOR_SCALES,
} from "../util/color";
import DrillDown from "./DrillDown.vue";
import {
  CoordinateInfo,
  TileInfo,
  PointInfo,
  VertexPrimitive,
  Coordinate,
  updateVertexPrimitiveColor,
} from "../util/rendering/coordinates";
import { isGeoLocatedType } from "../util/types";
import { overlayRouteEntry } from "../util/routes";
import { EventList } from "../util/events";

const SINGLE_FIELD = 1;
const SPLIT_FIELD = 2;

interface GeoField {
  type: number;
  latField?: string;
  lngField?: string;
  field?: string;
}

type LatLngBoundsLiteral = import("leaflet").LatLngBoundsLiteral;

interface Area {
  info: CoordinateInfo;
  imageUrl: string;
  item: TableRow;
}
// Bucket contains the clustered tile data
interface Bucket {
  coordinates: number[][]; // should be two points each with x,y expect -> number[2][2]
  meta: { selected: boolean; count: number }; // count num of tiles in bucket
}

// contains the state of the map for things such as event callbacks and the quads to render
// currently there is two states tiled and clustered
interface MapState {
  onHover(id: number); // onhover callback
  onClick(id: number); // onclick callback
  vertices(): VertexPrimitive[]; // get quads for rendering
  init(): void; // called when state becomes current state -- essentially put any inits stuff here
  drawMode(): any; // returns DRAW_MODES
  layerId(): string;
}
interface LumoPoint {
  x: number;
  y: number;
}

enum CoordinateType {
  TileBased,
  PointBased,
}
export default Vue.extend({
  name: "GeoPlot",

  components: {
    ImageLabel,
    ImagePreview,
    DrillDown,
    GeoplotToggleButtons,
  },

  props: {
    instanceName: String as () => string,
    dataItems: { type: Array as () => any[], default: [] },
    baselineItems: { type: Array as () => TableRow[], default: [] },
    baselineMap: { type: Object as () => Dictionary<number>, default: null },
    dataFields: Object as () => Dictionary<TableColumn>,
    summaries: {
      type: Array as () => VariableSummary[],
      default: Array as () => VariableSummary[],
    },
    quadOpacity: { type: Number, default: 0.8 },
    pointOpacity: { type: Number, default: 0.8 },
    zoomThreshold: { type: Number, default: 8 },
    areaOfInterestItems: {
      type: Object as () => { inner: TableRow[]; outer: TableRow[] },
      default: null,
    },
    maxZoom: { type: Number, default: 17 }, // defaults to max zoom
    enableSelectionToolEvent: {
      type: Boolean as () => boolean,
      default: false,
    },
    isExclude: { type: Boolean as () => boolean, default: false },
    maxDataSize: {
      type: Number as () => number,
      default: Number.MAX_SAFE_INTEGER,
    },
    variables: {
      type: Array as () => Variable[],
      default: (): Variable[] => {
        return [];
      },
    },
  },

  data() {
    return {
      map: null,
      tileRenderer: null,
      overlay: null,
      renderer: null,
      isSelectionMode: false,
      isImageDrilldown: false,
      imageUrl: null,
      item: null,
      quadLayerId: "quad-layer",
      pointLayerId: "point-layer",
      clusterLayerId: "cluster-layer",
      toastTitle: "",
      hoverItem: null,
      hoverUrl: "",
      imageWidth: 128,
      imageHeight: 128,
      previousStateTiled: false,
      currentState: null,
      drillDownState: {
        tiles: [],
        bounds: null,
        centerTile: null,
        numCols: 7, // should be odd
        numRows: 5, // should be odd
      },
      selectionToolData: {
        startPoint: null,
        currentPoint: null,
        startPointClient: null,
        exit: { top: 0, right: 0 },
      },
      selectionToolId: "selection-tool-layer",
      showExit: false,
      pointSize: 0.025,
      isClustering: false,
      isSatelliteView: false,
      tileAreaThreshold: 170, // area in pixels
      boundsInitialized: false,
      isHidingBaseline: false,
      areas: [],
      debounceKey: null,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    imageType(): string {
      return MULTIBAND_IMAGE_TYPE;
    },
    colorScaleByVar(): string {
      return routeGetters.getColorScaleVariable(this.$store);
    },
    coloringByVariable(): boolean {
      return this.colorScaleByVar.length > 0;
    },
    colorByVariable(): (item: TableRow, idx: number) => number {
      const findKey = (v) => {
        return v.key === this.colorScaleByVar;
      };
      if (!this.summaries.some(findKey)) {
        return (item: TableRow, idx: number) => {
          return 0;
        };
      }
      return colorByFacet(this.summaries.find(findKey));
    },
    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    mapID(): string {
      return `map-${this.instanceName}`;
    },
    toastId(): string {
      return `notifications-${this.instanceName}`;
    },
    showDrillDown(): boolean {
      return this.isImageDrilldown;
    },
    colorScale(): (t: number) => string {
      const colorScale = routeGetters.getColorScale(this.$store);
      return COLOR_SCALES.get(colorScale);
    },
    getCoordinateType(): CoordinateType {
      if (this.coordinateColumn) {
        return CoordinateType.TileBased;
      }
      if (this.fieldSpecs.length > 0) {
        return CoordinateType.PointBased;
      }
      return -1;
    },
    fieldSpecs(): GeoField[] {
      const variables = datasetGetters.getVariables(this.$store);

      const matches = variables.filter((v) => {
        return (
          (v.grouping && v.grouping.type === GEOCOORDINATE_TYPE) ||
          v.colType === LONGITUDE_TYPE ||
          v.colType === LATITUDE_TYPE ||
          v.colType === REAL_VECTOR_TYPE
        );
      });

      let lng = null;
      let lat = null;
      const fields = [];

      matches.forEach((match) => {
        if (match.grouping && match.grouping.type === GEOCOORDINATE_TYPE) {
          const grouping = match.grouping as GeoCoordinateGrouping;
          lng = grouping.xCol;
          lat = grouping.yCol;
        } else if (match.colType === REAL_VECTOR_TYPE) {
          fields.push({
            type: SINGLE_FIELD,
            field: match.key,
          });
        } else {
          if (match.colType === LONGITUDE_TYPE) {
            lng = match.key;
          }
          if (match.colType === LATITUDE_TYPE) {
            lat = match.key;
          }
        }

        // TODO: currently we pair any two random lat / lngs, we should
        // eventually use the groupings functionality to let the user
        // group the two vars into a single point field.
        if (lng && lat) {
          fields.push({
            type: SPLIT_FIELD,
            lngField: lng,
            latField: lat,
          });
          lng = null;
          lat = null;
        }
      });

      return fields;
    },
    bucketFeatures(): Bucket[] {
      if (!this.geoSummary.length) {
        return [];
      }
      const features = [];
      this.geoSummary.forEach((summary) => {
        // compute the bucket size in degrees
        const buckets =
          summary.filtered && highlightsExist()
            ? summary.filtered.buckets
            : summary.baseline.buckets;
        const xSize = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
        const ySize =
          _.toNumber(buckets[0].buckets[1].key) -
          _.toNumber(buckets[0].buckets[0].key);

        // create a feature collection from the server-supplied bucket data
        buckets.forEach((lonBucket) => {
          lonBucket.buckets.forEach((latBucket) => {
            // Don't include features with a count of 0.
            if (latBucket.count > 0) {
              const xCoord = _.toNumber(lonBucket.key);
              const yCoord = _.toNumber(latBucket.key);
              const feature = {
                coordinates: [
                  [xCoord, yCoord],
                  [xCoord + xSize, yCoord + ySize],
                ],
                meta: { selected: false, count: latBucket.count },
              };
              features.push(feature);
            }
          });
        });
      });
      return features;
    },
    minBucketCount(): number {
      return Math.min(
        ...this.bucketFeatures.map((bf) => {
          return bf.meta.count;
        })
      );
    },
    maxBucketCount(): number {
      return Math.max(
        ...this.bucketFeatures.map((bf) => {
          return bf.meta.count;
        })
      );
    },

    predictedField(): string {
      const predictions = requestGetters.getActivePredictions(this.$store);
      if (predictions) {
        return predictions.predictedKey;
      }

      const solution = requestGetters.getActiveSolution(this.$store);
      return solution ? solution.predictedKey : "";
    },

    targetField(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },
    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },
    rowSelectionMap(): Map<number, number> {
      return new Map(
        this.rowSelection?.d3mIndices.map((di) => {
          return [di, di];
        })
      );
    },
    geoSummary(): VariableSummary[] {
      return this.summaries.filter((cs) => {
        return isGeoLocatedType(cs.varType);
      });
    },
    isMultiBandImage(): boolean {
      const variables = datasetGetters.getVariables(this.$store);
      return variables.some((v) => {
        return v.colType === MULTIBAND_IMAGE_TYPE;
      });
    },

    // Return name of column containing geobounds associated with a multiband image
    coordinateColumn(): string {
      const coordinateColumns = datasetGetters
        .getVariables(this.$store)
        .filter((v) => v.colType === GEOBOUNDS_TYPE)
        .map((v) => (v.grouping as GeoBoundsGrouping).coordinatesCol);
      if (coordinateColumns.length > 1) {
        console.error("only 1 coordinate column is supported");
      }
      return coordinateColumns[0];
    },

    // Return name of column used as grouping column for the table data
    multibandImageGroupColumn(): string {
      const groupColumns = datasetGetters
        .getVariables(this.$store)
        .find((v) => v.colType === MULTIBAND_IMAGE_TYPE);
      return groupColumns.key;
    },

    band(): string {
      return routeGetters.getBandCombinationId(this.$store);
    },

    tileRequest(): (x: number, y: number, z: number) => string {
      return this.isSatelliteView
        ? (x: number, y: number, z: number) => {
            return `https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/${z}/${y}/${x}.png`;
          }
        : (x: number, y: number, z: number) => {
            const SUBDOMAINS = ["a", "b", "c", "d"];
            const s = SUBDOMAINS[(x + y + z) % SUBDOMAINS.length];
            return `https:/${s}.basemaps.cartocdn.com/light_all/${z}/${x}/${y}.png`;
          };
    },

    tileState(): MapState {
      return {
        onHover: (id: number) => {
          if (id > this.areas.length) {
            console.error(`id: ${id} is outside of this.areas bounds`);
            return; // id outside of bounds
          }
          if (this.areas[id].imageUrl === null) {
            return;
          }
          this.toastTitle = this.areas[id].imageUrl;
          this.hoverItem = this.areas[id].item;
          this.hoverUrl = this.areas[id].imageUrl;
          this.$bvToast.show(this.toastId);
          window.addEventListener("mousemove", this.fadeToast);
        },
        onClick: (id: number) => {
          this.onTileClick(id);
        },
        vertices: () => {
          return this.areaToQuads();
        },
        init: () => {
          this.renderer.setPointSize(1); // default
        },
        drawMode: () => {
          return DRAW_MODES.TRIANGLES;
        },
        layerId: () => {
          return this.quadLayerId;
        },
      };
    },

    clusterState(): MapState {
      return {
        onHover: () => {
          return;
        }, // onHover empty for cluster state
        onClick: (id: number) => {
          if (id > this.bucketFeatures.length || id < 0) {
            console.error(
              `id retrieved from buffer picker ${id} not within index bounds of areas.`
            );
            return;
          }
          const bucket = this.bucketFeatures[id];
          const point1 = this.renderer.latlngToNormalized([
            bucket.coordinates[0][1],
            bucket.coordinates[0][0],
          ]);
          const point2 = this.renderer.latlngToNormalized([
            bucket.coordinates[1][1],
            bucket.coordinates[1][0],
          ]);
          const center = {
            x: (point1.x + point2.x) / 2,
            y: (point1.y + point2.y) / 2,
          };
          this.map.zoomToPosition(this.zoomThreshold, center); // zoom to the center of the cluster clicked. Zoom to the point where the state switches
        },
        vertices: () => {
          return this.bucketsToQuads();
        },
        init: () => {
          this.renderer.setPointSize(1); // default
        },
        drawMode: () => {
          return DRAW_MODES.TRIANGLES;
        },
        layerId: () => {
          return this.clusterLayerId;
        },
      };
    },

    pointState(): MapState {
      return {
        onHover: (id: number) => {
          if (id > this.areas.length) {
            console.error(`id: ${id} is outside of this.areas bounds`);
            return; // id outside of bounds
          }
          if (this.areas[id].imageUrl === null) {
            return;
          }
          this.toastTitle = this.areas[id].imageUrl;
          this.hoverItem = this.areas[id].item;
          this.hoverUrl = this.areas[id].imageUrl;
          this.$bvToast.show(this.toastId);
          window.addEventListener("mousemove", this.fadeToast);
        },
        onClick: (id: number) => {
          this.onTileClick(id);
        },
        vertices: () => {
          return this.areaToPoints();
        },
        init: () => {
          this.renderer.setPointSize(this.pointSize);
        },
        drawMode: () => {
          return DRAW_MODES.POINTS;
        },
        layerId: () => {
          return this.pointLayerId;
        },
      };
    },
  },

  watch: {
    dataItems() {
      this.onNewData();
    },
    rowSelection() {
      clearTimeout(this.debounceKey);
      this.debounceKey = setTimeout(() => {
        this.onNewData();
      }, 2000);
    },
    baselineItems() {
      if (this.baselineItems && !this.areas.length) {
        this.areas = this.tableDataToAreas(this.baselineItems);
      }
    },
    colorScale() {
      this.onNewData();
    },
    colorScaleByVar() {
      this.onNewData();
    },
    areaOfInterestItems() {
      // if null return
      if (
        !this.areaOfInterestItems.inner?.length &&
        !this.areaOfInterestItems.outer?.length
      ) {
        return;
      }
      const tileMap = new Map<string, Area>();
      const innerArea = this.tableDataToAreas(
        this.areaOfInterestItems.inner
      ) as any[];
      innerArea.forEach((i) => {
        i.gray = 0;
        tileMap.set(i.imageUrl, i);
      });
      const outerArea = this.tableDataToAreas(
        this.areaOfInterestItems.outer
      ) as any[];
      outerArea.forEach((i) => {
        i.gray = 100;
        if (!tileMap.has(i.imageUrl)) {
          tileMap.set(i.imageUrl, i);
        }
      });
      if (!tileMap.size) {
        return;
      }
      this.drillDownState.tiles = [...tileMap.values()];
      this.isImageDrilldown = true;
    },
  },

  mounted() {
    this.createLumoMap();
    const maxVal = 255;
    const color = Color(GRAY).rgb().object();
    this.renderer.setFragmentToDiscard([
      color.r / maxVal,
      color.g / maxVal,
      color.b / maxVal,
    ]);
    this.fitBounds();
    const entry = overlayRouteEntry(this.$route, {
      dataSize: this.maxDataSize,
    });
    this.$router.push(entry).catch((err) => console.warn(err));
  },
  beforeDestroy() {
    this.map.remove(this.overlay);
    this.map.remove(this.tileRenderer);
  },
  methods: {
    fitBounds() {
      const mapBounds = this.getBounds(
        this.overlay.getQuad(this.clusterState.layerId())
      );
      this.map.fitToBounds(mapBounds);
      this.boundsInitialized = true;
    },
    baselineToggle() {
      this.isHidingBaseline = !this.isHidingBaseline;
      this.renderer.shouldDiscardFragment(this.isHidingBaseline);
      this.tileRenderer.draw();
      this.renderer.draw();
    },
    addPrimitives() {
      let vertices = this.tileState.vertices();
      if (vertices.length) {
        this.overlay.addQuad(
          this.tileState.layerId(),
          vertices,
          this.tileState.drawMode()
        );
      }
      vertices = this.pointState.vertices();
      this.overlay.addQuad(
        this.pointState.layerId(),
        vertices,
        this.pointState.drawMode()
      );
      this.renderer.setDrawList([
        this.currentState.layerId(),
        this.selectionToolId,
      ]);
      vertices = this.clusterState.vertices();
      this.overlay.addQuad(
        this.clusterState.layerId(),
        vertices,
        this.clusterState.drawMode()
      );
    },
    createLumoMap() {
      // create map
      this.map = new lumo.Plot("#" + this.mapID, {
        continuousZoom: true,
        inertia: true,
        wraparound: true,
        zoom: 2,
        maxZoom: this.maxZoom,
      });
      this.createMapLayers();
      // convert this.areas to quads in normalized space and add to overlay layer
      this.currentState = this.pointState;
      this.map.on(lumo.ZOOM_END, this.onZoom);
      this.currentState.init();
      this.areas = this.tableDataToAreas(this.baselineItems);
      if (!this.areas.length) {
        return; // no data
      }

      this.addPrimitives();

      // add listener for clicks on quads
      this.renderer.addListener(
        EVENT_TYPES.MOUSE_CLICK,
        this.currentState.onClick
      );
      this.renderer.addListener(
        EVENT_TYPES.MOUSE_HOVER,
        this.currentState.onHover
      );
      if (this.dataItems.length) {
        this.onNewData();
      }
    },
    createMapLayers(createOverlay = true) {
      // WebGL CARTO Image Layer
      this.tileRenderer = new lumo.TileLayer({
        renderer: new lumo.ImageTileRenderer(),
      });
      // tile request function
      this.tileRenderer.requestTile = (coord, done) => {
        const dim = Math.pow(2, coord.z); // this is done in lumo however there is no get function to get the correct y coordinate for requesting tiles
        const url = this.tileRequest(coord.x, dim - 1 - coord.y, coord.z);
        lumo.loadImage(url, done); // load the image to the map
      };
      this.map.add(this.tileRenderer);
      if (createOverlay) {
        // Quad layer
        this.overlay = new BatchQuadOverlay();
        this.renderer = new BatchQuadOverlayRenderer();
        this.overlay.setRenderer(this.renderer);
        this.map.add(this.overlay);
      }
    },
    getInterestBounds(area: Area): LatLngBoundsLiteral {
      const xDistance = (this.drillDownState.numCols - 1) / 2;
      const yDistance = (this.drillDownState.numRows - 1) / 2;
      const tileWidth =
        area.info.coordinates[1][Coordinate.lng] -
        area.info.coordinates[0][Coordinate.lng];
      const tileHeight =
        area.info.coordinates[1][Coordinate.lat] -
        area.info.coordinates[0][Coordinate.lat];
      const result = [
        [0, 0],
        [0, 0],
      ];
      result[0][0] =
        area.info.coordinates[1][Coordinate.lat] + yDistance * tileHeight; // top
      result[0][1] =
        area.info.coordinates[0][Coordinate.lng] - xDistance * tileWidth; // left
      result[1][0] =
        area.info.coordinates[0][Coordinate.lat] - yDistance * tileHeight; // bottom
      result[1][1] =
        area.info.coordinates[1][Coordinate.lng] + xDistance * tileWidth; // right
      return result as LatLngBoundsLiteral;
    },
    onFocusOut() {
      this.isImageDrilldown = false;
    },
    mapToggle() {
      this.isSatelliteView = !this.isSatelliteView;
      this.map.remove(this.tileRenderer); // remove old tile renderer to destroy the buffers hold the previous tile set
      this.map.remove(this.overlay);
      this.createMapLayers(false);
      this.map.add(this.overlay);
      this.tileRenderer.draw();
      this.renderer.draw();
    },
    /**
     * toggle clustering
     */
    toggleClustering() {
      this.isClustering = !this.isClustering;
      if (this.isClustering && this.map.getZoom() < this.zoomThreshold) {
        this.currentState = this.clusterState;
        this.updateMapState();
        this.tileRenderer.draw();
        this.renderer.draw();
        return;
      }
      if (!this.isClustering && this.map.getZoom() < this.zoomThreshold) {
        this.currentState = this.pointState;
        this.updateMapState();
        this.tileRenderer.draw();
        this.renderer.draw();
      }
    },
    /**
     * on selection tool toggle disable or enable the quad interactions such as click or hover
     */
    toggleSelectionTool() {
      this.isSelectionMode = !this.isSelectionMode;
      if (this.isSelectionMode) {
        // disable interactions so the selection tool can interact without triggering the other interactions
        this.renderer.disableInteractions();
        this.map.on("mousedown", this.selectionToolDown);
        this.map.disablePanning();
        return;
      }
      this.overlay.removeQuad(this.selectionToolId);
      this.map.removeListener("mousedown", this.selectionToolDown);
      this.showExit = false;
      // enable interactions
      this.renderer.enableInteractions();
      this.map.enablePanning();
    },
    // mouse move clear and redraw quad with new point
    selectionToolDraw(e) {
      this.selectionToolData.currentPoint = e.pos;
      // draw current selection
      this.overlay.removeQuad(this.selectionToolId);
      this.overlay.addQuad(
        this.selectionToolId,
        this.pointsToQuad(
          this.selectionToolData.startPoint,
          this.selectionToolData.currentPoint
        ),
        DRAW_MODES.TRIANGLES
      );
      this.renderer.setDrawList([
        this.currentState.layerId(),
        this.selectionToolId,
      ]);
    },
    // register mousemouve and up callbacks to draw the selection quad
    selectionToolDown(e) {
      this.selectionToolData.startPoint = e.pos;
      this.selectionToolData.startPointClient = e.originalEvent;
      this.showExit = false;
      this.overlay.removeQuad(this.selectionToolId);
      this.map.on("mousemove", this.selectionToolDraw);
      this.map.on("mouseup", this.selectionToolUp);
    },
    // add exit button and send selection to postgis to update data
    selectionToolUp(e) {
      this.selectionToolData.currentPoint = e.pos;
      this.map.removeListener("mousemove", this.selectionToolDraw);
      this.map.removeListener("mouseup", this.selectionToolUp);
      this.selectionToolData.exit.top = Math.min(
        e.originalEvent.layerY,
        this.selectionToolData.startPointClient.layerY
      ); // get top most y value
      const right = Math.max(
        e.originalEvent.layerX,
        this.selectionToolData.startPointClient.layerX
      ); // get right most x value
      this.selectionToolData.exit.right = e.target.canvas.clientWidth - right; // had to subtract width for some reason x is reversed in lumo
      this.showExit = true;
      // convert from normalized coordinate system to lat lng
      const p1 = this.renderer.normalizedPointToLatLng(
        this.selectionToolData.startPoint
      );
      const p2 = this.renderer.normalizedPointToLatLng(
        this.selectionToolData.currentPoint
      );

      const minX = Math.min(p1.lng, p2.lng);
      const maxX = Math.max(p1.lng, p2.lng);
      const minY = Math.min(p1.lat, p2.lat);
      const maxY = Math.max(p1.lat, p2.lat);
      // send selection to PostGis
      this.createHighlight({ minX, minY, maxX, maxY });
      this.overlay.removeQuad(this.selectionToolId);
    },
    getBounds(quads: VertexPrimitive[]) {
      // set mapBounds to a single tile to start
      const mapBounds = new lumo.Bounds(
        quads[0].x,
        quads[0].x,
        quads[0].y,
        quads[0].y
      );
      // extend bounds to fit the entire quad set
      quads.forEach((q) => {
        mapBounds.extend(q);
      });
      return mapBounds;
    },
    // fades toast after mouse is moved
    fadeToast() {
      this.$bvToast.hide(this.toastId);
      window.removeEventListener("mousemove", this.fadeToast); // remove event listener because toast is now faded
    },
    onTileClick(id: number) {
      if (id > this.areas.length || id < 0) {
        console.error(
          `id retrieved from buffer picker ${id} not within index bounds of areas.`
        );
        return;
      }
      if (this.areas[id].imageUrl === null) {
        return;
      }
      this.drillDownState.centerTile = this.areas[id];
      this.drillDownState.bounds = this.getInterestBounds(this.areas[id]);
      this.$emit(EventList.MAP.TILE_CLICKED_EVENT, {
        bounds: this.drillDownState.bounds,
        key: this.geoSummary[0].key,
        displayName: this.geoSummary[0].label,
        type: this.geoSummary[0].type,
      });
    },
    // assumes x and y are normalized points this function is for the selection tool
    pointsToQuad(p1: LumoPoint, p2: LumoPoint): VertexPrimitive[] {
      const result = [];
      const id = this.renderer.idToRGBA(0); // pass in 0 as the id, currently there is only ever one selection at a time.
      const color = Color(BLUE_PALETTE[0]).rgb().object();
      const maxColorVal = 256;
      // normalize color values
      color.a = this.pointOpacity;
      color.r /= maxColorVal;
      color.g /= maxColorVal;
      color.b /= maxColorVal;
      result.push({ ...p1, ...color, ...id });
      result.push({ x: p2.x, y: p1.y, ...color, ...id });
      result.push({ ...p2, ...color, ...id });
      result.push({ ...p1, ...color, ...id });
      result.push({ x: p1.x, y: p2.y, ...color, ...id });
      result.push({ ...p2, ...color, ...id });
      return result;
    },

    // packs all data into single aligned memory array
    bucketsToQuads(): VertexPrimitive[] {
      const maxVal = this.maxBucketCount;
      const minVal = this.minBucketCount;
      const d = (maxVal - minVal) / BLUE_PALETTE.length;
      const domain = BLUE_PALETTE.map((val, index) => minVal + d * (index + 1));
      const scaleColors = scaleThreshold()
        .range(BLUE_PALETTE as any)
        .domain(domain);
      const result = []; // packing array with
      this.bucketFeatures.forEach((bucket, idx) => {
        const p1 = this.renderer.latlngToNormalized([
          bucket.coordinates[0][1],
          bucket.coordinates[0][0],
        ]);
        const p2 = this.renderer.latlngToNormalized([
          bucket.coordinates[1][1],
          bucket.coordinates[1][0],
        ]);
        const color = Color(scaleColors(bucket.meta.count).toString(16))
          .rgb()
          .object(); // convert hex color to rgb
        const maxColorVal = 256;
        // normalize color values
        color.a = this.quadOpacity;
        color.r /= maxColorVal;
        color.g /= maxColorVal;
        color.b /= maxColorVal;
        const id = this.renderer.idToRGBA(idx); // separate index bytes into 4 channels iR,iG,iB,iA. Used to render the index of the object into webgl FBO
        // need to get rid of spread operators super slow
        result.push({ ...p1, ...color, ...id });
        result.push({ x: p2.x, y: p1.y, ...color, ...id });
        result.push({ ...p2, ...color, ...id });
        result.push({ ...p1, ...color, ...id });
        result.push({ x: p1.x, y: p2.y, ...color, ...id });
        result.push({ ...p2, ...color, ...id });
      });
      return result;
    },
    areaToPoints(): VertexPrimitive[] {
      let result = [];

      this.areas.forEach((area, idx) => {
        result = result.concat(
          area.info.toPoint(this.renderer, this.pointOpacity, idx)
            .vertexPrimitives
        );
      });
      return result;
    },
    // packs all data into single aligned memory array
    areaToQuads(): VertexPrimitive[] {
      let result = [];
      this.areas.forEach((area, idx) => {
        result = result.concat(
          area.info.toQuad(this.renderer, this.quadOpacity, idx)
            .vertexPrimitives
        );
      });
      return result;
    },
    pointGroups(tableData: any[]): Area[] {
      let areas = [];
      this.fieldSpecs.forEach((fieldSpec) => {
        const temp = tableData.map((item, i) => {
          const imageUrl = this.isMultiBandImage
            ? item[this.multibandImageGroupColumn].value
            : null;
          const color = this.tileColor(item, i);
          const lat = this.latValue(fieldSpec, item);
          const lng = this.lngValue(fieldSpec, item);

          if (lat !== undefined && lng !== undefined) {
            const coordinates = [[lat, lng]] as LatLngBoundsLiteral; // Corner A as [Lat, Lng]
            const info = new PointInfo(coordinates, color);
            return {
              imageUrl,
              item,
              info,
            } as Area;
          }

          return null;
        });

        areas = areas.concat(temp);
      });

      return areas;
    },
    tableDataToAreas(tableData: any[]): Area[] {
      if (this.getCoordinateType === CoordinateType.PointBased) {
        return this.pointGroups(tableData);
      }
      if (!tableData) {
        return [];
      }
      const areas = tableData.map((item, i) => {
        const imageUrl = this.isMultiBandImage
          ? item[this.multibandImageGroupColumn]?.value
          : null;
        const fullCoordinates = item[this.coordinateColumn].value;
        if (!fullCoordinates) {
        }
        if (fullCoordinates.some((x) => x === undefined)) return;

        /*
          Item store the coordinates as a list of 8 values being four pairs of [Lng, Lat],
          one for each corner of the isMultiBandImage-sensing image.

          [0,1]     [2,3]
            A-------B
            |       |
            |       |
            D-------C
          [6,7]     [4,5]
        */
        const coordinates = [
          [fullCoordinates[1], fullCoordinates[0]], // Corner A as [Lat, Lng]
          [fullCoordinates[5], fullCoordinates[4]], // Corner C as [Lat, Lng]
        ] as LatLngBoundsLiteral;

        const color = this.tileColor(item, i);
        const info = new TileInfo(coordinates, color);
        return { item, imageUrl, info } as Area;
      });

      return areas;
    },
    shouldTilesRender(): boolean {
      if (!this.areas.length) {
        return false;
      }
      return this.areas[0].info.shouldTile(
        this.renderer,
        this.map.getPixelExtent(),
        this.tileAreaThreshold
      );
    },
    // callback when zooming on map
    onZoom() {
      const shouldBeTiles = this.shouldTilesRender();
      const wasPoints = shouldBeTiles && !this.previousStateTiled;
      const wasTiled = !shouldBeTiles && this.previousStateTiled;
      this.previousStateTiled = shouldBeTiles;
      // check if map should be rendering clustered tiles
      if (wasPoints) {
        this.currentState = this.tileState;
        this.updateMapState();
        return;
      }
      if (wasTiled) {
        this.currentState = this.isClustering
          ? this.clusterState
          : this.pointState;
        this.updateMapState();
        return;
      }
    },
    // called after state changes and map needs to update
    updateMapState() {
      //this.overlay.removeQuad(this.quadLayerId);
      this.renderer.clearListeners();
      this.currentState.init();
      this.renderer.setDrawList([
        this.currentState.layerId(),
        this.selectionToolId,
      ]);
      this.renderer.addListener(
        EVENT_TYPES.MOUSE_CLICK,
        this.currentState.onClick
      );
      this.renderer.addListener(
        EVENT_TYPES.MOUSE_HOVER,
        this.currentState.onHover
      );
    },

    onEsc() {
      if (this.isSelectionMode) {
        this.toggleSelectionTool();
      }
    },

    createHighlight(value: {
      minX: number;
      maxX: number;
      minY: number;
      maxY: number;
    }) {
      if (this.highlights && this.highlights.length > 0) {
        const isExistingHighlight = this.highlights.reduce(
          (hasHighlight, highlight) => {
            return (
              hasHighlight ||
              (highlight.value &&
                highlight.value.minX === value.minX &&
                highlight.value.maxX === value.maxX &&
                highlight.value.minY === value.minY &&
                highlight.value.maxY === value.maxY)
            );
          },
          false
        );
        // dont push existing highlight
        if (isExistingHighlight) {
          return;
        }
      }

      // TODO: support filtering multiple vars?
      const fieldSpec = this.fieldSpecs[0];
      let key = "";
      if (!!fieldSpec) {
        key =
          fieldSpec.type === SINGLE_FIELD
            ? fieldSpec.field
            : this.fieldHash(fieldSpec);
      } else if (!!this.geoSummary[0].key) {
        const variable = this.variables.find((v) => {
          return v.key === this.geoSummary[0].key;
        });
        key = (variable.grouping as GeoBoundsGrouping).polygonCol;
      } else {
        console.error("Error createHighlight no available key");
        return;
      }
      if (this.enableSelectionToolEvent) {
        this.$emit(EventList.MAP.SELECTION_TOOL_EVENT, {
          context: this.instanceName,
          dataset: this.dataset,
          key: key,
          value: value,
        });
        return;
      }
      updateHighlight(this.$router, {
        context: this.instanceName,
        dataset: this.dataset,
        key: key,
        value: value,
      });
    },

    lngValue(fieldSpec: GeoField, row: TableRow): number {
      if (fieldSpec.type === SINGLE_FIELD) {
        return row[fieldSpec.field].Elements[0].Float;
      }
      return row[fieldSpec.lngField].value;
    },

    latValue(fieldSpec: GeoField, row: TableRow): number {
      if (fieldSpec.type === SINGLE_FIELD) {
        return row[fieldSpec.field].Elements[1].Float;
      }
      return row[fieldSpec.latField].value;
    },

    fieldHash(fieldSpec: GeoField): string {
      if (fieldSpec.type === SINGLE_FIELD) {
        return fieldSpec.field;
      }
      return fieldSpec.lngField + "_" + fieldSpec.latField;
    },

    showImageDrilldown(imageUrl: string, item: TableRow) {
      this.imageUrl = imageUrl ?? null;
      this.item = item ?? null;
      this.isImageDrilldown = true;
    },

    hideImageDrilldown() {
      this.isImageDrilldown = false;
    },

    tileColor(item: any, idx: number) {
      let color = this.isExclude ? BLACK : BLUE; // Default
      if (this.rowSelectionMap.has(item.d3mIndex)) {
        return SELECTION_RED;
      }
      if (this.coloringByVariable) {
        return this.colorScale(this.colorByVariable(item, idx));
      }
      if (item.isExcluded) {
        return GRAY;
      }
      if (item[this.targetField] && item[this.predictedField]) {
        color =
          item[this.targetField].value === item[this.predictedField].value
            ? RESULT_GREEN // Correct: green.
            : RESULT_RED; // Incorrect: red.
      }

      return color;
    },
    onNewData() {
      if (!this.overlay.getQuad(this.currentState.layerId())) {
        return;
      }
      if (this.overlay.getQuad(this.tileState.layerId())) {
        updateVertexPrimitiveColor(
          this.overlay.getQuad(this.tileState.layerId()),
          this.dataItems,
          this.tileColor.bind(this),
          this.areas.length,
          this.baselineMap
        );
      }
      updateVertexPrimitiveColor(
        this.overlay.getQuad(this.pointState.layerId()),
        this.dataItems,
        this.tileColor.bind(this),
        this.areas.length,
        this.baselineMap
      );
      // must happen to refresh webgl
      this.overlay.refresh(); // clips the geometry
      this.renderer.refreshBuffers(); // rebuilds webgl buffers
      this.tileRenderer.draw();
      this.renderer.draw(); // draw the newly rebuilt buffers
      if (!this.boundsInitialized) {
        this.fitBounds();
      }
      // don't show exit button
      this.showExit = false;
    },
  },
});
</script>

<style>
.geo-plot-container,
.geo-plot {
  position: relative;
  z-index: 0;
  width: 100%;
  height: 98%;
  bottom: 0;
  max-height: 98%;
}

.geo-close-button {
  position: absolute;
  width: 24px;
  height: 24px;
  text-align: center;
  line-height: 24px;

  left: 8px;
  top: -24px;
  border: 1px solid #ccc;
  border-radius: 4px;
  background-color: #fff;
  color: #000;
  cursor: pointer;
}
.selection-mode {
  cursor: cell;
}
.geo-close-button:hover {
  background-color: #f4f4f4;
}
.image-label {
  position: absolute;
  top: 0px;
  z-index: 1;
}
.selection-exit {
  position: absolute;
}
</style>
