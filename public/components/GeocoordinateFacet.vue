<template>
	<div class="facet-card">
		<div class="group-header">
			<span class="header-title">
				GEOCOORDINATE
			</span>
			<i class="fa fa-globe"></i>
		<type-change-menu
			:dataset="dataset"
			:field="target">
		</type-change-menu>
		</div>
	<div class="geo-plot-container">
		<div class="geo-plot" v-bind:id="mapID"></div>
	</div>

	</div>
</template>

<script lang="ts">
import _ from 'lodash';
import $ from 'jquery';
import leaflet from 'leaflet';
import Vue from 'vue';
import { scaleThreshold } from 'd3';
import { actions as datasetActions, getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { Dictionary } from '../util/dict';
import { VariableSummary, Bucket } from '../store/dataset/index';
import TypeChangeMenu from '../components/TypeChangeMenu';

import 'leaflet/dist/leaflet.css';

import helpers, { polygon, featureCollection } from '@turf/helpers';
import bbox from '@turf/bbox';

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
		TypeChangeMenu
	},

	props: {
		summary: Object as () => VariableSummary,
	},

	data() {
		return {
			map: null,
			baseLayer: null,
			bounds: null,
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
	},

	methods: {
		paint() {
			// NOTE: this component re-mounts on any change, so do everything in here

			// Lazy map instantiation with a default zoom position
			if (!this.map) {
				this.map = leaflet.map(this.mapID, {
					center: [30, 0],
					zoom: 2,
					scrollWheelZoom: false,
					zoomControl: false,
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
    font-size: 13.872px;
    font-size: .867rem;
    font-weight: 700;
    text-transform: uppercase;
    color: rgba(0,0,0,.54);
	background: white;
	padding: 4px 8px 6px;
	height: 50px;
}

.header-title{
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.geo-plot-container, .geo-plot {
	position: relative;
	z-index: 0;
	height: 300px;
	width: 100%;
}

.facet-card .group-header .type-change-dropdown-wrapper {
	float: right;
	bottom: 20px;
}

.geo-plot-container .type-change-dropdown-wrapper .dropdown-menu {
	z-index: 3;
}

.geo-plot-container, .geo-plot {
	position: relative;
	z-index: 0;
	height: 300px;
	width: 100%;
}
</style>
