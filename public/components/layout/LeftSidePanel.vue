<template>
  <aside class="d-flex flex-column p-2 h-100 left-side-panel-container">
    <header class="mb-1">
      <h5>{{ panelTitle }}</h5>
      <!-- <close-button @close="onClose" /> -->
    </header>
    <div class="panel-body">
      <slot />
    </div>
  </aside>
</template>

<script lang="ts">
import Vue from "vue";
// import CloseButton from "../layout/CloseButton.vue";

export default Vue.extend({
  name: "LeftSidePanel",

  // components: { CloseButton },

  props: {
    // isOpen: { type: Boolean, default: true },
    panelTitle: { type: String, default: "" },
  },

  // methods: {
  //   onClose(): void {
  //     this.$emit("close-panel");
  //   },
  // },
});
</script>

<style scoped>
.left-side-panel-container {
  background-color: var(--bg-color-side-panel);
  border-color: var(--border-color);
  border-style: solid;
  border-width: 0 1px;
  width: var(--width-left-side-panel);
  z-index: var(--z-index-left-side-panel);
}
</style>
