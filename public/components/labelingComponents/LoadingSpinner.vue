<template>
  <div class="w-100 h-100 d-flex justify-content-center align-items-center">
    <div class="d-flex justify-content-center align-items-center flex-column">
      <p>
        <strong>{{ state }}</strong>
      </p>
      <b-spinner />
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
export default Vue.extend({
  name: "loading-spinner",
  props: {
    state: { type: String as () => string, default: "" },
  },
});
</script>
