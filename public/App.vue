<template>
	<div id="app" class="container-fluid">
		<div class="row">
			<h1>Distil</h1>
		</div>
		<div class="row">
			<search-bar></search-bar class="col-md-12">
		</div>
		<div class="row">
			<search-results id="search-results" class="col-md-8"></search-results>
			<variable-summaries id="variable-summaries" class="col-md-4"></variable-summaries>
		</div>
		<div class="row">
			<data-table class="col-md-12"></data-table>
		</div>
	</div>
</template>

<script>

import SearchBar from './components/SearchBar';
import SearchResults from './components/SearchResults';
import VariableSummaries from './components/VariableSummaries';
import DataTable from './components/DataTable';

import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap-vue/dist/bootstrap-vue.css';

export default {
	name: 'app',
	components: {
		SearchBar,
		SearchResults,
		VariableSummaries,
		DataTable
	}
};
</script>

<style>
#app {
	text-align: left;
	margin-top: 5px;
}

#search-results {
	height:100%;
	overflow: auto;
	float: left;
}

#variable-summaries {
	height:100%;
	overflow: auto;
	float: left;
}
</style>
