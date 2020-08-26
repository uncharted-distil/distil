<template>
  <div class="container-fluid d-flex flex-column h-100">
    <div class="row flex-0-nav"></div>
    <div class="col-12 col-md-12">
      <h4 class="abort-header-label">Application aborted successfully</h4>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "abort-success",
});
</script>

<style>
.abort-header-label {
  padding: 28px 0 28px 0;
  text-align: center;
}
</style>
