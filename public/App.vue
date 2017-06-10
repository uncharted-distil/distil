<template>
	<div id="app">
		<h1>Distil Server</h1>
		<search-bar></search-bar>
		<search-results id="search-results"></search-results>
		<variable-summaries id="variable-summaries"></variable-summaries>
	</div>
</template>

<script>

import SearchBar from './components/SearchBar';
import SearchResults from './components/SearchResults';
import VariableSummaries from './components/VariableSummaries';

export default {
	name: 'app',
	components: {
		SearchBar,
		SearchResults,
		VariableSummaries
	}
};
</script>

<style>
#app {
	font-family: 'Avenir', Helvetica, Arial, sans-serif;
	-webkit-font-smoothing: antialiased;
	-moz-osx-font-smoothing: grayscale;
	text-align: left;
	color: #2c3e50;
	margin-top: 5px;
}

#search-results {
	width: 80%;
	height:100%;
	overflow: auto;
	float: left;
}

#variable-summaries {
	width: 20%;
	height:100%;
	overflow: auto;
	float: left;
	background-color: aqua;
}
</style>
