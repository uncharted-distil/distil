<template>
	<div class="explore-view">
		<flow-bar
			left-text="Return to Search"
			:on-left="gotoSearch"
			center-text="Explore the dataset"
			right-text="Continue to Select features"
			:on-right="gotoSelect">
		</flow-bar>
		<div class="explore-items">
			<variable-summaries class="explore-variable-summaries"
				:variables="summaries"
				:dataset="dataset"></variable-summaries>
			<explore-data-table class="explore-data-table"></explore-data-table>
		</div>
	</div>
</template>

<script lange="ts">
import FlowBar from '../components/FlowBar';
import ExploreDataTable from '../components/ExploreDataTable';
import VariableSummaries from '../components/VariableSummaries';
import { gotoSearch, gotoSelect } from '../util/nav';
import Vue from 'vue';
import { getters as dataGetters, actions } from '../store/data/module';
import { getters as routeGetters } from '../store/route/module';

export default Vue.extend({
	name: 'explore',

	components: {
		FlowBar,
		ExploreDataTable,
		VariableSummaries
	},

	computed: {
		dataset() {
			return routeGetters.getRouteDataset(this.$store);
		},
		variables() {
			return dataGetters.getVariables(this.$store);
		},
		summaries() {
			return dataGetters.getVariableSummaries(this.$store);
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		'$route.query.dataset'() {
			this.fetch();
		}
	},

	methods: {
		gotoSearch() {
			gotoSearch(this.$store, this.$router);
		},
		gotoSelect() {
			gotoSelect(this.$store, this.$router);
		},
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
		}
	}
});
</script>

<style>
.explore-view {
	display: flex;
	flex-direction: column;
	align-items: center;
}
.explore-items {
	display: flex;
	justify-content: space-around;
	padding: 8px;
	width: 100%;
}
.explore-variable-summaries {
	width: 25%;
}
.explore-data-table {
	display: flex;
	flex-direction: column;
	width: 75%;
}
</style>
