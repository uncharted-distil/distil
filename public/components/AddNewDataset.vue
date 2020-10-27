<template>
  <b-modal :id="id" title="Add New Dataset"> LOL </b-modal>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "AddNewDataset",

  props: {
    id: { type: String as () => string, required: true },
  },
});
</script>
