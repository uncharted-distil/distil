<template>
  <div>
    <b-modal v-model="show" cancel-disabled hide-header hide-footer>
      <div class="row justify-content-center">
        <i class="fa fa-exclamation-triangle fa-3x fail-icon"></i>
        <div>
          <b>{{ title }}:</b> {{ error }}
        </div>
      </div>
      <div class="row justify-content-center">
        <b-btn
          class="mt-3 close-modal-button"
          variant="outline-secondary"
          block
          @click="onClose"
          >OK</b-btn
        >
      </div>
    </b-modal>
  </div>
</template>

<script lang="ts">
import _ from "lodash";
import Vue from "vue";

export default Vue.extend({
  name: "error-model",

  props: {
    title: {
      type: String as () => string,
      default: "Error",
    },
    error: {
      type: String as () => string,
      default: "Internal server error",
    },
    show: Boolean as () => boolean,
  },

  methods: {
    onClose() {
      this.$emit("close");
    },
  },
});
</script>

<style>
.close-modal-button {
  margin: 0 8px;
  width: 25% !important;
  line-height: 32px !important;
}
</style>
