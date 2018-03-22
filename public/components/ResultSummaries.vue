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
			@facet-click="onCategoricalClick"
			@numerical-click="onNumericalClick"
			@range-change="onRangeChange"
			:groups="targetGroups"
			:highlights="highlights"></facets>
		<p class="nav-link font-weight-bold">Predictions by Model</p>
		<result-facets :regression="regressionEnabled"></result-facets>
		<b-btn v-b-modal.export variant="primary" class="check-button">Task 2: Export Model</b-btn>
		<b-modal id="export" title="Export" @ok="onExport">
			<div class="check-message-container">
				<i class="fa fa-check-circle fa-3x check-icon"></i>
				<div>This action will export pipeline <b>{{activePipelineName}}</b> and terminate the session.</div>
			</div>
		</b-modal>

		<b-modal id="export-failure-modal" ref="exportFailModal" title="Export Failed"
			cancel-disabled
			hide-header
			hide-footer>
			<div class="check-message-container">
				<i class="fa fa-exclamation-triangle fa-3x fail-icon"></i>
				<div><b>Export Failed:</b> {{exportFailureMsg}} </div>
				<b-btn class="mt-3 close-modal" variant="success" block @click="hideFailureModal">OK</b-btn>
			</div>
		</b-modal>
	</div>
</template>

<script lang="ts">

import ResultFacets from '../components/ResultFacets.vue';
import Facets from '../components/Facets.vue';
import { createGroups, Group } from '../util/facets';
import { overlayRouteEntry } from '../util/routes';
import { getPipelineById, getTask } from '../util/pipelines';
import { isTarget, getVarFromTarget, getTargetCol } from '../util/data';
import { getHighlights, updateHighlightRoot, clearHighlightRoot } from '../util/highlights';
import { VariableSummary, Extrema, Highlight } from '../store/data/index';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions as appActions, getters as appGetters } from '../store/app/module';
import { EXPORT_SUCCESS_ROUTE } from '../store/route/index';
import vueSlider from 'vue-slider-component';
import Vue from 'vue';
import _ from 'lodash';
import 'font-awesome/css/font-awesome.css';
import { PipelineInfo } from '../store/pipelines/index';
import { getters as pipelineGetters } from '../store/pipelines/module';

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

		highlights(): Highlight {
			// find var marked as 'target' and set associated values as highlights
			const highlights = _.cloneDeep(getHighlights(this.$store));
			if (_.isEmpty(highlights)) {
				return highlights;
			}
			_.forEach(highlights.values.samples, (values, varName) => {
				if (isTarget(varName)) {
					highlights.values.samples[getVarFromTarget(varName)] = values;
				}
			});
			if (highlights.root && isTarget(highlights.root.key)) {
				highlights.root.key = getVarFromTarget(highlights.root.key);
			}
			return highlights;
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
			const varSummaries = dataGetters.getResultSummaries(this.$store);
			return _.find(varSummaries, v => _.toLower(v.name) === _.toLower(this.target));
		},

		targetGroups(): Group[] {
			if (this.targetSummary) {
				const target = createGroups([ this.targetSummary ]);
				if (this.highlights.root) {
					const group = target[0];
					if (group.key === this.highlights.root.key) {
						group.facets.forEach(facet => {
							facet.filterable = true;
						});
					}
				}
				return target;
			}
			return [];
		},

		predictedSummaries(): VariableSummary[] {
			return dataGetters.getPredictedSummaries(this.$store);
		},

		residualsSummaries(): VariableSummary[] {
			return this.regressionEnabled ? dataGetters.getResidualsSummaries(this.$store) : [];
		},

		residualExtrema(): Extrema {
			return dataGetters.getResidualExtrema(this.$store);
		},

		regressionEnabled(): boolean {
			const targetVar = _.find(dataGetters.getVariables(this.$store), v => _.toLower(v.name) === _.toLower(this.target));
			if (_.isEmpty(targetVar)) {
				return false;
			}
			const task = getTask(targetVar.type);
			return task.schemaName === 'regression';
		},

		pipelineId(): string {
			return routeGetters.getRoutePipelineId(this.$store);
		},

		activePipeline(): PipelineInfo {
			return getPipelineById(this.$store.state.pipelineModule, this.pipelineId);
		},

		activePipelineName(): string {
			return this.activePipeline ? this.activePipeline.name : '';
		},

		sessionId(): string {
			return pipelineGetters.getPipelineSessionID(this.$store);
		},

		instanceName(): string {
			return 'groundTruth';
		},

		isAborted(): boolean {
			return appGetters.isAborted(this.$store);
		}
	},

	methods: {

		onCategoricalClick(context: string, key: string, value: string) {
			if (key && value) {
				// extract the var name from the key
				const colKey = getTargetCol(this.target);
				updateHighlightRoot(this, {
					context: context,
					key: colKey,
					value: value
				});
			} else {
				clearHighlightRoot(this);
			}
		},

		onNumericalClick(key: string) {
			if (!this.highlights.root || this.highlights.root.key !== key) {
				const colKey = getTargetCol(this.target);
				updateHighlightRoot(this, {
					context: this.instanceName,
					key: colKey,
					value: null
				});
			}
		},

		onRangeChange(context: string, key: string, value: { from: { label: string[] }, to: { label: string[] } }) {
			const colKey = getTargetCol(this.target);
			updateHighlightRoot(this, {
				context: context,
				key: colKey,
				value: value
			});
			this.$emit('range-change', key, value);
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
			appActions.exportPipeline(this.$store, {
				pipelineId: this.activePipeline.pipelineId,
				sessionId: this.sessionId
			}).then(err => {
				if (this.isAborted) {
					// the export was successful
					this.$router.replace(EXPORT_SUCCESS_ROUTE);
				} else {
					if (err) {
						// failed, this is because the wrong variable was selected
						const modal = this.$refs.exportFailModal as any;
						this.exportFailureMsg = err.message;
						modal.show();
					}
				}
			});
		},

		hideFailureModal() {
			const modal = this.$refs.exportFailModal as any;
			modal.hide();
		}
	}
});
</script>

<style>
.result-summaries {
	overflow-x: hidden;
	overflow-y: auto;
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

.fail-icon {
	display: flex;
	flex-shrink: 0;
	color:#dc3545;
	padding-right: 15px;
}

.check-button {
	width: 60%;
	margin: 0 20%;
}
</style>
