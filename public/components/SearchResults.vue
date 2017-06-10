<template>
	<div id="search-results">		
		<div v-for="dataset in datasets">
			<div v-on:click="setActiveDataset(dataset.name)"><h3>{{dataset.name}}</h3></div>
			<div class="description-text">{{dataset.description}}</div>
		</div>
	</div>
</template>

<script>

export default {
	name: 'search-results',

	//data change handlers
	computed: {	
	  	datasets() {
			return this.$store.getters.getDatasets();
		}
	},

	// 
	methods: {
		setActiveDataset(name) {
			this.$store.commit('setActiveDataset', name);
		}
	}
};
</script>

<style>
.description-text {	
	height: 150px;
	overflow: auto ;
}
</style>
