<template>
	<nav-bar class="navbar col-md-12"></nav-bar>
</template>

<script>
import NavBar from '../components/NavBar';
export default {
	name: 'navigation',
	components: {
		NavBar
	}
};
</script>

<style>
.navbar {
	background-color: #fff;
}
</style>
