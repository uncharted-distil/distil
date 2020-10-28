<template>
  <div>
    <!-- Waiting -->
    <b-alert :show="status === 'started'" variant="info">
      Importing <b>{{ name }}</b> as <b>{{ datasetId }}</b>
    </b-alert>

    <!-- Success -->
    <b-alert :show="status === 'success'" dismissible variant="success">
      <i class="fa fa-check-circle-o" /> Imported <b>{{ name }}</b> as
      <b>{{ datasetId }}</b>
      <template v-if="isSampling">
        &mdash; Because of its size, the dataset has been sampled to
        {{ rowCount }} rows.
        <b-button @click="onClick" variant="success" size="sm">
          <i class="fa fa-download" />
          Import the full dataset
        </b-button>
      </template>
    </b-alert>

    <!-- Error -->
    <b-alert :show="status === 'error'" dismissible variant="danger">
      <i class="fa fa-times-circle-o" /> An unexpected error has happened while
      importing <b>{{ name }}</b>
    </b-alert>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "ImportStatus",

  props: {
    datasetId: { type: String, required: true },
    name: { type: String, default: null },
    importResponse: { type: Object, default: null },
    numRows: { type: Number, default: null },
    status: { type: String, required: true },
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
