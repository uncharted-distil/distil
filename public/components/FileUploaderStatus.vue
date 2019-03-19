<template>

<div>
		<b-alert :show="status === 'started'" variant="info">
			Importing <b>{{ filename }}</b> as <b>{{ datasetID }}</b>... 
		</b-alert>
		<b-alert :show="status === 'success'" dismissible variant="success">
			<b>{{ filename }}</b> has been imported as <b>{{ datasetID }}</b>
		</b-alert>
		<b-alert :show="status === 'error'" dismissible variant="danger">
			An unexpected error has happened while importing <b>{{ filename }}</b>
		</b-alert>
</div>

</template>

<script>

import Vue from 'vue'

export default Vue.extend({
	name: 'file-uploader-status',
	props: ['status', 'filename', 'datasetID'],
});
</script>

<style>
</style>