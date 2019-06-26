<template>

	<div class="container-fluid d-flex flex-column h-100 select-view">
		<div class="row flex-0-nav"></div>
		
		<div class="row flex-shrink-0 align-items-center bg-white">
			<div class="col-4 offset-md-1">
				<h5 class="header-label">Select Feature to Predict</h5>
			</div>
			<div class="col-2 offset-md-4">
				<b-button class="grouping-button" variant="primary" @click="onGroupingClick">
					Create Variable Grouping
				</b-button>
			</div>
		</div>
		<div class="row justify-content-center pb-3 h-100">
			<div class="col-12 col-md-10 flex-column d-flex h-100">
				<available-target-variables>
				</available-target-variables>
			</div>
		</div>
		<timeseries-analysis-modal
			:show="showTimeseriesChoice"
			@close="onTimeseriesChoice">
		</timeseries-analysis-modal>
	</div>

</template>

<script lang="ts">

import Vue from 'vue';
import { Variable } from '../store/dataset/index';
import TimeseriesAnalysisModal from '../components/TimeseriesAnalysisModal';
import AvailableTargetVariables from '../components/AvailableTargetVariables';
import { actions as viewActions } from '../store/view/module';
import { getters as datasetGetters } from '../store/dataset/module';
import { getters as routeGetters } from '../store/route/module';
import { gotoVariableGrouping } from '../util/nav';
import { isTimeType } from '../util/types';
import { overlayRouteEntry } from '../util/routes';

export default Vue.extend({
	name: 'select-target-view',

	data() {
		return {
			showTimeseriesChoice: false,
			haveVariablesLoaded: false
		};
	},

	components: {
		AvailableTargetVariables,
		TimeseriesAnalysisModal
	},

	computed: {

		availableTargetVarsPage(): number {
			return routeGetters.getRouteAvailableTargetVarsPage(this.$store);
		},
		variables(): Variable[] {
			return datasetGetters.getVariables(this.$store);
		},
		timeseriesAnalysis(): string {
			return routeGetters.getRouteTimeseriesAnalysis(this.$store);
		},
		hasTimeVariable(): boolean {
			return this.variables.filter(v => isTimeType(v.colType)).length  > 0;
		}
	},

	watch: {
		availableTargetVarsPage() {
			viewActions.fetchSelectTargetData(this.$store, false);
		},
		timeseriesAnalysis() {
			viewActions.fetchSelectTargetData(this.$store, true);
		},
		variables() {
			if (this.variables.length > 0 && !this.timeseriesAnalysis && !this.haveVariablesLoaded) {
				if (this.hasTimeVariable) {
					this.showTimeseriesChoice = true;
				}
				this.haveVariablesLoaded = true;
			}
		}
	},

	beforeMount() {
		viewActions.fetchSelectTargetData(this.$store, false);
	},

	methods: {
		onTimeseriesChoice(event: any) {
			if (event) {
				const entry = overlayRouteEntry(routeGetters.getRoute(this.$store), {
					timeseriesAnalysis: event.col
				});
				this.$router.push(entry);
			}
			this.showTimeseriesChoice = false;
		},
		onGroupingClick() {
			gotoVariableGrouping(this.$router);
		}
	}
});
</script>

<style>
.select-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
	color: rgba(0,0,0,.87);
}
.select-view .variable-facets {
	height: 100%;
}
.select-view .nav-tabs .nav-item a {
	padding-left: 0.5rem;
	padding-right: 0.5rem;
}
.select-view .nav-tabs .nav-link {
	color: #757575;
}
.select-view .nav-tabs .nav-link.active {
	color: rgba(0, 0, 0, 0.87);
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.grouping-button {
	margin: 0 8px;
	width: 100%;
}
</style>
