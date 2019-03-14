<template>

<div>
	<b-button block variant="primary" v-b-modal.modal1>Import Local File</b-button>
	
	<!-- Modal Component -->
	<b-modal
		id="modal1"
		title="Import local file"
		@hide="clearFile()"
		@ok="handleOk()">
		<b-form-file
			ref="fileinput"
			v-model="file"
			:state="Boolean(file)"
			accept=".csv"
			plain/>
		<div class="mt-3">Selected file: {{ file ? file.name : '' }}</div>
	</b-modal>

</div>

</template>

<script>

import Vue from 'vue'
import { actions as datasetActions } from '../store/dataset/module';

export default Vue.extend({
	name: 'file-uploader',
	data() {
		return {
			file: null
		}
	},
	methods: {
		clearFile() {
			this.file = null;
			this.$refs.fileinput.reset();
		},
		handleOk() {
			if (!this.file) {
				return;
			}
			const fileNameTokens = this.file.name.split('.');
			const fname = fileNameTokens.length > 1
				? fileNameTokens.slice(0, -1).join('.')
				: fileNameTokens.join('.');
			const datasetID = fname.replace(' ', '_');
			console.log(datasetID);
			datasetActions.uploadDataFile(this.$store, { datasetID, file: this.file});
		}
	}
});
</script>

<style>
</style>