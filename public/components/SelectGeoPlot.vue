<template>

	<div class="select-geo-plot" id="map"></div>

</template>

<script lang="ts">

import leaflet from 'leaflet';
import Vue from 'vue';
import { getters as datasetGetters } from '../store/dataset/module';
import { Dictionary } from '../util/dict';
import { TableColumn, TableRow } from '../store/dataset/index';

import 'leaflet/dist/leaflet.css';
import 'leaflet/dist/images/marker-icon.png';
import 'leaflet/dist/images/marker-shadow.png';

export default Vue.extend({
	name: 'select-geo-plot',

	props: {
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
	},

	data() {
		return {
			map: null,
			layer: null,
			markers: null,
			internalLatLons: [],
			fieldName: 'lat_lon'
		};
	},

	mounted() {
		this.map = leaflet.map('map', {
			center: [30, 0],
			zoom: 2,
		});

		this.layer = leaflet.tileLayer('http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png');
		this.layer.addTo(this.map);

		this.markers = leaflet.layerGroup([]);
		this.markers.addTo(this.map);

		setInterval(() => {
			this.markers.clearLayers();
			this.internalLatLons.forEach(lonLat => {
				this.markers.addLayer(leaflet.marker(lonLat));
			});
		}, 1000);
	},

	computed: {

		fields(): Dictionary<TableColumn> {
			return this.includedActive ? datasetGetters.getIncludedTableDataFields(this.$store) : datasetGetters.getExcludedTableDataFields(this.$store);
		},

		items(): TableRow[] {
			return this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
		},

		lonLats(): number[][] {
			if (!this.items || !this.fields || !this.fields[this.fieldName]) {
				return [];
			}
			console.log('re-computing');

			return this.items.map(item => {
				return [
					item[this.fieldName].Elements[0].Float,
					item[this.fieldName].Elements[1].Float
				];
			});
		}
	},

	watch: {
		lonLats() {
			console.log('WATCH lonLats:', this.lonLats);
		}
	}
});

</script>

<style>

.select-geo-plot {
	position: relative;
	height: 100%;
	width: 100%;
}

</style>
