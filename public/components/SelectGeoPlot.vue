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
import { getters as routeGetters } from '../store/route/module';
import { Dictionary } from '../util/dict';
import { TableColumn, TableRow } from '../store/dataset/index';
import { HighlightRoot } from '../store/highlights/index';
import { updateHighlightRoot, clearHighlightRoot } from '../util/highlights';
import { overlayRouteEntry } from '../util/routes';

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
				//this.map.off('click', this.clearSelection);
				this.map.dragging.disable();
			}
		},
		onMouseUp(event: MouseEvent) {
			if (this.currentRect) {
				this.setSelection(this.currentRect);
				this.currentRect = null;

				// disable drawing mode
				this.map.dragging.enable();
				//this.map.on('click', this.clearSelection);
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
			this.closeButton.on('click', () => {
				this.clearSelection();
				this.selectedRect.remove();
				this.selectedRect = null;
				this.closeButton.remove();
				this.closeButton = null;
			});
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
				value: value
			});
		},
		drawHighlight() {
			if (this.highlightRoot &&
				this.highlightRoot.value.minX !== undefined &&
				this.highlightRoot.value.maxX !== undefined &&
				this.highlightRoot.value.minY !== undefined &&
				this.highlightRoot.value.maxY !== undefined) {

				const rect = leaflet.rectangle([
					[
						this.highlightRoot.value.minY,
						this.highlightRoot.value.minX
					],
					[
						this.highlightRoot.value.maxY,
						this.highlightRoot.value.maxX
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
		drawFilters() {

		},
		updateRoute() {
			const center = this.map.getCenter();
			const zoom  = this.map.getZoom();
			const arg = `${center.lng},${center.lat},${zoom}`;
			const entry = overlayRouteEntry(this.$route, {
				geo: arg,
			});
			this.$router.push(entry);
		}
	},

	mounted() {
		// NOTE: this component re-mounts on any change, so do everything in here
		this.map = leaflet.map('map', {
			center: [30, 0],
			zoom: 2,
		});
		if (this.mapZoom) {
			this.map.setZoom(this.mapZoom, {animate: false});
		}
		if (this.mapCenter) {
			this.map.panTo({
				lat: this.mapCenter[1],
				lng: this.mapCenter[0]
			}, {animate: false});
		}

		this.map.on('moveend', event => {
			this.updateRoute();
		});
		this.map.on('zoomend', event => {
			this.updateRoute();
		});

		//this.map.on('click', this.clearSelection);

		this.layer = leaflet.tileLayer('http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png');
		this.layer.addTo(this.map);

		this.markers = leaflet.layerGroup([]);
		this.markers.addTo(this.map);

		this.lonLats.forEach(lonLat => {
			this.markers.addLayer(leaflet.marker(lonLat));
		});

		this.drawHighlight();
		this.drawFilters();
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
		},

		hasGeoField(): boolean {
			return !!datasetGetters.getVariablesMap(this.$store)[this.fieldName];
		},

		highlightRoot(): HighlightRoot {
			return routeGetters.getDecodedHighlightRoot(this.$store);
		},

		mapCenter(): number[] {
			return routeGetters.getGeoCenter(this.$store);
		},

		mapZoom(): number {
			return routeGetters.getGeoZoom(this.$store);
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
