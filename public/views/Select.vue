<template>
	<div class="select-view">
		<b-modal id="target-modal" ref="targetModal" title="Select Target Feature"
			hide-header-close
			no-close-on-backdrop
			no-close-on-esc
			hide-footer>
			<available-target-variables></available-target-variables>
		</b-modal>
		<div class="left-container">
			<h5 class="header-label">Select the Training features</h5>
			<div class="select-items">
				<available-training-variables class="select-available-variables"></available-training-variables>
				<training-variables class="select-training-variables"></training-variables>
			</div>
		</div>
		<div class="right-container">
			<h5 class="header-label">Select the Target Feature</h5>
			<target-variable class="select-target-variables"></target-variable>
			<select-data-table class="select-data-table"></select-data-table>
			<h5 class="header-label">Create the Pipelines</h5>
			<create-pipelines-form class="select-create-pipelines"></create-pipelines-form>
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
	name: 'select',

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
.select-view {
	display: flex;
	justify-content: space-around;
	padding: 8px;
}
.header-label {
	color: #333;
	margin: 0.75rem 0;
}
.select-items {
	display: flex;
	justify-content: space-around;
	padding: 8px;
	width: 100%;
}
.select-available-variables {
	width: 50%;
}
.select-training-variables {
	width: 50%;
}
.left-container {
	display: flex;
	flex-direction: column;
	justify-content: space-around;
	padding: 8px;
	width: 50%;
}
.right-container {
	display: flex;
	flex-direction: column;
	width: 50%;
}
</style>
