<template>
  <div class="form-container">
    <div class="control-group">
      <b-button :disabled="isLoading" size="lg" @click="onEvent(applyEvent)">
        <template v-if="isLoading">
          <div v-html="spinnerHTML" />
        </template>
        <template v-else> Search Similar </template>
      </b-button>
    </div>
    <div>
      <b-button size="lg" @click="onEvent(exportEvent)">Export</b-button>
      <b-button size="lg" variant="primary" @click="onEvent(saveEvent)">
        Save
      </b-button>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { circleSpinnerHTML } from "../../util/spinner";

const enum COMPONENT_EVENT {
  EXPORT = "export",
  SAVE = "save",
  APPLY = "apply",
}
export default Vue.extend({
  name: "create-labeling-form",
  props: {
    isLoading: { type: Boolean as () => boolean, default: false },
  },
  computed: {
    spinnerHTML(): string {
      return circleSpinnerHTML();
    },
    saveEvent(): string {
      return COMPONENT_EVENT.SAVE;
    },
    applyEvent(): string {
      return COMPONENT_EVENT.APPLY;
    },
    exportEvent(): string {
      return COMPONENT_EVENT.EXPORT;
    },
  },
  methods: {
    onEvent(event: COMPONENT_EVENT) {
      this.$emit(event);
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
