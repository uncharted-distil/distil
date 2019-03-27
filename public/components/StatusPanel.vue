<template>

<div class="status-panel">

</div>
    
</template>

<script lang="ts">

import Vue from 'vue';

export default Vue.extend({
	name: 'status-panel',
});

</script>

<style>

.status-panel {
	box-shadow: 0 4px 5px 0 rgba(0,0,0,0.14), 0 1px 10px 0 rgba(0,0,0,0.12), 0 2px 4px -1px rgba(0,0,0,0.2);
	display: block;
	position: fixed;
	right: 0;
	top: 0;
	bottom: 0;
	z-index: 100;
	width: 300px;
	height: 100%;
	background: #fff;
	display: none;
}

</style>
