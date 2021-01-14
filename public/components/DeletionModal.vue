<template>
  <b-modal :id="modalId" @ok="okHandler">
    <template #modal-title> Deleting {{ target }}</template>
    This action can NOT be undone. Do you want to delete {{ target }}?
  </b-modal>
</template>

<script lang="ts">
import Vue from "vue";
export default Vue.extend({
  name: "deletion-modal",
  props: {
    target: { type: String as () => string, default: "" },
  },
  data() {
    return { modalId: "deletion-modal" };
  },
  watch: {
    target() {
      if (!this.target.length) {
        return;
      }
      // show deletion prompt if target changed
      this.$bvModal.show(this.modalId);
    },
  },
  methods: {
    okHandler() {
      this.$emit("ok");
    },
  },
});
</script>
