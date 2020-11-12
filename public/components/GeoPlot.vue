<template>
  <div
    id="geo-test"
    class="geo-plot-container"
    :class="{ 'selection-mode': isSelectionMode }"
  >
    <div
      ref="geoPlot"
      class="geo-plot"
      :id="mapID"
      @keydown.esc="onEsc"
      tabindex="0"
    ></div>

    <drill-down
      v-if="showDrillDown"
      :dataFields="dataFields"
      :imageType="imageType"
      :tiles="drillDownState.tiles"
      :centerTile="drillDownState.centerTile"
      :bounds="drillDownState.bounds"
      @close="onFocusOut"
    />
    <div
      class="selection-toggle toggle"
      :class="{ active: isSelectionMode }"
      @click="toggleSelectionTool"
    >
      <a
        class="selection-toggle-control"
        title="Select area"
        aria-label="Select area"
      >
        <icon-base width="100%" height="100%"> <icon-crop-free /> </icon-base>
      </a>
    </div>
    <div
      class="cluster-toggle toggle"
      :class="{ active: isClustering }"
      @click="toggleClustering"
    >
      <a class="cluster-icon" title="Cluster" aria-label="Cluster Tiles">
        <i class="fa fa-object-group fa-lg" aria-hidden="true" />
      </a>
    </div>
    <div
      v-if="dataHasConfidence"
      class="confidence-toggle toggle"
      :class="{ active: isColoringByConfidence }"
      @click="toggleConfidenceColoring"
    >
      <a
        :class="confidenceClass"
        title="confidence"
        aria-label="Color by Confidence"
        :style="colorGradient"
      >
        C
      </a>
    </div>
    <div
      class="map-toggle toggle"
      :class="{ active: isSatelliteView }"
      @click="mapToggle"
    >
      <a class="cluster-icon" title="Change Map" aria-label="Change Map">
        <i class="fa fa-globe" aria-hidden="true" />
      </a>
    </div>
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
          :dataFields="dataFields"
          includedActive
          shortenLabels
          alignHorizontal
          :item="hoverItem"
        />
        <image-preview
          class="image-preview"
          :row="hoverItem"
          :image-url="hoverUrl"
          :width="imageWidth"
          :height="imageHeight"
          :type="imageType"
        ></image-preview>
      </div>
    </b-toast>
    <button
      type="button"
      class="close selection-exit"
      aria-label="Close"
      v-show="showExit"
      :style="exitStyle"
    >
      <span aria-hidden="true">&times;</span>
    </button>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";
import IconBase from "./icons/IconBase.vue";
import IconCropFree from "./icons/IconCropFree.vue";
import ImageLabel from "./ImageLabel.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import viridisScale from "scale-color-perceptual/viridis";
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
} from "../store/dataset/index";
import { updateHighlight, highlightsExist } from "../util/highlights";
import ImagePreview from "../components/ImagePreview.vue";
import {
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  REAL_VECTOR_TYPE,
  GEOCOORDINATE_TYPE,
  MULTIBAND_IMAGE_TYPE,
} from "../util/types";
import { scaleThreshold } from "d3";
import Color from "color";
import "leaflet/dist/leaflet.css";
import "leaflet/dist/images/marker-icon.png";
import "leaflet/dist/images/marker-icon-2x.png";
import "leaflet/dist/images/marker-shadow.png";
import { BLUE_PALETTE } from "../util/color";
import DrillDown from "./DrillDown.vue";

const SINGLE_FIELD = 1;
const SPLIT_FIELD = 2;

interface GeoField {
  type: number;
  latField?: string;
  lngField?: string;
  field?: string;
}

interface Point {
  lat: number;
  lng: number;
  row?: TableRow;
  color?: string;
}

interface PointGroup {
  field: GeoField;
  points: Point[];
}

type TileLayer = import("leaflet").TileLayer;
type LatLngBoundsLiteral = import("leaflet").LatLngBoundsLiteral;

