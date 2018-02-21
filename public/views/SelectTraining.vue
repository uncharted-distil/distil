<template>
	<div class="container-fluid d-flex flex-column h-100 select-view">
		<div class="row flex-0-nav">
		</div>
		<div class="row flex-1 pb-3">
			<div class="col-12 col-md-6 d-flex flex-column border-gray-right">
				<div class="row flex-1 bg-white align-items-center">
					<div class="col-12 d-flex">
						<h5 class="header-label">Select Features That May Predict {{capitalize(target)}}</h5>
					</div>
				</div>
				<div class="row flex-12">
					<available-training-variables class="col-12 col-md-6 select-available-variables d-flex"></available-training-variables>
					<training-variables class="col-12 col-md-6 select-training-variables d-flex"></training-variables>
				</div>
			</div>
			<div class="col-12 col-md-6 d-flex flex-column">
				<div class="row flex-1 bg-white align-items-center">
					<div class="col-12">
					</div>
				</div>
				<div class="row flex-12">
					<div class="col-12 d-flex flex-column">
						<div class="row flex-4">
							<target-variable class="col-12 d-flex flex-column select-target-variables"></target-variable>
						</div>
						<div class="row responsive-flex pb-3">
							<select-data-table class="col-12 d-flex flex-column select-data-table"></select-data-table>
						</div>
						<div class="row flex-2 align-items-center">
							<div class="col-12 d-flex flex-column">
								<create-pipelines-form class="select-create-pipelines"></create-pipelines-form>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">

import _ from 'lodash';
import CreatePipelinesForm from '../components/CreatePipelinesForm.vue';
import SelectDataTable from '../components/SelectDataTable.vue';
import AvailableTrainingVariables from '../components/AvailableTrainingVariables.vue';
import TrainingVariables from '../components/TrainingVariables.vue';
import TargetVariable from '../components/TargetVariable.vue';
import { getters as dataGetters, actions } from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { Variable } from '../store/data/index';
import { HighlightRoot } from '../util/highlights';
import { Filter } from '../util/filters';
import Vue from 'vue';

export default Vue.extend({
	name: 'select-view',

	components: {
		CreatePipelinesForm,
		SelectDataTable,
		AvailableTrainingVariables,
		TrainingVariables,
		TargetVariable
	},

	computed: {
		dataset(): string {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables(): Variable[] {
			return dataGetters.getVariables(this.$store);
		},
		training(): string {
			return routeGetters.getRouteTrainingVariables(this.$store);
		},
		target(): string {
			return routeGetters.getRouteTargetVariable(this.$store);
		},
		filters(): Filter[] {
			return routeGetters.getDecodedFilters(this.$store);
		},
		filterStr(): string {
			return routeGetters.getRouteFilters(this.$store);
		},
		selectedFilters(): Filter[] {
			return dataGetters.getSelectedFilters(this.$store);
		},
		highlightRoot(): HighlightRoot {
			return routeGetters.getDecodedHighlightRoot(this.$store);
		},
		highlightRootStr(): string {
			return routeGetters.getRouteHighlightRoot(this.$store);
		}
	},

	watch: {
		highlightRootStr() {
			actions.fetchDataHighlightValues(this.$store, {
				dataset: this.dataset,
				filters: this.filters,
				highlightRoot: this.highlightRoot,
			});
		},
		filterStr() {
			actions.fetchDataHighlightValues(this.$store, {
				dataset: this.dataset,
				filters: this.filters,
				highlightRoot: this.highlightRoot,
			});
			actions.fetchSelectedTableData(this.$store, {
				dataset: this.dataset,
				filters: this.selectedFilters
			});
		}
	},

	beforeMount() {
		this.fetch();
	},

	methods: {
		fetch() {
			actions.fetchVariablesAndVariableSummaries(this.$store, {
				dataset: this.dataset
			});
			actions.fetchDataHighlightValues(this.$store, {
				dataset: this.dataset,
				filters: this.filters,
				highlightRoot: this.highlightRoot,
			});
			actions.fetchSelectedTableData(this.$store, {
				dataset: this.dataset,
				filters: this.selectedFilters
			});
		},
		capitalize(str) {
			return _.capitalize(str);
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
.header-label {
	padding: 1rem 0 0.5rem 0;
	font-weight: bold;
}
.select-view .responsive-flex {
	flex:4;
}
@media (min-width: 1200px) {
	.select-view .responsive-flex {
		flex:6;
	}
}
</style>
