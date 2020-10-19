<template>
  <div class="facet-card">
    <div class="group-header">
      <span class="header-title">
        {{ headerLabel }}
      </span>
      <i class="fa fa-globe"></i>
      <type-change-menu
        :geocoordinate="true"
        :dataset="dataset"
        :field="target"
        :expandCollapse="expandCollapse"
        :expand="expand"
      />
    </div>
    <div class="geofacet-container">
      <div
        class="geofacet"
        v-bind:id="mapID"
        v-on:mousedown="onMouseDown"
        v-on:mouseup="onMouseUp"
        v-on:mousemove="onMouseMove"
      ></div>
      <div v-if="isAvailableFeatures">
        <button
          class="action-btn btn btn-sm btn-outline-secondary ml-2 mr-1 mb-2"
          @click="selectFeature()"
        >
          Add
        </button>
      </div>
      <div v-if="isFeaturesToModel">
        <button
          class="action-btn btn btn-sm btn-outline-secondary ml-2 mr-1 mb-2"
          @click="removeFeature()"
        >
          Remove
        </button>
      </div>
    </div>
    <div v-if="expand" class="latlon">
      <facet-numerical
        :instanceName="latSummary.label"
        :summary="latSummary"
        :enabledTypeChanges="enabledTypeChanges"
        :enable-highlighting="enableHighlighting"
        :highlight="latHighlight"
        @numerical-click="latHistogramClick"
        @range-change="latRangeChange"
      />
      <facet-numerical
        :instanceName="lonSummary.label"
        :summary="lonSummary"
        :enabledTypeChanges="enabledTypeChanges"
        :enable-highlighting="enableHighlighting"
        :highlight="lonHighlight"
        @numerical-click="lonHistogramClick"
        @range-change="lonRangeChange"
      />
    </div>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import $ from "jquery";
import leaflet from "leaflet";
import Vue from "vue";
import IconBase from "../icons/IconBase.vue";
import IconCropFree from "../icons/IconCropFree.vue";
import { scaleThreshold } from "d3";
import {
  actions as datasetActions,
  getters as datasetGetters,
} from "../../store/dataset/module";
import { getters as routeGetters } from "../../store/route/module";
import { actions as appActions } from "../../store/app/module";
import { Dictionary } from "../../util/dict";
import {
  TableRow,
  VariableSummary,
  Bucket,
  Extrema,
  Highlight,
  NUMERICAL_SUMMARY,
  RowSelection,
  SummaryMode,
  TaskTypes,
} from "../../store/dataset";
import TypeChangeMenu from "../TypeChangeMenu.vue";
import FacetNumerical from "./FacetNumerical.vue";
import { updateHighlight, clearHighlight } from "../../util/highlights";
import {
  GEOCOORDINATE_TYPE,
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  REAL_VECTOR_TYPE,
  EXPAND_ACTION_TYPE,
  COLLAPSE_ACTION_TYPE,
} from "../../util/types";
import { overlayRouteEntry, varModesToString } from "../../util/routes";
import { Filter, removeFiltersByName } from "../../util/filters";
import { Feature, Activity, SubActivity } from "../../util/userEvents";

import "leaflet/dist/leaflet.css";

import helpers, { polygon, featureCollection, point } from "@turf/helpers";
import bbox from "@turf/bbox";
import booleanContains from "@turf/boolean-contains";
import { BLUE_PALETTE, BLACK_PALETTE } from "../../util/color";
const SINGLE_FIELD = 1;
const SPLIT_FIELD = 2;
const CLOSE_BUTTON_CLASS = "geo-close-button";
const CLOSE_ICON_CLASS = "fa-times";
const LON_LAT_KEY = "longitude:latitude";

interface GeoField {
  type: number;
  latField?: string;
  lngField?: string;
  field?: string;
}

interface GeoTableRow extends TableRow {
  latitude: number;
  longitude: number;
}

interface BucketData {
  extrema: Extrema;
  buckets: Bucket[];
}

const GEOCOORDINATE_LABEL = "longitude";