interface Area {
  coordinates: LatLngBoundsLiteral;
  color: string;
  imageUrl: string;
  item: TableRow;
}
// Bucket contains the clustered tile data
interface Bucket {
  coordinates: number[][]; // should be two points each with x,y expect -> number[2][2]
  meta: { selected: boolean; count: number }; // count num of tiles in bucket
}
interface Quad {
  x: number; // vertex x
  y: number; // vertex y
  r: number; // color r channel
  g: number; // color g channel
  b: number; // color b channel
  a: number; // color alpha channel
  // id's bytes is broken down into 4 channels
  iR: number; // id smallest byte
  iG: number; // id second smallest byte
  iB: number; // id second largest byte
  iA: number; // id largest byte
}
// contains the state of the map for things such as event callbacks and the quads to render
// currently there is two states tiled and clustered
interface MapState {
  onHover(id: number); // onhover callback
  onClick(id: number); // onclick callback
  quads(): Quad[]; // get quads for rendering
  init(): void; // called when state becomes current state -- essentially put any inits stuff here
  drawMode(): any; // returns DRAW_MODES
}
interface LumoPoint {
  x: number;
  y: number;
}

export interface TileClickData {
  bounds: number[][];
  key: string;
  displayName: string;
  type: string;
  callback: (inner: TableRow[], outer: TableRow[]) => void;
}
// Minimum pixels size of clickable target displayed on the map.
const TARGETSIZE = 6;

