<template>
  <div
    class="geo-plot-container"
    :class="{ 'selection-mode': isSelectionMode }"
  >
    <div
      class="geo-plot"
      :id="mapID"
      @mousedown="onMouseDown"
      @mouseup="onMouseUp"
      @mousemove="onMouseMove"
      @keydown.esc="onEsc"
    ></div>

    <image-drilldown
      v-if="isRemoteSensing"
      @hide="hideImageDrilldown"
      :dataFields="dataFields"
      :imageUrl="imageUrl"
      :item="item"
      :visible="isImageDrilldown"
    />

    <div
      class="selection-toggle"
      :class="{ active: isSelectionMode }"
      @click="isSelectionMode = !isSelectionMode"
    >
      <a
        class="selection-toggle-control"
        title="Select area"
        aria-label="Select area"
      >
        <icon-base width="100%" height="100%"> <icon-crop-free /> </icon-base>
      </a>
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import $ from "jquery";
import leaflet, {
  MarkerOptions,
  LatLngTuple,
  LatLngBounds,
  CircleMarkerOptions,
} from "leaflet";
import Vue from "vue";
import IconBase from "./icons/IconBase.vue";
import IconCropFree from "./icons/IconCropFree.vue";
import ImageDrilldown from "./ImageDrilldown.vue";
import ImageLabel from "./ImageLabel.vue";
import { getters as datasetGetters } from "../store/dataset/module";
import { getters as requestGetters } from "../store/requests/module";
import { getters as routeGetters } from "../store/route/module";
import { Dictionary } from "../util/dict";
import {
  TableColumn,
  TableRow,
  D3M_INDEX_FIELD,
  Highlight,
  RowSelection,
  GeoCoordinateGrouping,
} from "../store/dataset/index";
import { updateHighlight, clearHighlight } from "../util/highlights";
import {
  addRowSelection,
  removeRowSelection,
  isRowSelected,
} from "../util/row";
import {
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  REAL_VECTOR_TYPE,
  GEOCOORDINATE_TYPE,
} from "../util/types";

import "leaflet/dist/leaflet.css";
import "leaflet/dist/images/marker-icon.png";
import "leaflet/dist/images/marker-icon-2x.png";
import "leaflet/dist/images/marker-shadow.png";

const SINGLE_FIELD = 1;
const SPLIT_FIELD = 2;
const CLOSE_BUTTON_CLASS = "geo-close-button";
const CLOSE_ICON_CLASS = "fa-times";

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

// Minimum pixels size of clickable target displayed on the map.
const TARGETSIZE = 6;

