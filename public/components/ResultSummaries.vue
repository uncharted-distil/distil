<template>
	<div class='result-summaries'>
		<p class="nav-link font-weight-bold">Results<p>
		<div v-if="regressionEnabled" class="result-summaries-error">
			<div class="result-summaries-label">
				Error:
			</div>
			<div class="result-summaries-slider">
				<vue-slider ref="slider"
					:min="residualExtrema.min"
					:max="residualExtrema.max"
					:interval="interval"
					:value="initialValue"
					:formatter="formatter"
					:lazy="true"
					width=100%
					tooltip-dir="bottom"
					@callback="onSlide"/>
			</div>
		</div>
		<p class="nav-link font-weight-bold">Actual</p>
		<facets class="result-summaries-target"
			@histogram-click="onHistogramClick"
			@facet-click="onFacetClick"
			:groups="targetGroups"
			:highlights="highlights"></facets>
		<p class="nav-link font-weight-bold">Predictions by Model</p>
		<result-facets
			:regression="regressionEnabled"
			:result-extrema="resultExtrema"
			:residual-extrema="residualExtrema">
		</result-facets>
		<b-btn v-b-modal.export variant="outline-success" class="check-button">Export Pipeline</b-btn>
		<b-modal id="export" title="Export" @ok="onExport">
			<div class="check-message-container">
				<i class="fa fa-check-circle fa-3x check-icon"></i>
				<div>This action will export pipeline <b>{{activePipelineName}}</b> and terminate the session.</div>
			</div>
		</b-modal>
	</div>
</template>

<script lang="ts">

import ResultFacets from '../components/ResultFacets.vue';
import Facets from '../components/Facets.vue';
import { createGroups, Group } from '../util/facets';
import { overlayRouteEntry } from '../util/routes';
import { getPipelineById } from '../util/pipelines';
import { getTask } from '../util/pipelines';
import { isTarget, getVarFromTarget, getTargetCol } from '../util/data';
import { updateResultHighlights } from '../util/highlights';
import { VariableSummary, Extrema, Highlights, Range } from '../store/data/index';
import { NUMERICAL_FILTER, CATEGORICAL_FILTER } from '../util/filters';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { mutations as dataMutations } from '../store/data/module';
import { actions } from '../store/app/module';
import { Dictionary } from '../util/dict';
import vueSlider from 'vue-slider-component';
import Vue from 'vue';
import _ from 'lodash';
import 'font-awesome/css/font-awesome.css';

const DEFAULT_PERCENTILE = 0.25;
const NUM_STEPS = 100;
const RESULT_SUMMARY_CONTEXT = 'result_summary';

