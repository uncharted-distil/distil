<template>
	<div id="search-results">
		<div v-for="dataset in datasets">
			<div v-on:click="setActiveDataset(dataset.name)">
				<h4>{{dataset.name}}</h4>
			</div>
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

	methods: {
		setActiveDataset(name) {
			if (name !== this.$store.state.activeDataset) {
				this.$store.commit('setActiveDataset', name);
				this.$store.dispatch('getVariableSummaries', name);
			}			
		}
	}
};
</script>

<style>
.description-text {
	height: 150px;
	overflow: auto;
}
</style>
