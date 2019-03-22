<template>

<div>
	<b-button block variant="primary" v-b-modal.upload-modal>Import File</b-button>

	<!-- Modal Component -->
	<b-modal
		id="upload-modal"
		title="Import local file"
		:ok-disabled="!Boolean(file)"
		@hide="clearFile()"
		@ok="handleOk()">
		<p>Select a csv file to import</p>
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

<script lang="ts">

import Vue from 'vue';
import { actions as datasetActions } from '../store/dataset/module';
import { filterSummariesByDataset } from '../util/data';

export default Vue.extend({
	name: 'file-uploader',

	data() {
		return {
			file: null,
		};
	},

	computed: {
		filename(): string {
			return this.file ? this.file.name : '';
		},
		datasetID(): string {
			if (this.filename) {
				const fileNameTokens = this.filename.split('.');
				const fname = fileNameTokens.length > 1
					? fileNameTokens.slice(0, -1).join('.')
					: fileNameTokens.join('.');
				const datasetID = fname.replace(' ', '_');
				return datasetID;
			}
			return '';
		}
	},

	methods: {
		clearFile() {
			this.file = null;
			const $refs = this.$refs as any;
			$refs.fileinput.reset();
		},
		handleOk() {
			if (!this.file) {
				return;
			}
			this.$emit('uploadstart', {
				file: this.file,
				filename: this.filename,
				datasetID: this.datasetID,
			});
			let uploadError;
			datasetActions
				.uploadDataFile(this.$store, { datasetID: this.datasetID, file: this.file})
				.catch((err) => {
					uploadError = err;
				})
				.then(() => {
					this.$emit('uploadfinish', uploadError);
				});
		}
	}
});
</script>

<style>
</style>
