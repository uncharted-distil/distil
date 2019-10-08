<template>
	<div class="facet-card">
		<div class="group-header">
			<span class="header-title">
				{{headerLabel}}
			</span>
			<i class="fa fa-globe"></i>
		<type-change-menu
			:dataset="dataset"
			:field="target">
		</type-change-menu>
		</div>
	<div class="geo-plot-container">
		<div
			class="geo-plot"
			v-bind:id="mapID"
			v-on:mousedown="onMouseDown"
			v-on:mouseup="onMouseUp"
			v-on:mousemove="onMouseMove">
		</div>
		<div v-if="isAvailableFeatures">
			<button
				class="action-btn btn btn-sm btn-outline-secondary ml-2 mr-1 mb-2"
				@click="selectFeature()">
				Add
			</button>
		</div>
		<div  v-if="isFeaturesToModel">
			<button
				class="action-btn btn btn-sm btn-outline-secondary ml-2 mr-1 mb-2"
				@click="removeFeature()">
				Remove
			</button>
		</div>
	</div>

	</div>
</template>

<script lang="ts">
import _ from 'lodash';
import $ from 'jquery';
import leaflet from 'leaflet';
import Vue from 'vue';
import IconBase from './icons/IconBase';
import IconCropFree from './icons/IconCropFree';
import { scaleThreshold } from 'd3';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { Dictionary } from '../util/dict';
import { VariableSummary, Bucket, Highlight } from '../store/dataset/index';
import TypeChangeMenu from '../components/TypeChangeMenu';
import { updateHighlight, clearHighlight } from '../util/highlights';
import { GEOCOORDINATE_TYPE, LATITUDE_TYPE, LONGITUDE_TYPE, REAL_VECTOR_TYPE } from '../util/types';
import { overlayRouteEntry } from '../util/routes';
import { Filter, removeFiltersByName } from '../util/filters';

import 'leaflet/dist/leaflet.css';

import helpers, { polygon, featureCollection } from '@turf/helpers';
import bbox from '@turf/bbox';

const SINGLE_FIELD = 1;
const SPLIT_FIELD = 2;
const CLOSE_BUTTON_CLASS = 'geo-close-button';
const CLOSE_ICON_CLASS = 'fa-times';
const LON_LAT_KEY = 'longitude:latitude';

interface GeoField {
	type: number;
	latField?: string;
	lngField?: string;
	field?: string;
}

const GEOCOORDINATE_LABEL = 'longitude';

const BLACK_PALLETE = ['#000000'];

const BLUE_PALETTE = [
	'rgba(0,0,0,0)',
	'#F0FBFD',
	'#E2F8FB',
	'#D4F5FA',
	'#C6F2F8',
	'#B8EFF6',
	'#AAECF5',
	'#9BE8F3',
	'#8DE5F1',
	'#7FE2F0',
	'#71DFEE',
	'#63DCEC',
	'#55D9EB',
	'#46D5E9',
	'#38D2E7',
	'#2ACFE6',
	'#1CCCE4',
	'#0EC9E2',
	'#00C6E1'
];

const PALETTE =   [
	'#E2E2E2',
	'#D5D5D5',
	'#C8C8C8',
	'#BCBCBC',
	'#AFAFAF',
	'#A3A3A3',
	'#969696',
	'#8A8A8A',
	'#7D7D7D',
	'#717171',
	'#646464',
	'#575757',
	'#4B4B4B',
	'#3E3E3E',
	'#323232',
	'#252525',
	'#191919',
	'#0C0C0C',
	'#000000'
];

