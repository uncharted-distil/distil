<template>
	<div class='result-summaries'>
		<p class="nav-link font-weight-bold">Results<p>
		<div v-if="regressionEnabled" class="result-summaries-error">
			<div class="result-summaries-label">
				Error:
			</div>
			<div class="result-summaries-slider" v-if="showSlider">
				<div class="error-center-line"></div>
				<div class="error-center-label">0</div>
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
		<p class="nav-link font-weight-bold">Predictions by Model</p>
		<result-facets
			:regression="regressionEnabled">
			</result-facets>
		<b-btn v-b-modal.export variant="primary" class="check-button">Task 2: Export Model</b-btn>

		<b-modal id="export" title="Export" @ok="onExport">
			<div class="check-message-container">
				<i class="fa fa-check-circle fa-3x check-icon"></i>
				<div>This action will export solution <b>{{activeSolutionName}}</b> and terminate the session.</div>
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

import _ from 'lodash';
import ResultFacets from '../components/ResultFacets.vue';
import Facets from '../components/Facets.vue';
import { overlayRouteEntry } from '../util/routes';
import { getSolutionById, getTask } from '../util/solutions';
import { Extrema } from '../store/data/index';
import { getters as dataGetters} from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';
import { actions as appActions, getters as appGetters } from '../store/app/module';
import { EXPORT_SUCCESS_ROUTE } from '../store/route/index';
import vueSlider from 'vue-slider-component';
import Vue from 'vue';
import 'font-awesome/css/font-awesome.css';
import { SolutionInfo } from '../store/solutions/index';

const DEFAULT_PERCENTILE = 0.25;
const NUM_STEPS = 100;
const ERROR_DECIMALS = 0;

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
			},
			exportFailureMsg: '',
			symmetricSlider: true
		};
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},

		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},

		showSlider(): boolean {
			return !_.isNaN(this.interval);
		},

		initialValue(): number[] {
			const min = routeGetters.getRouteResidualThresholdMin(this.$store);
			const max = routeGetters.getRouteResidualThresholdMax(this.$store);
			if (min === undefined || max === undefined) {
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
			return this.range / NUM_STEPS;
		},

		residualExtrema(): Extrema {
			const extrema = dataGetters.getResidualExtrema(this.$store);
			if (!extrema) {
				return extrema;
			}
			return {
				min: _.round(extrema.min, ERROR_DECIMALS),
				max: _.round(extrema.max, ERROR_DECIMALS)
			};
		},

		regressionEnabled(): boolean {
			const targetVar = _.find(dataGetters.getVariables(this.$store), v => _.toLower(v.name) === _.toLower(this.target));
			if (_.isEmpty(targetVar)) {
				return false;
			}
			const task = getTask(targetVar.type);
			return task.schemaName === 'regression';
		},

		solutionId(): string {
			return routeGetters.getRouteSolutionId(this.$store);
		},

		activeSolution(): SolutionInfo {
			return getSolutionById(this.$store.state.solutionModule, this.solutionId);
		},

		activeSolutionName(): string {
			return this.activeSolution ? this.activeSolution.name : '';
		},

		instanceName(): string {
			return 'groundTruth';
		},

		isAborted(): boolean {
			return appGetters.isAborted(this.$store);
		}
	},

	methods: {

		updateThreshold(min: number, max: number) {
			const entry = overlayRouteEntry(this.$route, {
				residualThresholdMin: `${min}`,
				residualThresholdMax: `${max}`
			});
			this.$router.push(entry);
		},

		onSlide(value) {
			// if (this.symmetricSlider) {
			// 	console.log('before:', this.$refs.slider.getValue();
			// 	console.log('after:', value);
			// 	//this.$refs.slider.setValue([ ] true);
			// }
			this.updateThreshold(value[0], value[1]);
		},

		onExport() {
			appActions.exportSolution(this.$store, {
				solutionId: this.activeSolution.solutionId
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
	position: relative;
}

.result-summaries-slider .vue-slider-component .vue-slider-process {
	background-color: #9e9e9e;
}

.result-summaries-slider .vue-slider-component .vue-slider-tooltip {
	border: 1px solid #9e9e9e;
	background-color: #9e9e9e;
}

.result-summaries-slider .vue-slider-component .vue-slider-piecewise {
	background-color: #ee0701;
}


.facets-facet-vertical.select-highlight .facet-bar-selected {
	box-shadow: inset 0 0 0 1000px #007bff;
}

.error-center-line {
	position:absolute;
	left: 50%;
	height: 22px;
	width: 1px;
	background-color: #333;
}

.error-center-label {
	position:absolute;
	top: 22px;
	width: 100%;
	color: #333;
	text-align: center;
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
	color:#ee0701;
	padding-right: 15px;
}

.check-button {
	width: 60%;
	margin: 0 20%;
}
</style>
