<template>
	<div class="explore">
		<variable-summaries class="explore-variable-summaries"></variable-summaries>
		<explore-data-table class="explore-data-table"></explore-data-table>
	</div>
</template>

<script>
import ExploreDataTable from '../components/ExploreDataTable';
import VariableSummaries from '../components/VariableSummaries';

export default {
	name: 'explore',

	components: {
		ExploreDataTable,
		VariableSummaries
	},

	mounted() {
		const dataset = this.$store.getters.getRouteDataset();
		this.$store.dispatch('getVariableSummaries', dataset);
	},

	watch: {
		'$route.query.dataset'() {
			const dataset = this.$store.getters.getRouteDataset();
			this.$store.dispatch('getVariableSummaries', dataset);
		}
	}
};
</script>

<style>
.explore {
	display: flex;
	justify-content: space-around;
	padding: 8px;
}
.explore-variable-summaries {
	width: 30%;
}
.explore-data-table {
	display: flex;
	flex-direction: column;
	width: 60%;
}
</style>