export default Vue.extend({
	name: 'result-summaries',

	components: {
		ResultFacets,
		Facets,
		vueSlider,
	},

	data() {
		return {
			formatter(arg) {
				return arg ? arg.toFixed(2) : '';
			}
		};
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		initialValue(): number[] {
			const min = routeGetters.getRouteResidualThresholdMin(this.$store);
			const max = routeGetters.getRouteResidualThresholdMax(this.$store);
			if (min === undefined || min === '' ||
				max === undefined || max === '') {
				if (!_.isNaN(this.defaultValue[0]) && !_.isNaN(this.defaultValue[1])) {
					this.updateThreshold(this.defaultValue[0], this.defaultValue[1]);
				}
				return this.defaultValue;
			}
			const nmin = _.toNumber(min);
			const nmax = _.toNumber(max);
			// NOTE: the slider component discards the values if they are
			// not within the extrema. We have to read the extrema here so
			// that the values are recomputed when the extrema is computed.
			const extrema = this.residualExtrema;
			if (nmin < extrema.min || nmax > extrema.max) {
				return [ NaN, NaN ];
			}
			return [
				nmin,
				nmax
			];
		},

		highlights(): Highlights {
			// find var marked as 'target' and set associated values as highlights
			const highlights = dataGetters.getHighlightedFeatureValues(this.$store);
			const facetHighlights = <Highlights>{
				root: _.cloneDeep(highlights.root),
				values: <Dictionary<string[]>>{}
			};
			_.forEach(highlights.values, (values, varName) => {
				if (isTarget(varName)) {
					facetHighlights.values[getVarFromTarget(varName)] = values;
				}
			});
			if (highlights.root && isTarget(highlights.root.key)) {
				facetHighlights.root.key = getVarFromTarget(highlights.root.key);
			}
			return facetHighlights;
		},

		range(): number {
			if (_.isNaN(this.residualExtrema.min) ||
				_.isNaN(this.residualExtrema.max)) {
				return NaN;
			}
			return this.residualExtrema.max - this.residualExtrema.min;
		},

		defaultValue(): number[] {
			return [
				-this.range/2 * DEFAULT_PERCENTILE,
				this.range/2 * DEFAULT_PERCENTILE
			];
		},

		interval(): number {
			const interval = this.range / NUM_STEPS;
			return interval;
		},

		targetSummary() : VariableSummary {
			const targetVariable = routeGetters.getRouteTargetVariable(this.$store);
			const varSummaries = dataGetters.getVariableSummaries(this.$store);
			return _.find(varSummaries, v => _.toLower(v.name) === _.toLower(targetVariable));
		},

		targetGroups(): Group[] {
			if (this.targetSummary) {
				return createGroups([ this.targetSummary ], false, false, this.resultExtrema);
			}
			return [];
		},

		resultsSummaries():  VariableSummary[] {
			return dataGetters.getResultsSummaries(this.$store);
		},

		resultExtrema(): Extrema {
			if (this.targetSummary || this.resultsSummaries) {
				let min = Infinity;
				let max = -Infinity;
				if (this.targetSummary) {
					min = Math.min(this.targetSummary.extrema.min, min);
					max = Math.max(this.targetSummary.extrema.max, max);
				}
				if (this.resultsSummaries) {
					this.resultsSummaries.forEach(summary => {
						min = Math.min(summary.extrema.min, min);
						max = Math.max(summary.extrema.max, max);
					});
				}
				return {
					min: min,
					max: max
				};
			}
			return null;
		},

		residualsSummaries():  VariableSummary[] {
			return this.regressionEnabled ? dataGetters.getResidualsSummaries(this.$store) : [];
		},

		residualExtrema(): Extrema {
			let extrema = NaN;
			this.residualsSummaries.forEach(summary => {
				extrema = Math.max(
					Math.abs(summary.extrema.min),
					Math.abs(summary.extrema.max));
			});
			return {
				min: -extrema,
				max: extrema
			};
		},

		regressionEnabled(): boolean {
			const targetVarName = routeGetters.getRouteTargetVariable(this.$store);
			const targetVar = _.find(dataGetters.getVariables(this.$store), v => _.toLower(v.name) === _.toLower(targetVarName));
			if (_.isEmpty(targetVar)) {
				return false;
			}
			const task = getTask(targetVar.type);
			return task.schemaName === 'regression';
		},

		activePipelineName(): string {
			const pipelineId = routeGetters.getRoutePipelineId(this.$store);
			const result = getPipelineById(this.$store.state.pipelineModule, pipelineId);
			return result ? result.name : '';
		}
	},

	methods: {
		onHistogramClick(context: string, key: string, value: Range) {
			if (key && value) {
				const colKey = getTargetCol(routeGetters.getRouteTargetVariable(this.$store));
				const filter = {
					name: colKey,
					type: NUMERICAL_FILTER,
					enabled: true,
					context: RESULT_SUMMARY_CONTEXT,
					min: value.from,
					max: value.to
				};
				updateResultHighlights(this, context, colKey, value, filter);
			} else {
				dataMutations.clearFeatureHighlights(this.$store);
			}
		},

		onFacetClick(context: string, key: string, value: string) {
			// clear exiting highlights
			if (key && value) {
				// extract the var name from the key
				const colKey = getTargetCol(routeGetters.getRouteTargetVariable(this.$store));
				const filter = {
					name: colKey,
					type: CATEGORICAL_FILTER,
					enabled: true,
					context: RESULT_SUMMARY_CONTEXT,
					categories: [value]
				};
				updateResultHighlights(this, context, colKey, value, filter);
			} else {
				dataMutations.clearFeatureHighlights(this.$store);
			}
		},

		updateThreshold(min: number, max: number) {
			const entry = overlayRouteEntry(this.$route, {
				residualThresholdMin: `${min}`,
				residualThresholdMax: `${max}`
			});
			this.$router.push(entry);
		},

		onSlide(value) {
			this.updateThreshold(value[0], value[1]);
		},

		onExport() {
			this.$router.replace('/');
			const pipelineId = routeGetters.getRoutePipelineId(this.$store);
			const result = getPipelineById(this.$store.state.pipelineModule, pipelineId);
			actions.exportPipeline(this.$store, {
				pipelineId: result.pipelineId,
				sessionId: this.$store.state.session.id
			});
		}
	}
});
</script>

<style>
.result-summaries {
	overflow-x: hidden;
	overflow-y: auto;
}

.result-summaries .facet-range,
.result-summaries .facets-facet-horizontal {
	height: 35px;
}

.result-summaries .facets-facet-base {
	overflow: visible;
}

.result-summaries-target {
	margin-bottom: 12px;
}

.result-summaries-target .facets-group {
	box-shadow: none;
}

.result-summaries-target .facets-facet-horizontal .facet-histogram-bar-highlighted {
	fill: #00C851;
}

.result-summaries-target .facets-facet-horizontal .facet-histogram-bar-highlighted:hover {
	fill: #007E33;
}

.result-summaries-target .facets-facet-horizontal .facet-histogram-bar-highlighted.select-highlight {
	fill: #007bff;
}

.result-summaries-target .facets-facet-vertical .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #00C851;
}

.result-summaries-target .facets-facet-horizontal .facet-range-filter {
	box-shadow: inset 0 0 0 1000px rgba(0, 225, 11, 0.15);
}

.result-summaries-error {
	display: flex;
	flex-direction: row;
	justify-content: flex-start;
	margin-bottom: 30px;
}

.result-summaries-label {
	display: flex;
	flex-basis: auto;
	margin-left: 10px;
	margin-right: 15px;
}

.result-summaries-slider {
	display: flex;
	flex-grow: 1;
}

.result-summaries-slider .vue-slider-component .vue-slider-process {
	background-color: #00C851;
}

.result-summaries-slider .vue-slider-component .vue-slider-tooltip {
	border: 1px solid #00C851;
	background-color: #00C851;
}

.facets-facet-vertical.select-highlight .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #007bff;
}

.check-message-container {
	display: flex;
	justify-content: flex-start;
	flex-direction: row;
	align-items: center;
}

.check-icon {
	display: flex;
	flex-shrink: 0;
	color:#00C851;
	padding-right: 15px;
}

.check-button {
	width: 60%;
	margin: 0 20%;
}
</style>