export default Vue.extend({
  name: "geo-plot",

  components: {
    IconBase,
    IconCropFree,
    ImageDrilldown,
  },

  props: {
    instanceName: String as () => string,
    dataItems: Array as () => any[],
    dataFields: Object as () => Dictionary<TableColumn>,
  },

  data() {
    return {
      poiLayer: null,
      map: null,
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
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    /*
     Flag to decide if we display accurate areas based on coordinates, or if they are physically
     too small, we present a circle big enough for the user to interact with them.
    */
    displayCircleMarker(): boolean {
      const pointA = this.map.latLngToContainerPoint([0, 0]);
      const pointB = this.map.latLngToContainerPoint([0, this.areasMeanLng]);
      const distanceInPixel = Math.abs(pointB.x - pointA.x);
      return distanceInPixel < TARGETSIZE;
    },

    target(): string {
      return routeGetters.getRouteTargetVariable(this.$store);
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
                color: this.colorPrediction(item),
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

      // Array to store the longitude width (degrees) of each areas.
      const longitudes = [];

      const areas = this.dataItems.map((item) => {
        const imageUrl = item.group_id.value;
        const fullCoordinates = item.coordinates.value.Elements;
        if (fullCoordinates.some((x) => x === undefined)) return;

        /*
          Item store the coordinates as a list of 8 values being four pairs of [Lng, Lat],
          one for each corner of the remote-sensing image.

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

        const color = this.colorPrediction(item);

        longitudes.push(fullCoordinates[4].Float - fullCoordinates[0].Float); // Corner C Lng - Corner A Lng

        return { item, imageUrl, coordinates, color } as Area;
      });

      // Calculate the mean longitude of the areas.
      this.areasMeanLng =
        longitudes.reduce((acc, val) => acc + val, 0) / longitudes.length;

      return areas;
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

    isRemoteSensing(): boolean {
      return routeGetters.isRemoteSensing(this.$store);
    },

    /* Base layer for the map. */
    baseLayer(): TileLayer {
      const URL = "http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png";
      return leaflet.tileLayer(URL);
    },
  },

  methods: {
    clearSelectionRect() {
      if (this.selectedRect) {
        this.selectedRect.remove();
        this.selectedRect = null;
      }
      if (this.currentRect) {
        this.currentRect.remove();
        this.currentRect = null;
      }
      if (this.closeButton) {
        this.closeButton.remove();
        this.closeButton = null;
      }
    },

    onMouseDown(event: MouseEvent) {
      const mapEventTarget = event.target as HTMLElement;

      // check if mapEventTarget is the close button or icon
      if (
        mapEventTarget.classList.contains(CLOSE_BUTTON_CLASS) ||
        mapEventTarget.classList.contains(CLOSE_ICON_CLASS)
      ) {
        this.clearSelection();
        this.selectedRect.remove();
        this.selectedRect = null;
        this.closeButton.remove();
        this.closeButton = null;
        return;
      }

      if (this.isSelectionMode) {
        this.clearSelectionRect();

        const offset = $(this.map.getContainer()).offset();
        this.startingLatLng = this.map.containerPointToLatLng({
          x: event.pageX - offset.left,
          y: event.pageY - offset.top,
        });

        const bounds = [this.startingLatLng, this.startingLatLng];
        this.currentRect = leaflet.rectangle(bounds, {
          color: "#255DCC",
          weight: 1,
          bubblingMouseEvents: false,
        });
        this.currentRect.on("click", (e) => {
          this.setSelection(e.target);
        });
        this.currentRect.addTo(this.map);

        // enable drawing mode
        // this.map.off('click', this.clearSelection);
        this.map.dragging.disable();
      }
    },

    onMouseUp(event: MouseEvent) {
      if (this.currentRect) {
        this.setSelection(this.currentRect);
        this.currentRect = null;

        // disable drawing mode
        this.map.dragging.enable();
        // this.map.on('click', this.clearSelection);
      }
    },

    onMouseMove(event: MouseEvent) {
      if (this.currentRect) {
        const offset = $(this.map.getContainer()).offset();
        const latLng = this.map.containerPointToLatLng({
          x: event.pageX - offset.left,
          y: event.pageY - offset.top,
        });
        const bounds = [this.startingLatLng, latLng];
        this.currentRect.setBounds(bounds);
      }
    },

    onEsc() {
      if (this.currentRect) {
        this.clearSelectionRect();
        // disable drawing mode
        this.map.dragging.enable();
      }
    },

    setSelection(rect) {
      this.clearSelection();

      this.selectedRect = rect;
      const $selected = $(this.selectedRect._path);
      $selected.addClass("selected");

      const ne = rect.getBounds().getNorthEast();
      const sw = rect.getBounds().getSouthWest();
      const icon = leaflet.divIcon({
        className: CLOSE_BUTTON_CLASS,
        iconSize: null,
        html: `<i class="fa ${CLOSE_ICON_CLASS}"></i>`,
      });
      this.closeButton = leaflet.marker([ne.lat, ne.lng], {
        icon: icon,
      });
      this.closeButton.addTo(this.map);
      this.createHighlight({
        minX: sw.lng,
        maxX: ne.lng,
        minY: sw.lat,
        maxY: ne.lat,
      });
    },

    clearSelection() {
      if (this.selectedRect) {
        $(this.selectedRect._path).removeClass("selected");
        clearHighlight(this.$router);
      }
      if (this.closeButton) {
        this.closeButton.remove();
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
      const key =
        fieldSpec.type === SINGLE_FIELD
          ? fieldSpec.field
          : this.fieldHash(fieldSpec);

      updateHighlight(this.$router, {
        context: this.instanceName,
        dataset: this.dataset,
        key: key,
        value: value,
      });
    },

    drawHighlight() {
      if (
        this.highlight &&
        this.highlight.value.minX !== undefined &&
        this.highlight.value.maxX !== undefined &&
        this.highlight.value.minY !== undefined &&
        this.highlight.value.maxY !== undefined
      ) {
        const rect = leaflet.rectangle(
          [
            [this.highlight.value.minY, this.highlight.value.minX],
            [this.highlight.value.maxY, this.highlight.value.maxX],
          ],
          {
            color: "#255DCC",
            weight: 1,
            bubblingMouseEvents: false,
          }
        );
        rect.on("click", (e) => {
          this.setSelection(e.target);
        });
        rect.addTo(this.map);

        this.setSelection(rect);
      }
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

    clear() {
      if (this.selectedRect) {
        this.selectedRect.remove();
        this.selectedRect = null;
      }

      if (this.currentRect) {
        this.currentRect.remove();
        this.currentRect = null;
      }

      if (this.closeButton) {
        this.closeButton.remove();
        this.closeButton = null;
      }

      _.forIn(this.markers, (markerLayer) => {
        markerLayer.removeFrom(this.map);
      });

      if (this.map.hasLayer(this.poiLayer)) {
        this.map.removeLayer(this.poiLayer);
      }

      this.markers = {};
      this.startingLatLng = null;
    },

    toggleSelection(event) {
      const marker = event.target;
      const row = marker.options.row;
      if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
        addRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          row[D3M_INDEX_FIELD]
        );
      } else {
        removeRowSelection(
          this.$router,
          this.instanceName,
          this.rowSelection,
          row[D3M_INDEX_FIELD]
        );
      }
    },

    updateMarkerSelection(markers) {
      markers.forEach((marker) => {
        const row = marker.options.row;
        const markerElem = marker.getElement();
        const isSelected = isRowSelected(
          this.rowSelection,
          row[D3M_INDEX_FIELD]
        );
        markerElem.classList.toggle("selected", isSelected);
      });
    },

    showImageDrilldown(imageUrl: string, item: TableRow) {
      this.imageUrl = imageUrl ?? null;
      this.item = item ?? null;
      this.isImageDrilldown = true;
    },

    hideImageDrilldown() {
      this.isImageDrilldown = false;
    },

    colorPrediction(item: any) {
      let color = "#255DCC"; // Default

      if (item[this.targetField] && item[this.predictedField]) {
        color =
          item[this.targetField].value === item[this.predictedField].value
            ? "#03c003" // Correct: green.
            : "#be0000"; // Incorrect: red.
      }

      return color;
    },

    /* Create a Leaflet map, if it doesn't exist already, with basic defaults. */
    createMap() {
      if (this.map) {
        return;
      }

      // NOTE: this component re-mounts on any change, so do everything in here
      this.map = leaflet.map(this.mapID, {
        center: [30, 0],
        zoom: 2,
      });

      if (this.mapZoom) {
        this.map.setZoom(this.mapZoom, { animate: true });
      }

      if (this.mapCenter) {
        this.map.panTo(
          {
            lat: this.mapCenter[1],
            lng: this.mapCenter[0],
          },
          { animate: true }
        );
      }

      this.baseLayer.addTo(this.map);

      // this.map.on('click', this.clearSelection);
    },

    /* Create a Leaflet Group to contains the Point Of Interest (POI) if it doesn't exist already. */
    createPoiLayer(pois) {
      // Test if the area Layer is already on the map.
      if (this.map.hasLayer(this.poiLayer)) {
        // Let's clear all of it before adding new ones.
        this.poiLayer.clearLayers();
      } else {
        // Create a layer group to contain all the POIS to be displayed.
        this.poiLayer = leaflet.layerGroup();
        this.poiLayer.addTo(this.map);
      }

      // Extend the bounds of the map to include all coordinates.
      const bounds = leaflet.latLngBounds(null);
      pois.forEach((poi) => {
        if (poi.coordinates) {
          poi.coordinates.forEach((coordinate) => bounds.extend(coordinate));
        } else {
          bounds.extend([poi.lat, poi.lng]);
        }
      });
      if (bounds.isValid()) {
        this.map.fitBounds(bounds);
      }
    },

    /* Display areas as circleMarker or rectangle layers on the map. */
    displayAreas() {
      this.createPoiLayer(this.areas);

      // Add each area to the layer group.
      this.areas.forEach((area) => {
        const { color, coordinates, imageUrl, item } = area;

        // Create the layer (circleMarker or rectangle) for the user to interact.
        let layer: any;
        if (this.displayCircleMarker) {
          const centerOfCoordinates = [
            coordinates[0][0] + (coordinates[1][0] - coordinates[0][0]), // Lat
            coordinates[0][1] + (coordinates[1][1] - coordinates[0][1]), // Lng
          ] as LatLngTuple;
          const displayOptions = {
            color: color,
            radius: TARGETSIZE / 2,
            stroke: false,
            fillOpacity: 1.0,
          };
          layer = leaflet.circleMarker(centerOfCoordinates, displayOptions);
        } else {
          layer = leaflet.rectangle(coordinates, { color });
        }

        // Create a Vue tooltip for the area with the label for the image.
        const ImageLabelComponent = Vue.extend(ImageLabel);
        const tooltip = new ImageLabelComponent({
          parent: this,
          propsData: {
            dataFields: this.dataFields,
            includeActive: true,
            item: item,
          },
          store: this.$store,
        }).$mount();

        // Add interactivity to the layer.
        layer
          .bindTooltip(tooltip.$el as HTMLElement)
          .on("click", () => this.showImageDrilldown(imageUrl, item));

        // Add the rectangle to the layer group.
        this.poiLayer.addLayer(layer);
      });
    },

    /* Display point as circleMarker on the map. */
    displayPoints() {
      this.createPoiLayer(this.pointGroups[0].points);

      this.pointGroups.forEach((group) => {
        const hash = this.fieldHash(group.field);
        const layerGroup = leaflet.layerGroup([]);

        group.points.forEach((point) => {
          const coordinate = [point.lat, point.lng] as LatLngTuple;
          const displayOptions = {
            className: "markerPoint",
            fillColor: point.color,
            fillOpacity: 1.0,
            radius: TARGETSIZE / 2,
            row: (<any>point).row,
            stroke: false,
          } as CircleMarkerOptions;
          const layer = leaflet.circleMarker(coordinate, displayOptions);

          layer.bindTooltip(() => {
            const target = point.row[this.target].value;
            const values = [];
            const MAX_VALUES = 5;

            this.getTopVariables.forEach((v) => {
              if (point.row[v] && values.length <= MAX_VALUES) {
                values.push(`<b>${_.capitalize(v)}:</b> ${point.row[v].value}`);
              }
            });

            return [`<b>${_.capitalize(target)}</b>`]
              .concat(values)
              .join("<br>");
          });

          layer.on("click", this.toggleSelection);

          // Add the point to the layer group.
          layerGroup.addLayer(layer);
        });

        this.markers[hash] = layerGroup;
        layerGroup.on("add", () =>
          this.updateMarkerSelection(layerGroup.getLayers())
        );

        // Add the point to the layer group.
        this.poiLayer.addLayer(layerGroup);
      });
    },

    paint() {
      this.createMap();
      this.clear();

      if (this.isRemoteSensing) {
        // Display areas and update them on zoom to be sure they are selectable.
        this.displayAreas();
        this.map.on("zoomend", () => this.displayAreas());
      } else {
        this.displayPoints();
      }

      this.drawHighlight();
      this.drawFilters();
    },
  },

  watch: {
    dataItems() {
      this.paint();
    },

    rowSelection() {
      const markers = _.map(this.markers, (markerLayer) =>
        markerLayer.getLayers()
      ).reduce((prev, cur) => [...prev, ...cur], []);
      this.updateMarkerSelection(markers);
    },
  },

  mounted() {
    this.paint();
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

.geo-plot-container .selection-toggle:hover {
  background-color: #f4f4f4;
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
</style>
