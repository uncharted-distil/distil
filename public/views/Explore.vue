<template>
	<div class="row mt-2 mb-2">
		<variable-summaries class="col-md-4"></variable-summaries>
		<data-table class="col-md-8"></data-table>
	</div>
</template>

<script>
import DataTable from '../components/DataTable';
import VariableSummaries from '../components/VariableSummaries';

export default {
	name: 'explore',
	components: {
		DataTable,
		VariableSummaries
	}
};
</script>

<style>
</style>
