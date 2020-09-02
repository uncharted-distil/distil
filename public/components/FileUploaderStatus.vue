<template>
  <div>
    <b-alert :show="status === 'started'" variant="info">
      Importing <b>{{ filename }}</b> as <b>{{ datasetID }}</b
      >...
    </b-alert>
    <b-alert :show="status === 'success'" dismissible variant="success">
      <i class="fa fa-check-circle-o" aria-hidden="true"></i> Imported
      <b>{{ filename }}</b> as <b>{{ datasetID }}</b>
    </b-alert>
    <b-alert :show="status === 'error'" dismissible variant="danger">
      <i class="fa fa-times-circle-o" aria-hidden="true"></i> An unexpected
      error has happened while importing <b>{{ filename }}</b>
    </b-alert>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "file-uploader-status",
  props: {
    status: {
      type: String,
      required: true,
    },
    filename: {
      type: String,
    },
    datasetID: {
      type: String,
    },
  },
});
</script>

<style></style>