export default Vue.extend({
	name: 'geocoordinate-facet',

	components: {
		TypeChangeMenu,
		IconBase,
		IconCropFree
	},

	props: {
		summary: Object as () => VariableSummary,
		isAvailableFeatures: Boolean as () => boolean,
		isFeaturesToModel: Boolean as () => boolean,
	},

	data() {
		return {
			map: null,
			baseLayer: null,
			bounds: null,
			closeButton: null,
			startingLatLng: null,
			currentRect: null,
			selectedRect: null,
			baseLineLayer: null,
			filteredLayer: null,
		};
	},
	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		instanceName(): string {
			return 'unique-map';
		},

		mapID(): string {
			return `map-${this.instanceName}`;
		},
		updated(){
			console.log('excludedBucketFeature', this.excludedBucketFeatures);

		},
		// Computes the bounds of the summary data.
		bucketBounds(): helpers.BBox {
			return bbox(this.bucketFeatures);
		},

		// Creates a GeoJSON feature collection that can be passed directly to a Leaflet layer for rendering.  The collection represents
		// the baseline bucket set for geocoordinate, and does not change as filters / highlights are introduced.
		bucketFeatures(): helpers.FeatureCollection {
			// compute the bucket size in degrees
			const buckets  = this.summary.baseline.buckets;
			const xSize = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
			const ySize = _.toNumber(buckets[0].buckets[1].key) - _.toNumber(buckets[0].buckets[0].key);

			// create a feature collection from the server-supplied bucket data
			const features: helpers.Feature[] = [];
			this.summary.baseline.buckets.forEach(lonBucket => {

				lonBucket.buckets.forEach(latBucket => {
					// Don't include features with a count of 0.
					if (latBucket.count > 0) {
						const xCoord = _.toNumber(lonBucket.key);
						const yCoord = _.toNumber(latBucket.key);
						const feature = polygon([[
									[xCoord, yCoord],
									[xCoord, yCoord + ySize],
									[xCoord + xSize, yCoord + ySize],
									[xCoord + xSize, yCoord],
									[xCoord, yCoord]
								]], { selected: false,
									count: latBucket.count });
						features.push(feature);
					}
				});
			});

			return featureCollection(features);

		},

		// Creates a GeoJSON feature collection that can be passed directly to a Leaflet layer for rendering.  The collection
		// represents the subset of buckets to be rendered based on the currently applied filters and highlights.
		filteredBucketFeatures(): helpers.FeatureCollection {
			// compute the bucket size in degrees

			if (this.summary.filtered) {
				const buckets  = this.summary.filtered.buckets;
				const xSize = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
				const ySize = _.toNumber(buckets[0].buckets[1].key) - _.toNumber(buckets[0].buckets[0].key);

				// create a feature collection from the server-supplied bucket data
				const features: helpers.Feature[] = [];
				this.summary.filtered.buckets.forEach(lonBucket => {
					lonBucket.buckets.forEach(latBucket => {
						// Don't include features with a count of 0.
						if (latBucket.count > 0) {
							const xCoord = _.toNumber(lonBucket.key);
							const yCoord = _.toNumber(latBucket.key);
							const feature = polygon([[
										[xCoord, yCoord],
										[xCoord, yCoord + ySize],
										[xCoord + xSize, yCoord + ySize],
										[xCoord + xSize, yCoord],
										[xCoord, yCoord]
									]], { selected: false,
										count: latBucket.count });
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
<<<<<<< HEAD
		excludedBucketFeatures(): helpers.FeatureCollection {
			console.log('this.excludedSummaries', this.excludedSummaries);

			if (this.excludedSummaries.filtered) {
				const buckets  = this.excludedSummaries.filtered.buckets;

				const xSize = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
				const ySize = _.toNumber(buckets[0].buckets[1].key) - _.toNumber(buckets[0].buckets[0].key);
				// create a feature collection from the server-supplied bucket data
				const features: helpers.Feature[] = [];
				this.excludedSummaries.filtered.buckets.forEach(lonBucket => {
					lonBucket.buckets.forEach(latBucket => {
						// Don't include features with a count of 0.
						if (latBucket.count > 0) {
							const xCoord = _.toNumber(lonBucket.key);
							const yCoord = _.toNumber(latBucket.key);
							const feature = polygon([[
										[xCoord, yCoord],
										[xCoord, yCoord + ySize],
										[xCoord + xSize, yCoord + ySize],
										[xCoord + xSize, yCoord],
										[xCoord, yCoord]
									]], { selected: false,
										count: latBucket.count });
							features.push(feature);
						}
					});
				});
=======
>>>>>>> filter-selection-logic-fixes

		// Returns the minimum non-zero bucket count value
		minCount(): number {
			return this.bucketFeatures.features.reduce((min, feature) =>
				feature.properties.count < min ? feature.properties.count : min, Number.MAX_SAFE_INTEGER);
		},

		// Returns the maximum bucket count value
		maxCount(): number {
			return this.bucketFeatures.features.reduce((max, feature) =>
				feature.properties.count > max ? feature.properties.count : max, Number.MIN_SAFE_INTEGER);
		},
		filteredMinCount(): number {
			return this.filteredBucketFeatures.features.reduce((min, feature) =>
				feature.properties.count < min ? feature.properties.count : min, Number.MAX_SAFE_INTEGER);
		},

		// Returns the maximum bucket count value
		filteredMaxCount(): number {
			return this.filteredBucketFeatures.features.reduce((max, feature) =>
				feature.properties.count > max ? feature.properties.count : max, Number.MIN_SAFE_INTEGER);
		},
		headerLabel(): string {
			return GEOCOORDINATE_TYPE.toUpperCase();
		},
		// is the display in included (blue) or excluded (black) mode
		includedActive(): boolean {
			return routeGetters.getRouteInclude(this.$store);
		},
		// is data currently being highlighted
		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
		},
	},
	methods: {
		selectFeature() {
			const training = routeGetters.getDecodedTrainingVariableNames(this.$store);
			const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
				training: training.concat([ 'Longitude' ]).join(',')
			});
			this.$router.push(entry);
		},
		removeFeature() {
			const training = routeGetters.getDecodedTrainingVariableNames(this.$store);
			training.splice(training.indexOf('Longitude'), 1);
			const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
				training: training.join(',')
			});
			this.$router.push(entry);
			removeFiltersByName(this.$router, 'Longitude');
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
				this.setSelection(this.currentRect);
				this.currentRect = null;
			}
		},
		onMouseMove(event: MouseEvent) {
			if (this.currentRect) {
				const offset = $(this.map.getContainer()).offset();
				const latLng = this.map.containerPointToLatLng({
					x: event.pageX - offset.left,
					y: event.pageY - offset.top
				});
				const bounds = [
					this.startingLatLng,
					latLng
				];
				this.currentRect.setBounds(bounds);
			}
		},
		onMouseDown(event: MouseEvent) {
			const mapEventTarget = event.target as HTMLElement;

			// check if mapEventTarget is the close button or icon
			if (mapEventTarget.classList.contains(CLOSE_BUTTON_CLASS) ||  mapEventTarget.classList.contains(CLOSE_ICON_CLASS)) {
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

				this.startingLatLng = this.map.containerPointToLatLng({
					x: event.pageX - offset.left,
					y: event.pageY - offset.top
				});

				const bounds = [this.startingLatLng, this.startingLatLng];
				this.currentRect = leaflet.rectangle(bounds, {
					color: '#00c6e1',
					weight: 1,
					bubblingMouseEvents: false
				});
				this.currentRect.on('click', e => {
					this.setSelection(e.target);
				});
				this.currentRect.addTo(this.map);

				// enable drawing mode
				// this.map.off('click', this.clearSelection);
				this.map.dragging.disable();
			}
		},
		setSelection(rect) {

			this.clearSelection();

			this.selectedRect = rect;
			const $selected = $(this.selectedRect._path);
			$selected.addClass('selected');

			const ne = rect.getBounds().getNorthEast();
			const sw = rect.getBounds().getSouthWest();
			const icon = leaflet.divIcon({
				className: CLOSE_BUTTON_CLASS,
				iconSize: null,
				html: `<i class="fa ${CLOSE_ICON_CLASS}"></i>`
			});
			this.closeButton = leaflet.marker([ ne.lat, ne.lng ], {
				icon: icon
			});
			this.closeButton.addTo(this.map);
			this.createHighlight({
				minX: sw.lng,
				maxX: ne.lng,
				minY: sw.lat,
				maxY: ne.lat
			});
		},
		clearSelection() {
			if (this.selectedRect) {
				$(this.selectedRect._path).removeClass('selected');
				clearHighlight(this.$router);
			}
			if (this.closeButton) {
				this.closeButton.remove();
			}
		},
		createHighlight(value: { minX: number, maxX: number, minY: number, maxY: number }) {

			if (this.highlight &&
				this.highlight.value.minX === value.minX &&
				this.highlight.value.maxX === value.maxX &&
				this.highlight.value.minY === value.minY &&
				this.highlight.value.maxY === value.maxY) {
				return;
			}

			updateHighlight(this.$router, {
				context: this.instanceName,
				dataset: this.dataset,
				key: LON_LAT_KEY,
				value: value
			});
		},
		drawHighlight() {
			if (this.highlight &&
				this.highlight.value.minX !== undefined &&
				this.highlight.value.maxX !== undefined &&
				this.highlight.value.minY !== undefined &&
				this.highlight.value.maxY !== undefined) {

				const rect = leaflet.rectangle([
					[
						this.highlight.value.minY,
						this.highlight.value.minX
					],
					[
						this.highlight.value.maxY,
						this.highlight.value.maxX
					]], {
					color: '#00c6e1',
					weight: 1,
					bubblingMouseEvents: false
				});
				rect.on('click', e => {
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
					doubleClickZoom: false
				});
				this.map.dragging.disable();

				this.baseLayer = leaflet.tileLayer(
					'http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png'
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

<<<<<<< HEAD
				if (!this.isAvailableFeatures && !this.isFeaturesToModel || !this.highlight && !this.hasFilters) {

					if (this.baseLineLayer) {
						this.baseLineLayer.removeFrom(this.map);
					}
=======
				// Check to see if we're showing included or excluded mode, whichi based on the user's current
				// tab setting.  In included mode we render all the currently included data in blue, in excluded
				//  mode we show only excluded data and render it in black.
				if (this.includedActive) {
					if (!this.highlight) {
						// if there's no highlight active render from the baseline (all) set of buckets.
>>>>>>> filter-selection-logic-fixes
						const d = (maxVal - minVal) / BLUE_PALETTE.length;
						const domain = BLUE_PALETTE.map((val, index) => minVal + d * (index + 1));
						const scaleColors = scaleThreshold().range(BLUE_PALETTE as any).domain(domain);

						// Render the heatmap buckets as a GeoJSON layer
						this.baseLineLayer = leaflet.geoJSON(this.bucketFeatures, {
							style: feature => {
								return {
									fillColor: scaleColors(feature.properties.count),
									weight: 0,
									opacity: 1,
									color: 'rgba(0,0,0,0)',
									dashArray: '3',
									fillOpacity: 0.7
								};
							}
						});
						this.baseLineLayer.addTo(this.map);
					} else {
						// there's a highlight active - render from the set of features returned in the filter portion of the
						// variable summary strucure
						const filteredMaxVal = this.filteredMaxCount;
						const filteredMinVal = this.filteredMinCount;
						const dVal = (filteredMaxVal - filteredMinVal) / BLUE_PALETTE.length;
						const filteredDomain = BLUE_PALETTE.map((val, index) => minVal + dVal * (index + 1));
						const filteredScaleColors = scaleThreshold().range(BLUE_PALETTE as any).domain(filteredDomain);

						this.filteredLayer = leaflet.geoJSON(this.filteredBucketFeatures, {
							style: feature => {
								return {
									fillColor: filteredScaleColors(feature.properties.count),
									weight: 0,
									opacity: 1,
									color: 'rgba(0,0,0,0)',
									dashArray: '3',
									fillOpacity: 0.7
								};
							}
						});
						this.filteredLayer.addTo(this.map);
					}
				} else {
					// Excluded mode is active - render visuals using a black pallette.
					// Any data we need to render is in the filter portion of variable summary structure.
					this.filteredLayer = leaflet.geoJSON(this.filteredBucketFeatures, {
						style: feature => {
							return {
<<<<<<< HEAD
								fillColor: filteredScaleColors(feature.properties.count),
								weight: 0,
								opacity: 1,
								color: 'rgba(0,0,0,0)',
								dashArray: '3',
								fillOpacity: 0.7
							};
						}
					});


					this.filteredLayer.addTo(this.map);

				}

				if (this.hasFilters) {
					if (this.excludedLayer) {
						this.excludedLayer.removeFrom(this.map);
					}

					this.excludedLayer = leaflet.geoJSON(this.excludedBucketFeatures, {
						style: feature => {
							return {
=======
>>>>>>> filter-selection-logic-fixes
								fillColor: BLACK_PALLETE[0],
								weight: 0,
								opacity: 1,
								color: 'rgba(0,0,0,0)',
								dashArray: '3',
								fillOpacity: 1
							};
						}
					});
					this.filteredLayer.addTo(this.map);
					this.clearSelectionRect();
				}
			}
		}
	},

	watch: {
		bucketFeatures() {
			if (this.summary.baseline) {
				this.paint();
			}
		},
		filteredBucketFeatures() {
			if (this.summary.filtered) {
				this.paint();
			}
		}
	},

	mounted() {
		this.paint();
	}
});
</script>

<style>

.facet-card .group-header {
	font-family: inherit;
    font-size: .867rem;
    font-weight: 700;
    color: rgba(0,0,0,.54);
	background: white;
	padding: 4px 8px 6px;
	position: relative;
    top: 30px;
    z-index: 1;
}

.facet-card .geo-plot-container .selection-toggle {
	top: 55px;
}

.facet-card .geo-plot-container .action-btn {
	position: relative;
    bottom: 37px;
    background: white;
}

.facet-card .geo-plot-container .action-btn:hover {
	color: #fff;
    background-color: #9e9e9e;
    border-color: #9e9e9e;
}

.header-title{
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.geo-plot-container {
	bottom: 16px;
}

.geo-plot-container, .geo-plot {
	height: 214px;
}

.facet-card .group-header .type-change-dropdown-wrapper {
	float: right;
	bottom: 20px;
}

.geo-plot-container .type-change-dropdown-wrapper .dropdown-menu {
	z-index: 3;
}

.geo-close-button {
	z-index: 3;
}

</style>
