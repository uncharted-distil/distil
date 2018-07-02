<template>
	<div class='result-summaries'>
		<p class="nav-link font-weight-bold">Results<p>
		<div v-if="regressionEnabled" class="result-summaries-error">
			<error-threshold-slider></error-threshold-slider>
		</div>
		<p class="nav-link font-weight-bold">Predictions by Model</p>
		<result-facets
			:regression="regressionEnabled">
			</result-facets>
		<b-btn v-b-modal.export variant="primary" class="check-button" v-if="isExport">Task 2: Export Model</b-btn>

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
import ErrorThresholdSlider from '../components/ErrorThresholdSlider.vue';
import { getSolutionById, getTask } from '../util/solutions';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { actions as appActions, getters as appGetters } from '../store/app/module';
import { EXPORT_SUCCESS_ROUTE } from '../store/route/index';
import vueSlider from 'vue-slider-component';
import Vue from 'vue';
import 'font-awesome/css/font-awesome.css';
import { Solution } from '../store/solutions/index';

export default Vue.extend({
	name: 'result-summaries',

	components: {
		ResultFacets,
		Facets,
		ErrorThresholdSlider,
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

		regressionEnabled(): boolean {
			const targetVar = _.find(datasetGetters.getVariables(this.$store), v => _.toLower(v.key) === _.toLower(this.target));
			if (_.isEmpty(targetVar)) {
				return false;
			}
			const task = getTask(targetVar.type);
			return task.schemaName === 'regression';
		},

		solutionId(): string {
			return routeGetters.getRouteSolutionId(this.$store);
		},

		activeSolution(): Solution {
			return getSolutionById(this.$store.state.solutionModule, this.solutionId);
		},

		isExport(): boolean {
			return !appGetters.isDiscovery(this.$store);
		},

		activeSolutionName(): string {
			return this.activeSolution ? this.activeSolution.feature : '';
		},

		instanceName(): string {
			return 'groundTruth';
		},

		isAborted(): boolean {
			return appGetters.isAborted(this.$store);
		}
	},

	methods: {

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
	color:#ee0701;
	padding-right: 15px;
}

.check-button {
	width: 60%;
	margin: 0 20%;
}
</style>
