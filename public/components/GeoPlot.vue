<template>
	<div class="geo-plot" v-bind:id="mapID"
		v-on:mousedown="onMouseDown"
		v-on:mouseup="onMouseUp"
		v-on:mousemove="onMouseMove"
		v-on:keydown="onKeyDown"
		v-on:keyup="onKeyUp"></div>
</template>

<script lang="ts">

import _ from 'lodash';
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
import { LATITUDE_TYPE, LONGITUDE_TYPE, REAL_VECTOR_TYPE } from '../util/types';

import 'leaflet/dist/leaflet.css';
import 'leaflet/dist/images/marker-icon.png';
import 'leaflet/dist/images/marker-icon-2x.png';
import 'leaflet/dist/images/marker-shadow.png';

const SINGLE_FIELD = 1;
const SPLIT_FIELD = 2;

interface GeoField {
	type: number;
	latField?: string;
	lngField?: string;
	field?: string;
}

interface LatLng {
	lat: number;
	lng: number;
	row: TableRow;
}

interface PointGroup {
	field: GeoField;
	points: LatLng[];
}

export default Vue.extend({
	name: 'geo-plot',

	props: {
		instanceName: String as () => string,
		dataItems: Array as () => any[],
		dataFields: Object as () => Dictionary<TableColumn>,

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
			selectedRect: null
		};
	},


	computed: {

		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		getTopVariables(): string[] {
			const rankings = datasetGetters.getVariableRankings(this.$store)[this.dataset];
			if (!rankings) {
				return [];
			}
			return _.map(rankings, (ranking, variable) => {
					return {
						variable: variable,
						ranking: ranking
					};
				})
				.sort((a, b) => {
					return b.ranking - a.ranking;
				})
				.map(r => r.variable);
		},

		mapID(): string {
			return `map-${this.instanceName}`;
		},

		fieldSpecs(): GeoField[] {

			const variables = datasetGetters.getVariables(this.$store);

			const matches = variables.filter(v => {
				return v.colType === LONGITUDE_TYPE ||
					v.colType === LATITUDE_TYPE ||
					v.colType === REAL_VECTOR_TYPE;
			});

			let lng = null;
			let lat = null;
			const fields = [];
			matches.forEach(match => {
				if (match.colType === LONGITUDE_TYPE) {
					lng = match.colName;
				}
				if (match.colType === LATITUDE_TYPE) {
					lat = match.colName;
				}
				if (match.colType === REAL_VECTOR_TYPE) {
					fields.push({
						type: SINGLE_FIELD,
						field: match.colName
					});
				}
				// TODO: currently we pair any two random lat / lngs, we should
				// eventually use the groupings functionality to let the user
				// group the two vars into a single point field.
				if (lng && lat) {
					fields.push({
						type: SPLIT_FIELD,
						lngField: lng,
						latField: lat
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

			this.fieldSpecs.forEach(fieldSpec => {
				const group = {
					field: fieldSpec,
					points: []
				};
				group.points = this.dataItems.map(item => {
					const lat = this.latValue(fieldSpec, item);
					const lng = this.lngValue(fieldSpec, item);
					if (lat !== undefined && lng !== undefined) {
						return {
							lng: lng,
							lat: lat,
							row: item
						};
					}
					return null;
				}).filter(p => !!p);
				groups.push(group);
			});

			return groups;
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
				html: '<i class="fa fa-times"></i>'
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
			// TODO: support filtering multiple vars?
			const fieldSpec = this.fieldSpecs[0];
			const key = fieldSpec.type === SINGLE_FIELD ? fieldSpec.field : this.fieldHash(fieldSpec);
			updateHighlightRoot(this.$router, {
				context: this.instanceName,
				dataset: this.dataset,
				key: key,
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
			// TODO: impl this
		},
		updateRoute() {
			const center = this.map.getCenter();
			const zoom  = this.map.getZoom();
			const arg = `${center.lng},${center.lat},${zoom}`;
			const entry = overlayRouteEntry(this.$route, {
				geo: arg,
			});
			this.$router.push(entry);
		},

		lngValue(fieldSpec: GeoField, row: TableRow): number {
			if (fieldSpec.type === SINGLE_FIELD) {
				return row[fieldSpec.field].Elements[0].Float;
			}
			return row[fieldSpec.lngField];
		},

		latValue(fieldSpec: GeoField, row: TableRow): number {
			if (fieldSpec.type === SINGLE_FIELD) {
				return row[fieldSpec.field].Elements[1].Float;
			}
			return row[fieldSpec.latField];
		},

		fieldHash(fieldSpec: GeoField): string {
			if (fieldSpec.type === SINGLE_FIELD) {
				return fieldSpec.field;
			}
			return fieldSpec.lngField + ':' + fieldSpec.latField;
		},

		paint() {
			if (this.map) {
				this.map.remove();
				this.map = null;
			}

			// NOTE: this component re-mounts on any change, so do everything in here
			this.map = leaflet.map(this.mapID, {
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

			// this.map.on('click', this.clearSelection);

			this.layer = leaflet.tileLayer('http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png');
			this.layer.addTo(this.map);

			this.markers = {};
			this.pointGroups.forEach(group => {
				const hash = this.fieldHash(group.field);
				const layer = leaflet.layerGroup([]);
				group.points.forEach(p => {
					const marker =  leaflet.marker(p);
					marker.bindTooltip(() => {
						const target = p.row[this.target];
						const values = [];
						const MAX_VALUES = 5;
						this.getTopVariables.forEach(v => {
							if (p.row[v] && values.length <= MAX_VALUES) {
								values.push(`<b>${_.capitalize(v)}:</b> ${p.row[v]}`);
							}
						});
						return [ `<b>${_.capitalize(target)}</b>` ].concat(values).join('<br>');
					});

					marker.on('mouseover', (e) => {
						$(marker._icon).css('filter', 'brightness(1.2)');
					});

					marker.on('mouseout', () => {
						$(marker._icon).css('filter', '');
					});

					layer.addLayer(marker);
				});
				layer.addTo(this.map);
				this.markers[hash] = layer;
			});

			this.drawHighlight();
			this.drawFilters();
		}

	},

	watch: {
		dataItems() {
			this.paint();
		}
	}

	mounted() {
		this.paint();
	}
});

</script>

<style>

.geo-plot {
	position: relative;
	height: 100%;
	width: 100%;
}

path.selected {
	stroke-width: 2;
	fill-opacity: 0.4;
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
