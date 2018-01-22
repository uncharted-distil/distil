<template>
	<div class="container-fluid h-100 select-view">
		<b-modal id="target-modal" ref="targetModal" title="Select Target Feature"
			hide-header-close
			no-close-on-backdrop
			no-close-on-esc
			hide-footer>
			<available-target-variables></available-target-variables>
		</b-modal>
		<div class="row h-100">
			<div class="col-12 col-md-6 col-p-top d-flex flex-column">
				<div class="row mh-100">
					<div class="col-12 d-flex">
						<h5 class="header-label">Select Features That May Predict</h5>
					</div>
				</div>
				<div class="row mh-100">
					<available-training-variables class="col-12 col-md-6 select-available-variables d-flex"></available-training-variables>
					<training-variables class="col-12 col-md-6 select-training-variables d-flex"></training-variables>
				</div>
			</div>
			<div class="col-12 col-md-6 col-p-top d-flex flex-column">
				<div class="row mh-100">
					<div class="col-12 d-flex flex-column">
						<h5 class="header-label">Select Feature to Predict</h5>
					</div>
				</div>
				<div class="row mh-100">
						<target-variable class="col-12 d-flex flex-column select-target-variables"></target-variable>
				</div>
				<div class="row mh-45">
						<select-data-table class="col-12 d-flex flex-column select-data-table"></select-data-table>
				</div>
				<div class="row mh-100">
					<div class="col-12 d-flex flex-column">
						<h5 class="header-label">Create the Pipelines</h5>
					</div>
				</div>
				<div class="row mh-100">
					<create-pipelines-form class="col-12 d-flex flex-column select-create-pipelines"></create-pipelines-form>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts">

import FlowBar from '../components/FlowBar.vue';
import CreatePipelinesForm from '../components/CreatePipelinesForm.vue';
import SelectDataTable from '../components/SelectDataTable.vue';
import AvailableTargetVariables from '../components/AvailableTargetVariables.vue';
import AvailableTrainingVariables from '../components/AvailableTrainingVariables.vue';
import TrainingVariables from '../components/TrainingVariables.vue';
import TargetVariable from '../components/TargetVariable.vue';
import { getters as dataGetters, actions } from '../store/data/module';
import { getters as routeGetters} from '../store/route/module';
import { Variable } from '../store/data/index';
import Vue from 'vue';

export default Vue.extend({
	name: 'select-view',

	components: {
		FlowBar,
		CreatePipelinesForm,
		SelectDataTable,
		AvailableTargetVariables,
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
		}
	},

	mounted() {
		this.fetch();
		this.updateModal();
	},

	watch: {
		'$route.query.dataset'() {
			this.fetch();
		},
		'$route.query.training'() {
		},
		'$route.query.target'() {
			this.updateModal();
		}
	},

	methods: {
		fetch() {
			actions.getVariables(this.$store, {
				dataset: this.dataset
				})
				.then(() => {
					actions.getVariableSummaries(this.$store, {
						dataset: this.dataset,
						variables: this.variables
					});
				});
		},
		updateModal() {
			const target = routeGetters.getRouteTargetVariable(this.$store);
			const modal = this.$refs.targetModal as any;
			if (target) {
				modal.hide();
			} else {
				modal.show();
			}
		}
	}
});
</script>

<style>
.select-view .nav-link {
    padding: 1rem 0 0.5rem 0;
}
.mh-45{
	max-height: 35%!important;
}
.header-label {
	color: #333;
	padding: 1rem 0 0.5rem 0;
}

</style>
