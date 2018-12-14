<template>

	<div class="select-timeseries-view" @mousemove="mouseMove" @wheel="scroll">
		<div class="timeseries-row-header">
			<div class="timeseries-var-col pad-top"><b>VARIABLES</b></div>
			<div class="timeseries-min-col pad-top"><b>MIN</b></div>
			<div class="timeseries-max-col pad-top"><b>MAX</b></div>
			<div class="timeseries-chart-axis">
				<template v-if="!!timeseriesExtrema">
					<svg ref="svg" class="axis"></svg>
				</template>
			</div>
		</div>
		<div v-for="item in items">
			<sparkline-row
				:timeseries-url="item[timeseriesField]"
				:timeseries-extrema="microExtrema"
				:margin="margin">
			</sparkline-row>
		</div>
		<div class="vertical-line"></div>
	</div>

</template>

<script lang="ts">

import * as d3 from 'd3';
import _ from 'lodash';
import $ from 'jquery';
import Vue from 'vue';
import SparklineRow from './SparklineRow';
import { Dictionary } from '../util/dict';
import { Filter } from '../util/filters';
import { RowSelection } from '../store/highlights/index';
import { TableRow, TableColumn, TimeseriesExtrema } from '../store/dataset/index';
import { getters as routeGetters } from '../store/route/module';
import { getters as datasetGetters } from '../store/dataset/module';

const TICK_SIZE = 8;
const SELECTED_TICK_SIZE = 18;

