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

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getVariables();
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
		fetch() {
			this.$store.dispatch('getVariables', this.dataset)
				.then(() => {
					this.$store.dispatch('getVariableSummaries', {
						dataset: this.dataset,
						variables: this.variables
					});
				});
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
