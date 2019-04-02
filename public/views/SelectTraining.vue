<template>
	<div class="select-training-view d-flex h-100">
		<status-panel v-if="statusPanel.isOpen" 
			@close="statusPanel.isOpen = false"
			:statusType="statusPanel.type">
		</status-panel>
		<div class="sidebar-container d-flex flex-column h-100">
			<div class="padding-nav"></div>
			<status-sidebar @statusIconClick="openStatusPanel"></status-sidebar>
		</div>
		<div class="container-fluid d-flex flex-column h-100 select-view">
			<div class="row flex-0-nav"></div>

			<div class="row align-items-center justify-content-center bg-white">

				<div class="col-12 col-md-6 d-flex flex-column">
					<h5 class="header-label">Select Features That May Predict {{target.toUpperCase()}}</h5>

					<div class="row col-12 pl-4">
						<div>
							{{target.toUpperCase()}} is being modeled as a
						</div>
						<div class="pl-2">
							<type-change-menu
								:dataset="dataset"
								:field="target"
								:values="targetSampleValues"></type-change-menu>
						</div>
					</div>
					<div class="row col-12 pl-4">
						<p>
							<b>Select Features That May Predict {{target.toUpperCase()}}</b> Use interactive feature highlighting to analyze relationships or to exclude samples from the model. Features which appear to have stronger relation are listed first.
						</p>
					</div>
				</div>

				<div class="col-12 col-md-6 d-flex flex-column">
					<target-variable class="col-12 d-flex flex-column select-target-variables"></target-variable>
				</div>
			</div>

			<div class="row flex-1 pb-3">
				<available-training-variables class="col-12 col-md-3 d-flex h-100"></available-training-variables>
				<training-variables class="col-12 col-md-3 nopadding d-flex h-100"></training-variables>

				<div class="col-12 col-md-6 d-flex flex-column h-100">
					<select-data-slot class="flex-1 d-flex flex-column pb-3 pt-2"></select-data-slot>
					<create-solutions-form class="select-create-solutions"></create-solutions-form>
				</div>
			</div>
		</div>

	</div>
</template>

<script lang="ts">

import Vue from 'vue';
import StatusPanel from '../components/StatusPanel';
import StatusSidebar from '../components/StatusSidebar';
import CreateSolutionsForm from '../components/CreateSolutionsForm';
import SelectDataSlot from '../components/SelectDataSlot';
import AvailableTrainingVariables from '../components/AvailableTrainingVariables';
import TrainingVariables from '../components/TrainingVariables';
import TargetVariable from '../components/TargetVariable';
import TypeChangeMenu from '../components/TypeChangeMenu';
import { actions as viewActions } from '../store/view/module';
import { getters as routeGetters } from '../store/route/module';
import { DatasetPendingRequest, DatasetPendingRequestType } from '../store/dataset/index';

export default Vue.extend({
	name: 'select-view',

	data() {
		return {
			statusPanel: {
				isOpen: false,
				type: undefined as DatasetPendingRequestType
			},
		};
	},
	components: {
		CreateSolutionsForm,
		SelectDataSlot,
		AvailableTrainingVariables,
		TrainingVariables,
		TargetVariable,
		TypeChangeMenu,
		StatusPanel,
		StatusSidebar,
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		training(): string {
			return routeGetters.getRouteTrainingVariables(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		filtersStr(): string {
			return routeGetters.getRouteFilters(this.$store);
		},
		highlightRootStr(): string {
			return routeGetters.getRouteHighlightRoot(this.$store);
		},
		targetSampleValues(): any[] {
			const summaries = routeGetters.getTargetVariableSummaries(this.$store);
			if (summaries.length > 0) {
				const summary = summaries[0];
				return summary.buckets;
			}
			return [];
		},
		availableTrainingVarsPage(): number {
			return routeGetters.getRouteAvailableTrainingVarsPage(this.$store);
		},
		trainingVarsPage(): number {
			return routeGetters.getRouteTrainingVarsPage(this.$store);
		}
	},

	watch: {
		highlightRootStr() {
			viewActions.updateSelectTrainingData(this.$store);
		},
		training() {
			viewActions.updateSelectTrainingData(this.$store);
		},
		filtersStr() {
			viewActions.updateSelectTrainingData(this.$store);
		},
		availableTrainingVarsPage() {
			viewActions.updateSelectTrainingData(this.$store);
		},
		trainingVarsPage() {
			viewActions.updateSelectTrainingData(this.$store);
		}
	},

	methods: {
		openStatusPanel(statusType: DatasetPendingRequestType) {
			this.statusPanel.isOpen = true;
			this.statusPanel.type = statusType;
		},
	},

	beforeMount() {
		viewActions.fetchSelectTrainingData(this.$store);
	}
});

</script>

<style>
.select-training-view {
	flex-direction: row-reverse;
}

.select-view .nav-link {
	padding: 1rem 0 0.25rem 0;
	border-bottom: 1px solid #E0E0E0;
	color: rgba(0,0,0,.87);
}
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.select-data-container {
	flex: 1;
	z-index: 1; /* to show the scroll bar */
}
</style>
