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
        @type-change="onTypeChange"
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
    <div
      v-show="displayFooter"
      class="facet-footer-custom-html padding-right-12"
      ref="footer"
    />
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
import { datasetGetters, appActions } from "../../store";
import { getters as routeGetters } from "../../store/route/module";
import {
  TableRow,
  VariableSummary,
  Bucket,
  Extrema,
  Highlight,
  NUMERICAL_SUMMARY,
  RowSelection,
} from "../../store/dataset";
import TypeChangeMenu from "../TypeChangeMenu.vue";
import FacetNumerical from "./FacetNumerical.vue";
import { updateHighlight, clearHighlight } from "../../util/highlights";
import {
  LATITUDE_TYPE,
  LONGITUDE_TYPE,
  EXPAND_ACTION_TYPE,
  COLLAPSE_ACTION_TYPE,
  DISTIL_ROLES,
} from "../../util/types";
import { Feature, Activity, SubActivity } from "../../util/userEvents";

import "leaflet/dist/leaflet.css";
import helpers, { polygon, featureCollection, point } from "@turf/helpers";
import bbox from "@turf/bbox";
import booleanContains from "@turf/boolean-contains";
import { BLUE_PALETTE, BLACK_PALETTE } from "../../util/color";
import { EventList } from "../../util/events";