export default Vue.extend({
	name: 'select-timeseries-view',

	components: {
		SparklineRow
	},

	props: {
		margin: {
			type: Object as () => any,
			default: () => ({
				top: 2,
				right: 16,
				bottom: 2,
				left: 16
			})
		},
		instanceName: String as () => string,
		includedActive: Boolean as () => boolean
	},

	data() {
		return {
			macroScale: null,
			microScale: null,
			microRangeSelection: null,
			selectedMicroMin: null,
			selectedMicroMax: null
		};
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		items(): TableRow[] {
			return this.includedActive ? datasetGetters.getIncludedTableDataItems(this.$store) : datasetGetters.getExcludedTableDataItems(this.$store);
		},

		fields(): Dictionary<TableColumn> {
			return this.includedActive ? datasetGetters.getIncludedTableDataFields(this.$store) : datasetGetters.getExcludedTableDataFields(this.$store);
		},

		timeseriesField(): string {
			const fields = _.map(this.fields, (field, key) => {
					return {
						key: key,
						type: field.type
					};
				})
				.filter(field => field.type === 'timeseries')
				.map(field => field.key);
			return fields[0];
		},

		filters(): Filter[] {
			if (this.includedActive) {
				return this.invertFilters(routeGetters.getDecodedFilters(this.$store));
			}
			return routeGetters.getDecodedFilters(this.$store);
		},

		rowSelection(): RowSelection {
			return routeGetters.getDecodedRowSelection(this.$store);
		},

		timeseriesExtrema(): TimeseriesExtrema {
			const extrema = datasetGetters.getTimeseriesExtrema(this.$store);
			return extrema[this.dataset];
		},

		microExtrema(): TimeseriesExtrema {
			return {
				x: {
					min: this.microMin,
					max: this.microMax
				},
				y: {
					min: this.timeseriesExtrema ? this.timeseriesExtrema.y.min : 0,
					max: this.timeseriesExtrema ? this.timeseriesExtrema.y.max : 1
				}
			};
		},

		width(): number {
			const $svg = this.$refs.svg as any;
			const dims = $svg.getBoundingClientRect();
			return dims.width - this.margin.left - this.margin.right;
		},

		height(): number {
			const $svg = this.$refs.svg as any;
			const dims = $svg.getBoundingClientRect();
			return dims.height - this.margin.top - this.margin.bottom;
		},

		svg(): d3.Selection<SVGElement, {}, HTMLElement, any> {
			const $svg = this.$refs.svg as any;
			return d3.select($svg);
		},

		microMin(): number {
			if (this.selectedMicroMin) {
				return this.selectedMicroMin;
			}
			if (this.timeseriesExtrema) {
				return this.timeseriesExtrema.x.min;
			}
			return 0;
		},

		microMax(): number {
			if (this.selectedMicroMax) {
				return this.selectedMicroMax;
			}
			if (this.timeseriesExtrema) {
				return this.timeseriesExtrema.x.max;
			}
			return 1;
		}
	},

	methods: {
		invertFilters(filters: Filter[]): Filter[] {
			// TODO: invert filters
			return filters;
		},
		mouseMove(event) {
			const parentOffset = $('.select-timeseries-view').offset();
			const chartBounds = $('.timeseries-chart-axis').offset();
			const chartWidth = $('.timeseries-chart-axis').width();
			const chartScroll = $('.select-timeseries-view').parent().scrollTop();

			const relX = event.pageX - parentOffset.left;

			const chartLeft = chartBounds.left - parentOffset.left;
			if (relX >= chartLeft && relX <= chartLeft + chartWidth) {
				$('.vertical-line').show();
				$('.vertical-line').css({
					'left': relX,
					'top': chartScroll
				});
			} else {
				$('.vertical-line').hide();
			}
		},
		scroll(event) {
			const chartScroll = $('.select-timeseries-view').parent().scrollTop();
			$('.vertical-line').css('top', chartScroll);
		},
		injectSVG() {

			this.clearSVG();

			if (!this.timeseriesExtrema) {
				return;
			}

			this.macroScale = d3.scaleLinear()
				.domain([this.timeseriesExtrema.x.min, this.timeseriesExtrema.x.max])
				.range([0, this.width]);

			this.microScale = d3.scaleLinear()
				.domain([this.microMin, this.microMax])
				.range([0, this.width]);

			this.microRangeSelection = d3.axisTop(this.macroScale)
				.tickSize(SELECTED_TICK_SIZE)
				.tickValues([
					this.microMin,
					this.microMax
				]);

			this.svg.append('g')
				.attr('class', 'macro-axis')
				.attr('transform', `translate(${this.margin.left}, ${this.margin.top + SELECTED_TICK_SIZE + TICK_SIZE * 2})`)
				.call(d3.axisTop(this.macroScale));

			this.svg.append('g')
				.attr('class', 'micro-axis')
				.attr('transform', `translate(${this.margin.left}, ${-this.margin.bottom + this.height - TICK_SIZE * 2})`)
				.call(d3.axisBottom(this.microScale));

			this.svg.append('g')
				.attr('class', 'axis-selection')
				.attr('transform', `translate(${this.margin.left}, ${this.margin.top + SELECTED_TICK_SIZE + TICK_SIZE * 2})`)
				.call(this.microRangeSelection);

			this.attachHandlers();
		},
		attachHandlers() {

			const dragstarted = (d, index, elem) => {
				d3.select(elem[index]).raise().classed('active', true);
			};

			const dragged = (d, index, elem) => {
				const MIN = 0;
				const MAX = this.width;

				const px = Math.max(Math.min(d3.event.x, MAX), MIN);
				const x = this.macroScale.invert(px);
				if (index === 0) {
					this.selectedMicroMin = x;
				} else {
					this.selectedMicroMax = x;
				}
				d3.select(elem[index]).attr('transform', `translate(${px}, 0)`);
				d3.select(elem[index]).select('text').text(x.toFixed(2));

				this.svg.select('.micro-axis').remove();

				this.microScale = d3.scaleLinear()
					.domain([this.microMin, this.microMax])
					.range([0, this.width]);

				this.svg.append('g')
					.attr('class', 'micro-axis')
					.attr('transform', `translate(${this.margin.left}, ${-this.margin.bottom + this.height - TICK_SIZE * 2})`)
					.call(d3.axisBottom(this.microScale));

			};

			const dragended = (d, index, elem) => {
				d3.select(elem[index]).classed('active', false);
			};

			this.svg.selectAll('.axis-selection .tick')
				.call(d3.drag()
					.on('start', dragstarted)
					.on('drag', dragged)
					.on('end', dragended));
		},
		clearSVG() {
			this.svg.selectAll('*').remove();
		}
	},

	watch: {
		timeseriesExtrema: {
			handler() {
				Vue.nextTick(() => {
					this.injectSVG();
				});
			},
			deep: true
		}
	},

	mounted() {
		//this.injectSVG();
	}

});
</script>

<style>
svg.axis {
	position: relative;
	max-height: 64px;
	width: 100%;
}
.select-timeseries-view {
	position: relative;
	flex: 1;
}
.timeseries-row-header {
	height: 64px;
	line-height: 32px;
	border-bottom: 1px solid #999;
	padding: 0 8px;
}
.timeseries-var-col {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: 156px;
}
.timeseries-min-col {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: 48px;
}
.timeseries-max-col {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: 48px;
}
.timeseries-chart-col {
	float: left;
	position: relative;
	line-height: 32px;
	height: 32px;
	width: calc(100% - 276px);
}
.timeseries-chart-axis {
	float: left;
	position: relative;
	line-height: 32px;
	height: 64px;
	width: calc(100% - 276px);
}
.pad-top {
	padding-top: 32px;
}
.vertical-line {
	position: absolute;
	display: none;
	top: 0;
	left: 0;
	width: 1px;
	height: 100%;
	border-left: 1px solid #00c6e1;
	box-shadow: 0px 0px 5px #00c6e1;
	pointer-events: none;
}
.axis-selection {
}

.axis-selection .tick {
	cursor: pointer;
	stroke-width: 3;
}

.axis-selection path.domain {
	visibility: hidden;
}

</style>
