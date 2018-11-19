<template>
	<div class="select-geo-plot" id="map"
		v-on:mousedown="onMouseDown"
		v-on:mouseup="onMouseUp"
		v-on:mousemove="onMouseMove"
		v-on:keydown="onKeyDown"
		v-on:keyup="onKeyUp"></div>
</template>

<script lang="ts">

import $ from 'jquery';
import leaflet from 'leaflet';
import Vue from 'vue';
import { getters as datasetGetters } from '../store/dataset/module';
import { Dictionary } from '../util/dict';
import { TableColumn, TableRow } from '../store/dataset/index';
import { updateHighlightRoot, clearHighlightRoot } from '../util/highlights';

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
			rect: null,
			closeButton: null,
			ctrlDown: false,
			startingLatLng: null,
			currentRect: null,
			selectedRect: null,
			fieldName: 'lat_lon'
		};
	},

	methods: {
		onMouseDown(event: MouseEvent) {
			if (this.ctrlDown) {
				const offset = $(this.map.getContainer()).offset();
				this.startingLatLng = this.map.containerPointToLatLng({
					x: event.pageX - offset.left,
					y: event.pageY - offset.top
				});

				const bounds = [this.startingLatLng, this.startingLatLng];
				this.currentRect = leaflet.rectangle(bounds, {
					color: 'blue',
					weight: 1,
					bubblingMouseEvents: false
				});
				this.currentRect.on('click', e => {
					this.setSelection(e.target);
				});
				this.currentRect.addTo(this.map);
				this.map.off('click', this.clearSelection);
				this.map.dragging.disable();
			}
		},
		onMouseUp(event: MouseEvent) {
			if (this.currentRect) {
				this.setSelection(this.currentRect);
			}
			this.currentRect = null;
			this.map.dragging.enable();
			this.map.on('click', this.clearSelection);
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
		onKeyDown(event: KeyboardEvent) {
			const CTRL = 17;
			if (event.keyCode === CTRL) {
				this.ctrlDown = true;
			}
		},
		onKeyUp(event: KeyboardEvent) {
			const CTRL = 17;
			if (event.keyCode === CTRL) {
				this.ctrlDown = false;
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
				className: 'geo-close-button',
				iconSize: null,
				html:'<i class="fa fa-times"></i>'
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
				clearHighlightRoot(this.$router);
			}
			if (this.closeButton) {
				this.closeButton.remove();
			}
		},
		createHighlight(value: { minX: number, maxX: number, minY: number, maxY: number }) {
			updateHighlightRoot(this.$router, {
				context: this.instanceName,
				key: this.fieldName,
				value: {
					minX: value.minX,
					maxX: value.maxX,
					minY: value.minY,
					maxY: value.maxY
				}
			});
		}
	},

	mounted() {
		// NOTE: this component re-mounts on any change, so do everything in here
		this.map = leaflet.map('map', {
			center: [30, 0],
			zoom: 2,
		});
		this.map.on('click', this.clearSelection);

		this.layer = leaflet.tileLayer('http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png');
		this.layer.addTo(this.map);

		this.markers = leaflet.layerGroup([]);
		this.markers.addTo(this.map);

		this.lonLats.forEach(lonLat => {
			this.markers.addLayer(leaflet.marker(lonLat));
		});
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

			return this.items.map(item => {
				return [
					item[this.fieldName].Elements[0].Float,
					item[this.fieldName].Elements[1].Float
				];
			});
		}
	},
});

</script>

<style>

.select-geo-plot {
	position: relative;
	height: 100%;
	width: 100%;
}

path.selected {
	stroke-width: 2;
	fill-opacity: 0.4;
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
