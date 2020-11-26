<template>
  <div class="form-container">
    <div class="control-group">
      <b-button @click="onApply">Apply</b-button>
      <b-form-checkbox class="check-box" @input="onSearchSimilarToggle"
        >Search Similar</b-form-checkbox
      >
    </div>
    <b-button @click="onExport">Export</b-button>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
export default Vue.extend({
  name: "create-labeling-form",
  methods: {
    onExport() {
      this.$emit("export");
    },
    onApply() {
      this.$emit("apply");
    },
    onSearchSimilarToggle(toggled: boolean) {
      this.$emit("toggle", toggled);
    },
  },
});
</script>

<style scoped>
.form-container {
  display: flex;
  justify-content: space-between;
}

.control-group {
  display: flex;
  flex-direction: row;
  align-items: center;
}
.check-box {
  margin-left: 16px;
}
</style>
