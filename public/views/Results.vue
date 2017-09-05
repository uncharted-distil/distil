<template>
	<div class="row mt-2 mb-2">
		<variable-summaries class="col-md-2"></variable-summaries>
		<data-table class="col-md-8"></data-table>
		<result-summaries class="col-md-2"></result-summaries>
	</div>
</template>

<script>
import DataTable from '../components/DataTable';
import VariableSummaries from '../components/VariableSummaries';
import ResultSummaries from '../components/ResultSummaries';

export default {
	name: 'results',
	components: {
		DataTable,
		VariableSummaries,
		ResultSummaries
	}
};
</script>

<style>
</style>
