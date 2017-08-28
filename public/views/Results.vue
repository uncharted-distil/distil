<template>
	<div class="row mt-2 mb-2">
		<data-table class="col-md-8"></data-table>
		<variable-summaries class="col-md-4"></variable-summaries>
	</div>
</template>

<script>
import DataTable from '../components/DataTable';
import VariableSummaries from '../components/VariableSummaries';

export default {
	name: 'results',
	components: {
		DataTable,
		VariableSummaries
	}
};
</script>

<style>
</style>