const CLOSE_BUTTON_CLASS = "geo-close-button";
const CLOSE_ICON_CLASS = "fa-times";

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
    datasetName: { type: String as () => string, default: null },
    include: { type: Boolean as () => boolean, default: true },
    html: [
      String as () => string,
      Object as () => any,
      Function as () => Function,
    ],
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
      return this.datasetName ?? routeGetters.getRouteDataset(this.$store);
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
      if (this.hasValidGeoHighlights) {
        const latLonRanges = this.getLatLonRanges();
        return {
          value: {
            from: this.calcBucketKey(latLonRanges.minY, LATITUDE_TYPE),
            to: this.calcBucketKey(latLonRanges.maxY, LATITUDE_TYPE),
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
      if (this.hasValidGeoHighlights) {
        const latLonRanges = this.getLatLonRanges();
        return {
          value: {
            from: this.calcBucketKey(latLonRanges.minX, LONGITUDE_TYPE),
            to: this.calcBucketKey(latLonRanges.maxX, LONGITUDE_TYPE),
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
      return `-facet-${this.target}-${this.dataset}`;
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
    },
    displayFooter(): boolean {
      return !!this.html && this.summary.distilRole != DISTIL_ROLES.Augmented;
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

    // is data currently being highlighted
    highlights(): Highlight[] {
      return routeGetters.getDecodedHighlights(this.$store);
    },

    hasValidGeoHighlights(): Boolean {
      return this.highlights.reduce((hasGeoHighlight, highlight) => {
        return (
          hasGeoHighlight ||
          (!!highlight &&
            !!highlight.value &&
            !!highlight.value.minX &&
            !!highlight.value.minY &&
            !!highlight.value.maxX &&
            !!highlight.value.maxY)
        );
      }, false);
    },

    selectedRows(): RowSelection {
      return routeGetters.getDecodedRowSelection(this.$store);
    },
    data(): TableRow[] {
      return this.include
        ? datasetGetters.getIncludedTableDataItems(this.$store)
        : datasetGetters.getExcludedTableDataItems(this.$store);
    },
    selectedPoints(): helpers.Point[] {
      if (this.selectedRows) {
        const tableItems = this.data;
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
    computeCustomHTML(): HTMLElement | null {
      // hack to get the custom html buttons showing up
      // changing this would mean to change how the instantiation of the facets works
      // right now they are wrapped by other components like
      // available-target-variables, available-training-variables, etc
      // those components inject HTML into the facets through their `html` function
      // we might want to change that in the future though
      if (this.html) {
        return _.isFunction(this.html)
          ? this.html({
              key: this.summary.key,
            })
          : this.html;
      }
      return null;
    },
  },

  methods: {
    getLatLonRanges(): {
      minX: number;
      maxX: number;
      minY: number;
      maxY: number;
    } {
      const latLonRanges = {
        minX: null,
        maxX: null,
        minY: null,
        maxY: null,
      };
      const keys = Object.keys(latLonRanges);
      this.highlights.forEach((highlight) => {
        keys.forEach((key) => {
          if (
            highlight &&
            highlight.value[key] &&
            (latLonRanges[key] === null ||
              (key.includes("min") &&
                latLonRanges[key] > highlight.value[key]) ||
              (key.includes("max") && latLonRanges[key] < highlight.value[key]))
          ) {
            latLonRanges[key] = highlight.value[key];
          }
        });
      });
      return latLonRanges;
    },

    calcBucketKey(value: number, type: string): string {
      const buckets =
        type === LONGITUDE_TYPE
          ? this.lonSummary.baseline.buckets
          : this.latSummary.baseline.buckets;
      const step = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
      return _.toString(value - (value % step));
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
      if (this.hasValidGeoHighlights) {
        this.highlights.forEach((highlight) => {
          if (
            highlight.value &&
            highlight.value.minX &&
            highlight.value.minY &&
            highlight.value.maxX &&
            highlight.value.maxY
          ) {
            const currentValue = highlight.value;
            const highlightValue = {
              minX: currentValue.minX,
              maxX: currentValue.maxX,
              minY: currentValue.minY,
              maxY: currentValue.maxY,
            };
            if (value === null) {
              clearHighlight(this.$router, key);
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
          }
        });
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
    onTypeChange() {
      this.$emit(EventList.VARIABLES.TYPE_CHANGE);
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
          color: this.include ? "#255DCC" : "black",
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
        minY: sw.lat,
        maxY: ne.lat,
        minX: sw.lng,
        maxX: ne.lng,
      });
    },

    clearSelection() {
      if (this.selectedRect) {
        const rectPath = (<any>this.selectedRect)._path;
        $(rectPath).removeClass("selected");
        clearHighlight(this.$router, this.summary.key);
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
        if (isExistingHighlight) {
          return;
        }
      }

      updateHighlight(this.$router, {
        context: this.instanceName,
        dataset: this.dataset,
        key: this.summary.key,
        value: value,
      });
    },

    drawHighlight() {
      if (this.highlights && this.highlights.length > 0) {
        this.highlights.forEach((highlight) => {
          if (
            highlight.value.minX !== undefined &&
            highlight.value.maxX !== undefined &&
            highlight.value.minY !== undefined &&
            highlight.value.maxY !== undefined
          ) {
            const rect = leaflet.rectangle(
              [
                [highlight.value.minY, highlight.value.minX],
                [highlight.value.maxY, highlight.value.maxX],
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
        });
      }
    },

    paint() {
      // NOTE: this component re-mounts on any change, so do everything in here
      if (!this.highlights || this.highlights.length < 1) {
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

        if (this.include) {
          if (
            (!this.highlights || this.highlights.length < 1) &&
            !this.hasFilters
          ) {
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
            const d = (maxVal - minVal) / BLUE_PALETTE.length;
            const filteredDomain = BLUE_PALETTE.map(
              (val, index) => minVal + d * (index + 1)
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

          const d = (maxVal - minVal) / BLUE_PALETTE.length;
          const filteredDomain = BLUE_PALETTE.map(
            (val, index) => minVal + d * (index + 1)
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
      return (rows as GeoTableRow[])[0]?.latitude !== undefined;
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

    include() {
      if (!this.include) {
        this.clearSelectionRect();
      }
    },
    computeCustomHTML() {
      if (this.displayFooter) {
        const footerRef = this.$refs["footer"] as HTMLElement;
        footerRef.innerHTML = "";
        footerRef.append(this.computeCustomHTML);
      }
    },
  },

  mounted() {
    this.paint();
    if (this.displayFooter) {
      const footerRef = this.$refs["footer"] as HTMLElement;
      footerRef.append(this.computeCustomHTML);
    }
  },
});
</script>

<style>
.facet-card {
  color: var(--color-text-second);
  background: var(--white);
}
.facet-card .group-header {
  font-family: inherit;
  font-size: 0.867rem;
  font-weight: 700;
  color: var(--color-text-second);
  background: var(--white);
  padding: 4px 12px 6px;
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
.facet-footer-custom-html {
  color: var(--color-text-second);
  background: var(--white);
  padding: 6px 8px 5px 5px;
  font-family: inherit;
  font-size: 0.867rem;
  font-weight: 700;
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
.padding-right-12 {
  padding-right: 12px;
}
.facet-card .group-header .type-change-dropdown-wrapper {
  float: right;
}

.geofacet-container .type-change-dropdown-wrapper .dropdown-menu {
  z-index: 3;
}

.latlon .facets-root.highlighting-enabled {
  padding-left: 0px;
}
</style>
