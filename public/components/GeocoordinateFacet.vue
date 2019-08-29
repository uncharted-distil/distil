<template>
	<div class="facet-card">
	<div class="geo-plot-container" v-bind:class="{ 'selection-mode': isSelectionMode }">
		<type-change-menu
			:dataset="dataset"
			:field="target">
		</type-change-menu>
		<div class="geo-plot" v-bind:id="mapID"></div>
	</div>

	</div>
</template>

<script lang="ts">
import _ from 'lodash';
import $ from 'jquery';
import leaflet from 'leaflet';
import Vue from 'vue';
import * as turf from '@turf/turf';
import IconBase from './icons/IconBase';
import IconCropFree from './icons/IconCropFree';
import { scaleThreshold } from 'd3';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { Dictionary } from '../util/dict';
import {
	TableColumn,
	TableRow,
	D3M_INDEX_FIELD,
	Highlight,
	RowSelection,
	VariableSummary,
	Bucket
} from '../store/dataset/index';
import { updateHighlight, clearHighlight } from '../util/highlights';
import {
	addRowSelection,
	removeRowSelection,
	isRowSelected
} from '../util/row';
import { LATITUDE_TYPE, LONGITUDE_TYPE, REAL_VECTOR_TYPE, GEOCOORDINATE_TYPE } from '../util/types';
import { DUMMY_GEODATA } from '../util/data';
import TypeChangeMenu from '../components/TypeChangeMenu';
import { SELECT_TARGET_ROUTE } from '../store/route';
import { createRouteEntry } from '../util/routes';
import { Filter, addFilterToRoute, removeFilterFromRoute, FilterParams, INCLUDE_FILTER, GEOCOORDINATE_FILTER } from '../util/filters';


import 'leaflet/dist/leaflet.css';
import 'leaflet/dist/images/marker-icon.png';
import 'leaflet/dist/images/marker-icon-2x.png';
import 'leaflet/dist/images/marker-shadow.png';