export default Vue.extend({
  name: "geo-plot",

  components: {
    IconBase,
    IconCropFree,
    ImageLabel,
    ImagePreview,
    DrillDown,
  },

  props: {
    instanceName: String as () => string,
    dataItems: Array as () => any[],
    dataFields: Object as () => Dictionary<TableColumn>,
    summaries: {
      type: Array as () => VariableSummary[],
      default: Array as () => VariableSummary[],
    },
    quadOpacity: { type: Number, default: 0.8 },
    pointOpacity: { type: Number, default: 0.8 },
    zoomThreshold: { type: Number, default: 8 },
    maxZoom: { type: Number, default: 18 },
    colorScale: { type: Function, default: viridisScale },
  },

  data() {
    return {
      poiLayer: null,
      map: null,
      tileRenderer: null,
      overlay: null,
      renderer: null,
      markers: null,
      areasMeanLng: 0,
      closeButton: null,
      startingLatLng: null,
      currentRect: null,
      selectedRect: null,
      isSelectionMode: false,
      isImageDrilldown: false,
      imageUrl: null,
      item: null,
      quadLayerId: "quad-layer",
      toastTitle: "",
      hoverItem: null,
      toastImg: "",
      hoverUrl: "",
      imageWidth: 128,
      imageHeight: 128,
      previousZoom: 0,
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
      isColoringByConfidence: false,
      confidenceIconClass: "confidence-icon",
      isSatelliteView: false,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },
    imageType(): string {
      return MULTIBAND_IMAGE_TYPE;
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },
    dataHasConfidence(): boolean {
      return "confidence" in this.dataItems[0];
    },
    getTopVariables(): string[] {
      const variables = datasetGetters
        .getVariables(this.$store)
        .filter((v) => v.datasetName === this.dataset);
      return variables
        .map((variable) => ({
          variable: variable.colName,
          order: _.isNumber(variable.ranking)
            ? variable.ranking
            : variable.importance,
        }))
        .sort((a, b) => b.order - a.order)
        .map((r) => r.variable);
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
    colorGradient(): string {
      return this.isColoringByConfidence
        ? `background-image:linear-gradient(${[
            0.0, // padding
            0.0, // padding
            1.0,
            0.9,
            0.8,
            0.7,
            0.6,
            0.5,
            0.4,
            0.3,
            0.2,
            0.1,
            0.0,
            0.0, // padding
            0.0, // padding
          ]
            .map(this.colorScale)
            .join(",")})`
        : "";
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
            field: match.colName,
          });
        } else {
          if (match.colType === LONGITUDE_TYPE) {
            lng = match.colName;
          }
          if (match.colType === LATITUDE_TYPE) {
            lat = match.colName;
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
      if (!this.summaries.length) {
        return [];
      }
      const features = [];
      this.summaries.forEach((summary) => {
        // compute the bucket size in degrees
        const buckets =
          summary.filtered && highlightsExist(this.$router)
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
      // console.log(duplicateCheck);
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
    pointGroups(): PointGroup[] {
      const groups = [];

      if (!this.dataItems) {
        return groups;
      }

      this.fieldSpecs.forEach((fieldSpec) => {
        const group = {
          field: fieldSpec,
          points: [],
        };

        group.points = this.dataItems
          .map((item) => {
            const lat = this.latValue(fieldSpec, item);
            const lng = this.lngValue(fieldSpec, item);

            if (lat !== undefined && lng !== undefined) {
              return {
                lng: lng,
                lat: lat,
                row: item,
                color: this.tileColor(item),
              };
            }

            return null;
          })
          .filter((point) => !!point);

        groups.push(group);
      });

      return groups;
    },

    predictedField(): string {
      const predictions = requestGetters.getActivePredictions(this.$store);
      if (predictions) {
        return predictions.predictedKey;
      }

      const solution = requestGetters.getActiveSolution(this.$store);
      return solution ? `${solution.predictedKey}` : "";
    },

    targetField(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
    },

    /* Data with multiple geocordinates to be displayed as an area on the map. */
    areas(): Area[] {
      if (!this.dataItems) {
        return [];
      }
      return this.tableDataToAreas(this.dataItems);
    },

    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    mapCenter(): number[] {
      return routeGetters.getGeoCenter(this.$store);
    },

    mapZoom(): number {
      return routeGetters.getGeoZoom(this.$store);
    },

    rowSelection(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    isMultiBandImage(): boolean {
      return routeGetters.isMultiBandImage(this.$store);
    },
    isGeoSpatial(): boolean {
      return routeGetters.isGeoSpatial(this.$store);
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
          this.toastTitle = this.areas[id].imageUrl;
          this.hoverItem = this.areas[id].item;
          this.hoverUrl = this.areas[id].imageUrl;
          this.$bvToast.show(this.toastId);
          window.addEventListener("mousemove", this.fadeToast);
        },
        onClick: (id: number) => {
          this.onTileClick(id);
        },
        quads: () => {
          return this.areaToQuads();
        },
        init: () => {
          this.renderer.setPointSize(1); // default
        },
        drawMode: () => {
          return DRAW_MODES.TRIANGLES;
        },
      };
    },
    clusterState(): MapState {
      return {
        onHover: (id: number) => {
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
        quads: () => {
          return this.bucketsToQuads();
        },
        init: () => {
          this.renderer.setPointSize(1); // default
        },
        drawMode: () => {
          return DRAW_MODES.TRIANGLES;
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
          this.toastTitle = this.areas[id].imageUrl;
          this.hoverItem = this.areas[id].item;
          this.hoverUrl = this.areas[id].imageUrl;
          this.$bvToast.show(this.toastId);
          window.addEventListener("mousemove", this.fadeToast);
        },
        onClick: (id: number) => {
          this.onTileClick(id);
        },
        quads: () => {
          return this.areaToPoints();
        },
        init: () => {
          this.renderer.setPointSize(this.pointSize);
        },
        drawMode: () => {
          return DRAW_MODES.POINTS;
        },
      };
    },
    confidenceClass(): string {
      return this.confidenceIconClass;
    },
    exitStyle(): string {
      return `top:${this.selectionToolData.exit.top}px; right:${this.selectionToolData.exit.right}px;`;
    },
  },
  methods: {
    createLumoMap() {
      // create map
      this.map = new lumo.Plot("#" + this.mapID, {
        continuousZoom: true,
        inertia: true,
        wraparound: true,
        zoom: this.maxZoom,
        maxZoom: 11,
      });
      this.createMapLayers();
      // convert this.areas to quads in normalized space and add to overlay layer
      this.currentState = this.pointState;
      this.map.on(lumo.ZOOM_END, this.onZoom);
      this.currentState.init();
      if (!this.areas.length) {
        return; // no data
      }
      const quads = this.currentState.quads();
      // get quad set bounds
      const mapBounds = this.getBounds(quads);
      this.overlay.addQuad(
        this.quadLayerId,
        quads,
        this.currentState.drawMode()
      );

      // add listener for clicks on quads
      this.renderer.addListener(
        EVENT_TYPES.MOUSE_CLICK,
        this.currentState.onClick
      );
      this.renderer.addListener(
        EVENT_TYPES.MOUSE_HOVER,
        this.currentState.onHover
      );
      this.map.fitToBounds(mapBounds);
    },
    createMapLayers() {
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
      // Quad layer
      this.overlay = new BatchQuadOverlay();
      this.renderer = new BatchQuadOverlayRenderer();
      this.overlay.setRenderer(this.renderer);
      this.map.add(this.overlay);
    },
    getInterestBounds(area: Area): LatLngBoundsLiteral {
      const xDistance = (this.drillDownState.numCols - 1) / 2;
      const yDistance = (this.drillDownState.numRows - 1) / 2;
      const tileWidth = area.coordinates[1][1] - area.coordinates[0][1];
      const tileHeight = area.coordinates[1][0] - area.coordinates[0][0];
      const result = [
        [0, 0],
        [0, 0],
      ];
      result[0][0] = area.coordinates[1][0] + yDistance * tileHeight; // top
      result[0][1] = area.coordinates[0][1] - xDistance * tileWidth; // left
      result[1][0] = area.coordinates[0][0] - yDistance * tileHeight; // bottom
      result[1][1] = area.coordinates[1][1] + xDistance * tileWidth; // right
      return result as LatLngBoundsLiteral;
    },
    onFocusOut() {
      this.isImageDrilldown = false;
    },
    mapToggle() {
      this.isSatelliteView = !this.isSatelliteView;
      this.map.remove(this.tileRenderer); // remove old tile renderer to destroy the buffers hold the previous tile set
      this.map.remove(this.overlay);
      this.createMapLayers();
      this.updateMapState(); // trigger a tile render
    },
    /**
     * toggle clustering
     */
    toggleClustering() {
      this.isClustering = !this.isClustering;
      if (this.isClustering && this.map.getZoom() < this.zoomThreshold) {
        this.currentState = this.clusterState;
        this.updateMapState();
        return;
      }
      if (!this.isClustering && this.map.getZoom() < this.zoomThreshold) {
        this.currentState = this.pointState;
        this.updateMapState();
      }
    },
    /**
     * toggles coloring tiles by confidence (only available in result screen)
     */
    toggleConfidenceColoring() {
      this.isColoringByConfidence = !this.isColoringByConfidence;
      if (this.isColoringByConfidence) {
        this.confidenceIconClass = "toggled-confidence-icon";
        this.updateMapState();
        return;
      }
      this.confidenceIconClass = "confidence-icon";
      this.updateMapState();
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
      this.map.enableZooming();
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
    },
    getBounds(quads: Quad[]) {
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
      this.drillDownState.centerTile = this.areas[id];
      this.drillDownState.bounds = this.getInterestBounds(this.areas[id]);
      this.$emit("tileClicked", {
        bounds: this.drillDownState.bounds,
        key: this.summaries[0].key,
        displayName: this.summaries[0].label,
        type: this.summaries[0].type,
        callback: (inner: TableRow[], outer: TableRow[]) => {
          const innerArea = this.tableDataToAreas(inner) as any[];
          innerArea.forEach((i) => {
            i.gray = 0;
          });
          const outerArea = this.tableDataToAreas(outer) as any[];
          outerArea.forEach((i) => {
            i.gray = 100;
          });
          this.drillDownState.tiles = innerArea.concat(outerArea);
          this.isImageDrilldown = true;
        },
      });
    },
    // assumes x and y are normalized points this function is for the selection tool
    pointsToQuad(p1: LumoPoint, p2: LumoPoint): Quad[] {
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
    bucketsToQuads(): Quad[] {
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
    areaToPoints(): Quad[] {
      const result = [];
      this.areas.forEach((area, idx) => {
        const p1 = this.renderer.latlngToNormalized(area.coordinates[0]);
        const p2 = this.renderer.latlngToNormalized(area.coordinates[1]);
        const centerPoint = { x: (p1.x + p2.x) / 2, y: (p1.y + p2.y) / 2 };
        const color = Color(area.color).rgb().object(); // convert hex color to rgb
        const maxVal = 255;
        // normalize color values
        color.a = this.pointOpacity;
        color.r /= maxVal;
        color.g /= maxVal;
        color.b /= maxVal;
        const id = this.renderer.idToRGBA(idx); // separate index bytes into 4 channels iR,iG,iB,iA. Used to render the index of the object into webgl FBO
        // need to get rid of spread operators super slow
        result.push({ ...centerPoint, ...color, ...id });
      });
      return result;
    },
    // packs all data into single aligned memory array
    areaToQuads(): Quad[] {
      const result = [];
      this.areas.forEach((area, idx) => {
        const p1 = this.renderer.latlngToNormalized(area.coordinates[0]);
        const p2 = this.renderer.latlngToNormalized(area.coordinates[1]);
        const color = Color(area.color).rgb().object(); // convert hex color to rgb
        const maxVal = 255;
        // normalize color values
        color.a = this.quadOpacity;
        color.r /= maxVal;
        color.g /= maxVal;
        color.b /= maxVal;
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
    tableDataToAreas(tableData: any[]): Area[] {
      const areas = tableData.map((item) => {
        const imageUrl = this.isMultiBandImage ? item.group_id.value : null;
        const fullCoordinates = item.coordinates.value.Elements;
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
          [fullCoordinates[1].Float, fullCoordinates[0].Float], // Corner A as [Lat, Lng]
          [fullCoordinates[5].Float, fullCoordinates[4].Float], // Corner C as [Lat, Lng]
        ] as LatLngBoundsLiteral;

        const color = this.tileColor(item);

        return { item, imageUrl, coordinates, color } as Area;
      });

      return areas;
    },
    // callback when zooming on map
    onZoom() {
      const zoom = this.map.getZoom();
      const wasPoints =
        zoom >= this.zoomThreshold && this.previousZoom < this.zoomThreshold;
      const wasTiled =
        zoom < this.zoomThreshold && this.previousZoom >= this.zoomThreshold;
      this.previousZoom = this.map.getZoom();
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
      this.overlay.removeQuad(this.quadLayerId);
      this.renderer.clearListeners();
      this.currentState.init();
      this.overlay.addQuad(
        this.quadLayerId,
        this.currentState.quads(),
        this.currentState.drawMode()
      );
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
      if (
        this.highlight &&
        this.highlight.value.minX === value.minX &&
        this.highlight.value.maxX === value.maxX &&
        this.highlight.value.minY === value.minY &&
        this.highlight.value.maxY === value.maxY
      ) {
        // dont push existing highlight
        return;
      }

      // TODO: support filtering multiple vars?
      const fieldSpec = this.fieldSpecs[0];
      let key = "";
      if (!!fieldSpec) {
        key =
          fieldSpec.type === SINGLE_FIELD
            ? fieldSpec.field
            : this.fieldHash(fieldSpec);
      } else if (!!this.summaries[0].key) {
        key = this.summaries[0].key;
      } else {
        console.error("Error createHighlight no available key");
        return;
      }
      updateHighlight(this.$router, {
        context: this.instanceName,
        dataset: this.dataset,
        key: key,
        value: value,
      });
    },

    drawFilters() {
      // TODO: impl this
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
      return fieldSpec.lngField + ":" + fieldSpec.latField;
    },

    showImageDrilldown(imageUrl: string, item: TableRow) {
      this.imageUrl = imageUrl ?? null;
      this.item = item ?? null;
      this.isImageDrilldown = true;
    },

    hideImageDrilldown() {
      this.isImageDrilldown = false;
    },

    tileColor(item: any) {
      let color = "#255DCC"; // Default
      if (this.isColoringByConfidence) {
        return this.colorScale(item.confidence.value);
      }
      if (item[this.targetField] && item[this.predictedField]) {
        color =
          item[this.targetField].value === item[this.predictedField].value
            ? "#03c003" // Correct: green.
            : "#be0000"; // Incorrect: red.
      }
      if (item.isExcluded) {
        return "#999999";
      }

      return color;
    },
    onNewData() {
      // clear quads
      this.overlay.clearQuads();
      // don't show exit button
      this.showExit = false;
      // create quads from latlng
      const quads = this.currentState.quads();
      if (!quads.length) {
        return;
      }
      // get bounds of quad set
      const mapBounds = this.getBounds(quads);
      // add the batched quads to a single layer on the overlay
      this.overlay.addQuad(
        this.quadLayerId,
        quads,
        this.currentState.drawMode()
      );
      // fit map to the quad set
      this.map.fitToBounds(mapBounds);
    },
  },

  watch: {
    dataItems() {
      this.onNewData();
    },
    summaries(cur, prev) {
      if (!prev.length && this.isClustering) {
        if (this.map.getZoom() < this.zoomThreshold) {
          this.currentState = this.clusterState;
          this.updateMapState();
        }
      } else {
        this.onNewData();
      }
    },
  },

  mounted() {
    this.createLumoMap();
  },
});
</script>

<style>
.geo-plot-container,
.geo-plot {
  position: relative;
  z-index: 0;
  width: 100%;
  height: 100%;
  bottom: 0;
}

.geo-plot-container .selection-toggle {
  position: absolute;
  z-index: 999;
  top: 80px;
  left: 10px;
  width: 34px;
  height: 34px;
  background-color: #fff;
  border: 2px solid rgba(0, 0, 0, 0.2);
  background-clip: padding-box;
  text-align: center;
  border-radius: 4px;
}

.cluster-toggle {
  position: absolute;
  z-index: 999;
  top: 40px;
  left: 10px;
  width: 34px;
  height: 34px;
  background-color: #fff;
  border: 2px solid rgba(0, 0, 0, 0.2);
  background-clip: padding-box;
  text-align: center;
  border-radius: 4px;
}
.map-toggle {
  position: absolute;
  z-index: 999;
  top: 120px;
  left: 10px;
  width: 34px;
  height: 34px;
  background-color: #fff;
  border: 2px solid rgba(0, 0, 0, 0.2);
  background-clip: padding-box;
  text-align: center;
  border-radius: 4px;
}
.confidence-toggle {
  position: absolute;
  z-index: 999;
  top: 160px;
  left: 10px;
  width: 34px;
  height: 34px;
  background-color: #fff;
  border: 2px solid rgba(0, 0, 0, 0.2);
  background-clip: padding-box;
  text-align: center;
  border-radius: 4px;
}
.cluster-icon {
  width: 30px;
  height: 30px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}
.confidence-icon {
  width: 30px;
  height: 30px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  font-weight: bolder;
  font-size: xx-large;
}
.toggled-confidence-icon {
  width: 30px;
  height: 30px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  font-weight: bolder;
  font-size: xx-large;
  background-size: 100%;
  background-clip: text;
  -webkit-background-clip: text;
  -moz-background-clip: text;
  background-image: linear-gradient(0deg, #f3ec78, #af4261);
  -webkit-text-fill-color: transparent;
  -moz-text-fill-color: transparent;
}
.confidence-toggle.active:hover::after {
  content: "----Less Confidence";
  position: absolute;
  white-space: nowrap;
  left: 30px;
  top: 15px; /*works out to 4 pixels from bottom (this is based off the font size)*/
  display: inline;
  position: absolute;
}
.confidence-toggle.active:hover::before {
  content: "----More Confidence";
  white-space: nowrap;
  left: 30px;
  top: -7px; /*works out to 4 pixels from top (this is based off the font size)*/
  display: inline;
  position: absolute;
}
.toggle {
}
.toggle:hover {
  background-color: #f4f4f4;
}

.toggle.active {
  color: #26b8d1;
}

.geo-plot-container .selection-toggle-control {
  text-decoration: none;
  color: black;
  cursor: pointer;
}

.geo-plot-container .selection-toggle-control:hover {
  text-decoration: none;
  color: black;
}

.geo-plot-container .selection-toggle.active {
  position: absolute;
}

.geo-plot-container .selection-toggle.active .selection-toggle-control {
  color: #26b8d1;
}

.geo-plot-container.selection-mode .geo-plot {
  cursor: crosshair;
}

path.selected {
  stroke-width: 2;
  fill-opacity: 0.4;
}

.geo-plot .markerPoint:hover {
  filter: brightness(1.2);
}

.geo-plot .markerPoint.selected {
  filter: hue-rotate(150deg);
}

.leaflet-tooltip {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 300px !important;
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

.geo-close-button:hover {
  background-color: #f4f4f4;
}
.geo-toast {
  position: absolute;
  top: 0px;
  right: 0px;
}
.image-label {
  position: absolute;
  left: 2px;
  top: 2px;
  z-index: 1;
}
.selection-exit {
  position: absolute;
}
</style>
