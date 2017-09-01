<template>
	<div class="explore">
		<variable-summaries class="explore-variable-summaries"></variable-summaries>
		<data-table class="explore-data-table"></data-table>
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
.explore {
	display: flex;
	justify-content: space-around;
	padding: 8px;
}
.explore-variable-summaries {
	width: 30%;
}
.explore-data-table {
	width: 60%;
}
</style>