const SINGLE_FIELD = 1;
const SPLIT_FIELD = 2;
const CLOSE_BUTTON_CLASS = 'geo-close-button';
const CLOSE_ICON_CLASS = 'fa-times';

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
	name: 'geocoordinate-facet',

	components: {
		IconBase,
		IconCropFree,
		TypeChangeMenu
		},

	props: {
		summary: Object as () => VariableSummary,
	},

	data() {
		return {
			map: null,
			baseLayer: null,
			markers: null,
			closeButton: null,
			startingLatLng: null,
			currentRect: null,
			selectedRect: null,
			bounds: null,
			isSelectionMode: false,
			currentFilter: null
		};
		},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
			},
		datasummary(): any {
			const buckets = this.summary.baseline.buckets;
			const validPoints = buckets.filter((bucket) => {
				return bucket.count > 0;
			});

			const dataPoints = validPoints.reduce((acc, curr) => {
				const coordinates = curr.key.split(',');
				const lon = coordinates[0];
				const lat = coordinates[1];
				const points = Array(curr.count).fill({
					latitude: lat,
					longitude: lon
				});
				acc.push(...points);
				return acc;
				}, [] as any);

			return dataPoints;
		},
		instanceName(): string {
			return 'unique-map';
		},
		dataItems(): any {
			return this.datasummary;
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		targetSampleValues(): any[] {
			const summaries = routeGetters.getTargetVariableSummaries(this.$store);
			if (summaries.length > 0) {
				const summary = summaries[0];
				if (summary.baseline) {
					return summary.baseline.buckets;
				}
			}
			return [];
		},
		getTopVariables(): string[] {
			const variables = datasetGetters
				.getVariables(this.$store)
				.filter(v => v.datasetName === this.dataset);
			return variables
				.map(variable => ({
					variable: variable.colName,
					order: _.isNumber(variable.ranking)
						? variable.ranking
						: variable.importance
				}))
				.sort((a, b) => b.order - a.order)
				.map(r => r.variable);
		},

		mapID(): string {
			return `map-${this.instanceName}`;
		},

		fieldSpecs(): GeoField[] {
			const fields = [{
				latField: `${LATITUDE_TYPE}`,
				lngField: `${LONGITUDE_TYPE}`,
				type: 2
				}];
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
				group.points = this.dataItems
					.map(item => {
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
					})
					.filter(p => !!p);
				groups.push(group);
			});

			return groups;
		},

		highlight(): Highlight {
			return routeGetters.getDecodedHighlight(this.$store);
		},

		mapCenter(): number[] {
			return routeGetters.getGeoCenter(this.$store);
		},

		mapZoom(): number {
			return routeGetters.getGeoZoom(this.$store);
		},
		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},
		availableTargetVarsPage(): number {
			return routeGetters.getRouteAvailableTargetVarsPage(this.$store);
		},
	},

	methods: {
		variableSummaryToGeocoordinate(key: string, label: string, buckets: Bucket[]): any {
			const geocoordinates = buckets.map(b => [ _.parseInt(b.key), b.count ]);

			const summaries = [{
				label: label,
				key: key,
				geocoordinates: geocoordinates
			}];

			return summaries;
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

			if (this.isSelectionMode) {
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
				const bounds = [this.startingLatLng, latLng];
				this.currentRect.setBounds(bounds);
			}
		},
		onEsc() {
			if (this.currentRect) {
				this.clearSelectionRect();
				// disable drawing mode
				this.map.dragging.enable();
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
				html: `<i class='fa ${CLOSE_ICON_CLASS}'></i>`
			});
			this.closeButton = leaflet.marker([ne.lat, ne.lng], {
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
		createHighlight(value: {
			minX: number;
			maxX: number;
			minY: number;
			maxY: number;
		}) {
			if (
				this.highlight &&
				this.highlight.value.minX === value.minX &&
				this.highlight.value.maxX === value.maxX &&
				this.highlight.value.minY === value.minY &&
				this.highlight.value.maxY === value.maxY
			) {
				// dont push existing highlight
				return;
			}

			// TODO: support filtering multiple vars?
			const fieldSpec = this.fieldSpecs[0];
			const key =
				fieldSpec.type === SINGLE_FIELD
					? fieldSpec.field
					: this.fieldHash(fieldSpec);

			updateHighlight(this.$router, {
				context: this.instanceName,
				dataset: this.dataset,
				key: key,
				value: value
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
						[this.highlight.value.maxY, this.highlight.value.maxX]
					],
					{
						color: '#00c6e1',
						weight: 1,
						bubblingMouseEvents: false
					}
				);
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

		clear() {
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
			_.forIn(this.markers, markerLayer => {
				markerLayer.removeFrom(this.map);
			});
			this.markers = {};
			this.startingLatLng = null;
		},

		toggleSelection(event) {
			const marker = event.target;
			const row = marker.options.row;
			if (!isRowSelected(this.rowSelection, row[D3M_INDEX_FIELD])) {
				addRowSelection(
					this.$router,
					this.instanceName,
					this.rowSelection,
					row[D3M_INDEX_FIELD]
				);
			} else {
				removeRowSelection(
					this.$router,
					this.instanceName,
					this.rowSelection,
					row[D3M_INDEX_FIELD]
				);
			}
		},

		updateMarkerSelection(markers) {
			markers.forEach(marker => {
				const row = marker.options.row;
				const markerElem = marker.getElement();
				const isSelected = isRowSelected(
					this.rowSelection,
					row[D3M_INDEX_FIELD]
				);
				markerElem.classList.toggle('selected', isSelected);
			});
		},

		createPointGroup(points) {
			const count = points.length;
			const features = [];
			for (let i = 0; i < count; i++) {
				features.push(turf.point(points[i]));
			}
			return turf.featureCollection(features);
		},
		paint() {
			if (!this.map) {
				// NOTE: this component re-mounts on any change, so do everything in here
				this.map = leaflet.map(this.mapID, {
					center: [30, 0],
					zoom: 2,
					scrollWheelZoom: false
				});
				if (this.mapZoom) {
					this.map.setZoom(this.mapZoom, { animate: true });
				}
				if (this.mapCenter) {
					this.map.panTo(
						{
							lat: this.mapCenter[1],
							lng: this.mapCenter[0]
						},
						{ animate: true }
					);
				}

				this.baseLayer = leaflet.tileLayer(
					'http://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png'
				);
				this.baseLayer.addTo(this.map);
			}

			this.clear();

			const bounds = leaflet.latLngBounds();
			this.bounds = leaflet.latLngBounds();

			const pointLength = this.pointGroups.length;

			this.pointGroups.forEach(group => {
				const hash = this.fieldHash(group.field);

				const layer = leaflet.layerGroup([]);

				group.points.forEach(p => {
					const marker = leaflet.marker(p, { row: p.row });
					bounds.extend([p.lat, p.lng]);
					marker.bindTooltip(() => {
						const target = p.row[this.target];
						const values = [];
						const MAX_VALUES = 5;
						this.getTopVariables.forEach(v => {
							if (p.row[v] && values.length <= MAX_VALUES) {
								values.push(
									`<b>${_.capitalize(v)}:</b> ${p.row[v]}`
								);
							}
						});
						return [`<b>${_.capitalize(target)}</b>`]
							.concat(values)
							.join('<br>');
					});

					marker.on('click', this.toggleSelection);

					layer.addLayer(marker);
				});
				this.markers[hash] = layer;
				layer.on('add', () =>
					this.updateMarkerSelection(layer.getLayers())
				);
			});

			if (bounds.isValid()) {
				this.map.fitBounds(bounds);

				// create a turf BBox
				const bbox = turf.square(
					turf.square(
						bounds
							.toBBoxString()
							.split(',')
							.map(Number)
					)
				);

				// get distance of one side of bbox
				const bboxVertexCoordA = turf.point([bbox[0], bbox[1]]);
				const bboxVertexCoordB = turf.point([bbox[0], bbox[3]]);

				const sideDistance = turf.distance(
					bboxVertexCoordA,
					bboxVertexCoordB,
					{
						units: 'kilometers'
					}
				);

				// round to nearest 10
				const roundedDistance = Math.ceil(sideDistance / 10) * 10;

				const points = [];

				this.pointGroups.forEach(group => {
					if (group.points.length) {
						group.points.forEach(latLng => {
							points.push([
								Number(latLng.lng),
								Number(latLng.lat)
							]);
						});
					}
				});

				const multiPointFeature = this.createPointGroup(points);

				multiPointFeature.features.forEach(d => {
					d.properties.z = 1;
				});

				const squareGrid = turf.squareGrid(bbox, sideDistance / 10, {
					units: 'kilometers'
				});

//  @ts-ignore
				const count = turf.collect(squareGrid, multiPointFeature, 'z', 'z');

				const pallete = [
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

				const maxVal = _.maxBy(count.features, i => i.properties.z)
					.properties.z.length;
				const minVal = _.minBy(count.features, i => i.properties.z)
					.properties.z.length;

				const d = (maxVal - minVal) / pallete.length;
				const domain = pallete.map(
					(val, index) => minVal + d * (index + 1)
				);

//  @ts-ignore
const scaleColors = scaleThreshold().range(pallete).domain(domain);

				const gridLayer = leaflet
					.geoJSON(count, {
						style: feature => {
							return {
								fillColor: scaleColors(
									feature.properties.z.length
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
			this.drawHighlight();
			this.drawFilters();
		}
	},

	watch: {
		dataItems() {
			this.paint();
		},
		rowSelection() {
			const markers = _.map(this.markers, markerLayer =>
				markerLayer.getLayers()
			).reduce((prev, cur) => [...prev, ...cur], []);
			this.updateMarkerSelection(markers);
		}
	},

	mounted() {
		this.paint();
		console.log('this.map.getBounds', this.map.getBounds());

			// map action events


// 			export interface Filter {
// 	type: string;
// 	mode: string;
// 	key?: string;
// 	min?: number;
// 	max?: number;
// 	minX?: number;
// 	maxX?: number;
// 	minY?: number;
// 	maxY?: number;
// 	categories?: string[];
// 	d3mIndices?: string[];
// }

// export interface FilterParams {
// 	highlight: Filter;
// 	filters: Filter[];
// 	variables: string[];
// 	size?: number;
// }
		this.map.on('zoomend', () => {

				if (this.currentFilter) {
					removeFilterFromRoute(this.$router, this.currentFilter);
				}

				this.bounds = this.map.getBounds();
				const maxY = this.bounds.getNorthEast().lat.toString();
				const maxX = this.bounds.getNorthEast().lng.toString();
				const minY = this.bounds.getSouthWest().lat.toString();
				const minX = this.bounds.getSouthWest().lng.toString();

				const filter: Filter = {
					type: GEOCOORDINATE_FILTER,
					mode: INCLUDE_FILTER,
					minX: minX,
					maxX: maxX,
					minY: minY,
					maxY: maxY
				};

				this.currentFilter = filter;

				addFilterToRoute(this.$router, this.currentFilter);

			});

		this.map.on('moveend', () => {

				if (this.currentFilter) {
					removeFilterFromRoute(this.$router, this.currentFilter);
				}

				this.bounds = this.map.getBounds();
				const maxY = this.bounds.getNorthEast().lat;
				const maxX = this.bounds.getNorthEast().lng;
				const minY = this.bounds.getSouthWest().lat;
				const minX = this.bounds.getSouthWest().lng;

				const filter: Filter = {
					type: GEOCOORDINATE_FILTER,
					mode: INCLUDE_FILTER,
					minX: minX,
					maxX: maxX,
					minY: minY,
					maxY: maxY
				};

				this.currentFilter = filter;

				addFilterToRoute(this.$router, this.currentFilter);
			});
	}
});
</script>

<style>

.geo-plot-container, .geo-plot {
	position: relative;
	z-index: 0;
	height: 300px;
	width: 100%;
}


.geo-plot-container .type-change-dropdown-wrapper {
	float: right;
	z-index: 3;
	top: 22px;
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