<template>
  <div>
    <!-- Waiting -->
    <b-alert :show="status === 'started'" variant="info">
      Importing <b>{{ filename }}</b> as <b>{{ datasetID }}</b
      >...
    </b-alert>

    <!-- Success -->
    <b-alert :show="status === 'success'" dismissible variant="success">
      <i class="fa fa-check-circle-o" aria-hidden="true"></i> Imported
      <b>{{ filename }}</b> as <b>{{ datasetID }}</b>
      <template v-if="isSampling">
        &mdash; Because of its size, the dataset has been sampled to
        {{ rowCount }} rows.
        <b-button @click="onClick" variant="success" size="sm">
          <i class="fa fa-download" aria-hidden="true"></i>
          Import the full dataset
        </b-button>
      </template>
    </b-alert>

    <!-- Error -->
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
    datasetID: String,
    filename: String,
    importResponse: Object,
    numRows: Number,
    status: {
      type: String,
      required: true,
    },
  },

  computed: {
    isSampling(): boolean {
      return !!this.importResponse?.sampled;
    },

    rowCount(): number {
      return this.importResponse?.rowCount ?? 0;
    },
  },

  methods: {
    onClick() {
      this.$emit("importfull");
    },
  },
});
</script>
