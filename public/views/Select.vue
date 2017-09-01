<template>
	<div class="select">
		<variable-summaries class="select-variable-summaries"></variable-summaries>
		<data-table class="select-data-table"></data-table>
	</div>
</template>

<script>
import DataTable from '../components/DataTable';
import VariableSummaries from '../components/VariableSummaries';

export default {
	name: 'select',
	components: {
		DataTable,
		VariableSummaries
	}
};
</script>

<style>
.select {
	display: flex;
	justify-content: space-around;
	padding: 8px;
}
.select-variable-summaries {
	width: 30%;
}
.select-data-table {
	width: 60%;
}
</style>
