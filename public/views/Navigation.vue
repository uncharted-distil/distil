<template>
	<div class="navigation">
		<nav-bar></nav-bar>
	</div>
</template>

<script lang="ts">
import NavBar from '../components/NavBar.vue';
import Vue from 'vue';

export default Vue.extend({
	name: 'navigation-view',

	components: {
		NavBar
	}
});
</script>

<style>
.navigation {
}
</style>
