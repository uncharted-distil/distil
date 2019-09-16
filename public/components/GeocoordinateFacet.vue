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
		<div
			class="selection-toggle"
			v-bind:class="{ active: isSelectionMode }"
			v-on:click="isSelectionMode = !isSelectionMode">
			<a
				class="selection-toggle-control"
				title="Select area"
				aria-label="Select area">
			<icon-base width="100%" height="100%"> <icon-crop-free /> </icon-base>
			</a>
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
import { GEOCOORDINATE_TYPE } from '../util/types';

import 'leaflet/dist/leaflet.css';

import helpers, { polygon, featureCollection } from '@turf/helpers';
import bbox from '@turf/bbox';

const CLOSE_BUTTON_CLASS = 'geo-close-button';
const CLOSE_ICON_CLASS = 'fa-times';

const PALETTE = [
	'rgba(0,0,0,0)',
	'#F4F8FB',
	'#E9F2F8',
	'#DEEBF5',
	'#D3E5F1',
	'#C8DFEE',
	'#BDD8EB',
	'#B2D2E8',
	'#A7CCE4',
	'#9CC5E1',
	'#91BFDE',
	'#86B8DB',
	'#7BB2D7',
	'#70ACD4',
	'#65A5D1',
	'#5A9FCE',
	'#4F99CA',
	'#4492C7',
	'#398CC4',
	'#2E86C1'
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
			isSelectionMode: false,
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

		// Computes the bounds of the summary data.
		bucketBounds(): helpers.BBox {
			return bbox(this.bucketFeatures);
		},

		// Creates a GeoJSON feature collection that can be passed directly to a Leaflet layer for rendering.
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
								]], { count: latBucket.count });
						features.push(feature);
					}
				});
			});

			return featureCollection(features);
		},

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
		headerLabel(): string {
			return GEOCOORDINATE_TYPE.toUpperCase();
		},
		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
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
		onMouseUp(event: MouseEvent) {
			if (this.currentRect) {
				this.setSelection(this.currentRect);
				this.currentRect = null;

				// disable drawing mode
				// this.map.dragging.enable();
				// this.map.on('click', this.clearSelection);
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
			console.log('selection mode', this.isSelectionMode);

			if (this.isSelectionMode) {

				this.clearSelectionRect();

				const offset = $(this.map.getContainer()).offset();
				console.log('offset', offset);

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
				// clearHighlight(this.$router);
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
				// dont push existing highlight
				return;
			}

			// TODO: support filtering multiple vars?
			// const fieldSpec = this.fieldSpecs[0];
			// const key = fieldSpec.type === SINGLE_FIELD ? fieldSpec.field : this.fieldHash(fieldSpec);

			// updateHighlight(this.$router, {
			// 	context: this.instanceName,
			// 	dataset: this.dataset,
			// 	key: key,
			// 	value: value
			// });
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
				const d = (maxVal - minVal) / PALETTE.length;
				const domain = PALETTE.map((val, index) => minVal + d * (index + 1));
				const scaleColors = scaleThreshold().range(PALETTE as any).domain(domain);

				// Render the heatmap buckets as a GeoJSON layer
				leaflet.geoJSON(this.bucketFeatures, {
					style: feature => {
						return {
							fillColor: scaleColors(feature.properties.count),
							weight: 2,
							opacity: 1,
							color: 'rgba(0,0,0,0)',
							dashArray: '3',
							fillOpacity: 0.7
						};
					}
				})
				.addTo(this.map);
			}
		}
	},

	watch: {
		bucketFeatures() {
			this.paint();
		},
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

</style>
