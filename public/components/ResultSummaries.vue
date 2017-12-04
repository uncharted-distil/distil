<template>
	<div class='result-summaries'>
		<h6 class="nav-link">Results</h6>
		<div v-if="regressionEnabled" class="result-summaries-error">
			<div class="result-summaries-label">
				Error:
			</div>
			<div class="result-summaries-slider">
				<vue-slider ref="slider"
					:v-model="value"
					:min="minVal"
					:max="maxVal"
					:interval="interval"
					:value="value"
					:formatter="formatter"
					:lazy="true"
					width=100%
					tooltip-dir="bottom"
					@callback="onSlide"/>
			</div>
		</div>
		<h6 class="nav-link">Actual</h6>
		<facets class="result-summaries-target"
			:groups="targetSummaries"
			:highlights="highlights"></facets>
		<h6 class="nav-link">Predicted</h6>
		<result-facets
			v-on:activePipelineChange="onPipelineUpdate($event)"
			enable-group-collapse
			enable-facet-filtering
			:variables="variables"
			:dataset="dataset"
			:groups="targetSummaries"></result-facets>
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
import { createRouteEntryFromRoute } from '../util/routes';
import { getTask } from '../util/pipelines';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions } from '../store/app/module';
import { Dictionary, VariableSummary } from '../store/data/index';
import vueSlider from 'vue-slider-component';
import Vue from 'vue';
import _ from 'lodash';
import 'font-awesome/css/font-awesome.css';

const DEFAULT_PERCENTILE = 0.25;
const NUM_STEPS = 100;

export default Vue.extend({
	name: 'result-summaries',

	components: {
		ResultFacets,
		Facets,
		vueSlider,
	},

	data() {
		return {
			activePipelineName: null as string,
			activePipelineId: null as string,
			formatter(arg) {
				return arg.toFixed(2);
			}
		};
	},

	computed: {

		value: {
			set(value) {
				this.updateThreshold(value);
			},
			get(): number {
				const value = routeGetters.getRouteResidualThreshold(this.$store);
				if (value === undefined || value === '') {
					this.updateThreshold(this.defaultValue);
					return this.defaultValue;
				}
				return _.toNumber(value);
			}
		},

		highlights(): Dictionary<any> {
			return dataGetters.getHighlightedFeatureValues(this.$store);
		},

		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		minVal(): number {
			const resultItems = dataGetters.getResultDataItems(this.$store) as { [name: string]: any }[];
			if (!_.isEmpty(resultItems) && _.has(resultItems[0], 'error')) {
				const minErr = Math.abs(_.minBy(resultItems, r => Math.abs(r.error)).error);
				// round to closest 2 decimal places, otherwise interval computation makes the slider angry
				return Math.ceil(100 * minErr) / 100;
			}
			return 0.0;
		},

		maxVal(): number {
			const resultItems = dataGetters.getResultDataItems(this.$store) as { [name: string]: any }[]	;
			if (!_.isEmpty(resultItems) && _.has(resultItems[0], 'error')) {
				const maxErr = Math.abs(_.maxBy(resultItems, r => Math.abs(r.error)).error);
				// round to closest 2 decimal places, otherwise interval computation makes the slider angry
				return Math.ceil(100 * maxErr) / 100;
			}
			return 1.0;
		},

		range(): number {
			return this.maxVal - this.minVal;
		},

		defaultValue(): number {
			return this.minVal + (this.range * DEFAULT_PERCENTILE);
		},

		interval(): number {
			const interval = this.range / NUM_STEPS;
			return interval;
		},

		targetSummaries(): Group[] {
			// Get the current target variable and the summary associated with it
			const targetVariable = routeGetters.getRouteTargetVariable(this.$store);
			const varSummaries = dataGetters.getVariableSummaries(this.$store);
			const targetSummary = _.find(varSummaries, v => _.toLower(v.name) === _.toLower(targetVariable));
			// Create a facet for it - this will act as a basis of comparison for the result sets
			if (!_.isEmpty(targetSummary)) {
				return createGroups([targetSummary], false, false);
			}
			return [];
		},

		variables(): VariableSummary[] {
			return dataGetters.getResultsSummaries(this.$store);
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
	},

	methods: {
		updateThreshold(value: number) {
			const entry = createRouteEntryFromRoute(this.$route, {
				residualThreshold: value
			});
			this.$router.push(entry);
		},
		onSlide(value) {
			const entry = createRouteEntryFromRoute(this.$route, { residualThreshold: value });
			this.$router.replace(entry);
		},
		onExport() {
			this.$router.replace('/');
			actions.exportPipeline(this.$store, {
				pipelineId: this.activePipelineId,
				sessionId: this.$store.state.pipelineSession.id
			});
		},
		onPipelineUpdate(args) {
			this.activePipelineName = args.name;
			this.activePipelineId = args.id;
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
	height: 55px;
}

.result-summaries .facets-facet-horizontal-abbreviated {
	height: 40px;
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
