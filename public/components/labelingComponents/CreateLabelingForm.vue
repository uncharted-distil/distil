<template>
  <div class="form-container">
    <div class="control-group">
      <b-button @click="onApply" size="lg">
        <template v-if="isLoading">
          <div v-html="spinnerHTML"></div>
        </template>
        <template v-else> Apply </template>
      </b-button>
    </div>
    <b-button @click="onExport" size="lg">Export</b-button>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { circleSpinnerHTML } from "../../util/spinner";

export default Vue.extend({
  name: "create-labeling-form",
  props: {
    isLoading: { type: Boolean as () => boolean, default: false },
  },
  computed: {
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
  },
  methods: {
    onExport() {
      this.$emit("export");
    },
    onApply() {
      this.$emit("apply");
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
