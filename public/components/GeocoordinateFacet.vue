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
import 'leaflet/dist/images/marker-icon.png';
import 'leaflet/dist/images/marker-icon-2x.png';
import 'leaflet/dist/images/marker-shadow.png';

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
			startingLatLng: null,
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

		// Computes the bounds of the summary data.
		featureBounds(): helpers.BBox {
			return bbox(this.features);
		},

		// Creates a GeoJSON feature collection that can be passed directly to a Leaflet layer for rendering.
		features(): helpers.FeatureCollection {
			// compute the bucket size in degrees
			const buckets  = this.summary.baseline.buckets;
			const xSize = _.toNumber(buckets[1].key) - _.toNumber(buckets[0].key);
			const ySize = _.toNumber(buckets[0].buckets[1].key) - _.toNumber(buckets[0].buckets[0].key);

			const features: helpers.Feature[] = [];
			this.summary.baseline.buckets.forEach(lonBucket => {
				lonBucket.buckets.forEach(latBucket => {

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
				});
			});
			return featureCollection(features);
		},

		instanceName(): string {
			return 'unique-map';
		},

		mapID(): string {
			return `map-${this.instanceName}`;
		},
	},

	methods: {
		paint() {
			if (!this.map) {
				// NOTE: this component re-mounts on any change, so do everything in here
				this.map = leaflet.map(this.mapID, {
					center: [30, 0],
					zoom: 2,
					// scrollWheelZoom: false,
					zoomControl: false,
				});
				// this.map.dragging.disable();

				this.baseLayer = leaflet.tileLayer(
					'http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png'
				);
				this.baseLayer.addTo(this.map);
			}

			const bounds = this.featureBounds;
			const northEast = leaflet.latLng(bounds[3], bounds[2]);
			const southWest = leaflet.latLng(bounds[1], bounds[0]);
			this.bounds = leaflet.latLngBounds(northEast, southWest);

			if (this.bounds.isValid()) {
				this.map.fitBounds(this.bounds);

				const maxVal = 1000;
				const minVal = 0;
				const d = (maxVal - minVal) / PALETTE.length;
				const domain = PALETTE.map(
					(val, index) => minVal + d * (index + 1)
				);
				const scaleColors = scaleThreshold().range(PALETTE as any).domain(domain);

				leaflet.geoJSON(this.features, {
					style: feature => {
						return {
							fillColor: scaleColors(
								feature.properties.count
							),
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
		dataItems() {
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

.geo-plot .leaflet-marker-icon:hover {
	filter: brightness(1.2);
}

.geo-plot .leaflet-marker-icon.selected {
	filter: hue-rotate(150deg);
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
