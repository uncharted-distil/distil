<template>
  <div class="close-button" @click="onClick">
    <i class="fa fa-times" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";
export default Vue.extend({
  name: "CloseButton",
  methods: {
    onClick(): void {
      this.$emit("close");
    },
  },
});
</script>

<style scoped>
.close-button {
  align-items: center;
  border-radius: 3px;
  color: var(--gray-800);
  cursor: pointer;
  display: flex;
  font-size: 16px;
  height: 32px;
  justify-content: center;
  position: absolute;
  right: 12px;
  top: 12px;
  width: 32px;
}

.close-button:hover {
  background-color: var(--bg-color-button-hover);
}
</style>