/**
 * Geocoordinate Facet.
 * @param {Boolean} [expanded=false] - To display the facet expanded; Collapsed by default.
 */
export default Vue.extend({
  name: "geocoordinate-facet",

  components: {
    TypeChangeMenu,
    IconBase,
    IconCropFree,
    FacetNumerical,
  },

  props: {
    summary: Object as () => VariableSummary,
    isAvailableFeatures: Boolean as () => boolean,
    isFeaturesToModel: Boolean as () => boolean,
    enableHighlighting: Boolean as () => boolean,
    ignoreHighlights: Boolean as () => boolean,
    logActivity: {
      type: String as () => Activity,
      default: Activity.DATA_PREPARATION,
    },
    expanded: { type: Boolean, default: false },
  },

  data() {
    return {
      map: null as leaflet.Map,
      baseLayer: null as leaflet.Layer,
      bounds: null as leaflet.LatLngBounds,
      closeButton: null,
      startingLatLng: null as leaflet.LatLng,
      currentRect: null as leaflet.Rectangle,
      selectedRect: null as leaflet.Rectangle,
      baseLineLayer: null as leaflet.Layer,
      filteredLayer: null as leaflet.Layer,
      expand: this.expanded,
      enabledTypeChanges: new Array(0),
      blockNextEvent: false,
    };
  },

  computed: {
    dataset(): string {
      return routeGetters.getRouteDataset(this.$store);
    },

    latSummary(): VariableSummary {
      const latSummary: VariableSummary = {
        label: LATITUDE_TYPE,
        description: this.summary.description,
        type: NUMERICAL_SUMMARY,
        key: this.summary.key,
        dataset: this.summary.dataset,
        baseline: this.latitudeToNumeric("baseline"),
        filtered: this.latitudeToNumeric("filtered"),
      };
      return latSummary;
    },

    lonSummary(): VariableSummary {
      const lonSummary: VariableSummary = {
        label: LONGITUDE_TYPE,
        description: this.summary.description,
        type: NUMERICAL_SUMMARY,
        key: this.summary.key,
        dataset: this.summary.dataset,
        baseline: this.longitudeToNumeric("baseline"),
        filtered: this.longitudeToNumeric("filtered"),
      };
      return lonSummary;
    },

    latHighlight(): Object {
      if (this.hasValidGeoHighlight) {
        return {
          value: {
            from: this.calcBucketKey(this.highlight.value.minY, LATITUDE_TYPE),
            to: this.calcBucketKey(this.highlight.value.maxY, LATITUDE_TYPE),
          },
          context: LATITUDE_TYPE,
          key: this.summary.key,
          dataset: this.dataset,
        };
      } else {
        return null;
      }
    },

    lonHighlight(): Object {
      if (this.hasValidGeoHighlight) {
        return {
          value: {
            from: this.calcBucketKey(this.highlight.value.minX, LONGITUDE_TYPE),
            to: this.calcBucketKey(this.highlight.value.maxX, LONGITUDE_TYPE),
          },
          context: LONGITUDE_TYPE,
          key: this.summary.key,
          dataset: this.dataset,
        };
      } else {
        return null;
      }
    },

    target(): string {
      return this.summary.key;
    },

    instanceName(): string {
      return "unique-map";
    },

    mapID(): string {
      return `map-${this.instanceName}`;
    },

    // Computes the bounds of the summary data.
    bucketBounds(): helpers.BBox {
      return bbox(this.bucketFeatures);
    },

    // Creates a GeoJSON feature collection that can be passed directly to a Leaflet layer
    // for rendering.  The collection represents the baseline bucket set for geocoordinate,
    // and does not change as filters / highlights are introduced.
    bucketFeatures(): helpers.FeatureCollection {
      // compute the bucket size in degrees
      const buckets = this.summary.baseline.buckets;
      const xSize = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
      const ySize =
        _.toNumber(buckets[0].buckets[1].key) -
        _.toNumber(buckets[0].buckets[0].key);

      // create a feature collection from the server-supplied bucket data
      const features: helpers.Feature[] = [];
      this.summary.baseline.buckets.forEach((lonBucket) => {
        lonBucket.buckets.forEach((latBucket) => {
          // Don't include features with a count of 0.
          if (latBucket.count > 0) {
            const xCoord = _.toNumber(lonBucket.key);
            const yCoord = _.toNumber(latBucket.key);
            const feature = polygon(
              [
                [
                  [yCoord, xCoord],
                  [yCoord + ySize, xCoord],
                  [yCoord + ySize, xCoord + xSize],
                  [yCoord, xCoord + xSize],
                  [yCoord, xCoord],
                ], // leaflet and most map frameworks use latlng which is y,x
              ],
              { selected: false, count: latBucket.count }
            );
            features.push(feature);
          }
        });
      });

      return featureCollection(features);
    },

    // Creates a GeoJSON feature collection that can be passed directly to a Leaflet layer
    // for rendering.  The collection represents the subset of buckets to be rendered based
    // on the currently applied filters and highlights.
    filteredBucketFeatures(): helpers.FeatureCollection {
      if (this.summary.filtered) {
        // compute the bucket size in degrees
        const buckets = this.summary.filtered.buckets;
        const xSize = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
        const ySize =
          _.toNumber(buckets[0].buckets[1].key) -
          _.toNumber(buckets[0].buckets[0].key);

        // create a feature collection from the server-supplied bucket data
        const features: helpers.Feature[] = [];
        this.summary.filtered.buckets.forEach((lonBucket) => {
          lonBucket.buckets.forEach((latBucket) => {
            // Don't include features with a count of 0.
            if (latBucket.count > 0) {
              const xCoord = _.toNumber(lonBucket.key);
              const yCoord = _.toNumber(latBucket.key);
              const feature = polygon(
                [
                  [
                    [xCoord, yCoord],
                    [xCoord, yCoord + ySize],
                    [xCoord + xSize, yCoord + ySize],
                    [xCoord + xSize, yCoord],
                    [xCoord, yCoord],
                  ],
                ],
                { selected: false, count: latBucket.count }
              );
              features.push(feature);
            }
          });
        });

        return featureCollection(features);
      } else {
        const features: helpers.Feature[] = [];
        return featureCollection(features);
      }
    },

    // Returns the minimum non-zero bucket count value
    minCount(): number {
      return this.bucketFeatures.features.reduce(
        (min, feature) =>
          feature.properties.count < min ? feature.properties.count : min,
        Number.MAX_SAFE_INTEGER
      );
    },

    // Returns the maximum bucket count value
    maxCount(): number {
      return this.bucketFeatures.features.reduce(
        (max, feature) =>
          feature.properties.count > max ? feature.properties.count : max,
        Number.MIN_SAFE_INTEGER
      );
    },

    filteredMinCount(): number {
      return this.filteredBucketFeatures.features.reduce(
        (min, feature) =>
          feature.properties.count < min ? feature.properties.count : min,
        Number.MAX_SAFE_INTEGER
      );
    },

    headerLabel(): string {
      return this.summary.label.toUpperCase();
    },

    hasFilters(): boolean {
      return routeGetters.getDecodedFilters(this.$store).length > 0;
    },

    // is the display in included (blue) or excluded (black) mode
    includedActive(): boolean {
      return routeGetters.getRouteInclude(this.$store);
    },

    // is data currently being highlighted
    highlight(): Highlight {
      return routeGetters.getDecodedHighlight(this.$store);
    },

    hasValidGeoHighlight(): Boolean {
      return (
        !!this.highlight &&
        !!this.highlight.value &&
        !!this.highlight.value.minX &&
        !!this.highlight.value.minY &&
        !!this.highlight.value.maxX &&
        !!this.highlight.value.maxY
      );
    },

    selectedRows(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },

    selectedPoints(): helpers.Point[] {
      if (this.selectedRows) {
        const tableItems = this.includedActive
          ? datasetGetters.getIncludedTableDataItems(this.$store)
          : datasetGetters.getExcludedTableDataItems(this.$store);
        if (this.isGeoTableRows(tableItems)) {
          const selectedItems = this.selectedRows.d3mIndices.flatMap(
            (index) => {
              return tableItems.filter((item) => item.d3mIndex === index);
            }
          );
          const selectedPoints = selectedItems.map((item) =>
            point([Number(item.longitude), Number(item.latitude)])
          );
          return selectedPoints.map((p) => p.geometry);
        }
      }
      return [];
    },
  },

  methods: {
    calcBucketKey(value: string, type: string): string {
      const numValue = _.toNumber(value);
      const buckets =
        type === LONGITUDE_TYPE
          ? this.lonSummary.baseline.buckets
          : this.latSummary.baseline.buckets;
      const step = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
      return _.toString(numValue - (numValue % step));
    },

    numericWithMetadata(buckets: Bucket[]): BucketData {
      const extrema = {
        min: _.toNumber(buckets[0].key),
        max:
          _.toNumber(buckets[buckets.length - 1].key) +
          _.toNumber(buckets[buckets.length - 1].key) -
          _.toNumber(buckets[buckets.length - 2].key),
      };
      return {
        extrema,
        buckets,
      };
    },

    longitudeToNumeric(bucketType: string): BucketData {
      if (this.summary[bucketType]) {
        const lonBuckets = this.summary[bucketType].buckets.reduce(
          (lbs, lonBucket) => {
            lbs.push({
              key: lonBucket.key,
              count: lonBucket.buckets.reduce((total, latBucket) => {
                return (total += latBucket.count);
              }, 0),
            });
            return lbs;
          },
          []
        );
        return this.numericWithMetadata(lonBuckets);
      } else {
        return null;
      }
    },

    latitudeToNumeric(bucketType: string): BucketData {
      if (this.summary[bucketType]) {
        const latBuckets = this.summary[bucketType].buckets.reduce(
          (lbs, lonBucket) => {
            if (lbs.length > 0) {
              lonBucket.buckets.forEach((latBucket, ind) => {
                lbs[ind].count += latBucket.count;
              });
            } else {
              lbs = lonBucket.buckets.map((b) => {
                return {
                  key: b.key,
                  count: b.count,
                };
              });
            }
            return lbs;
          },
          []
        );
        return this.numericWithMetadata(latBuckets);
      } else {
        return null;
      }
    },

    latHistogramClick(
      context: string,
      key: string,
      value: { from: number; to: number; type: string },
      dataset: string
    ) {
      if (this.blockNextEvent) {
        this.blockNextEvent = false;
        return;
      }
      this.onHistogramAction(
        context,
        key,
        value,
        dataset,
        LATITUDE_TYPE,
        "numerical-click"
      );
    },

    lonHistogramClick(
      context: string,
      key: string,
      value: { from: number; to: number; type: string },
      dataset: string
    ) {
      if (this.blockNextEvent) {
        this.blockNextEvent = false;
        return;
      }
      this.onHistogramAction(
        context,
        key,
        value,
        dataset,
        LONGITUDE_TYPE,
        "numerical-click"
      );
    },

    latRangeChange(
      context: string,
      key: string,
      value: { from: number; to: number; type: string },
      dataset: string
    ) {
      this.blockNextEvent = true;
      this.onHistogramAction(
        context,
        key,
        value,
        dataset,
        LATITUDE_TYPE,
        "range-change"
      );
    },

    lonRangeChange(
      context: string,
      key: string,
      value: { from: number; to: number; type: string },
      dataset: string
    ) {
      this.blockNextEvent = true;
      this.onHistogramAction(
        context,
        key,
        value,
        dataset,
        LONGITUDE_TYPE,
        "range-change"
      );
    },

    onHistogramAction(
      context: string,
      key: string,
      value: { from: number; to: number; type: string },
      dataset: string,
      geocoordinateComponent: string,
      actionType
    ) {
      if (this.hasValidGeoHighlight) {
        const currentValue = this.highlight.value;
        const highlightValue = {
          minX: currentValue.minX,
          maxX: currentValue.maxX,
          minY: currentValue.minY,
          maxY: currentValue.maxY,
        };
        if (value === null) {
          clearHighlight(this.$router);
        } else {
          if (geocoordinateComponent === LONGITUDE_TYPE) {
            highlightValue.minX = value.from;
            highlightValue.maxX = value.to;
          } else {
            highlightValue.minY = value.from;
            highlightValue.maxY = value.to;
          }
          this.createHighlight(highlightValue);
        }
      } else {
        this.createHighlight({
          minX: this.lonSummary.baseline.extrema.min,
          maxX: this.lonSummary.baseline.extrema.max,
          minY: this.latSummary.baseline.extrema.min,
          maxY: this.latSummary.baseline.extrema.max,
        });
      }
      this.clearSelectionRect();
      this.$emit(actionType, key, value);
      appActions.logUserEvent(this.$store, {
        feature: Feature.CHANGE_HIGHLIGHT,
        activity: this.logActivity,
        subActivity: SubActivity.DATA_TRANSFORMATION,
        details: { key: key, value: value },
      });
    },

    expandCollapse(action) {
      if (action === EXPAND_ACTION_TYPE) {
        this.expand = true;
      } else if (action === COLLAPSE_ACTION_TYPE) {
        this.expand = false;
      }
    },

    async selectFeature() {
      const training = routeGetters
        .getDecodedTrainingVariableNames(this.$store)
        .concat([this.summary.key]);

      // update task based on the current training data
      const taskResponse = await datasetActions.fetchTask(this.$store, {
        dataset: routeGetters.getRouteDataset(this.$store),
        targetName: routeGetters.getRouteTargetVariable(this.$store),
        variableNames: training,
      });

      const task = taskResponse.data.task.join(",");
      const varModesMap = routeGetters.getDecodedVarModes(this.$store);

      if (task.includes(TaskTypes.REMOTE_SENSING)) {
        const available = routeGetters.getAvailableVariables(this.$store);

        training.forEach((v) => {
          varModesMap.set(v, SummaryMode.MultiBandImage);
        });

        available.forEach((v) => {
          varModesMap.set(v.colName, SummaryMode.MultiBandImage);
        });

        varModesMap.set(
          routeGetters.getRouteTargetVariable(this.$store),
          SummaryMode.MultiBandImage
        );
      }
      const varModesStr = varModesToString(varModesMap);

      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        training: training.join(","),
        task: task,
        varModes: varModesStr,
      });

      this.$router.push(entry).catch((err) => console.warn(err));
    },

    async removeFeature() {
      const training = routeGetters.getDecodedTrainingVariableNames(
        this.$store
      );
      _.remove(training, (t) => t === this.summary.key);

      // update task based on the current training data
      const taskResponse = await datasetActions.fetchTask(this.$store, {
        dataset: routeGetters.getRouteDataset(this.$store),
        targetName: routeGetters.getRouteTargetVariable(this.$store),
        variableNames: training,
      });

      const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
        training: training.join(","),
        task: taskResponse.data.task.join(","),
      });

      this.$router.push(entry).catch((err) => console.warn(err));

      removeFiltersByName(this.$router, this.summary.key);
    },

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

    onMouseUp(event: MouseEvent) {
      if (this.currentRect) {
        // prevent creation of a single point highlight via click
        const pxBounds = (<any>this.currentRect)._pxBounds as leaflet.Bounds;
        const rectangleSize = pxBounds.max.subtract(pxBounds.min);
        const singlePoint = leaflet.point(1, 1);

        if (!rectangleSize.equals(singlePoint)) {
          this.setSelection(this.currentRect);
        } else {
          this.clearSelection();
          this.clearSelectionRect();
        }

        this.currentRect = null;
      }
    },

    onMouseMove(event: MouseEvent) {
      if (this.currentRect) {
        const offset = $(this.map.getContainer()).offset();
        const latLng = this.map.containerPointToLatLng(
          leaflet.point(event.pageX - offset.left, event.pageY - offset.top)
        );
        const bounds = leaflet.latLngBounds(this.startingLatLng, latLng);
        this.currentRect.setBounds(bounds);
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
      if (this.isFeaturesToModel) {
        this.clearSelectionRect();

        const offset = $(this.map.getContainer()).offset();

        this.startingLatLng = this.map.containerPointToLatLng(
          leaflet.point(event.pageX - offset.left, event.pageY - offset.top)
        );

        const bounds = leaflet.latLngBounds(
          this.startingLatLng,
          this.startingLatLng
        );

        this.currentRect = leaflet.rectangle(bounds, {
          color: this.includedActive ? "#255DCC" : "black",
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

    setSelection(rect: leaflet.Rectangle) {
      this.clearSelection();

      this.selectedRect = rect;
      const rectPath = (<any>this.selectedRect)._path;
      const $selected = $(rectPath);
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
        const rectPath = (<any>this.selectedRect)._path;
        $(rectPath).removeClass("selected");
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
        this.highlight.value &&
        this.highlight.value.minX === value.minX &&
        this.highlight.value.maxX === value.maxX &&
        this.highlight.value.minY === value.minY &&
        this.highlight.value.maxY === value.maxY
      ) {
        return;
      }

      updateHighlight(this.$router, {
        context: this.instanceName,
        dataset: this.dataset,
        key: this.summary.key,
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

    paint() {
      // NOTE: this component re-mounts on any change, so do everything in here
      if (!this.highlight) {
        this.clearSelectionRect();
      }

      // remove previously added layers
      if (this.baseLineLayer) {
        this.baseLineLayer.removeFrom(this.map);
      }
      if (this.filteredLayer) {
        this.filteredLayer.removeFrom(this.map);
      }

      // Lazy map instantiation with a default zoom position
      if (!this.map) {
        this.map = leaflet.map(this.mapID, {
          center: [30, 0],
          zoom: 2,
          scrollWheelZoom: false,
          zoomControl: false,
          doubleClickZoom: false,
        });
        this.map.dragging.disable();

        this.baseLayer = leaflet.tileLayer(
          "http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png"
        );
        this.baseLayer.addTo(this.map);
      }

      // Restrict the bounds of the map to the bucket set
      const bounds = this.bucketBounds;
      const northEast = leaflet.latLng(bounds[3], bounds[2]);
      const southWest = leaflet.latLng(bounds[1], bounds[0]);
      this.bounds = leaflet.latLngBounds(northEast, southWest);

      if (this.bounds.isValid()) {
        this.map.fitBounds(this.bounds);

        // Generate the colour ramp scaling function
        const maxVal = this.maxCount;
        const minVal = this.minCount;

        // Check to see if we're showing included or excluded mode, whichi based on the user's current
        // tab setting.  In included mode we render all the currently included data in blue, in excluded
        //  mode we show only excluded data and render it in black.

        if (this.includedActive) {
          if (!this.highlight && !this.hasFilters) {
            // if there's no highlight active render from the baseline (all) set of buckets.
            const d = (maxVal - minVal) / BLUE_PALETTE.length;
            const domain = BLUE_PALETTE.map(
              (val, index) => minVal + d * (index + 1)
            );
            const scaleColors = scaleThreshold()
              .range(BLUE_PALETTE as any)
              .domain(domain);

            // Render the heatmap buckets as a GeoJSON layer
            this.baseLineLayer = leaflet.geoJSON(this.bucketFeatures, {
              style: (feature) => {
                let containsSelected = false;

                for (const point of this.selectedPoints) {
                  if (booleanContains(feature, point)) {
                    containsSelected = true;
                  }
                }

                const fill = containsSelected
                  ? "rgba(255,0,103,.2)"
                  : scaleColors(feature.properties.count).toString(16);

                return {
                  fillColor: fill,
                  weight: 0,
                  opacity: 1,
                  color: "rgba(0,0,0,0)",
                  dashArray: "3",
                  fillOpacity: 0.7,
                };
              },
            });
            this.baseLineLayer.addTo(this.map);
          } else {
            // there's a highlight active - render from the set of features returned in the filter portion of the
            // variable summary strucure
            const filteredMinVal = this.filteredMinCount;
            const dVal = (maxVal - minVal) / BLUE_PALETTE.length;
            const filteredDomain = BLUE_PALETTE.map(
              (val, index) => filteredMinVal + dVal * (index + 1)
            );
            const filteredScaleColors = scaleThreshold()
              .range(BLUE_PALETTE as any)
              .domain(filteredDomain);

            this.filteredLayer = leaflet.geoJSON(this.filteredBucketFeatures, {
              style: (feature) => {
                let containsSelected = false;

                for (const point of this.selectedPoints) {
                  if (booleanContains(feature, point)) {
                    containsSelected = true;
                  }
                }

                const fill = containsSelected
                  ? "rgba(255,0,103,.2)"
                  : filteredScaleColors(feature.properties.count).toString(16);

                return {
                  fillColor: fill,
                  weight: 0,
                  opacity: 1,
                  color: "rgba(0,0,0,0)",
                  dashArray: "3",
                  fillOpacity: 0.7,
                };
              },
            });
            this.filteredLayer.addTo(this.map);
          }
        } else if (this.hasFilters) {
          // Excluded mode is active - render visuals using a black pallette.
          // Any data we need to render is in the filter portion of variable summary structure.

          const filteredMinVal = this.filteredMinCount;
          const dVal = (maxVal - minVal) / BLACK_PALETTE.length;
          const filteredDomain = BLACK_PALETTE.map(
            (val, index) => filteredMinVal + dVal * (index + 1)
          );
          const filteredScaleColors = scaleThreshold()
            .range(BLACK_PALETTE as any)
            .domain(filteredDomain);

          this.filteredLayer = leaflet.geoJSON(this.filteredBucketFeatures, {
            style: (feature) => {
              return {
                fillColor: filteredScaleColors(
                  feature.properties.count
                ).toString(16),
                weight: 0,
                opacity: 1,
                color: "rgba(0,0,0,0)",
                dashArray: "3",
                fillOpacity: 0.7,
              };
            },
          });
          this.filteredLayer.addTo(this.map);
          this.clearSelectionRect();
        }
      }
    },

    // type guard for geo table data
    isGeoTableRows(rows: TableRow[]): rows is GeoTableRow[] {
      return (rows as GeoTableRow[])[0].latitude !== undefined;
    },
  },

  watch: {
    selectedPoints() {
      this.paint();
    },

    bucketFeatures() {
      if (this.summary.baseline) {
        this.paint();
      }
    },

    filteredBucketFeatures() {
      if (this.summary.filtered) {
        this.paint();
      }
    },

    includedActive() {
      if (!this.includedActive) {
        this.clearSelectionRect();
      }
    },
  },

  mounted() {
    this.paint();
  },
});
</script>

<style>
.facet-card .group-header {
  font-family: inherit;
  font-size: 0.867rem;
  font-weight: 700;
  color: var(--color-text-second);
  background: var(--white);
  padding: 4px 8px 6px;
  position: relative;
  z-index: 1;
}

.facet-card .geofacet-container .selection-toggle {
  top: 55px;
}

.facet-card .geofacet-container .action-btn {
  position: relative;
  bottom: 37px;
  background: var(--white);
}

.facet-card .geofacet-container .action-btn:hover {
  color: var(--white);
  background-color: var(--gray-600);
  border-color: var(--gray-600);
}

.header-title {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.geofacet-container {
  bottom: 16px;
}

.geofacet,
.geofacet-container {
  position: relative;
  z-index: 0;
  height: 214px;
  width: 100%;
}

.facet-card .group-header .type-change-dropdown-wrapper {
  float: right;
  bottom: 20px;
}

.geofacet-container .type-change-dropdown-wrapper .dropdown-menu {
  z-index: 3;
}

.latlon .facets-root.highlighting-enabled {
  padding-left: 0px;
}
</style>
