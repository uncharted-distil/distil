<template>
  <div class="action-column-container">
    <slot name="actions" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
export default Vue.extend({
  name: "ActionColumn",
});
</script>

<style scoped>
.action-column-container {
  --width: var(--width-action-column, 3rem);
  height: 100%;
  position: relative;
  width: var(--width);
}

.action-column-nav-bar {
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  width: var(--width);
  margin: 0;
  padding: 0;
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  background-color: $secondary-bar-bg;
  border: 1px solid rgba(207, 216, 220, 0.5);
}

.action-column-nav-bar li {
  position: relative;
  display: block;
}

.action-column-nav-bar button {
  width: var(--width);
  height: var(--width);
  background-color: transparent;
  color: $icon-color;
}

.action-column-nav-bar.active button {
  border-left: 5px solid $selection;
  color: $selection;
}
</style>
