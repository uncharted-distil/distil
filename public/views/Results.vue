<template>
	<div class="results">
		<variable-summaries class="results-variable-summaries"></variable-summaries>
		<results-comparison class="results-result-comparison"></results-comparison>
		<result-summaries class="results-result-summaries"></result-summaries>
	</div>
</template>

<script>
import ResultsComparison from '../components/ResultsComparison';
import VariableSummaries from '../components/VariableSummaries';
import ResultSummaries from '../components/ResultSummaries';

export default {
	name: 'results',

	components: {
		ResultsComparison,
		VariableSummaries,
		ResultSummaries
	},

	computed: {
		dataset() {
			return this.$store.getters.getRouteDataset();
		},
		variables() {
			return this.$store.getters.getVariables();
		},
		requestId() {
			return this.$store.getters.getRouteCreateRequestId();
		}
	},

	mounted() {
		this.fetch();
	},

	watch: {
		// watch the route and update the results if its modified
		'$route.query.dataset'() {
			this.fetch();
		},
		'$route.query.requestId'() {
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
					this.$store.dispatch('getResultsSummaries', {
						dataset: this.dataset,
						requestId: this.requestId()
					});
				});
		}
	}
};
</script>

<style>
.results {
	display: flex;
	justify-content: space-around;
	padding: 8px;
}
.results-variable-summaries {
	width: 25%;
}
.results-result-summaries {
	width: 25%;
}
.results-result-comparison {
	width: 50%;
}
</style>
