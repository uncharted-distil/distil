<template>
	<div class="results-tables">
		<results-data-table
			class="results-data-table"
			title="Correct Predictions"
			:filterFunc="classificationMatchFilter"
			:decorateFunc="classificationMatchDecorate"
		></results-data-table>
		<results-data-table class="results-data-table" title="Incorrect Predictions"
			:filterFunc="classificationNoMatchFilter"
			:decorateFunc="classificationNoMatchDecorate"
		></results-data-table>
	</div>
</template>

<script>
import ResultsDataTable from '../components/ResultsDataTable';

export default {
	name: 'results-comparison',

	components: {
		ResultsDataTable,
	},

	methods: {
		// Methods passed to result table instances to filter their displays.
		classificationMatchFilter(dataItem) {
			return dataItem[dataItem._target.truth] === dataItem[dataItem._target.predicted];
		},
		classificationNoMatchFilter(dataItem) {
			return dataItem[dataItem._target.truth] !== dataItem[dataItem._target.predicted];
		},
		// Methods passed to result table instance to update their row visuals post-filter
		classificationMatchDecorate(dataItem) {
			dataItem._cellVariants = { [dataItem._target.predicted]: 'success' };
			return dataItem;
		},
		classificationNoMatchDecorate(dataItem) {
			dataItem._cellVariants = { [dataItem._target.predicted]: 'danger' };
			return dataItem;
		}
	}
};
</script>

<style>
.results-tables {
	display: flex;
	flex-direction: column;
}
.results-data-table {
	display: flex;
	flex-direction: column;
}
</style>
